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

// Package getsetv defines the logic that implements the GetSetValidate tests
// for the gnmitest framework. These tests are described in tests.proto in
// detail. The procedure implemented by the test is
//
//   1. Perform an initial Set operation, which can be used to initialise the
//      target system to a known good state.
//   2. Perform a Get operation, whose result is compared against a specified
//      GetResponse.
//   3. Perform a Set operation, which can be used to test a particular
//      behaviour of the system.
//   4. Perform a Get operation, whose result is again compared against a
//      specified GetResponse.
//
// The operations are paired such that 1+2, and 3+4 are considered together.
// Optionally, in each pair the Get or Set operation can be omitted. The
// initial operations (i.e., 1+2) may also be omitted. This test therefore
// allows a sequence of Get+Set tests, as well as individual tests for Get and
// Set if required.
//
// All operations are performed sequentially. The failure of any one of the
// operations is considered fatal for the test.
package getsetv

import (
	"fmt"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmitest/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	rpb "github.com/openconfig/gnmitest/proto/report"
	spb "github.com/openconfig/gnmitest/proto/suite"
	tpb "github.com/openconfig/gnmitest/proto/tests"
)

// Specification defines the parameters for a GetSetValidate test.
type Specification struct {
	Target         *common.ConnectionArgs // Target defines the connection to the target to be used.
	Instance       *spb.Instance          // Instance is the test instance that is being executed.
	Result         *rpb.Instance          // Result is the result of the test instance that should be written to.
	CommonRequests *spb.CommonMessages    // CommonRequests is the library of common messages that can be referenced by the test.
}

// getSetValidationInternal runs the GetSetValidate test specified in testInst
// using the specification provided.  It returns an error if encountered in the
// test, and writes it result to the Instance protobuf in the Specification.
func getSetValidationInternal(ctx context.Context, testInst *tpb.GetSetValidationTest, spec *Specification) error {
	conn, cleanup, err := common.Connect(ctx, spec.Target)
	if err != nil {
		return fmt.Errorf("cannot connect to target: %v", err)
	}
	defer cleanup()

	o, err := resolveOper(testInst.GetInitialiseOper(), spec.CommonRequests)
	if err != nil {
		return fmt.Errorf("cannot resolve initialise oper, %v", err)
	}

	t, err := resolveOper(testInst.GetTestOper(), spec.CommonRequests)
	if err != nil {
		return fmt.Errorf("cannot resolve test oper, %v", err)
	}

	ti := proto.Clone(testInst).(*tpb.GetSetValidationTest)
	ti.InitialiseOper = o
	ti.TestOper = t

	iRes, err := doOper(ctx, conn, ti.InitialiseOper)
	if err != nil {
		return fmt.Errorf("cannot run initialise oper, got error: %v", err)
	}

	tRes, err := doOper(ctx, conn, ti.TestOper)
	if err != nil {
		return fmt.Errorf("cannot run test oper, got err: %v", err)
	}

	var res rpb.Status
	// TODO(robjs): Report more status of the test in the result.
	switch {
	case iRes.isPass() && tRes.isPass():
		res = rpb.Status_SUCCESS
	case !iRes.isPass() || !tRes.isPass():
		res = rpb.Status_FAIL
	}

	spec.Result.Test = &rpb.TestResult{
		Test:   spec.Instance.GetTest(),
		Result: res,
	}

	return nil
}

// resolveOper looks up the common messages that are specified in the supplied
// oper in the lib provided. The fully resolved GetSetValidationOper is returned
// to the caller.
func resolveOper(oper *tpb.GetSetValidationOper, lib *spb.CommonMessages) (*tpb.GetSetValidationOper, error) {
	gn := oper.GetCommonGetrequest()
	sn := oper.GetCommonSetrequest()
	switch {
	case gn == "" && sn == "":
		return oper, nil
	case (gn != "" || sn != "") && lib == nil:
		return nil, fmt.Errorf("cannot look up common requests (Get: %s, Set: %s), nil library", gn, sn)
	case gn != "" && lib.GetRequests == nil:
		return nil, fmt.Errorf("cannot look up common GetRequest %s, nil GetRequest library", gn)
	case sn != "" && lib.SetRequests == nil:
		return nil, fmt.Errorf("cannot look up common SetRequest %s, nil SetRequest library", sn)
	}

	rt := proto.Clone(oper).(*tpb.GetSetValidationOper)
	if gn != "" {
		req, ok := lib.GetRequests[gn]
		if !ok {
			return nil, fmt.Errorf("common GetRequest %s does not exist", gn)
		}
		rt.Getrequest = &tpb.GetSetValidationOper_Get{req}
	}

	if sn != "" {
		req, ok := lib.SetRequests[sn]
		if !ok {
			return nil, fmt.Errorf("common SetRequest %s does not exist", sn)
		}
		rt.Setrequest = &tpb.GetSetValidationOper_Set{req}
	}
	return rt, nil
}

// operResult records the result of a GetValidationOper sub-test.
type operResult struct {
	getResult   codes.Code       // getResult records the RPC return of the Get within the operation.
	getResponse *gpb.GetResponse // getResponse stores the gNMI GetResponse received from the target.
	setResult   codes.Code       // setResult records the RPC return of the Set within the operation.
	setResponse *gpb.SetResponse // setResponse stores the gNMI SetResponse received from the target.
}

// isPass specifies whether a particular operation has passed or failed.
func (o *operResult) isPass() bool {
	return o.getResult == codes.OK && o.setResult == codes.OK
}

// doOper runs the specified oper test operation, and returns its result.
func doOper(ctx context.Context, conn gpb.GNMIClient, oper *tpb.GetSetValidationOper) (*operResult, error) {
	// TODO(robjs): Enhance doOper to store the GetResponse and SetResponse. This
	// initial implementation only stores the error codes to allow for an initial
	// test to be implemented.
	r := &operResult{}
	if g := oper.GetGet(); g != nil {
		_, err := conn.Get(ctx, g)
		r.getResult = toCode(err)
	}

	if s := oper.GetSet(); s != nil {
		_, err := conn.Set(ctx, s)
		r.setResult = toCode(err)
	}

	return r, nil
}

// toCode returns the error supplied as a status code. If the error is not a
// valid status then the Invalid error code is used.
func toCode(err error) codes.Code {
	s, ok := status.FromError(err)
	if !ok {
		return codes.Unknown
	}
	return s.Code()
}
