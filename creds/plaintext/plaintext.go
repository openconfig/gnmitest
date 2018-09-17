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

// Package plaintext registers plaintext resolver with the "plaintext" key into
// global resolvers table.
package plaintext

import (
	"context"
	"github.com/openconfig/gnmitest/creds"

	tpb "github.com/openconfig/gnmitest/proto/tests"
)

const (
	// empty is the resolver that matches when resolver is left empty.
	empty = ""
	// Key is the unique key to register plaintext resolver.
	Key = "plaintext"
)

func init() {
	resolver.Set(empty, &plainTextResolver{})
	resolver.Set(Key, &plainTextResolver{})
}

type plainTextResolver struct {
}

// Credentials returns the username and password in plaintext. They are
// specified in in tpb.Credentials message. Plaintext resolver doesn't perform
//  any resolution on the provided username and password.
func (*plainTextResolver) Credentials(_ context.Context, creds *tpb.Credentials) (*resolver.Credentials, error) {
	if creds == nil {
		return nil, nil
	}
	return &resolver.Credentials{
		Username: creds.GetUsername(),
		Password: creds.GetPassword(),
	}, nil
}
