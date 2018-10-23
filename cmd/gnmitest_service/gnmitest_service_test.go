package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/openconfig/gnmi/client/gnmi"
	"github.com/openconfig/gnmi/errdiff"
	"github.com/openconfig/gnmitest/common"
	"github.com/openconfig/gnmitest/schemafake"
	"github.com/openconfig/gnmitest/schemas/openconfig"
	"github.com/openconfig/gnmitest/service"
	"github.com/openconfig/ygot/testutil"
	"github.com/openconfig/ygot/ytypes"
	"google.golang.org/grpc"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	gtpb "github.com/openconfig/gnmitest/proto/gnmitest"
	rpb "github.com/openconfig/gnmitest/proto/report"
	spb "github.com/openconfig/gnmitest/proto/suite"
)

var (
	// certFile specifies the path to the certificate that should be
	// used by the test service and fake within the tests.
	certFile = filepath.Join("testdata", "cert.crt")
	// keyFile specifies the path to the certificate key that should
	// be used by the test service and fake in the tests.
	keyFile = filepath.Join("testdata", "key.key")
)

const (
	// The set of magic words that should be replaced in a suite
	// proto. This allows the test to override parameters based on
	// runtime setup.
	portMagicWord = "%%PORT%%"
	hostMagicWord = "%%HOST%%"
)

func TestIntegration(t *testing.T) {

	ocSchema, err := gostructs.Schema()
	if err != nil {
		t.Fatalf("cannot extract schema from gostructs, %v", err)
	}

	tests := []struct {
		name             string
		inSuiteFile      string
		inSchema         map[string]*ytypes.Schema
		inFakeDataFiles  map[string]string
		wantReportFile   string
		wantErrSubstring string
	}{{
		name:           "unimplemented Subscribe test",
		inSuiteFile:    filepath.Join("testdata", "unimplemented-suite.txtpb"),
		inSchema:       map[string]*ytypes.Schema{"openconfig": ocSchema},
		wantReportFile: filepath.Join("testdata", "unimplemented-report.txtpb"),
	}, {
		name:        "simple get test",
		inSuiteFile: filepath.Join("testdata", "simple-get-suite.txtpb"),
		inSchema:    map[string]*ytypes.Schema{"openconfig": ocSchema},
		inFakeDataFiles: map[string]string{
			"openconfig": filepath.Join("testdata", "simple-get.json"),
		},
		wantReportFile: filepath.Join("testdata", "simple-get-report.txtpb"),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			schemafake.Timestamp = func() int64 { return 42 }
			target, err := schemafake.New(tt.inSchema)
			if err != nil {
				t.Fatalf("cannot create fake, %v", err)
			}

			for origin, dataF := range tt.inFakeDataFiles {
				fd, err := ioutil.ReadFile(dataF)
				if err != nil {
					t.Fatalf("cannot read fakedata for origin %s: %v", origin, err)
				}

				if err := target.Load(fd, origin); err != nil {
					t.Fatalf("cannot load data into fake origin %s: %v", origin, err)
				}
			}

			// Start the fake.
			port, stop, err := target.Start(certFile, keyFile)
			if err != nil {
				t.Fatalf("cannot start fake, %v", err)
			}
			defer stop()

			sbyte, err := ioutil.ReadFile(tt.inSuiteFile)
			if err != nil {
				t.Fatalf("cannot read suite file, %v", err)
			}

			rp := strings.NewReplacer(
				portMagicWord, fmt.Sprintf("%d", port),
				hostMagicWord, "localhost",
			)

			ss := rp.Replace(string(sbyte))

			in := &spb.Suite{}
			if err := proto.UnmarshalText(ss, in); err != nil {
				t.Fatalf("cannot unmarshal suite proto, %v", err)
			}

			rbyte, err := ioutil.ReadFile(tt.wantReportFile)
			if err != nil {
				t.Fatalf("cannot read report file, %v", err)
			}

			rs := rp.Replace(string(rbyte))

			want := &rpb.Report{}
			if err := proto.UnmarshalText(rs, want); err != nil {
				t.Fatalf("cannot unmarshal report proto, %v", err)
			}

			// Start the test service
			testSrv, testLis, testPort := createTestServer(t)
			go testSrv.Serve(testLis)

			conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", testPort), grpc.WithInsecure())
			if err != nil {
				t.Fatalf("cannot connect client to local gnmitest service, %v", err)
			}

			cl := gtpb.NewGNMITestClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			got, err := cl.Run(ctx, in)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("did not get expected error from gnmitest server, %s", diff)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(got, want, cmp.FilterPath(func(p cmp.Path) bool {
				if p.Last().Type() == reflect.TypeOf(&gpb.GetResponse{}) {
					return true
				}
				return false
			}, cmp.Comparer(testutil.GetResponseEqual))); diff != "" {
				t.Fatalf("did not get expected report proto, diff(-got,+want):\n%s", diff)
			}
		})
	}

}

func createTestServer(t *testing.T) (*grpc.Server, net.Listener, uint64) {
	srv := grpc.NewServer()
	testSrv, err := service.NewServer(client.Type)
	if err != nil {
		t.Fatalf("cannot create a gnmitest server instance, %v", err)
	}

	gtpb.RegisterGNMITestServer(srv, testSrv)

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("cannot listen, %v", err)
	}

	port, err := common.ListenerTCPPort(lis)
	if err != nil {
		t.Fatalf("cannot determine TCP port, %v", err)
	}

	return srv, lis, port
}
