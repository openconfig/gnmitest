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

// Package common defines operations that are used within the gNMITest
// framework for multiple tests.
package common

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"context"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc"

	tpb "github.com/openconfig/gnmi/client"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

// ConnectionArgs stores the connection parameters used for a gNMI
// target.
type ConnectionArgs struct {
	Target  string           // The target name for device under test.
	Address string           // Address is the connection string for the target in the form host:port.
	Timeout int32            // Dial timeout while connecting to target.
	Creds   *tpb.Credentials // Creds is the authentication credentials that should be used to connect to the target.
}

// Connect opens a new gRPC connection to the target speciifed by the
// ConnectionArgs. It returns the gNMI Client connection, and a function
// which can be called to close the connection. If an error is encountered
// during opening the connection, it is returned.
func Connect(ctx context.Context, a *ConnectionArgs) (gpb.GNMIClient, func(), error) {
	if a.Address == "" {
		return nil, nil, errors.New("an address must be specified")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(a.Timeout)*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, a.Address, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot dial target %s, %v", a.Address, err)
	}

	return gpb.NewGNMIClient(conn), func() { conn.Close() }, nil
}
