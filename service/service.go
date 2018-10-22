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

// Package service implements the gnmitest.proto Run service API.
package service

import (
	"bytes"
	"fmt"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmitest/config"
	"github.com/openconfig/gnmitest/runner"

	// Register openconfig schema as the default schema.
	_ "github.com/openconfig/gnmitest/schemas/openconfig/register"

	// Tests below register themselves to test framework. They can be used while
	// composing Suite proto.
	_ "github.com/openconfig/gnmitest/tests/subscribe/haskeys"
	_ "github.com/openconfig/gnmitest/tests/subscribe/pathvalidation"
	_ "github.com/openconfig/gnmitest/tests/subscribe/schemapathc/schemapathc"

	rpb "github.com/openconfig/gnmitest/proto/report"
	spb "github.com/openconfig/gnmitest/proto/suite"
)

// Server struct stores the information that are same across all requests.
type Server struct {
	clientType string
}

// NewServer returns an instance of Server struct.
func NewServer(clientType string) (*Server, error) {
	s := &Server{
		clientType: clientType,
	}

	return s, nil
}

// Run RPC accepts a Suite proto and runs it. It returns the Report proto
// as result.
func (s *Server) Run(ctx context.Context, suite *spb.Suite) (*rpb.Report, error) {
	b := &bytes.Buffer{}
	if err := proto.MarshalText(b, suite); err != nil {
		return nil, fmt.Errorf("marshalling proto failed: %v", err)
	}
	// TODO(yusufsn): Change the signature of New to accept Suite proto.
	cfg, err := config.New(b.Bytes(), s.clientType)
	if err != nil {
		return nil, fmt.Errorf("config.New failed: %v", err)
	}

	report := &rpb.Report{}
	r := runner.New(cfg, func(ir *rpb.InstanceGroup) {
		report.Results = append(report.Results, ir)
	})

	if err := r.Start(ctx); err != nil {
		return nil, fmt.Errorf("error occurred during test execution %v", err)
	}

	return report, nil
}
