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

package getsetv

import (
	"testing"
	"time"

	"context"
	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/gnmi/errdiff"
	"github.com/openconfig/gnmi/testing/fake/gnmi"
	"github.com/openconfig/gnmitest/common"
	"github.com/openconfig/ygot/ygot"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	fpb "github.com/openconfig/gnmi/testing/fake/proto"
	rpb "github.com/openconfig/gnmitest/proto/report"
	spb "github.com/openconfig/gnmitest/proto/suite"
	tpb "github.com/openconfig/gnmitest/proto/tests"
)

func mustPath(s string) *gpb.Path {
	p, err := ygot.StringToStructuredPath(s)
	if err != nil {
		panic(err)
	}
	return p
}

func TestResolveOper(t *testing.T) {
	tests := []struct {
		name             string
		inOper           *tpb.GetSetValidationOper
		inLib            *spb.CommonMessages
		want             *tpb.GetSetValidationOper
		wantErrSubstring string
	}{{
		name: "no messages to resolve",
		inOper: &tpb.GetSetValidationOper{
			Setrequest: &tpb.GetSetValidationOper_Set{
				&gpb.SetRequest{
					Delete: []*gpb.Path{
						mustPath("/interfaces"),
					},
				},
			},
			Getrequest: &tpb.GetSetValidationOper_Get{
				&gpb.GetRequest{
					Path: []*gpb.Path{
						mustPath("/interfaces"),
					},
				},
			},
		},
		want: &tpb.GetSetValidationOper{
			Setrequest: &tpb.GetSetValidationOper_Set{
				&gpb.SetRequest{
					Delete: []*gpb.Path{
						mustPath("/interfaces"),
					},
				},
			},
			Getrequest: &tpb.GetSetValidationOper_Get{
				&gpb.GetRequest{
					Path: []*gpb.Path{
						mustPath("/interfaces"),
					},
				},
			},
		},
	}, {
		name: "resolve common setrequest",
		inOper: &tpb.GetSetValidationOper{
			Setrequest: &tpb.GetSetValidationOper_CommonSetrequest{"setname"},
		},
		inLib: &spb.CommonMessages{
			SetRequests: map[string]*gpb.SetRequest{
				"setname": {
					Delete: []*gpb.Path{
						mustPath("/interfaces"),
					},
				},
			},
		},
		want: &tpb.GetSetValidationOper{
			Setrequest: &tpb.GetSetValidationOper_Set{
				&gpb.SetRequest{
					Delete: []*gpb.Path{
						mustPath("/interfaces"),
					},
				},
			},
		},
	}, {
		name: "resolve common getrequest",
		inOper: &tpb.GetSetValidationOper{
			Getrequest: &tpb.GetSetValidationOper_CommonGetrequest{"getname"},
		},
		inLib: &spb.CommonMessages{
			GetRequests: map[string]*gpb.GetRequest{
				"getname": {
					Path: []*gpb.Path{
						mustPath("/interfaces"),
					},
				},
			},
		},
		want: &tpb.GetSetValidationOper{
			Getrequest: &tpb.GetSetValidationOper_Get{
				&gpb.GetRequest{
					Path: []*gpb.Path{
						mustPath("/interfaces"),
					},
				},
			},
		},
	}, {
		name: "missing setrequest",
		inOper: &tpb.GetSetValidationOper{
			Setrequest: &tpb.GetSetValidationOper_CommonSetrequest{"invalid"},
		},
		inLib:            &spb.CommonMessages{},
		wantErrSubstring: "cannot look up common SetRequest",
	}, {
		name: "missing getrequest",
		inOper: &tpb.GetSetValidationOper{
			Getrequest: &tpb.GetSetValidationOper_CommonGetrequest{"invalid"},
		},
		inLib:            &spb.CommonMessages{},
		wantErrSubstring: "cannot look up common GetRequest",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveOper(tt.inOper, tt.inLib)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("did not get expected error, %s", diff)
			}

			if diff := pretty.Compare(got, tt.want); diff != "" {
				t.Fatalf("did not get expected operation, diff(-got,+want):\n%s", diff)
			}
		})
	}
}

func TestGetSetValidateInternal(t *testing.T) {
	a, err := gnmi.New(&fpb.Config{
		Target: "test",
	}, nil)
	if err != nil {
		t.Fatalf("cannot start fake gNMI server, %v", err)
	}

	go func() {
		if err := a.Serve(); err != nil {
			t.Fatalf("gNMI agent failed, err: %v", err)
		}
	}()
	defer a.Close()

	commonConnectionArgs := &common.ConnectionArgs{
		Address: a.Address(),
		Timeout: 3,
	}

	tests := []struct {
		name             string
		inTestInst       *tpb.GetSetValidationTest
		inSpec           *Specification
		wantResult       *rpb.Instance
		wantErrSubstring string
	}{{
		name: "simple get failed - fake does not support Get",
		inTestInst: &tpb.GetSetValidationTest{
			TestOper: &tpb.GetSetValidationOper{
				Getrequest: &tpb.GetSetValidationOper_Get{
					&gpb.GetRequest{
						Path: []*gpb.Path{
							mustPath("/interfaces"),
						},
					},
				},
			},
		},
		inSpec: &Specification{
			Target: commonConnectionArgs,
			Result: &rpb.Instance{},
		},
		wantResult: &rpb.Instance{
			Test: &rpb.TestResult{
				Result: rpb.Status_FAIL,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := getSetValidationInternal(ctx, tt.inTestInst, tt.inSpec); err != nil {
				if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
					t.Fatalf("did not get expected error, %s", diff)
				}
			}

			if diff := pretty.Compare(tt.inSpec.Result, tt.wantResult); diff != "" {
				t.Fatalf("did not get expected result, diff(-got,+want):\n%s", diff)
			}
		})
	}
}
