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

package schema

import (
	"reflect"
	"testing"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/gnmi/errdiff"
	"github.com/openconfig/gnmitest/schemas/openconfig"
)

func reset() {
	goStructs = make(map[string]goStruct)
}

func TestSingleSchema(t *testing.T) {
	defer reset()
	key := "oc"
	rootStruct := &gostructs.Device{}
	schemaTree := map[string]*yang.Entry{
		"x": &yang.Entry{},
	}

	if err := Set(key, rootStruct, schemaTree, gostructs.Unmarshal); err != nil {
		t.Fatalf("Schema(%v, %v, %v) got error %v", key, rootStruct, schemaTree, err)
	}

	c, err := Get(key)
	if err != nil {
		t.Fatalf("got %v, want no error", err)
	}

	if nr := c.NewRoot(); nr == rootStruct {
		t.Error("NewRoot, got identical ygot.GoStruct, want not the same")
	} else if !reflect.DeepEqual(nr, rootStruct) {
		t.Error("NewRoot, got not equal ygot.GoStruct, want equal")
	}
}

func TestMultipleSchema(t *testing.T) {
	defer reset()
	if err := Set("oc", &gostructs.Device{}, gostructs.SchemaTree, gostructs.Unmarshal); err != nil {
		t.Fatalf("got %v, want nil", err)
	}

	if err := Set("oc", &gostructs.Device{}, gostructs.SchemaTree, gostructs.Unmarshal); err == nil {
		t.Fatalf("got nil, want err")
	}

	if err := Set("oc2", &gostructs.Device{}, gostructs.SchemaTree, gostructs.Unmarshal); err != nil {
		t.Fatalf("got %v, want nil", err)
	}
}

func TestGet(t *testing.T) {
	defer reset()
	key := "arbitrary"

	// test getting schema witha an unregistered key
	wantErr := "no schema found corresponding to"
	_, err := Get(key)
	if diff := errdiff.Substring(err, wantErr); diff != "" {
		t.Fatalf("got %v, want %v", err, wantErr)
	}

	// register a schema
	if err := Set(key, &gostructs.Device{}, gostructs.SchemaTree, gostructs.Unmarshal); err != nil {
		t.Fatalf("got %v, want nil", err)
	}

	// "arbitrary" key is now registered, Get must succeed
	if _, err := Get(key); err != nil {
		t.Fatalf("got %v, want no error", err)
	}
}
