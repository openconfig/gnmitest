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

// Package pathvalidation implements subscribe.Test interface and registers
// factory function to registry. It validates whether paths in received gNMI
// notifications are OpenConfig schema compliant.
package pathvalidation

import (
	"fmt"
	"reflect"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
	"github.com/openconfig/gnmitest/register"
	"github.com/openconfig/gnmitest/schemas"
	"github.com/openconfig/gnmitest/subscribe"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	tpb "github.com/openconfig/gnmitest/proto/tests"
)

// test stores the information that doesn't change during the course
// of message stream.
type test struct {
	subscribe.Test
	destStruct ygot.GoStruct
	schema     *yang.Entry
}

// init registers the factory function of the test to global tests registry.
func init() {
	register.NewSubscribeTest(&tpb.SubscribeTest_PathValidation{}, newTest)
}

// newTest is used as a callback by registry to instantiate the test when needed.
// It receives the SubscribeTest proto which contains the arguments to the test
// as well gNMI SubscribeRequest. This test uses neither any arguments nor the
// subscription request.
func newTest(st *tpb.Test) (subscribe.Subscribe, error) {
	goStruct, err := schema.Get(st.GetSchema())
	if err != nil {
		return nil, fmt.Errorf("failed to get %v schema; %v", st.GetSchema(), err)
	}
	// Device GoStruct is the root container within generated GoStructs.
	destStruct := goStruct.NewRoot()
	tn := reflect.TypeOf(destStruct).Elem().Name()
	schema := goStruct.Schema(tn)
	if schema == nil {
		return nil, fmt.Errorf("schema not found; %v", tn)
	}

	return &test{
		destStruct: destStruct,
		schema:     schema,
	}, nil
}

// Process function is the interface function for subscribe.Test. It checks whether
// received gNMI SubscribeResponse is OpenConfig compliant.
func (t *test) Process(sr *gpb.SubscribeResponse) (subscribe.Status, error) {
	validatePath := func(p *gpb.Path) error {
		node, sch, err := ytypes.GetOrCreateNode(t.schema, t.destStruct, p)
		if err != nil {
			return err
		}
		if sch.IsLeaf() || sch.IsLeafList() {
			return nil
		}
		return fmt.Errorf("path doesn't point to a leaf node; %T", node)
	}
	switch v := sr.Response.(type) {
	case *gpb.SubscribeResponse_Update:
		var pe []*gpb.PathElem
		// Join prefix path and update/delete path.
		if v.Update.Prefix != nil {
			pe = append(pe, v.Update.Prefix.Elem...)
		}
		for _, u := range v.Update.Update {
			if u != nil && u.Path != nil {
				if err := validatePath(&gpb.Path{Elem: append(pe, u.Path.GetElem()...)}); err != nil {
					return subscribe.Running, err
				}
			}
		}
		for _, d := range v.Update.Delete {
			if d != nil {
				if err := validatePath(&gpb.Path{Elem: append(pe, d.GetElem()...)}); err != nil {
					return subscribe.Running, err
				}
			}
		}
		return subscribe.Running, nil
	case *gpb.SubscribeResponse_SyncResponse:
		// Sync response isn't used by this test, so just ignore it.
		return subscribe.Running, nil
	}
	return subscribe.Running, fmt.Errorf("unexpected message; %T", sr.Response)
}
