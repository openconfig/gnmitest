/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package runner has functions to be able to run given suite of tests. It runs
// InstanceGroups sequentially, with each Instance within a group being run in
// parallel. See the comments in suite.proto for further description of the test
// suite execution logic.
package runner

import (
	"fmt"
	"log"
	"time"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmi/client"
	"github.com/openconfig/gnmi/match"
	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/gnmitest/common/report"
	"github.com/openconfig/gnmitest/config"
	"github.com/openconfig/gnmitest/creds"
	"github.com/openconfig/gnmitest/register"
	"github.com/openconfig/gnmitest/subscribe"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	rpb "github.com/openconfig/gnmitest/proto/report"
	spb "github.com/openconfig/gnmitest/proto/suite"
	tpb "github.com/openconfig/gnmitest/proto/tests"
)

// PartialReportFunc is used by framework to notify caller when running
// a single spb.InstanceGroup is finished.
type PartialReportFunc func(*rpb.InstanceGroup)

// Inlined functions are overridden in the tests.
var (
	createSubscription = func(ctx context.Context, sr *gpb.SubscribeRequest, pHandler client.ProtoHandler, conn *tpb.Connection, clientType string) error {
		q, err := client.NewQuery(sr)
		if err != nil {
			return err
		}
		q.Addrs = []string{conn.GetAddress()}
		q.Timeout = time.Duration(conn.GetTimeout()) * time.Second
		q.ProtoHandler = pHandler

		r, err := resolver.Get(conn.GetCredentials().GetResolver())
		if err != nil {
			return err
		}
		credentials, err := r.Credentials(ctx, conn.GetCredentials())
		if err != nil {
			return err
		}
		if credentials != nil {
			q.Credentials = &client.Credentials{
				Username: credentials.Username,
				Password: credentials.Password,
			}
		}

		c := client.BaseClient{}
		defer c.Close()
		return c.Subscribe(ctx, q, clientType)
	}
)

// Runner object encapsulates the config, report and logging to run
// a Suite of tests.
type Runner struct {
	*log.Logger // logger is used to log framework events and errors.

	cfg    *config.Config    // object that contains the Suite proto.
	report PartialReportFunc // used to update caller incrementally.
}

// New creates an instance of runner. It receives;
// - logger to log framework events
// - config that contains the Suite proto
// - update function to notify caller about the partial results
func New(l *log.Logger, cfg *config.Config, r PartialReportFunc) *Runner {
	return &Runner{Logger: l, cfg: cfg, report: r}
}

// Start runs all the tests in the Suite. Start blocks caller until all tests
// finish. Reporting is done by calling PartialReportFunc as rpb.InstanceGroup
// ready.
func (r *Runner) Start(pCtx context.Context) error {
	for igIndex, ig := range r.cfg.Suite.InstanceGroupList {
		// Create an InstanceGroup report.
		igResult := &rpb.InstanceGroup{Description: ig.Description}

		// Error value of all the tests are pushed into error channel
		errC := make(chan error)
		ctx, cancelFunc := context.WithCancel(pCtx)
		defer cancelFunc()

		for _, ins := range ig.Instance {
			// Create a rpb.Instance report and append it to rpb.IntanceGroup.
			// So, while each test is running, no synchronization is needed
			// as each is given a rpb.Instance to update.
			insResult := &rpb.Instance{Description: ins.Description}
			igResult.Instance = append(igResult.Instance, insResult)

			// Run test and continue to next one without waiting this to finish.
			go func(gIns *spb.Instance) {
				select {
				case errC <- r.runTest(ctx, gIns, insResult):
				case <-ctx.Done():
				}
			}(ins)
		}

		for i := 0; i < len(ig.Instance); i++ {
			err := <-errC
			if err != nil {
				return err
			}
		}

		// Update caller with the result of running spb.InstanceGroup
		r.report(igResult)

		// If the instance group is marked as fatal, and any test within
		// it failed, then stop processing.
		if ig.Fatal && report.InstGroupFailed(igResult) {
			// Mark all other instance groups as not run.
			for i := igIndex + 1; i < len(r.cfg.Suite.InstanceGroupList); i++ {
				g := r.cfg.Suite.InstanceGroupList[i]
				r.report(&rpb.InstanceGroup{Description: g.Description, Skipped: true})
			}
			break
		}
	}
	return nil
}

// runTest runs parent test and its extensions.
func (r *Runner) runTest(ctx context.Context, ins *spb.Instance, ir *rpb.Instance) error {
	switch v := ins.Test.Type.(type) {
	case *tpb.Test_FakeTest:
		return r.runFakeTest(ins, ir)
	case *tpb.Test_Subscribe:
		return r.runSubscribeTest(ctx, ins, ir)
	default:
		return fmt.Errorf("runner doesn't know how to run %T test", v)
	}
}

// runFakeTest allows a fake test to be run to check the execution logic of the
// test framework without external dependencies.
func (r *Runner) runFakeTest(ins *spb.Instance, res *rpb.Instance) error {
	res.Description = ins.Description
	switch t := ins.GetTest().GetFakeTest(); t.Pass {
	case true:
		res.Test = &rpb.TestResult{Result: rpb.Status_SUCCESS}
	default:
		res.Test = &rpb.TestResult{Result: rpb.Status_FAIL}
	}
	return nil
}

func (r *Runner) runSubscribeTest(ctx context.Context, ins *spb.Instance, insRes *rpb.Instance) error {
	v := ins.Test.GetSubscribe()

	// set the subscribe test report
	insRes.Description = ins.Description
	subRes := &rpb.SubscribeTestResult{}
	insRes.Test = &rpb.TestResult{Type: &rpb.TestResult_Subscribe{Subscribe: subRes}, Test: ins.Test}

	// get the registered test by its oneof type from the SubscribeTest message
	// in tests.proto.
	ti, err := register.GetSubscribeTest(v.Args, ins.Test)
	if err != nil {
		return fmt.Errorf("failed getting an instance of %T test: %v", v.Args, err)
	}

	// create a match tree to set extension tests as clients with query paths being
	// subsription paths of the tests
	m := match.New()
	var exts []updater
	for _, el := range ins.ExtensionList {
		ext, err := setExtensions(m, r.cfg.Suite.ExtensionList[el], insRes)
		if err != nil {
			return fmt.Errorf("failed setting extensions; %v", err)
		}
		exts = append(exts, ext...)
	}

	// create a child context with the test timeout
	tCtx, cancel := context.WithTimeout(ctx, time.Duration(ins.GetTest().GetTimeout())*time.Second)
	defer cancel()
	ctx = nil

	// add parent test into match tree as well, but with glob query path
	// note that parent has a callback to cancel context when it is done.
	pt := &parentTest{ti: ti, subRes: subRes, finish: cancel}
	m.AddQuery([]string{"*"}, pt)
	err = createSubscription(tCtx, v.Request, func(msg proto.Message) error {
		sr, ok := msg.(*gpb.SubscribeResponse)
		if !ok {
			return fmt.Errorf("update has unknown type: %T", msg)
		}
		p, err := getQueryPath(sr)
		if err != nil {
			return fmt.Errorf("query path couldn't be extracted for %v: %v", sr, err)
		}
		// dispatch message to parent and extensions
		m.Update(sr, p)
		return nil
	}, ins.GetTest().GetConnection(), r.cfg.ClientType)

	switch {
	// A test may return Complete status which in turn triggers context
	// to be cancelled. This doesn't need to be reported as an RPC error.
	case tCtx.Err() == context.Canceled:
	// A test may timeout. This doesn't need to be reported as an RPC error.
	case tCtx.Err() == context.DeadlineExceeded:
		subRes.Status = rpb.CompletionStatus_TIMEOUT
	case err != nil:
		subRes.Status = rpb.CompletionStatus_RPC_ERROR
		return err
	}

	// end the extensions that haven't ended before.
	for _, e := range exts {
		e.End()
	}
	pt.End()

	setTestResult(insRes)

	return nil
}

// getQueryPath returns the gNMI path as indexed strings for the given SubscribeResponse.
// If the response is a sync response, "*" is returned to indicate that the message will
// be dispatched to all extensions.
func getQueryPath(sr *gpb.SubscribeResponse) ([]string, error) {
	switch v := sr.Response.(type) {
	case *gpb.SubscribeResponse_Update:
		pr := path.ToStrings(v.Update.Prefix, true)
		if len(v.Update.Update) > 0 {
			return append(pr, path.ToStrings(v.Update.Update[0].Path, false)...), nil
		} else if len(v.Update.Delete) > 0 {
			return append(pr, path.ToStrings(v.Update.Delete[0], false)...), nil
		}
		return pr, nil
	case *gpb.SubscribeResponse_SyncResponse:
		return []string{"*"}, nil
	case *gpb.SubscribeResponse_Error:
		return nil, fmt.Errorf("error occurred: %v", v.Error.Message)
	}
	return nil, fmt.Errorf("%T type isn't expected as subscription response", sr)
}

// setExtensions sets the extension tests in the given match tree. Test results
// are appended in the given rpb.Instance into Extensions field.
func setExtensions(m *match.Match, extensions *spb.ExtensionList, insRes *rpb.Instance) ([]updater, error) {
	var extTests []updater
	if extensions != nil {
		for _, e := range extensions.Extension {
			switch v := e.Type.(type) {
			case *tpb.Test_Subscribe:
				// get the registered test by its oneof type from the SubscribeTest message
				// in tests.proto.
				ti, err := register.GetSubscribeTest(v.Subscribe.Args, e)
				if err != nil {
					return nil, fmt.Errorf("error while getting an instance of %T test; %v", v.Subscribe.Args, err)
				}

				// Create test report instance for individual extension
				extRes := &rpb.SubscribeTestResult{}

				insRes.Extensions = append(insRes.Extensions, &rpb.TestResult{
					Test: e,
					Type: &rpb.TestResult_Subscribe{Subscribe: extRes},
				})
				et := &extensionTest{t: ti, r: extRes}
				extTests = append(extTests, et)
				addSubscription(m, v.Subscribe.Request.GetSubscribe(), et)
			}
		}
	}
	return extTests, nil
}

// addSubscription registers given extensionTest as match tree client.
func addSubscription(m *match.Match, s *gpb.SubscriptionList, et updater) {
	pr := path.ToStrings(s.Prefix, true)
	for _, p := range s.Subscription {
		if p.Path == nil {
			continue
		}
		path := append(pr, path.ToStrings(p.Path, false)...)
		// TODO(yusufsn): Removing queries individually can be supported for multi-subscription
		// requests.
		m.AddQuery(path, et)
	}
}

type updater interface {
	Update(interface{})
	End()
}

// extensionTest is the match tree client. It is used to dispatch indidividual
// messages to the tests which are clients in match tree.
type extensionTest struct {
	// Subscribe test instance to dispatch messages.
	t subscribe.Subscribe
	// Result object that is meant to be populated by the extension.
	r *rpb.SubscribeTestResult
	// Indicates whether extension test is done with receiving updates.
	done bool
}

// Update function is used to deliver gNMI message to matching test.
func (e *extensionTest) Update(l interface{}) {
	// check whether test is done with receiving updates.
	if e.done {
		return
	}

	sr := l.(*gpb.SubscribeResponse)

	// dispatch message to extension test.
	status, err := e.t.Process(sr)

	// update extension test report.
	srr := &rpb.SubscribeResponseResult{Response: sr}
	if err != nil {
		srr.Error = err.Error()
	}
	e.r.Responses = append(e.r.Responses, srr)

	if status == subscribe.Complete {
		if err := e.t.Check(); err != nil {
			e.r.Error = err.Error()
		}
		e.r.Status = rpb.CompletionStatus_EARLY_FINISHED

		// don't dispatch messages any more to the encapsulated test.
		e.done = true
	}
}

// End is called by parent test when the parent test is getting prepared to finish.
// This gives a chance to extensions to evaluate their situation and return a final
// result.
func (e *extensionTest) End() {
	if e.done {
		return
	}
	if err := e.t.Check(); err != nil {
		e.r.Error = err.Error()
	}
	e.r.Status = rpb.CompletionStatus_FINISHED
}

type parentTest struct {
	ti     subscribe.Subscribe
	subRes *rpb.SubscribeTestResult
	finish func()
}

func (p *parentTest) Update(l interface{}) {
	sr := l.(*gpb.SubscribeResponse)
	status, err := p.ti.Process(sr)

	// update parent test report.
	srr := &rpb.SubscribeResponseResult{Response: sr}
	if err != nil {
		srr.Error = err.Error()
	}
	p.subRes.Responses = append(p.subRes.Responses, srr)

	if status == subscribe.Complete {
		p.subRes.Status = rpb.CompletionStatus_EARLY_FINISHED
		p.finish()
	}
}

func (p *parentTest) End() {
	if err := p.ti.Check(); err != nil {
		p.subRes.Error = err.Error()
	}
	if p.subRes.Status == rpb.CompletionStatus_UNKNOWN {
		p.subRes.Status = rpb.CompletionStatus_FINISHED
	}
}

func setTestResult(ir *rpb.Instance) {
	anyFailed := false
	// set results of extension tests.
	for _, extRes := range ir.Extensions {
		var err error
		if v, ok := extRes.Type.(*rpb.TestResult_Subscribe); ok {
			err = getSubscribeTestResult(v.Subscribe)
		}

		if err == nil {
			extRes.Result = rpb.Status_SUCCESS
		} else {
			anyFailed = true
			extRes.Result = rpb.Status_FAIL
		}
	}

	// set parent test result.
	var err error
	if v, ok := ir.Test.Type.(*rpb.TestResult_Subscribe); ok {
		err = getSubscribeTestResult(v.Subscribe)
	}

	switch {
	case err == nil && !anyFailed:
		ir.Test.Result = rpb.Status_SUCCESS
	default:
		// if any of the extensions failed, parent test is reported as failed.
		ir.Test.Result = rpb.Status_FAIL
	}
}

func getSubscribeTestResult(sRes *rpb.SubscribeTestResult) error {
	for _, res := range sRes.Responses {
		if res.Error != "" {
			return fmt.Errorf("failure returned for %v; %v", res.Response, res.Error)
		}
	}
	if sRes.Error != "" {
		return fmt.Errorf("failure returned; %v", sRes.Error)
	}
	return nil
}
