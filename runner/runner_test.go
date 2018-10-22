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

package runner

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/gnmi/client"
	"github.com/openconfig/gnmitest/config"
	"github.com/openconfig/gnmitest/register"
	"github.com/openconfig/gnmitest/subscribe"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	rpb "github.com/openconfig/gnmitest/proto/report"
	spb "github.com/openconfig/gnmitest/proto/suite"
	tpb "github.com/openconfig/gnmitest/proto/tests"
)

var (
	tests = map[string]subscribe.Subscribe{
		"test11": &test1{},
		"test12": &test1{},
	}
)

func TestMain(m *testing.M) {
	register.NewSubscribeTest(&tpb.SubscribeTest_FakeTest{}, newFake)
	os.Exit(m.Run())
}

func newFake(t *tpb.Test) (subscribe.Subscribe, error) {
	s := t.GetSubscribe().GetFakeTest()
	test, ok := tests[s]
	if !ok {
		return nil, fmt.Errorf("no such test: %q", s)
	}
	return test, nil
}

type test1 struct {
	subscribe.Test
}

// Process is placeholder to satisfy subscribe.Subscribe interface.
func (test1) Process(sr *gpb.SubscribeResponse) (subscribe.Status, error) {
	return subscribe.Running, nil
}

func createTest(test, target string, sub *gpb.Path, mode gpb.SubscriptionList_Mode, logResponses bool) *tpb.Test {
	return &tpb.Test{
		Description: test,
		Type: &tpb.Test_Subscribe{
			Subscribe: &tpb.SubscribeTest{
				Request: &gpb.SubscribeRequest{
					Request: &gpb.SubscribeRequest_Subscribe{
						Subscribe: &gpb.SubscriptionList{
							Prefix: &gpb.Path{
								Target: target,
							},
							Mode:         mode,
							Subscription: []*gpb.Subscription{{Path: sub}},
						},
					},
				},
				LogResponses: logResponses,
				Args: &tpb.SubscribeTest_FakeTest{
					FakeTest: test,
				},
			},
		},
	}
}

func createNoti(t string, p *gpb.Path, sync bool) *gpb.SubscribeResponse {
	if sync {
		return &gpb.SubscribeResponse{
			Response: &gpb.SubscribeResponse_SyncResponse{
				SyncResponse: true,
			},
		}
	}
	return &gpb.SubscribeResponse{
		Response: &gpb.SubscribeResponse_Update{
			Update: &gpb.Notification{
				Prefix: &gpb.Path{Target: t},
				Update: []*gpb.Update{{Path: p}},
			},
		},
	}
}

func configFromProto(s *spb.Suite) (*config.Config, error) {
	b := &bytes.Buffer{}
	if err := proto.MarshalText(b, s); err != nil {
		return nil, fmt.Errorf("cannot unmarshal configuration, %v", err)
	}
	cfg, err := config.New(b.Bytes(), "<client type>")
	if err != nil {
		return nil, fmt.Errorf("cannot create config, %v", err)
	}
	return cfg, nil
}

func TestRunnerSequentialInstances(t *testing.T) {
	tests := []struct {
		desc    string
		spb     *spb.Suite
		upd     []*gpb.SubscribeResponse
		want    map[string][]*gpb.SubscribeResponse
		timeout int
		comSt   rpb.CompletionStatus
	}{
		{
			desc: "test finishes before timeout",
			spb: &spb.Suite{
				Name:       "",
				Connection: &tpb.Connection{},
				Timeout:    4,
				InstanceGroupList: []*spb.InstanceGroup{
					{
						Description: "instance group",
						Instance: []*spb.Instance{
							{
								Description:   "instance with query path a",
								ExtensionList: []string{"exts"},
								Test:          createTest("test11", "dev1", &gpb.Path{Element: []string{"a"}}, gpb.SubscriptionList_ONCE, true),
							},
						},
					},
				},
				ExtensionList: map[string]*spb.ExtensionList{
					"exts": {
						Extension: []*tpb.Test{
							createTest("test11", "dev1", &gpb.Path{Element: []string{"a", "b"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test12", "dev1", &gpb.Path{Element: []string{"a", "c"}}, gpb.SubscriptionList_STREAM, true),
						},
					},
				},
			},
			upd: []*gpb.SubscribeResponse{
				createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "c"}}, false),
				createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "d"}}, false),
				createNoti("dev1", &gpb.Path{Element: []string{"a", "c", "d"}}, false),
				createNoti("dev1", nil, true),
			},
			want: map[string][]*gpb.SubscribeResponse{
				"test11": {
					createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "c"}}, false),
					createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "d"}}, false),
					createNoti("dev1", nil, true)},
				"test12": {
					createNoti("dev1", &gpb.Path{Element: []string{"a", "c", "d"}}, false),
					createNoti("dev1", nil, true)},
			},
			timeout: 2,
			comSt:   rpb.CompletionStatus_FINISHED,
		},
		{
			desc: "test finishes before timeout - no logging",
			spb: &spb.Suite{
				Name:       "",
				Connection: &tpb.Connection{},
				Timeout:    4,
				InstanceGroupList: []*spb.InstanceGroup{
					{
						Description: "instance group",
						Instance: []*spb.Instance{
							{
								Description:   "instance with query path a",
								ExtensionList: []string{"exts"},
								Test:          createTest("test11", "dev1", &gpb.Path{Element: []string{"a"}}, gpb.SubscriptionList_ONCE, false),
							},
						},
					},
				},
				ExtensionList: map[string]*spb.ExtensionList{
					"exts": {
						Extension: []*tpb.Test{
							createTest("test11", "dev1", &gpb.Path{Element: []string{"a", "b"}}, gpb.SubscriptionList_STREAM, false),
							createTest("test12", "dev1", &gpb.Path{Element: []string{"a", "c"}}, gpb.SubscriptionList_STREAM, false),
						},
					},
				},
			},
			upd: []*gpb.SubscribeResponse{
				createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "c"}}, false),
				createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "d"}}, false),
				createNoti("dev1", &gpb.Path{Element: []string{"a", "c", "d"}}, false),
				createNoti("dev1", nil, true),
			},
			want: map[string][]*gpb.SubscribeResponse{
				"test11": {},
				"test12": {},
			},
			timeout: 2,
			comSt:   rpb.CompletionStatus_FINISHED,
		},
		{
			desc: "test timeouts",
			spb: &spb.Suite{
				Timeout:    2,
				Connection: &tpb.Connection{},
				InstanceGroupList: []*spb.InstanceGroup{
					{
						Description: "instance group",
						Instance: []*spb.Instance{
							{
								Description:   "instance with query path a",
								ExtensionList: []string{"exts"},
								Test:          createTest("test11", "dev1", &gpb.Path{Element: []string{"a"}}, gpb.SubscriptionList_STREAM, true),
							},
						},
					},
				},
				ExtensionList: map[string]*spb.ExtensionList{
					"exts": {
						Extension: []*tpb.Test{
							createTest("test11", "dev1", &gpb.Path{Element: []string{"a", "b"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test12", "dev1", &gpb.Path{Element: []string{"a", "c"}}, gpb.SubscriptionList_STREAM, true),
						},
					},
				},
			},
			upd: []*gpb.SubscribeResponse{
				createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "c"}}, false),
				createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "d"}}, false),
				createNoti("dev1", &gpb.Path{Element: []string{"a", "c", "d"}}, false),
				createNoti("dev1", nil, true),
			},
			want: map[string][]*gpb.SubscribeResponse{
				"test11": {
					createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "c"}}, false),
					createNoti("dev1", &gpb.Path{Element: []string{"a", "b", "d"}}, false),
					createNoti("dev1", nil, true)},
				"test12": {
					createNoti("dev1", &gpb.Path{Element: []string{"a", "c", "d"}}, false),
					createNoti("dev1", nil, true)},
			},
			timeout: 4,
			comSt:   rpb.CompletionStatus_TIMEOUT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cfg, err := configFromProto(tt.spb)
			if err != nil {
				t.Fatalf("config.New failed: %v", err)
			}

			createSubscription = func(ctx context.Context, _ *gpb.SubscribeRequest, pHandler client.ProtoHandler, _ *tpb.Connection, _ string) error {
				// send all the updates upon calling createSubscription
				for _, r := range tt.upd {
					pHandler(r)
				}
				// wait until either context times out or time set for testing times out
				select {
				case <-ctx.Done():
				case <-time.After(time.Duration(tt.timeout) * time.Second):
				}
				return nil
			}
			var reports []*rpb.InstanceGroup
			r := New(cfg, func(ir *rpb.InstanceGroup) {
				reports = append(reports, ir)
			})
			if err := r.Start(context.Background()); err != nil {
				t.Fatalf("error occurred during test execution %v", err)
			}
			if len(reports) != len(tt.spb.InstanceGroupList) {
				t.Errorf("got %d, want %d instance group reports", len(reports), len(tt.spb.InstanceGroupList))
			}
			for _, rep := range reports {
				if len(rep.Instance) != len(tt.spb.InstanceGroupList[0].Instance) {
					t.Fatalf("got %d instance report want %d instance report", len(rep.Instance), len(tt.spb.InstanceGroupList[0].Instance))
				}
				for i, ext := range rep.Instance[0].Extensions {
					got := ext.GetSubscribe().Responses
					want := tt.want[ext.Test.Description]
					if len(got) != len(want) {
						t.Fatalf("extension #%d: got %v, want %v updates", i, len(got), len(want))
					}
					for j, r := range got {
						if diff := cmp.Diff(r.Response, want[j]); diff != "" {
							t.Errorf("extension #%d: got %v, want %v:\n%v", i, r.Response, want[j], diff)
						}
					}
				}
				if rep.Instance[0].Test.GetSubscribe().Status != tt.comSt {
					t.Errorf("got %v, want %v completion status", rep.Instance[0].Test.GetSubscribe().Status, tt.comSt)
				}
			}
		})
	}
}

func processHookGenerator(m map[string]int, stateful bool) func(string) (subscribe.Status, error) {
	return func(test string) (subscribe.Status, error) {
		i, ok := m[test]
		if !ok {
			return subscribe.Running, nil
		}
		if i == 0 {
			delete(m, test)
			if stateful {
				return subscribe.Complete, nil
			}
			return subscribe.Running, errors.New("closure parameter is exhausted")
		}
		m[test]--
		return subscribe.Running, nil
	}
}

type test2 struct {
	subscribe.Test
	name string
}

var processHook func(string) (subscribe.Status, error)

func (t test2) Process(sr *gpb.SubscribeResponse) (subscribe.Status, error) {
	return processHook(t.name)
}

func TestParentResult(t *testing.T) {
	for i := 0; i < 10; i++ {
		tests[fmt.Sprintf("test2%d", i)] = &test2{name: fmt.Sprintf("test2%d", i)}
	}
	defer func() {
		for i := 0; i < 10; i++ {
			delete(tests, fmt.Sprintf("test2%d", i))
		}
	}()

	tests := []struct {
		inDesc  string
		inSuite *spb.Suite
		// Specifies after how many messages test returns error.
		inStats map[string]int
		// Specifies how many messages are going to be sent to the given path.
		numMessage map[string]int
		// Triggers test to return Complete status.
		stateful bool
		// Specifies the test result of the Instance.
		outResult rpb.Status
	}{
		{
			inDesc: "success - all tests pass",
			inSuite: &spb.Suite{
				Name:       "",
				Timeout:    3,
				Connection: &tpb.Connection{},
				InstanceGroupList: []*spb.InstanceGroup{
					{
						Description: "instance group",
						Instance: []*spb.Instance{
							{
								Description:   "instance with query path a",
								ExtensionList: []string{"exts"},
								Test:          createTest("test21", "dev1", &gpb.Path{Element: []string{"a"}}, gpb.SubscriptionList_STREAM, true),
							},
						},
					},
				},
				ExtensionList: map[string]*spb.ExtensionList{
					"exts": {
						Extension: []*tpb.Test{
							createTest("test22", "dev1", &gpb.Path{Element: []string{"a", "b"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test23", "dev1", &gpb.Path{Element: []string{"a", "c"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test24", "dev1", &gpb.Path{Element: []string{"a", "d"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test25", "dev1", &gpb.Path{Element: []string{"a", "e"}}, gpb.SubscriptionList_STREAM, true),
						},
					},
				},
			},
			numMessage: map[string]int{"a/b": 10, "a/c": 10, "a/d": 10, "a/e": 10},
			outResult:  rpb.Status_SUCCESS,
		},
		{
			inDesc: "fail - parent test fails",
			inSuite: &spb.Suite{
				Name:       "",
				Timeout:    3,
				Connection: &tpb.Connection{},
				InstanceGroupList: []*spb.InstanceGroup{
					{
						Description: "instance group",
						Instance: []*spb.Instance{
							{
								Description:   "instance with query path a",
								ExtensionList: []string{"exts"},
								Test:          createTest("test21", "dev1", &gpb.Path{Element: []string{"a"}}, gpb.SubscriptionList_STREAM, true),
							},
						},
					},
				},
				ExtensionList: map[string]*spb.ExtensionList{
					"exts": {
						Extension: []*tpb.Test{
							createTest("test22", "dev1", &gpb.Path{Element: []string{"a", "b"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test23", "dev1", &gpb.Path{Element: []string{"a", "c"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test24", "dev1", &gpb.Path{Element: []string{"a", "d"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test25", "dev1", &gpb.Path{Element: []string{"a", "e"}}, gpb.SubscriptionList_STREAM, true),
						},
					},
				},
			},
			inStats:    map[string]int{"test21": 30},
			numMessage: map[string]int{"a/b": 10, "a/c": 10, "a/d": 10, "a/e": 10},
			outResult:  rpb.Status_FAIL,
		},
		{
			inDesc: "fail - extension test fails",
			inSuite: &spb.Suite{
				Name:       "",
				Timeout:    3,
				Connection: &tpb.Connection{},
				InstanceGroupList: []*spb.InstanceGroup{
					{
						Description: "instance group",
						Instance: []*spb.Instance{
							{
								Description:   "instance with query path a",
								ExtensionList: []string{"exts"},
								Test:          createTest("test21", "dev1", &gpb.Path{Element: []string{"a"}}, gpb.SubscriptionList_STREAM, true),
							},
						},
					},
				},
				ExtensionList: map[string]*spb.ExtensionList{
					"exts": {
						Extension: []*tpb.Test{
							createTest("test22", "dev1", &gpb.Path{Element: []string{"a", "b"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test23", "dev1", &gpb.Path{Element: []string{"a", "c"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test24", "dev1", &gpb.Path{Element: []string{"a", "d"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test25", "dev1", &gpb.Path{Element: []string{"a", "e"}}, gpb.SubscriptionList_STREAM, true),
						},
					},
				},
			},
			inStats:    map[string]int{"test22": 5},
			numMessage: map[string]int{"a/b": 10, "a/c": 10, "a/d": 10, "a/e": 10},
			outResult:  rpb.Status_FAIL,
		},
		{
			inDesc: "success - parent test finishes early due to returning Complete status",
			inSuite: &spb.Suite{
				Name:       "",
				Timeout:    3,
				Connection: &tpb.Connection{},
				InstanceGroupList: []*spb.InstanceGroup{
					{
						Description: "instance group",
						Instance: []*spb.Instance{
							{
								Description:   "instance with query path a",
								ExtensionList: []string{"exts"},
								Test:          createTest("test21", "dev1", &gpb.Path{Element: []string{"a"}}, gpb.SubscriptionList_STREAM, true),
							},
						},
					},
				},
				ExtensionList: map[string]*spb.ExtensionList{
					"exts": {
						Extension: []*tpb.Test{
							createTest("test22", "dev1", &gpb.Path{Element: []string{"a", "b"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test23", "dev1", &gpb.Path{Element: []string{"a", "c"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test24", "dev1", &gpb.Path{Element: []string{"a", "d"}}, gpb.SubscriptionList_STREAM, true),
							createTest("test25", "dev1", &gpb.Path{Element: []string{"a", "e"}}, gpb.SubscriptionList_STREAM, true),
						},
					},
				},
			},
			inStats:    map[string]int{"test21": 5},
			numMessage: map[string]int{"a/b": 10, "a/c": 10, "a/d": 10, "a/e": 10},
			stateful:   true,
			outResult:  rpb.Status_SUCCESS,
		},
	}

	for testIndex, tt := range tests {
		t.Run(tt.inDesc, func(t *testing.T) {
			b := &bytes.Buffer{}
			if err := proto.MarshalText(b, tt.inSuite); err != nil {
				t.Fatalf("marshalling proto failed: %v", err)
			}
			cfg, err := config.New(b.Bytes(), "<client type>")
			if err != nil {
				t.Fatalf("config.New failed: %v", err)
			}

			createSubscription = func(ctx context.Context, _ *gpb.SubscribeRequest, pHandler client.ProtoHandler, _ *tpb.Connection, _ string) error {
				// send all the updates upon calling createSubscription
				for k, v := range tt.numMessage {
					for i := 0; i < v; i++ {
						pHandler(createNoti("dev1", &gpb.Path{Element: strings.Split(k, "/")}, false))
					}
				}

				// wait until either context times out or time set for testing times out
				<-ctx.Done()
				return ctx.Err()
			}
			var reports []*rpb.InstanceGroup
			r := New(cfg, func(ir *rpb.InstanceGroup) {
				reports = append(reports, ir)
			})
			processHook = processHookGenerator(tt.inStats, tt.stateful)
			if err := r.Start(context.Background()); err != nil {
				t.Fatalf("error occurred during test execution %v", err)
			}
			switch {
			case len(reports) != len(tt.inSuite.InstanceGroupList):
				t.Fatalf("#%d %s got %d, want %d instance groups", testIndex, tt.inDesc, len(reports), len(tt.inSuite.InstanceGroupList))
			case len(reports[0].Instance) != len(tt.inSuite.InstanceGroupList[0].Instance):
				t.Fatalf("#%d %s got %d, want %d instances", testIndex, tt.inDesc, len(reports[0].Instance), len(tt.inSuite.InstanceGroupList[0].Instance))
			}
			if reports[0].Instance[0].Test.Result != tt.outResult {
				t.Fatalf("#%d %s got %v, want %v report status", testIndex, tt.inDesc, reports[0].Instance[0].Test.Result, tt.outResult)
			}
		})
	}
}

func TestSimpleRunner(t *testing.T) {
	tests := []struct {
		name       string
		inSuite    *spb.Suite
		wantReport *rpb.Report
	}{{
		name: "one instance group - pass",
		inSuite: &spb.Suite{
			Name:       "simple suite",
			Connection: &tpb.Connection{},
			InstanceGroupList: []*spb.InstanceGroup{{
				Description: "instance group 1",
				Instance: []*spb.Instance{{
					Description: "group 1, test 1",
					Test: &tpb.Test{
						Description: "group 1, test1a",
						Type:        &tpb.Test_FakeTest{&tpb.FakeTest{Pass: true}},
					},
				}, {
					Description: "group 1, test 2",
					Test: &tpb.Test{
						Description: "group1, test2a",
						Type:        &tpb.Test_FakeTest{&tpb.FakeTest{Pass: true}},
					},
				}},
			}},
		},
		wantReport: &rpb.Report{
			Results: []*rpb.InstanceGroup{{
				Description: "instance group 1",
				Instance: []*rpb.Instance{{
					Description: "group 1, test 1",
					Test:        &rpb.TestResult{Result: rpb.Status_SUCCESS},
				}, {
					Description: "group 1, test 2",
					Test:        &rpb.TestResult{Result: rpb.Status_SUCCESS},
				}},
			}},
		},
	}, {
		name: "two instance groups - pass",
		inSuite: &spb.Suite{
			Name:       "simple suite",
			Connection: &tpb.Connection{},
			InstanceGroupList: []*spb.InstanceGroup{{
				Description: "instance group 1",
				Instance: []*spb.Instance{{
					Description: "group 1, test 1",
					Test: &tpb.Test{
						Description: "group 1, test 1a",
						Type:        &tpb.Test_FakeTest{&tpb.FakeTest{Pass: true}},
					},
				}},
			}, {
				Description: "instance group 2",
				Instance: []*spb.Instance{{
					Description: "group 2, test 1",
					Test: &tpb.Test{
						Description: "group 2, test 1a",
						Type:        &tpb.Test_FakeTest{&tpb.FakeTest{Pass: true}},
					},
				}},
			}},
		},
		wantReport: &rpb.Report{
			Results: []*rpb.InstanceGroup{{
				Description: "instance group 1",
				Instance: []*rpb.Instance{{
					Description: "group 1, test 1",
					Test:        &rpb.TestResult{Result: rpb.Status_SUCCESS},
				}},
			}, {
				Description: "instance group 2",
				Instance: []*rpb.Instance{{
					Description: "group 2, test 1",
					Test:        &rpb.TestResult{Result: rpb.Status_SUCCESS},
				}},
			}},
		},
	}, {
		name: "two instance groups, first fails, not fatal",
		inSuite: &spb.Suite{
			Name:       "simple suite",
			Connection: &tpb.Connection{},
			InstanceGroupList: []*spb.InstanceGroup{{
				Description: "instance group 1",
				Instance: []*spb.Instance{{
					Description: "group 1, test 1",
					Test: &tpb.Test{
						Description: "group 1, test 1a",
						Type:        &tpb.Test_FakeTest{&tpb.FakeTest{Pass: false}},
					},
				}},
			}, {
				Description: "instance group 2",
				Instance: []*spb.Instance{{
					Description: "group 2, test 1",
					Test: &tpb.Test{
						Description: "group 2, test 1a",
						Type:        &tpb.Test_FakeTest{&tpb.FakeTest{Pass: true}},
					},
				}},
			}},
		},
		wantReport: &rpb.Report{
			Results: []*rpb.InstanceGroup{{
				Description: "instance group 1",
				Instance: []*rpb.Instance{{
					Description: "group 1, test 1",
					Test:        &rpb.TestResult{Result: rpb.Status_FAIL},
				}},
			}, {
				Description: "instance group 2",
				Instance: []*rpb.Instance{{
					Description: "group 2, test 1",
					Test:        &rpb.TestResult{Result: rpb.Status_SUCCESS},
				}},
			},
			}},
	}, {
		name: "two instance groups, first fails, fatal",
		inSuite: &spb.Suite{
			Name:       "simple suite",
			Connection: &tpb.Connection{},
			InstanceGroupList: []*spb.InstanceGroup{{
				Description: "instance group 1",
				Fatal:       true,
				Instance: []*spb.Instance{{
					Description: "group 1, test 1",
					Test: &tpb.Test{
						Description: "group 1, test 1a",
						Type:        &tpb.Test_FakeTest{&tpb.FakeTest{Pass: false}},
					},
				}},
			}, {
				Description: "instance group 2",
				Instance: []*spb.Instance{{
					Description: "group 2, test 1",
					Test: &tpb.Test{
						Description: "group 2, test 1a",
						Type:        &tpb.Test_FakeTest{&tpb.FakeTest{Pass: true}},
					},
				}},
			}},
		},
		wantReport: &rpb.Report{
			Results: []*rpb.InstanceGroup{{
				Description: "instance group 1",
				Instance: []*rpb.Instance{{
					Description: "group 1, test 1",
					Test:        &rpb.TestResult{Result: rpb.Status_FAIL},
				}},
			}, {
				Description: "instance group 2",
				Skipped:     true,
			}},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := configFromProto(tt.inSuite)
			if err != nil {
				t.Fatalf("cannot create config, got: %v, want: nil", err)
			}

			got := &rpb.Report{}
			r := New(cfg, func(ig *rpb.InstanceGroup) {
				got.Results = append(got.Results, ig)
			})

			if err := r.Start(context.Background()); err != nil {
				t.Fatalf("error occurred during test execution, %v", err)
			}

			if diff := pretty.Compare(got, tt.wantReport); diff != "" {
				t.Fatalf("did not get expected result, diff(-got,+want):\n%s", diff)
			}
		})
	}
}
