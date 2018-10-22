// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package schemafake defines a fake implementation of a gNMI device with
// a known schema which is used to validate tests within the gNMITest
// framework.
package schemafake

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmi/unimplemented"
	"github.com/openconfig/gnmitest/common"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/util"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

var (
	// Timestamp is a function used to specify the timestamp to be used in Notifications
	// returned from the fake. It can be overridden in calling code to ensure that a
	// deterministic result is returned.
	Timestamp = func() int64 { return time.Now().UnixNano() }
)

// Target defines the gNMI fake target.
type Target struct {
	unimplemented.Server                    // Implement the gNMI server interface.
	schema               map[string]*origin // schema is a map, keyed by origin name, of the schemas supported by the fake.
}

// origin stores the internal state for a gNMI target's origins, supporting mixed schema
// operation.
type origin struct {
	mu            sync.RWMutex           // mu is a mutex used to protect access to the origin
	data          ygot.ValidatedGoStruct // data is the datatree stored at the origin.
	unmarshalFunc ytypes.UnmarshalFunc   // unmarshalFunc is a function which can be used to unmarshal into the root.
	rootSchema    *yang.Entry            // rootSchema is the schema of the root node for the YANG schematree.
}

// unmarshal unmarshals the contents of d, which must be valid RFC7951 JSON, to the
// origin with the specified options.
func (o *origin) unmarshal(d []byte, opts ...ytypes.UnmarshalOpt) error {
	if o.unmarshalFunc == nil {
		return errors.New("invalid (nil) unmarshal function")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.unmarshalFunc(d, o.data, opts...)
}

// get returns the value stored at the specified path from the origin.
func (o *origin) get(path *gpb.Path) ([]*ytypes.TreeNode, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return ytypes.GetNode(o.rootSchema, o.data, path, &ytypes.GetPartialKeyMatch{})
}

// New creates a new schemaefake using the specified map of schemas. The schemas map is keyed by the origin
// name, with the value being a Schema specification.
func New(schemas map[string]*ytypes.Schema) (*Target, error) {
	t := &Target{
		schema: map[string]*origin{},
	}

	if len(schemas) == 0 {
		return nil, fmt.Errorf("target must have more than %d schemas specified", len(schemas))
	}

	for sn, s := range schemas {
		sroot, ok := s.SchemaTree[reflect.TypeOf(s.Root).Elem().Name()]
		if !ok {
			return nil, fmt.Errorf("could not find schema for %T in the supplied schema tree for origin %s", sroot, sn)
		}

		t.schema[sn] = &origin{
			data:          reflect.New(reflect.TypeOf(s.Root).Elem()).Interface().(ygot.ValidatedGoStruct),
			rootSchema:    sroot,
			unmarshalFunc: s.Unmarshal,
		}
	}

	return t, nil
}

// Start starts the fake gNMI server with the specified certificate and key. It returns
// the TCP port the fake is listening on, a function to stop the server and an optional
// error if the server cannot be started.
func (t *Target) Start(cert, key string) (uint64, func(), error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, nil, fmt.Errorf("cannot create listener, %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile(cert, key)
	if err != nil {
		return 0, nil, fmt.Errorf("Failed to generate credentials %v", err)
	}

	server := grpc.NewServer(grpc.Creds(creds))

	tcpPort, err := common.ListenerTCPPort(l)
	if err != nil {
		return 0, nil, err
	}

	gpb.RegisterGNMIServer(server, t)
	go server.Serve(l)
	return tcpPort, server.Stop, nil
}

// getOrigin returns the definition from the origin named n from the target's stored
// schemas.
func (t *Target) getOrigin(n string) (*origin, error) {
	if n == "" {
		// According to the specification, "" is equal to "openconfig".
		n = "openconfig"
	}

	originNames := func(o map[string]*origin) []string {
		names := []string{}
		for s := range o {
			names = append(names, s)
		}
		return names
	}

	s, ok := t.schema[n]
	if !ok {
		return nil, fmt.Errorf("could not find origin %s, supported origins %v", n, originNames(t.schema))
	}

	return s, nil
}

// Load unmarshals the JSON supplied in b into the supplied origin,  using the options
// specified into the target's root using the stored unmarshal function.
func (t *Target) Load(b []byte, origin string, opts ...ytypes.UnmarshalOpt) error {
	orig, err := t.getOrigin(origin)
	if err != nil {
		return fmt.Errorf("cannot load data into origin: %v", err)
	}

	if err := orig.unmarshal(b, opts...); err != nil {
		return fmt.Errorf("cannot unmarshal JSON data, %v", err)
	}
	return nil
}

// Get implements the gNMI Get RPC. The request received from the client is extracted from the
// GetRequest received from the client. Each path is retrieved from the target's data tree,
// and subsequently marshalled into a gNMI Notification. Each path in the GetRequest is
// handled separately, such that there is no guarantee of consistency across separate
// paths within the GetRequest. Prefixing is performed within the results of each
// path expansion within the request.
func (t *Target) Get(ctx context.Context, r *gpb.GetRequest) (*gpb.GetResponse, error) {
	var notifications []*gpb.Notification
	for _, p := range r.Path {
		fullPath := p
		if r.Prefix != nil {
			fullPath = &gpb.Path{
				Elem: append(r.Prefix.Elem, p.Elem...),
			}

			if p.Origin == "" && r.Prefix.Origin != "" {
				fullPath.Origin = r.Prefix.Origin
			}
		}

		// Capture the timestamp for the Notification.
		ts := Timestamp()

		orig, err := t.getOrigin(p.Origin)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "cannot find origin %s on target, %v", p.Origin, err)
		}

		nodes, err := orig.get(fullPath)
		if err != nil {
			return nil, err
		}

		var prefix *gpb.Path
		if p.Origin != "" || p.Target != "" {
			prefix = &gpb.Path{
				Origin: p.Origin,
				Target: p.Target,
			}
		}

		if len(nodes) > 1 {
			var paths []*gpb.Path
			for _, n := range nodes {
				paths = append(paths, n.Path)
			}
			prefix = util.FindPathElemPrefix(paths)
		}

		var u []*gpb.Update
		for _, n := range nodes {
			v, err := ygot.EncodeTypedValue(n.Data, r.Encoding)
			if err != nil {
				return nil, status.Errorf(codes.Unavailable, "could not encode value at %s, %v", proto.MarshalTextString(n.Path), err)
			}

			u = append(u, &gpb.Update{
				Path: util.TrimGNMIPathElemPrefix(n.Path, prefix),
				Val:  v,
			})
		}

		notifications = append(notifications, &gpb.Notification{
			Timestamp: ts,
			Prefix:    prefix,
			Update:    u,
		})

	}

	return &gpb.GetResponse{Notification: notifications}, nil
}
