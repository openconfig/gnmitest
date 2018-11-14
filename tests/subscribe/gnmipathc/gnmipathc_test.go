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

package gnmipathc

import (
	"testing"

	"github.com/openconfig/gnmi/errdiff"
	"github.com/openconfig/gnmitest/schemas/openconfig/register"
	"github.com/openconfig/ygot/ygot"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	tpb "github.com/openconfig/gnmitest/proto/tests"
)

func mustPath(p string, slicePath bool) *gpb.Path {
	var path *gpb.Path
	var err error
	if slicePath {
		path, err = ygot.StringToStringSlicePath(p)
	} else {
		path, err = ygot.StringToStructuredPath(p)
	}
	if err != nil {
		panic(err)
	}
	return path
}

func noti(target, origin, prefixPath, updatePath string, v *gpb.TypedValue, slicePath bool) *gpb.SubscribeResponse {
	prefix := mustPath(prefixPath, slicePath)
	prefix.Target, prefix.Origin = target, origin
	return &gpb.SubscribeResponse{
		Response: &gpb.SubscribeResponse_Update{
			Update: &gpb.Notification{
				Prefix: prefix,
				Update: []*gpb.Update{
					&gpb.Update{
						Path: mustPath(updatePath, slicePath),
						Val:  v,
					},
				},
			},
		},
	}
}

func TestGNMIPathCompliance(t *testing.T) {
	tests := []struct {
		desc    string
		upd     *gpb.SubscribeResponse
		inArgs  *tpb.SubscribeTest
		wantErr string
	}{
		{
			desc: "target and origin must exist, but can be anything",
			upd:  noti("dev", "oc", "", "", nil, false),
			inArgs: &tpb.SubscribeTest{
				Args: &tpb.SubscribeTest_GnmipathCompliance{
					&tpb.GNMIPathCompliance{
						CheckTarget: "*",
						CheckOrigin: "*",
					},
				},
			},
		},
		{
			desc: "target and origin may be missing, but this is fine",
			upd:  noti("", "", "", "", nil, false),
			inArgs: &tpb.SubscribeTest{
				Args: &tpb.SubscribeTest_GnmipathCompliance{
					&tpb.GNMIPathCompliance{},
				},
			},
		},
		{
			desc: "target must exist, but missing",
			upd:  noti("", "", "", "", nil, false),
			inArgs: &tpb.SubscribeTest{
				Args: &tpb.SubscribeTest_GnmipathCompliance{
					&tpb.GNMIPathCompliance{
						CheckTarget: "*",
					},
				},
			},
			wantErr: "target isn't set in prefix gNMI Path",
		},
		{
			desc: "target and origin must exist, but they are missing",
			upd:  noti("", "", "", "", nil, false),
			inArgs: &tpb.SubscribeTest{
				Args: &tpb.SubscribeTest_GnmipathCompliance{
					&tpb.GNMIPathCompliance{
						CheckTarget: "*",
						CheckOrigin: "*",
					},
				},
			},
			wantErr: "target isn't set in prefix gNMI Path , origin isn't set in prefix gNMI Path ",
		},
		{
			desc: `target must exist and equal to "myDev"`,
			upd:  noti("dev", "", "", "", nil, false),
			inArgs: &tpb.SubscribeTest{
				Args: &tpb.SubscribeTest_GnmipathCompliance{
					&tpb.GNMIPathCompliance{
						CheckTarget: "myDev",
					},
				},
			},
			wantErr: `target in gNMI Path target:"dev"  is "dev", expect "myDev"`,
		},
		{

			desc: `origin must exist and equal to "orig"`,
			upd:  noti("dev", "orig", "", "", nil, false),
			inArgs: &tpb.SubscribeTest{
				Args: &tpb.SubscribeTest_GnmipathCompliance{
					&tpb.GNMIPathCompliance{
						CheckOrigin: "orig",
					},
				},
			},
		},
		{
			desc: "target and origin must be set, but PathElem isn't used",
			upd:  noti("dev", "orig", "interfaces", "interface[name=arbitrary_key]/state/admin-status", nil, true),
			inArgs: &tpb.SubscribeTest{
				Args: &tpb.SubscribeTest_GnmipathCompliance{
					&tpb.GNMIPathCompliance{
						CheckOrigin: "*",
						CheckTarget: "*",
						CheckElem:   true,
					},
				},
			},
			wantErr: "element field is used in gNMI Path",
		},
		{
			desc: "target, origin and elem are all correctly used",
			upd:  noti("dev", "orig", "interfaces", "interface[name=arbitrary_key]/state/admin-status", nil, false),
			inArgs: &tpb.SubscribeTest{
				Args: &tpb.SubscribeTest_GnmipathCompliance{
					&tpb.GNMIPathCompliance{
						CheckOrigin: "*",
						CheckTarget: "*",
						CheckElem:   true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			gt, err := newTest(&tpb.Test{Schema: openconfig.Key, Type: &tpb.Test_Subscribe{Subscribe: tt.inArgs}})
			if err != nil {
				t.Fatalf("newTest failed: %v", err)
			}

			_, err = gt.Process(tt.upd)
			if diff := errdiff.Substring(err, tt.wantErr); diff != "" {
				t.Fatalf("did not get expected error, %s", diff)
			}
		})
	}
}
