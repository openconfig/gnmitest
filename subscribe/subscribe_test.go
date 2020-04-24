package subscribe

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openconfig/gnmi/errdiff"
	"github.com/openconfig/gnmi/value"
	"github.com/openconfig/gnmitest/schemas/openconfig"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
)

func mustRootSchema(s *ytypes.Schema, err error) *yang.Entry {
	if err != nil {
		panic(err)
	}

	if !s.IsValid() {
		panic(fmt.Sprintf("invalid schema returned, %v", err))
	}

	return s.RootSchema()
}

func mustPath(s string) *gnmipb.Path {
	p, err := ygot.StringToStructuredPath(s)
	if err != nil {
		panic(fmt.Sprintf("cannot make path from %s, %v", s, err))
	}
	return p
}

func mustValue(v interface{}) *gnmipb.TypedValue {
	tv, err := value.FromScalar(v)
	if err != nil {
		panic(fmt.Sprintf("cannot make value from %v, %v", v, err))
	}
	return tv
}

func mustResponse(path string, val interface{}) *gnmipb.SubscribeResponse {
	return &gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			&gnmipb.Notification{
				Update: []*gnmipb.Update{{
					Path: mustPath(path),
					Val:  mustValue(val),
				}},
			},
		},
	}
}

func TestOneShotSetNode(t *testing.T) {
	tests := []struct {
		desc             string
		inSchema         *yang.Entry
		inRoot           ygot.GoStruct
		inUpdate         *gnmipb.SubscribeResponse
		inArgs           OneShotSetNodeArgs
		wantStatus       Status
		wantErrSubstring string
	}{{
		desc:     "successful deserialisation",
		inSchema: mustRootSchema(gostructs.Schema()),
		inRoot:   &gostructs.Device{},
		inUpdate: mustResponse("/interfaces/interface[name=eth0]/state/counters/in-octets", uint64(42)),
		inArgs: OneShotSetNodeArgs{
			YtypesArgs: []ytypes.SetNodeOpt{
				&ytypes.InitMissingElements{},
			},
		},
	}, {
		desc:     "error in deserialisation - invalid path",
		inSchema: mustRootSchema(gostructs.Schema()),
		inRoot:   &gostructs.Device{},
		inUpdate: mustResponse("/interfaces/interface[name=eth0]/qos/invalid-path", uint64(42)),
		inArgs: OneShotSetNodeArgs{
			YtypesArgs: []ytypes.SetNodeOpt{
				&ytypes.InitMissingElements{},
			},
		},
		wantErrSubstring: "no match found",
	}, {
		desc:     "ignored error in deserialisation - invalid path",
		inSchema: mustRootSchema(gostructs.Schema()),
		inRoot:   &gostructs.Device{},
		inUpdate: mustResponse("/interfaces/interface[name=eth0]/qos/invalid-path", uint64(42)),
		inArgs: OneShotSetNodeArgs{
			YtypesArgs: []ytypes.SetNodeOpt{
				&ytypes.InitMissingElements{},
			},
			IgnoreInvalidPaths: true,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := OneShotSetNode(tt.inSchema, tt.inRoot, tt.inUpdate, tt.inArgs)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("did not get expected error, %s", diff)
			}
			if diff := cmp.Diff(got, tt.wantStatus); diff != "" {
				t.Fatalf("did not get expected status, diff(-got,+want):\n%s", diff)
			}
		})
	}
}
