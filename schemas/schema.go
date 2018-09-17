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

// Package schema exports functions to register given schema information
// into lookup table as well as to retrieve from lookup table.
package schema

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
)

// unmarshalFunc defines a type which represents the Unmarshal function in
// ygot generated code.
type unmarshalFunc func([]byte, ygot.GoStruct, ...ytypes.UnmarshalOpt) error

// goStruct stores the factory function to create a new root as well as the
// keyed list of schemas that is shared across the test infrastructure for all
// tests.
type goStruct struct {
	// root stores the GoStruct which is the root of the tree.
	root ygot.GoStruct
	// schemaTree stores a map, keyed by Go type name, to the yang.Entry schema
	// that describes the struct.
	schemaTree map[string]*yang.Entry
	// unmarshalFunc defines the function that can be used to JSON unmarshal
	// into the tree.
	unmarshal unmarshalFunc
}

// TestGoStruct stores a cached copy of the global schema tree that is local to
// a given test so as to prevent the possiblity of cross-test schema contamination.
type TestGoStruct struct {
	goStruct
	localSchemaTree map[string]*yang.Entry
}

// NewRoot creates a new instance of root ygot.GoStruct.
func (s *TestGoStruct) NewRoot() ygot.GoStruct {
	return reflect.New(reflect.TypeOf(s.root).Elem()).Interface().(ygot.GoStruct)
}

// Schema returns the *yang.Entry corresponding to given key.
func (s *TestGoStruct) Schema(k string) *yang.Entry {
	if _, ok := s.localSchemaTree[k]; !ok {
		// TODO(yusufsn): return a copy of the *yang.Entry.
		s.localSchemaTree[k] = s.schemaTree[k]
	}
	return s.localSchemaTree[k]
}

var (
	// mu is used to synchronize registration of multiple GoStruct
	// into goStructs table.
	mu sync.Mutex
	// goStructs is the lookup table that maintains different goStructs
	// to choose from.
	goStructs = make(map[string]goStruct)
)

// Set registers given key and GoStruct into the registration table.
// If key already exists, it doesn't register GoStruct and returns an error.
func Set(key string, root ygot.GoStruct, schemaTree map[string]*yang.Entry, ufn unmarshalFunc) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := goStructs[key]; ok {
		return fmt.Errorf("another schema is registered with the key %v", key)
	}
	goStructs[key] = goStruct{
		root:       root,
		schemaTree: schemaTree,
		unmarshal:  ufn,
	}
	return nil
}

// Get returns reference to testGoStruct so that test can
// have access to underlying schema corresponding to given key.
func Get(key string) (*TestGoStruct, error) {
	mu.Lock()
	defer mu.Unlock()
	gs, ok := goStructs[key]
	if !ok {
		return nil, fmt.Errorf("no schema found corresponding to %q key", key)
	}
	return &TestGoStruct{
		goStruct:        gs,
		localSchemaTree: make(map[string]*yang.Entry),
	}, nil
}
