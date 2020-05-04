package datatreepaths

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/gnmi/errdiff"
	"github.com/openconfig/gnmitest/schemas"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/exampleoc"
	"github.com/openconfig/ygot/testutil"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	tpb "github.com/openconfig/gnmitest/proto/tests"
)

func mustPath(s string) *gpb.Path {
	p, err := ygot.StringToStructuredPath(s)
	if err != nil {
		panic(err)
	}
	return p
}

func noti(p ...string) *gpb.Notification {
	n := &gpb.Notification{}
	for _, s := range p {
		n.Update = append(n.Update, &gpb.Update{Path: mustPath(s)})
	}
	return n
}

type pathVal struct {
	p *gpb.Path
	v *gpb.TypedValue
}

func notiVal(ts int64, pfx *gpb.Path, upd ...pathVal) *gpb.Notification {
	n := &gpb.Notification{
		Prefix:    pfx,
		Timestamp: ts,
	}

	for _, u := range upd {
		n.Update = append(n.Update, &gpb.Update{
			Path: u.p,
			Val:  u.v,
		})
	}

	return n
}

func TestCheck(t *testing.T) {
	s, err := exampleoc.Schema()
	if err != nil {
		t.Fatalf("cannot get schema, %v", err)
	}

	if err := schema.Set("", s.Root, exampleoc.UnzipSchema, exampleoc.Unmarshal); err != nil {
		t.Fatalf("cannot register new schema, %v", err)
	}

	tests := []struct {
		name                    string
		inSpec                  *tpb.Test
		inSubscribeResponses    []*gpb.SubscribeResponse
		wantProcessErrSubstring string
		wantCheckErrSubstring   string
	}{{
		name: "simple data tree path",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
									Key:  map[string]string{"name": "eth0"},
								}},
								Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
									&tpb.DataTreePaths_RequiredPaths{
										Prefix: mustPath("state/counters"),
										Paths: []*gpb.Path{
											mustPath("in-pkts"),
											mustPath("out-pkts"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				noti(
					"/interfaces/interface[name=eth0]/state/counters/in-pkts",
					"/interfaces/interface[name=eth0]/state/counters/out-pkts",
				),
			},
		}},
	}, {
		name: "ignore erroneous path",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
									Key:  map[string]string{"name": "eth0"},
								}},
								Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
									&tpb.DataTreePaths_RequiredPaths{
										Prefix: mustPath("state/counters"),
										Paths: []*gpb.Path{
											mustPath("in-pkts"),
											mustPath("out-pkts"),
										},
									},
								},
							},
						},
					},
					IgnoreInvalidPaths: true,
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				noti(
					"/interfaces/interface[name=eth0]/state/counters/in-pkts",
					"/interfaces/interface[name=eth0]/state/counters/out-pkts",
					"/interfaces/interface[name=eth0]/qos/state/invalid-path",
				),
			},
		}},
	}, {
		name: "error due to erroneous path",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
									Key:  map[string]string{"name": "eth0"},
								}},
								Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
									&tpb.DataTreePaths_RequiredPaths{
										Prefix: mustPath("state/counters"),
										Paths: []*gpb.Path{
											mustPath("in-pkts"),
											mustPath("out-pkts"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				noti(
					"/interfaces/interface[name=eth0]/state/counters/in-pkts",
					"/interfaces/interface[name=eth0]/state/counters/out-pkts",
					"/interfaces/interface[name=eth0]/qos/state/invalid-path",
				),
			},
		}},
		wantProcessErrSubstring: "no match found",
	}, {
		name: "simple data tree value check",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
									Key:  map[string]string{"name": "eth0"},
								}},
								Type: &tpb.DataTreePaths_TestQuery_RequiredValues{
									&tpb.DataTreePaths_RequiredValues{
										Prefix: mustPath("state/counters"),
										Matches: []*tpb.PathValueMatch{{
											Path: mustPath("in-pkts"),
											Criteria: &tpb.PathValueMatch_Equal{
												&gpb.TypedValue{
													Value: &gpb.TypedValue_UintVal{42},
												},
											},
										}, {
											Path: mustPath("out-pkts"),
											Criteria: &tpb.PathValueMatch_Equal{
												&gpb.TypedValue{
													Value: &gpb.TypedValue_UintVal{84},
												},
											},
										}},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				notiVal(42, mustPath("/interfaces/interface[name=eth0]/state/counters"),
					pathVal{
						p: mustPath("in-pkts"),
						v: &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{42}},
					},
					pathVal{
						p: mustPath("out-pkts"),
						v: &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{84}},
					},
				),
			},
		}},
	}, {
		name: "unequal data value check",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
									Key:  map[string]string{"name": "eth0"},
								}},
								Type: &tpb.DataTreePaths_TestQuery_RequiredValues{
									&tpb.DataTreePaths_RequiredValues{
										Prefix: mustPath("state/counters"),
										Matches: []*tpb.PathValueMatch{{
											Path: mustPath("in-pkts"),
											Criteria: &tpb.PathValueMatch_Equal{
												&gpb.TypedValue{
													Value: &gpb.TypedValue_UintVal{42},
												},
											},
										}, {
											Path: mustPath("out-pkts"),
											Criteria: &tpb.PathValueMatch_Equal{
												&gpb.TypedValue{
													Value: &gpb.TypedValue_UintVal{84},
												},
											},
										}},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				notiVal(42, mustPath("/interfaces/interface[name=eth0]/state/counters"),
					pathVal{
						p: mustPath("in-pkts"),
						v: &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{96}},
					},
					pathVal{
						p: mustPath("out-pkts"),
						v: &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{84}},
					},
				),
			},
		}},
		wantCheckErrSubstring: "did not match expected value",
	}, {
		name: "simple data tree path with enum",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
									Key:  map[string]string{"name": "eth0"},
								}},
								Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
									&tpb.DataTreePaths_RequiredPaths{
										Prefix: mustPath("state"),
										Paths: []*gpb.Path{
											mustPath("admin-status"),
											mustPath("oper-status"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				notiVal(42, mustPath("/interfaces/interface[name=eth0]/state"),
					pathVal{mustPath("admin-status"), &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"UP"}}},
					pathVal{mustPath("oper-status"), &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"DORMANT"}}},
				),
			},
		}},
	}, {
		name: "unset enum",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
									Key:  map[string]string{"name": "eth0"},
								}},
								Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
									&tpb.DataTreePaths_RequiredPaths{
										Prefix: mustPath("state"),
										Paths: []*gpb.Path{
											mustPath("admin-status"),
											mustPath("oper-status"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				notiVal(42, mustPath("/interfaces/interface[name=eth0]/state"),
					pathVal{mustPath("admin-status"), &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"UP"}}},
				),
			},
		}},
		wantCheckErrSubstring: "enum type exampleoc.E_OpenconfigInterfaces_Interface_OperStatus was UNSET",
	}, {
		name: "iterative test",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
								}},
								Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
									&tpb.DataTreePaths_ListQuery{
										VarName: "%%interface%%",
										NextQuery: &tpb.DataTreePaths_TestQuery{
											Steps: []*tpb.DataTreePaths_QueryStep{{
												Name: "interfaces",
											}, {
												Name:    "interface",
												KeyName: "%%interface%%",
											}},

											Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
												&tpb.DataTreePaths_RequiredPaths{
													Prefix: mustPath("state/counters"),
													Paths: []*gpb.Path{
														mustPath("in-pkts"),
														mustPath("out-pkts"),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				noti(
					"/interfaces/interface[name=eth0]/state/counters/in-pkts",
					"/interfaces/interface[name=eth0]/state/counters/out-pkts",
					"/interfaces/interface[name=eth1]/state/counters/in-pkts",
					"/interfaces/interface[name=eth1]/state/counters/out-pkts",
				),
			},
		}},
	}, {
		name: "iterative test - failed",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						DataTreePaths: &tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "interfaces",
								}, {
									Name: "interface",
								}},
								Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
									&tpb.DataTreePaths_ListQuery{
										VarName: "%%interface%%",
										NextQuery: &tpb.DataTreePaths_TestQuery{
											Steps: []*tpb.DataTreePaths_QueryStep{{
												Name: "interfaces",
											}, {
												Name:    "interface",
												KeyName: "%%interface%%",
											}},

											Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
												&tpb.DataTreePaths_RequiredPaths{
													Prefix: mustPath("state/counters"),
													Paths: []*gpb.Path{
														mustPath("in-pkts"),
														mustPath("out-pkts"),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				noti(
					"/interfaces/interface[name=eth0]/state/counters/in-pkts",
					"/interfaces/interface[name=eth1]/state/counters/in-pkts",
					"/interfaces/interface[name=eth1]/state/counters/out-pkts",
				),
			},
		}},
		wantCheckErrSubstring: "got nil data for path",
	}, {
		name: "filtered test",
		inSpec: &tpb.Test{
			Type: &tpb.Test_Subscribe{
				&tpb.SubscribeTest{
					Args: &tpb.SubscribeTest_DataTreePaths{
						&tpb.DataTreePaths{
							TestOper: &tpb.DataTreePaths_TestQuery{
								Steps: []*tpb.DataTreePaths_QueryStep{{
									Name: "components",
								}, {
									Name: "component",
								}},
								Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
									&tpb.DataTreePaths_ListQuery{
										VarName: "%%component_name%%",
										Filter: &tpb.PathValueMatch{
											Path: mustPath("state/type"),
											Criteria: &tpb.PathValueMatch_Equal{
												&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"TRANSCEIVER"}},
											},
										},
										NextQuery: &tpb.DataTreePaths_TestQuery{
											Steps: []*tpb.DataTreePaths_QueryStep{{
												Name: "components",
											}, {
												Name:    "component",
												KeyName: "%%component_name%%",
											}},
											Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
												&tpb.DataTreePaths_RequiredPaths{
													Prefix: mustPath("state"),
													Paths: []*gpb.Path{
														mustPath("mfg-name"),
														mustPath("serial-no"),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		inSubscribeResponses: []*gpb.SubscribeResponse{{
			Response: &gpb.SubscribeResponse_Update{
				notiVal(42, nil,
					pathVal{
						p: mustPath("/components/component[name=xcvr1]/state/type"),
						v: &gpb.TypedValue{
							Value: &gpb.TypedValue_StringVal{"TRANSCEIVER"},
						},
					},
					pathVal{
						p: mustPath("/components/component[name=port0]/state/type"),
						v: &gpb.TypedValue{
							Value: &gpb.TypedValue_StringVal{"PORT"},
						},
					},
					pathVal{
						p: mustPath("/components/component[name=xcvr1]/state/mfg-name"),
						v: &gpb.TypedValue{
							Value: &gpb.TypedValue_StringVal{"MANUFACTURER"},
						},
					},
					pathVal{
						p: mustPath("/components/component[name=xcvr1]/state/serial-no"),
						v: &gpb.TypedValue{
							Value: &gpb.TypedValue_StringVal{"SERIAL"},
						},
					}),
			},
		}},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := newTest(tt.inSpec)
			if err != nil {
				t.Fatalf("cannot initialise test, %v", err)
			}

			for _, sr := range tt.inSubscribeResponses {
				_, err := ts.Process(sr)
				if diff := errdiff.Substring(err, tt.wantProcessErrSubstring); diff != "" {
					t.Fatalf("cannot process SubscribeResponse %s, %s", sr, diff)
				}
			}

			err = ts.Check()
			if diff := errdiff.Substring(err, tt.wantCheckErrSubstring); diff != "" {
				t.Fatalf("did not get expected error, %v", err)
			}
		})
	}
}

func TestQueries(t *testing.T) {
	tests := []struct {
		name             string
		inTestSpec       *tpb.DataTreePaths
		inSchema         *yang.Entry
		inDevice         *exampleoc.Device
		want             *resolvedOperation
		wantErrSubstring string
	}{{
		name: "foreach subinterface of each interface",
		inTestSpec: &tpb.DataTreePaths{
			TestOper: &tpb.DataTreePaths_TestQuery{
				Steps: []*tpb.DataTreePaths_QueryStep{{
					Name: "interfaces",
				}, {
					Name: "interface",
				}},
				Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
					&tpb.DataTreePaths_ListQuery{
						VarName: "%%interface%%",
						NextQuery: &tpb.DataTreePaths_TestQuery{
							Steps: []*tpb.DataTreePaths_QueryStep{{
								Name: "interfaces",
							}, {
								Name:    "interface",
								KeyName: "%%interface%%",
							}, {
								Name: "subinterfaces",
							}, {
								Name: "subinterface",
							}},
							Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
								&tpb.DataTreePaths_ListQuery{
									VarName: "%%subinterface%%",
									NextQuery: &tpb.DataTreePaths_TestQuery{
										Steps: []*tpb.DataTreePaths_QueryStep{{
											Name: "interfaces",
										}, {
											Name:    "interface",
											KeyName: "%%interface%%",
										}, {
											Name: "subinterfaces",
										}, {
											Name:    "subinterface",
											KeyName: "%%subinterface%%",
										}},
										Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
											&tpb.DataTreePaths_RequiredPaths{
												Paths: []*gpb.Path{
													mustPath("state/index"),
													mustPath("state/description"),
													mustPath("ipv4/addresses/address[ip=192.168.1.2]/state/ip"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		inDevice: func() *exampleoc.Device {
			oc := &exampleoc.Device{}
			oc.GetOrCreateInterface("eth0").GetOrCreateSubinterface(1)
			oc.GetOrCreateInterface("eth0").GetOrCreateSubinterface(2)
			oc.GetOrCreateInterface("eth1").GetOrCreateSubinterface(10)
			oc.GetOrCreateInterface("eth1").GetOrCreateSubinterface(20)
			return oc
		}(),
		want: &resolvedOperation{
			paths: []*gpb.Path{
				mustPath("/interfaces/interface[name=eth0]/subinterfaces/subinterface[index=1]/state/index"),
				mustPath("/interfaces/interface[name=eth0]/subinterfaces/subinterface[index=1]/state/description"),
				mustPath("/interfaces/interface[name=eth0]/subinterfaces/subinterface[index=1]/ipv4/addresses/address[ip=192.168.1.2]/state/ip"),
				mustPath("/interfaces/interface[name=eth0]/subinterfaces/subinterface[index=2]/state/index"),
				mustPath("/interfaces/interface[name=eth0]/subinterfaces/subinterface[index=2]/state/description"),
				mustPath("/interfaces/interface[name=eth0]/subinterfaces/subinterface[index=2]/ipv4/addresses/address[ip=192.168.1.2]/state/ip"),
				mustPath("/interfaces/interface[name=eth1]/subinterfaces/subinterface[index=10]/state/index"),
				mustPath("/interfaces/interface[name=eth1]/subinterfaces/subinterface[index=10]/state/description"),
				mustPath("/interfaces/interface[name=eth1]/subinterfaces/subinterface[index=10]/ipv4/addresses/address[ip=192.168.1.2]/state/ip"),
				mustPath("/interfaces/interface[name=eth1]/subinterfaces/subinterface[index=20]/state/index"),
				mustPath("/interfaces/interface[name=eth1]/subinterfaces/subinterface[index=20]/state/description"),
				mustPath("/interfaces/interface[name=eth1]/subinterfaces/subinterface[index=20]/ipv4/addresses/address[ip=192.168.1.2]/state/ip"),
			},
		},
	}, {
		name: "simple data tree paths, no queries",
		inTestSpec: &tpb.DataTreePaths{
			TestOper: &tpb.DataTreePaths_TestQuery{
				Steps: []*tpb.DataTreePaths_QueryStep{{
					Name: "interfaces",
				}, {
					Name: "interface",
					Key:  map[string]string{"name": "eth0"},
				}},
				Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
					&tpb.DataTreePaths_RequiredPaths{
						Prefix: mustPath("state/counters"),
						Paths: []*gpb.Path{
							mustPath("in-pkts"),
							mustPath("out-pkts"),
						},
					},
				},
			},
		},
		inDevice: &exampleoc.Device{},
		want: &resolvedOperation{
			paths: []*gpb.Path{
				mustPath("/interfaces/interface[name=eth0]/state/counters/in-pkts"),
				mustPath("/interfaces/interface[name=eth0]/state/counters/out-pkts"),
			},
		},
	}, {
		name: "filtered resolution of a list",
		inTestSpec: &tpb.DataTreePaths{
			TestOper: &tpb.DataTreePaths_TestQuery{
				Steps: []*tpb.DataTreePaths_QueryStep{{
					Name: "interfaces",
				}, {
					Name: "interface",
				}},
				Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
					&tpb.DataTreePaths_ListQuery{
						VarName: "%%interface%%",
						Filter: &tpb.PathValueMatch{
							Path: mustPath("config/name"),
							Criteria: &tpb.PathValueMatch_Equal{
								&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"eth0"}},
							},
						},
						NextQuery: &tpb.DataTreePaths_TestQuery{
							Steps: []*tpb.DataTreePaths_QueryStep{{
								Name: "interfaces",
							}, {
								Name:    "interface",
								KeyName: "%%interface%%",
							}},
							Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
								&tpb.DataTreePaths_RequiredPaths{
									Paths: []*gpb.Path{
										mustPath("ethernet/state/duplex-mode"),
										mustPath("ethernet/state/port-speed"),
									},
								},
							},
						},
					},
				},
			},
		},
		inDevice: func() *exampleoc.Device {
			d := &exampleoc.Device{}
			d.GetOrCreateInterface("eth0").GetOrCreateSubinterface(1)
			d.GetOrCreateInterface("eth1").GetOrCreateSubinterface(1)
			d.GetOrCreateInterface("eth2").GetOrCreateSubinterface(1)
			return d
		}(),
		want: &resolvedOperation{
			paths: []*gpb.Path{
				mustPath("/interfaces/interface[name=eth0]/ethernet/state/duplex-mode"),
				mustPath("/interfaces/interface[name=eth0]/ethernet/state/port-speed"),
			},
		},
	}, {
		name: "filtered resolution on a non-key field",
		inTestSpec: &tpb.DataTreePaths{
			TestOper: &tpb.DataTreePaths_TestQuery{
				Steps: []*tpb.DataTreePaths_QueryStep{{
					Name: "components",
				}, {
					Name: "component",
				}},
				Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
					&tpb.DataTreePaths_ListQuery{
						VarName: "%%component_name%%",
						Filter: &tpb.PathValueMatch{
							Path: mustPath("state/type"),
							Criteria: &tpb.PathValueMatch_Equal{
								&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"TRANSCEIVER"}},
							},
						},
						NextQuery: &tpb.DataTreePaths_TestQuery{
							Steps: []*tpb.DataTreePaths_QueryStep{{
								Name: "components",
							}, {
								Name:    "component",
								KeyName: "%%component_name%%",
							}},
							Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
								&tpb.DataTreePaths_RequiredPaths{
									Prefix: mustPath("state"),
									Paths: []*gpb.Path{
										mustPath("mfg-name"),
										mustPath("serial-no"),
									},
								},
							},
						},
					},
				},
			},
		},
		inDevice: func() *exampleoc.Device {
			d := &exampleoc.Device{}
			x := d.GetOrCreateComponent("xcvr1")
			x.Type = &exampleoc.Component_Type_Union_E_OpenconfigPlatformTypes_OPENCONFIG_HARDWARE_COMPONENT{
				exampleoc.OpenconfigPlatformTypes_OPENCONFIG_HARDWARE_COMPONENT_TRANSCEIVER,
			}
			e := d.GetOrCreateComponent("cpu0")
			e.Type = &exampleoc.Component_Type_Union_E_OpenconfigPlatformTypes_OPENCONFIG_HARDWARE_COMPONENT{
				exampleoc.OpenconfigPlatformTypes_OPENCONFIG_HARDWARE_COMPONENT_CPU,
			}
			return d
		}(),
		want: &resolvedOperation{
			paths: []*gpb.Path{
				mustPath("/components/component[name=xcvr1]/state/mfg-name"),
				mustPath("/components/component[name=xcvr1]/state/serial-no"),
			},
		},
	}, {
		name: "list query for non-list",
		inTestSpec: &tpb.DataTreePaths{
			TestOper: &tpb.DataTreePaths_TestQuery{
				Steps: []*tpb.DataTreePaths_QueryStep{{
					Name: "interfaces",
				}, {
					Name: "interface",
				}, {
					Name: "ethernet",
				}},
				Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
					&tpb.DataTreePaths_ListQuery{
						VarName: "%%foo%%",
						NextQuery: &tpb.DataTreePaths_TestQuery{
							Steps: []*tpb.DataTreePaths_QueryStep{{
								Name: "interfaces",
							}, {
								Name:    "interface",
								KeyName: "%%foo%%",
							}},
							Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
								&tpb.DataTreePaths_RequiredPaths{
									Prefix: mustPath("state/counters"),
									Paths: []*gpb.Path{
										mustPath("in-pkts"),
									},
								},
							},
						},
					},
				},
			},
		},
		inDevice: func() *exampleoc.Device {
			d := &exampleoc.Device{}
			d.GetOrCreateInterface("eth0")
			return d
		}(),
		wantErrSubstring: "was not a list",
	}, {
		name: "list query for empty list",
		inTestSpec: &tpb.DataTreePaths{
			TestOper: &tpb.DataTreePaths_TestQuery{
				Steps: []*tpb.DataTreePaths_QueryStep{{
					Name: "interfaces",
				}, {
					Name: "interface",
				}},
				Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
					&tpb.DataTreePaths_ListQuery{
						VarName: "%%foo%%",
						NextQuery: &tpb.DataTreePaths_TestQuery{
							Steps: []*tpb.DataTreePaths_QueryStep{{
								Name: "interfaces",
							}, {
								Name:    "interface",
								KeyName: "%%foo%%",
							}},
							Type: &tpb.DataTreePaths_TestQuery_RequiredPaths{
								&tpb.DataTreePaths_RequiredPaths{
									Prefix: mustPath("state/counters"),
									Paths: []*gpb.Path{
										mustPath("in-pkts"),
									},
								},
							},
						},
					},
				},
			},
		},
		wantErrSubstring: "code = NotFound",
	}, {
		name: "nil next_query in list query",
		inTestSpec: &tpb.DataTreePaths{
			TestOper: &tpb.DataTreePaths_TestQuery{
				Steps: []*tpb.DataTreePaths_QueryStep{{
					Name: "interfaces",
				}, {
					Name: "interface",
				}},
				Type: &tpb.DataTreePaths_TestQuery_GetListKeys{
					&tpb.DataTreePaths_ListQuery{
						VarName: "%%foo%%",
					},
				},
			},
		},
		wantErrSubstring: "specified nil next_query",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tinst := &test{
				dataTree: tt.inDevice,
				schema:   exampleoc.SchemaTree[reflect.TypeOf(tt.inDevice).Elem().Name()],
				testSpec: tt.inTestSpec,
			}

			got, err := tinst.queries()
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("did not get expected error, %s", diff)
			}

			if err != nil {
				return
			}

			neq := func(a, b *resolvedOperation) bool {
				return cmp.Equal(a.paths, b.paths, cmpopts.SortSlices(testutil.PathLess), cmpopts.EquateEmpty(), cmp.Comparer(proto.Equal)) &&
					cmp.Equal(a.vals, b.vals, cmp.Comparer(proto.Equal))
			}

			if !neq(got, tt.want) {
				diff := pretty.Compare(got, tt.want)
				t.Fatalf("did not get expected result, diff(-got,+want):\n%s", diff)
			}
		})
	}
}

func TestMakeQuery(t *testing.T) {
	tests := []struct {
		name             string
		inQuery          []*tpb.DataTreePaths_QueryStep
		inKnownVars      keyQuery
		want             []*gpb.Path
		wantErrSubstring string
	}{{
		name: "query with no expansions",
		inQuery: []*tpb.DataTreePaths_QueryStep{{
			Name: "one",
		}, {
			Name: "two",
			Key:  map[string]string{"value": "forty-two"},
		}},
		want: []*gpb.Path{
			mustPath("/one/two[value=forty-two]"),
		},
	}, {
		name: "query with expansions",
		inQuery: []*tpb.DataTreePaths_QueryStep{{
			Name: "one",
		}, {
			Name:    "two",
			KeyName: "%%vars%%",
		}},
		inKnownVars: keyQuery{
			"%%vars%%": []map[string]string{
				{"val": "one"},
				{"val": "two"},
			},
		},
		want: []*gpb.Path{
			mustPath("/one/two[val=one]"),
			mustPath("/one/two[val=two]"),
		},
	}, {
		name: "query with multiple expansions",
		inQuery: []*tpb.DataTreePaths_QueryStep{{
			Name:    "one",
			KeyName: "%%keyone%%",
		}, {
			Name:    "two",
			KeyName: "%%keytwo%%",
		}},
		inKnownVars: keyQuery{
			"%%keyone%%": []map[string]string{
				{"v1": "one"},
				{"v1": "two"},
			},
			"%%keytwo%%": []map[string]string{
				{"v2": "one"},
				{"v2": "two"},
			},
		},
		want: []*gpb.Path{
			mustPath("/one[v1=one]/two[v2=one]"),
			mustPath("/one[v1=one]/two[v2=two]"),
			mustPath("/one[v1=two]/two[v2=one]"),
			mustPath("/one[v1=two]/two[v2=two]"),
		},
	}, {
		name: "query with unresolvable step",
		inQuery: []*tpb.DataTreePaths_QueryStep{{
			Name:    "one",
			KeyName: "%%val%%",
		}},
		wantErrSubstring: "cannot resolve step",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeQuery(tt.inQuery, tt.inKnownVars)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("got unexpected error, %s", diff)
			}

			if err != nil {
				return
			}

			neq := func(a, b []*gpb.Path) bool {
				return cmp.Equal(a, b, cmpopts.SortSlices(testutil.PathLess), cmpopts.EquateEmpty(), cmp.Comparer(proto.Equal))
			}

			if !neq(got, tt.want) {
				diff := pretty.Compare(got, tt.want)
				t.Fatalf("did not get expected value, diff(-got,+want):\n%s", diff)
			}
		})
	}
}

func TestMakeStep(t *testing.T) {
	tests := []struct {
		name             string
		inQueryStep      *tpb.DataTreePaths_QueryStep
		inKnownVars      keyQuery
		inKnownPaths     []*gpb.Path
		want             []*gpb.Path
		wantErrSubstring string
	}{{
		name: "no paths with simple element append",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name: "last-element",
		},
		want: []*gpb.Path{mustPath("last-element")},
	}, {
		name: "specified path with simple element append",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name: "last-element",
		},
		inKnownPaths: []*gpb.Path{mustPath("path-one")},
		want:         []*gpb.Path{mustPath("/path-one/last-element")},
	}, {
		name: "multiple specified paths with simple element append",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name: "last-element",
		},
		inKnownPaths: []*gpb.Path{
			mustPath("/path-one"),
			mustPath("/path-two"),
		},
		want: []*gpb.Path{
			mustPath("/path-one/last-element"),
			mustPath("/path-two/last-element"),
		},
	}, {
		name: "specified path with element expansion",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name:    "list-element",
			KeyName: "%%keys%%",
		},
		inKnownVars: keyQuery{
			"%%keys%%": []map[string]string{
				{"name": "val1"},
			},
		},
		want: []*gpb.Path{
			mustPath("/list-element[name=val1]"),
		},
	}, {
		name: "specified paths with multiple element expansion",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name:    "list-element",
			KeyName: "%%keyname%%",
		},
		inKnownVars: keyQuery{
			"%%keyname%%": []map[string]string{
				{"name": "val1"},
				{"name": "val2"},
			},
		},
		want: []*gpb.Path{
			mustPath("/list-element[name=val1]"),
			mustPath("/list-element[name=val2]"),
		},
	}, {
		name: "error expanding step",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name:    "list-element",
			KeyName: "%%invalid%%",
		},
		wantErrSubstring: "cannot resolve step",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeStep(tt.inQueryStep, tt.inKnownVars, tt.inKnownPaths)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("did not get expected error, %s", diff)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Fatalf("did not get expected output, diff(-got,+want):\n%s", diff)
			}
		})
	}
}

func TestResolvedPathElem(t *testing.T) {
	tests := []struct {
		name             string
		inQueryStep      *tpb.DataTreePaths_QueryStep
		inKeyQuery       keyQuery
		want             []*gpb.PathElem
		wantErrSubstring string
	}{{
		name: "no expansion",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name: "element",
		},
		want: []*gpb.PathElem{{Name: "element"}},
	}, {
		name: "with key, no expansion",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name: "element",
			Key:  map[string]string{"key-name": "value"},
		},
		want: []*gpb.PathElem{{Name: "element", Key: map[string]string{"key-name": "value"}}},
	}, {
		name: "with key, expanded to one value",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name:    "element",
			KeyName: "%%test%%",
		},
		inKeyQuery: keyQuery{"%%test%%": {
			map[string]string{
				"value": "forty-two",
			},
		}},
		want: []*gpb.PathElem{{
			Name: "element",
			Key:  map[string]string{"value": "forty-two"},
		}},
	}, {
		name: "with key, expanded to >1 value",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name:    "element",
			KeyName: "%%test%%",
		},
		inKeyQuery: keyQuery{"%%test%%": {
			map[string]string{
				"value": "forty-two",
			},
			map[string]string{
				"value": "forty-three",
			},
		}},
		want: []*gpb.PathElem{{
			Name: "element",
			Key:  map[string]string{"value": "forty-two"},
		}, {
			Name: "element",
			Key:  map[string]string{"value": "forty-three"},
		}},
	}, {
		name: "with invalid key_name",
		inQueryStep: &tpb.DataTreePaths_QueryStep{
			Name:    "element",
			KeyName: "%%invalid%%",
		},
		wantErrSubstring: "could not substitute for key name %%invalid%%, no specified values",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvedPathElem(tt.inQueryStep, tt.inKeyQuery)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("did not get expected error, %s", diff)
			}

			if err != nil {
				return
			}

			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Fatalf("did not get expected result, diff(-got,+want):\n%s", diff)
			}
		})
	}
}

func TestValueMatched(t *testing.T) {
	tests := []struct {
		name             string
		inVal            *ytypes.TreeNode
		inSpec           *tpb.PathValueMatch
		want             bool
		wantErrSubstring string
	}{{
		name: "simple equal test",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname: ygot.String("hostname"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname"}},
			},
		},
		want: true,
	}, {
		name: "simple non-equal test",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname: ygot.String("host1"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"host2"}},
			},
		},
		want: false,
	}, {
		name: "simple AND equal test",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:   ygot.String("hostname"),
				MotdBanner: ygot.String("motd"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname"}},
			},
			And: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"motd"}},
				},
			}},
		},
		want: true,
	}, {
		name: "simple AND unequal test",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:   ygot.String("hostname"),
				MotdBanner: ygot.String("fish"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname"}},
			},
			And: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"motd"}},
				},
			}},
		},
		want: false,
	}, {
		name: "simple OR test - true from parent",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:   ygot.String("hostname"),
				MotdBanner: ygot.String("fish"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname"}},
			},
			Or: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"motd"}},
				},
			}},
		},
		want: true,
	}, {
		name: "simple OR test - true from child",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:   ygot.String("hostname"),
				MotdBanner: ygot.String("fish"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"box00"}},
			},
			Or: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"fish"}},
				},
			}},
		},
		want: true,
	}, {
		name: "simple or test -- false",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:   ygot.String("hostname"),
				MotdBanner: ygot.String("fish"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"box00"}},
			},
			Or: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"sturgeon"}},
				},
			}},
		},
		want: false,
	}, {
		name: "nested or in and",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:    ygot.String("hostname"),
				MotdBanner:  ygot.String("motd"),
				LoginBanner: ygot.String("login"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname"}},
			},
			And: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"sturgeon"}},
				},
				Or: []*tpb.PathValueMatch{{
					Path: mustPath("config/login-banner"),
					Criteria: &tpb.PathValueMatch_Equal{
						&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"login"}},
					},
				}},
			}},
		},
		want: true,
	}, {
		name: "nested and in or",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:    ygot.String("hostname"),
				MotdBanner:  ygot.String("motd"),
				LoginBanner: ygot.String("login"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname"}},
			},
			Or: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"motd"}},
				},
				And: []*tpb.PathValueMatch{{
					Path: mustPath("config/login-banner"),
					Criteria: &tpb.PathValueMatch_Equal{
						&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"login"}},
					},
				}},
			}},
		},
		want: true,
	}, {
		name: "multiple ands",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:    ygot.String("hostname"),
				MotdBanner:  ygot.String("motd"),
				LoginBanner: ygot.String("login"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname"}},
			},
			And: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"motd"}},
				},
			}, {
				Path: mustPath("config/login-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"login"}},
				},
			}},
		},
		want: true,
	}, {
		name: "multiple ors",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:    ygot.String("hostname"),
				MotdBanner:  ygot.String("motd"),
				LoginBanner: ygot.String("login"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname-fish"}},
			},
			Or: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"motd-chips"}},
				},
			}, {
				Path: mustPath("config/login-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"login"}},
				},
			}},
		},
		want: true,
	}, {
		name: "a AND b OR c",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname:    ygot.String("hostname"),
				MotdBanner:  ygot.String("motd"),
				LoginBanner: ygot.String("login"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_Equal{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname"}},
			},
			And: []*tpb.PathValueMatch{{
				Path: mustPath("config/motd-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"not-equal"}},
				},
			}},
			Or: []*tpb.PathValueMatch{{
				Path: mustPath("config/login-banner"),
				Criteria: &tpb.PathValueMatch_Equal{
					&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"login"}},
				},
			}},
		},
		want: true,
	}, {
		name: "nil spec",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname: ygot.String("hostname"),
			},
		},
		want: true,
	}, {
		name: "spec with a nil node",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/motd-banner"),
		},
		wantErrSubstring: "tried to apply match against a nil node",
	}, {
		name: "spec against a leaf",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system/config/hostname"),
			Data:   ygot.String("hostname"),
		},
		inSpec:           &tpb.PathValueMatch{Path: mustPath("config/hostname")},
		wantErrSubstring: "matches can only be applied against YANG containers or lists",
	}, {
		name: "bad criteria",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname: ygot.String("hostname"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: &gpb.Path{
				Elem: []*gpb.PathElem{{
					Key: map[string]string{"key": "val"},
				}},
			},
		},
		wantErrSubstring: "invalid criteria type",
	}, {
		name: "bad path in equal",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname: ygot.String("hostname"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: &gpb.Path{
				Elem: []*gpb.PathElem{{
					Key: map[string]string{"key": "val"},
				}},
			},
			Criteria: &tpb.PathValueMatch_Equal{},
		},
		wantErrSubstring: "could not query path",
	}, {
		name: "zero results for query in equal",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface"),
			Criteria: &tpb.PathValueMatch_Equal{},
		},
		want:             false,
		wantErrSubstring: "could not query path",
	}, {
		name: "multiple results for query",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/interfaces/interface[name=eth0]"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				d.GetOrCreateInterface("eth0")
				d.GetOrCreateInterface("eth1")
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface"),
			Criteria: &tpb.PathValueMatch_Equal{},
		},
		wantErrSubstring: "query criteria was invalid",
	}, {
		name: "container - is set - match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				e := d.GetOrCreateInterface("eth0").GetOrCreateEthernet()
				e.PortSpeed = exampleoc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/ethernet"),
			Criteria: &tpb.PathValueMatch_IsSet{true},
		},
		want: true,
	}, {
		name: "container - is set - no match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				e := d.GetOrCreateInterface("eth0").GetOrCreateEthernet()
				// This sets config/port-speed with path compression enabled.
				e.PortSpeed = exampleoc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/state/counters"),
			Criteria: &tpb.PathValueMatch_IsSet{true},
		},
		want: false,
	}, {
		name: "container - is unset - match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				// This sets config/port-speed with path compression enabled.
				e := d.GetOrCreateInterface("eth0").GetOrCreateEthernet()
				e.PortSpeed = exampleoc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/state/counters"),
			Criteria: &tpb.PathValueMatch_IsUnset{true},
		},
		want: true,
	}, {
		name: "container - is unset - no match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				e := d.GetOrCreateInterface("eth0").GetOrCreateEthernet()
				// This sets config/port-speed with path compression enabled.
				e.PortSpeed = exampleoc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/ethernet"),
			Criteria: &tpb.PathValueMatch_IsUnset{true},
		},
		want: false,
	}, {
		name: "leaf - is set - match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				// This sets config/port-speed with path compression enabled.
				d.GetOrCreateInterface("eth0").GetOrCreateEthernet().PortSpeed = exampleoc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/ethernet/config/port-speed"),
			Criteria: &tpb.PathValueMatch_IsSet{true},
		},
		want: true,
	}, {
		name: "leaf - is set - no match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				d.GetOrCreateInterface("eth0").GetOrCreateEthernet()
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/ethernet/config/port-speed"),
			Criteria: &tpb.PathValueMatch_IsSet{true},
		},
		want: false,
	}, {
		name: "leaf - is unset - match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				d.GetOrCreateInterface("eth0").GetOrCreateEthernet()
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/ethernet/config/port-speed"),
			Criteria: &tpb.PathValueMatch_IsUnset{true},
		},
		want: true,
	}, {
		name: "leaf - is unset - no match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				e := d.GetOrCreateInterface("eth0").GetOrCreateEthernet()
				// This sets config/port-speed with path compression enabled.
				e.PortSpeed = exampleoc.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/ethernet/config/port-speed"),
			Criteria: &tpb.PathValueMatch_IsUnset{true},
		},
		want: false,
	}, {
		name: "is set with or",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				d.GetOrCreateSystem().Hostname = ygot.String("foo")
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/state/oper-status"),
			Criteria: &tpb.PathValueMatch_IsSet{true},
			Or: []*tpb.PathValueMatch{{
				Path:     mustPath("/system/config/hostname"),
				Criteria: &tpb.PathValueMatch_IsSet{true},
			}},
		},
		want: true,
	}, {
		name: "is unset with or",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				d.GetOrCreateInterface("eth0").OperStatus = exampleoc.OpenconfigInterfaces_Interface_OperStatus_UP
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path:     mustPath("/interfaces/interface[name=eth0]/state/oper-status"),
			Criteria: &tpb.PathValueMatch_IsUnset{true},
			Or: []*tpb.PathValueMatch{{
				Path:     mustPath("/system/config/hostname"),
				Criteria: &tpb.PathValueMatch_IsUnset{true},
			}},
		},
		want: true,
	}, {
		name: "not_equal - match",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname: ygot.String("hostname"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_NotEqual{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"hostname42"}},
			},
		},
		want: true,
	}, {
		name: "not_equal - not matched",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["System"],
			Path:   mustPath("/system"),
			Data: &exampleoc.System{
				Hostname: ygot.String("host1"),
			},
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("config/hostname"),
			Criteria: &tpb.PathValueMatch_NotEqual{
				&gpb.TypedValue{Value: &gpb.TypedValue_StringVal{"host1"}},
			},
		},
		want: false,
	}, {
		name: "not equal with or",
		inVal: &ytypes.TreeNode{
			Schema: exampleoc.SchemaTree["Device"],
			Path:   mustPath("/"),
			Data: func() *exampleoc.Device {
				d := &exampleoc.Device{}
				i := d.GetOrCreateInterface("eth0")
				i.OperStatus = exampleoc.OpenconfigInterfaces_Interface_OperStatus_UP
				i.Description = ygot.String("fish")
				return d
			}(),
		},
		inSpec: &tpb.PathValueMatch{
			Path: mustPath("/interfaces/interface[name=eth0]/state/oper-status"),
			Criteria: &tpb.PathValueMatch_NotEqual{
				&gpb.TypedValue{
					Value: &gpb.TypedValue_StringVal{"UP"},
				},
			},
			Or: []*tpb.PathValueMatch{{
				Path: mustPath("/interfaces/interface[name=eth0]/config/description"),
				Criteria: &tpb.PathValueMatch_NotEqual{
					&gpb.TypedValue{
						Value: &gpb.TypedValue_StringVal{"chips"},
					},
				},
			}},
		},
		want: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := valueMatched(tt.inVal, tt.inSpec)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("did not get expected error, %s", diff)
			}

			if got != tt.want {
				t.Fatalf("did not get expected result, got: %v, want: %v", got, tt.want)
			}
		})
	}
}
