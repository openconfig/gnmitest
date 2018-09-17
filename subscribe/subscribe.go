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

// Package subscribe contains test interface definitions for gnmi Subscribe
// RPC.
package subscribe

import (
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

// Status is a type that is used by test to notify test framework.
type Status int

const (
	// Running indicates that test can accept more proto.Message.
	Running Status = iota
	// Complete indicates that test is finished, so it can be unregistered.
	Complete
)

// Subscribe is the interface of a test for gnmi Subscribe RPC.
type Subscribe interface {
	// Process is called for each individual message received. Status returned by
	// Process may have Running or Complete status. When Complete is returned,
	// test framework calls Check function of the test to get the holistic test
	// result and to unregister test.
	Process(sr *gpb.SubscribeResponse) (Status, error)
	// Check is called to get the holistic test result in the following cases;
	// - Process function returns Complete
	// - Test times out
	// - GNMI RPC request fails
	Check() error
}

// Test must be embedded by each gnmi Subscribe RPC test.
type Test struct {
	sr *gpb.SubscribeRequest
}

// Check is the default implementation for a test that only evaluates
// individual messages via Process and does not return a stateful result across
// multiple messages.
func (s *Test) Check() error {
	return nil
}

// GetRequest returns the gnmi SubscribeRequest stored in the receiver object.
func (s *Test) GetRequest() *gpb.SubscribeRequest {
	return s.sr
}

// SetRequest stores the gnmi SubscribeRequest in the struct.
func (s *Test) SetRequest(sr *gpb.SubscribeRequest) {
	s.sr = sr
}
