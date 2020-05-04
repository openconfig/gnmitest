// Package datatreepaths implements a test which can check the contents
// of the data tree for particular path. The query specification described
// in tests.proto is used to recursively iterate through the tree performing
// list key substitution.
package datatreepaths

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/golang/protobuf/proto"

	"github.com/openconfig/gnmitest/common/testerror"
	"github.com/openconfig/gnmitest/register"
	"github.com/openconfig/gnmitest/schemas"
	"github.com/openconfig/gnmitest/subscribe"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/util"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	rpb "github.com/openconfig/gnmitest/proto/report"
	tpb "github.com/openconfig/gnmitest/proto/tests"
)

// test implements the subscribe.Test interface for the DataTreePaths test.
type test struct {
	subscribe.Test

	// dataTree is the tree into which Notifications are deserialised
	dataTree ygot.GoStruct
	// schema is the root entry for the schema stored in dataTree
	schema *yang.Entry

	// testSpec is the configuration for the test specified in the
	// suite protobuf.
	testSpec *tpb.DataTreePaths
	// ignoreInvalidPaths specifies whether the test has been asked
	// to ignore paths that do not deserialise correctly.
	ignoreInvalidPaths bool
}

// init statically registers the test against the gnmitest framework.
func init() {
	register.NewSubscribeTest(&tpb.SubscribeTest_DataTreePaths{}, newTest)
}

// newTest creates a new instance eof the DataTreePaths test.
func newTest(st *tpb.Test) (subscribe.Subscribe, error) {
	goStruct, err := schema.Get(st.GetSchema())
	if err != nil {
		return nil, fmt.Errorf("failed to get %v schema: %v", st.GetSchema(), err)
	}

	root := goStruct.NewRoot()
	tn := reflect.TypeOf(root).Elem().Name()
	schema, err := goStruct.Schema(tn)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for %q: %v", tn, err)
	}

	return &test{
		dataTree:           root,
		schema:             schema,
		testSpec:           st.GetSubscribe().GetDataTreePaths(),
		ignoreInvalidPaths: st.GetSubscribe().GetIgnoreInvalidPaths(),
	}, nil
}

// Check builds the queries that are specified by the input test definition,
// and validates them against the dataTree stored in test. It returns an error
// if the required paths in the test are not found in the datatree.
func (t *test) Check() error {
	queries, err := t.queries()
	if err != nil {
		return fmt.Errorf("cannot resolve paths to query, %v", err)
	}

	errs := &testerror.List{}
	// Check the required paths that are specified in the operation.
	for _, q := range queries.paths {
		nodes, err := ytypes.GetNode(t.schema, t.dataTree, q)
		switch {
		case err != nil:
			errs.AddTestErr(&rpb.TestError{
				Path:    q,
				Message: fmt.Sprintf("cannot retrieve node, %v", err),
			})
		case len(nodes) == 1:
			_, isGoEnum := nodes[0].Data.(ygot.GoEnum)
			vv := reflect.ValueOf(nodes[0].Data)
			switch {
			case util.IsValuePtr(vv) && (util.IsValueNil(vv.Elem()) || !vv.Elem().IsValid()):
				errs.AddTestErr(&rpb.TestError{
					Path:    q,
					Message: "got nil data for path",
				})
			case isGoEnum:
				// This is an enumerated value -- check whether it is set to 0
				// which means it was not set.
				if vv.Int() == 0 {
					errs.AddTestErr(&rpb.TestError{
						Path:    q,
						Message: fmt.Sprintf("enum type %T was UNSET", vv.Interface()),
					})
				}
			}
		case len(nodes) == 0:
			errs.AddTestErr(&rpb.TestError{
				Path:    q,
				Message: "no matches for path",
			})
		}
	}

	for _, v := range queries.vals {
		ok, err := valueMatched(&ytypes.TreeNode{
			Schema: t.schema,
			Data:   t.dataTree,
		}, v)

		switch {
		case err != nil:
			errs.AddTestErr(&rpb.TestError{
				Path:    v.Path,
				Message: fmt.Sprintf("cannot check node value, %v", err),
			})
		case !ok:
			errs.AddTestErr(&rpb.TestError{
				Path:    v.Path,
				Message: fmt.Sprintf("did not match expected value, %v", v),
			})
		}
	}

	if len(errs.Errors()) == 0 {
		return nil
	}

	sortedErrs := &testerror.List{}
	paths := []string{}
	errMap := map[string]*rpb.TestError{}
	for _, e := range errs.Errors() {
		s, err := ygot.PathToString(e.Path)
		if err != nil {
			// If there's no way we can sort the errors, then just prefer to
			// ensure that we return some error condition.
			return errs
		}
		paths = append(paths, s)
		errMap[s] = e
	}

	sort.Strings(paths)
	for _, p := range paths {
		sortedErrs.AddTestErr(errMap[p])
	}

	return sortedErrs
}

// Process is called for each response received from the target for the test.
// It returns the current status of the test (running, or complete) based
// on the contents of the sr SubscribeResponse.
func (t *test) Process(sr *gpb.SubscribeResponse) (subscribe.Status, error) {
	return subscribe.OneShotSetNode(t.schema, t.dataTree, sr,
		subscribe.OneShotSetNodeArgs{
			YtypesArgs: []ytypes.SetNodeOpt{
				&ytypes.InitMissingElements{},
			},
			IgnoreInvalidPaths: t.ignoreInvalidPaths,
		},
	)
}

// queries resolves the contents of the testSpec into the exact paths to be
// queried from the data tree. It should be called after the data tree has been
// fully populated.
func (t *test) queries() (*resolvedOperation, error) {
	cfg := t.testSpec.GetTestOper()
	if cfg == nil {
		return nil, fmt.Errorf("invalid nil test specification")
	}
	knownVars := keyQuery{}

	queryPaths, err := t.resolveQuery(cfg, knownVars)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve query, %v", err)
	}

	return queryPaths, nil
}

// resolvedOperation stores a fully defined test operation that has been resolved
// from the query specification.
type resolvedOperation struct {
	// paths stores a resolved set of gNMI paths. Each path that is stored
	// in the paths set is from the required_paths argument of the test. The
	// pass/fail criteria for these paths is that they must be set to a non-nil
	// value within the datatree after the SubscribeResponse messages received
	// on the subscription are processed into the datatree.
	paths []*gpb.Path
	// vals stores a set of fully resolved path, value specifications. These
	// criteria are extracted from the required_values argument of the test. The
	// pass/fail criteria for each path is that it conforms to the value criteria
	// that are specified within the test after each of the SubscribeResponse messages
	// received from the subscription are processed into the datatree.
	vals []*tpb.PathValueMatch
}

// resolveQuery resolves an individual query into the set of paths that it
// corresponds to. The query is specified by the op specified, and the
// knownVars are used to extract values that have already been queried from the
// data tree. It returns the set of paths.
func (t *test) resolveQuery(op *tpb.DataTreePaths_TestQuery, knownVars keyQuery) (*resolvedOperation, error) {
	q, err := makeQuery(op.Steps, knownVars)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve query %s, %v", op, err)
	}

	rOps := &resolvedOperation{}

	for _, path := range q {
		// Make sure we append to a new map.
		newVars := joinVars(knownVars, nil)

		switch v := op.GetType().(type) {
		case *tpb.DataTreePaths_TestQuery_RequiredPaths:
			for _, rp := range v.RequiredPaths.GetPaths() {
				tp := path.GetElem()
				tp = append(tp, v.RequiredPaths.GetPrefix().GetElem()...)
				tp = append(tp, rp.GetElem()...)
				rOps.paths = append(rOps.paths, &gpb.Path{Elem: tp})
			}
		case *tpb.DataTreePaths_TestQuery_RequiredValues:
			for _, rv := range v.RequiredValues.GetMatches() {
				tp := path.GetElem()
				tp = append(tp, v.RequiredValues.GetPrefix().GetElem()...)
				tp = append(tp, rv.GetPath().GetElem()...)
				newOper := proto.Clone(rv).(*tpb.PathValueMatch)
				newOper.Path = &gpb.Path{Elem: tp}
				rOps.vals = append(rOps.vals, newOper)
			}
		case *tpb.DataTreePaths_TestQuery_GetListKeys:
			nextQ := v.GetListKeys.GetNextQuery()
			if nextQ == nil {
				return nil, fmt.Errorf("get_list_keys query %s specified nil next_query", v)
			}

			queriedKeys, err := t.queryListKeys(path, v.GetListKeys.GetFilter())
			if err != nil {
				return nil, fmt.Errorf("cannot resolve query, failed get_list_keys, %v", err)
			}

			for _, key := range queriedKeys {
				retp, err := t.resolveQuery(nextQ, joinVars(newVars, keyQuery{v.GetListKeys.VarName: []map[string]string{key}}))
				if err != nil {
					return nil, fmt.Errorf("cannot resolve query %s, %v", nextQ, err)
				}
				// A resolved operation for a GetListKey cannot specify a value required, and
				// hence we just append the paths.
				rOps.paths = append(rOps.paths, retp.paths...)
			}
		default:
			return nil, fmt.Errorf("got unhandled type in operation type, %T", v)
		}
	}

	return rOps, nil
}

// joinVars merges the contents of the two keyQuery maps into a single map, overwriting
// any value in the first map with the value in the second map if the keys overlap.
func joinVars(a, b keyQuery) keyQuery {
	nm := keyQuery{}
	for _, kq := range []keyQuery{a, b} {
		for k, v := range kq {
			nm[k] = v
		}
	}
	return nm
}

// queryListKeys queries the dataTree stored in the test receiver for the path
// specified by p, returning the keys of the list found at p. If a filter is
// specified, only list entries that meet the specified criteria are returned.
// If the value found at the specified path is not a list, an error is returned.
func (t *test) queryListKeys(path *gpb.Path, filter *tpb.PathValueMatch) ([]map[string]string, error) {
	nodes, err := ytypes.GetNode(t.schema, t.dataTree, path, &ytypes.GetPartialKeyMatch{})
	if err != nil {
		return nil, fmt.Errorf("cannot query for path %s, %v", path, err)
	}

	keys := []map[string]string{}
	for _, n := range nodes {
		if !n.Schema.IsList() {
			return nil, fmt.Errorf("path %s returned by query %s was not a list, was: %v", path, n.Path, n.Schema.Kind)
		}

		if filter != nil {
			match, err := valueMatched(n, filter)
			switch {
			case err != nil:
				return nil, fmt.Errorf("invalid filter criteria for %s, %v", path, err)
			case match == false:
				continue
			}
		}
		keys = append(keys, n.Path.GetElem()[len(n.Path.GetElem())-1].Key)
	}

	return keys, nil
}

// valueMatched determines whether the specified val matches the spec specified.
// It returns true if the value matches.
func valueMatched(val *ytypes.TreeNode, spec *tpb.PathValueMatch) (bool, error) {
	if spec == nil {
		return true, nil
	}

	if val.Data == nil {
		return false, fmt.Errorf("tried to apply match against a nil node")
	}

	specMatched, matchErr := nodeMatchesCriteria(val.Schema, val.Data, spec)

	andMatched := true
	// First check whether all AND criteria match. If there are any that do not match,
	// then we return false.
	for _, and := range spec.And {
		match, err := valueMatched(val, and)
		if err != nil {
			return false, fmt.Errorf("cannot parse match criteria %v, %v", and, err)
		}
		if !match {
			andMatched = false
		}
	}

	if len(spec.Or) == 0 {
		// If there are no OR criteria, we can return immediately if the spec
		// and all its AND conditions matched.
		return specMatched && andMatched, matchErr
	}

	if !specMatched {
		// Check any of the OR criteria specified in the query - we only check if
		// the initial critiera was not matched.
		for _, or := range spec.Or {
			match, err := valueMatched(val, or)
			if err != nil {
				return false, fmt.Errorf("cannot parse match criteria %v, %v", or, err)
			}
			if match {
				return true, nil
			}
		}
	}

	return specMatched, matchErr

}

// nodeMatchesCriteria evaluates whether the spec supplied is matched for the provided root.
// AND and OR criteria are not matched.
func nodeMatchesCriteria(rootSchema *yang.Entry, root interface{}, spec *tpb.PathValueMatch) (bool, error) {
	// nil criteria are considered to match.
	if spec == nil {
		return true, nil
	}

	goStruct, ok := root.(ygot.ValidatedGoStruct)
	if !ok {
		return false, fmt.Errorf("matches can only be applied against YANG containers or lists, invalid root type %T", root)
	}

	nodes, err := ytypes.GetNode(rootSchema, goStruct, spec.Path, &ytypes.GetPartialKeyMatch{})
	getNodeErr, ok := status.FromError(err)
	if !ok {
		return false, fmt.Errorf("got invalid error from GetNode, %T", err)
	}

	switch v := spec.Criteria.(type) {
	case *tpb.PathValueMatch_Equal:
		return matchNodeEqual(nodes, getNodeErr, v.Equal)
	case *tpb.PathValueMatch_NotEqual:
		m, err := matchNodeEqual(nodes, getNodeErr, v.NotEqual)
		return !m, err
	case *tpb.PathValueMatch_IsSet:
		return matchNodeIsSet(nodes, getNodeErr)
	case *tpb.PathValueMatch_IsUnset:
		return matchNodeIsUnset(nodes, getNodeErr)
	default:
		return false, fmt.Errorf("invalid criteria type specified %T", v)
	}
}

// matchNodeEqual determines whether the single node in the nodes slice
// supplied is equal to testVal. If there is more than one node in the slice an
// error is returned. The getNodeStatus supplied is used to handle the response
// of ytypes.GetNode specifically in the context of testing for equality. It
// returns a bool indicating whether the values are equal, and an error if they
// are not equal and an error was encountered whilst trying to test for
// equality.
func matchNodeEqual(nodes []*ytypes.TreeNode, getNodeStatus *status.Status, testVal *gpb.TypedValue) (bool, error) {
	switch {
	case getNodeStatus.Code() != codes.OK:
		// All errors for GetNode are fatal when checking for equality.
		return false, fmt.Errorf("could not query path, %s", getNodeStatus.Proto())
	case len(nodes) == 0:
		// No nodes were found, so this cannot be equal.
		return false, fmt.Errorf("no data tree node")
	case len(nodes) > 1:
		// Too many nodes returned for an equality check to be relevant.
		return false, fmt.Errorf("query criteria was invalid, %d nodes returned", len(nodes))
	default:
		typedVal, err := ygot.EncodeTypedValue(nodes[0].Data, gpb.Encoding_JSON_IETF)
		if err != nil {
			return false, fmt.Errorf("cannot encode received value %v as TypedValue, %v", nodes[0], err)
		}
		return proto.Equal(typedVal, testVal), nil
	}
}

// matchNodeIsSet determines whether the single node in the nodes slice supplied is
// set to a non-nil value. If there is more than one value in the nodes slice, the
// an error is returned. The getNodeStatus supplied is used to handle the response
// of ytypes.GetNode in the context of testing for a non-nil returned node. A bool
// indicating whether the node is set is returned, along with an error indicating
// if invalid input data was supplied.
func matchNodeIsSet(nodes []*ytypes.TreeNode, getNodeStatus *status.Status) (bool, error) {
	switch {
	case getNodeStatus.Code() != codes.OK:
		// All errors are fatal when checking for a set node.
		return false, fmt.Errorf("could not retrieve query path, %s", getNodeStatus.Proto())
	case len(nodes) == 0:
		// The node cannot be set if there is no value returned.
		return false, fmt.Errorf("no data tree node")
	case len(nodes) > 1:
		// Too many nodes returned for a set check to be valid.
		return false, fmt.Errorf("query criteria was invalid, %d nodes returned", len(nodes))
	default:
		return !util.IsValueNilOrDefault(nodes[0].Data), nil
	}
}

// modeNodesIsUnset determines whether the single node in the nodes slice supplied
// is set to a nil value, or there are no supplied nodes. The supplied getNodeStatus
// is used to perform handling of the return of ytypes.GetNodes in the context
// of testing for a nil value. A bool indicating whether the node is unset is returned,
// along with an error if it is not possible to check whether the node is nil.
func matchNodeIsUnset(nodes []*ytypes.TreeNode, getNodeStatus *status.Status) (bool, error) {
	switch {
	case getNodeStatus.Code() != codes.OK:
		// codes.NotFound is an OK return status if the node is not set.
		if getNodeStatus.Code() == codes.NotFound {
			return true, nil
		}
		return false, fmt.Errorf("could not retrieve query path, %s", getNodeStatus.Proto())
	case len(nodes) == 0:
		return true, nil
	case len(nodes) > 1:
		return false, fmt.Errorf("query criteria was invalid, %d nodes returned", len(nodes))
	default:
		return util.IsValueNilOrDefault(nodes[0].Data), nil
	}
}

// keyQuery is a type that can be used to store a set of key specifications
// for a query. The outer map is keyed by a user-defined variable name, and
// the value is a slice of maps specifying the keys in a gNMI PathElem message.
type keyQuery map[string][]map[string]string

// makeQuery takes an input slice of QuerySteps and resolves them into the set of
// gNMI paths that should be tested, using the knownVars keyQuery to resolve any
// variables that are specified.
func makeQuery(steps []*tpb.DataTreePaths_QueryStep, knownVars keyQuery) ([]*gpb.Path, error) {
	var (
		paths []*gpb.Path
		err   error
	)

	for _, s := range steps {
		paths, err = makeStep(s, knownVars, paths)
		if err != nil {
			return nil, err
		}
	}
	return paths, nil
}

// makeStep takes an input QueryStep (step), a set of currently known variables
// (knownVars), and the set of paths being processed in the current context, and
// resolves them into a fully qualified set of gNMI Paths.
func makeStep(step *tpb.DataTreePaths_QueryStep, knownVars keyQuery, knownPaths []*gpb.Path) ([]*gpb.Path, error) {
	paths := knownPaths
	if len(paths) == 0 {
		paths = []*gpb.Path{{}} // seed the paths with one path to be appended to.
	}

	resolvedElems, err := resolvedPathElem(step, knownVars)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve step %s, %v", step, err)
	}

	np := []*gpb.Path{}
	switch len(resolvedElems) {
	case 1:
		// Handle the case that we did not expand the path elements
		// out, and simply had one returned.
		for _, p := range paths {
			np = append(np, &gpb.Path{Elem: append(p.Elem, resolvedElems[0])})
		}
	default:
		for _, p := range paths {
			for _, e := range resolvedElems {
				expPath := proto.Clone(p).(*gpb.Path)
				expPath.Elem = append(expPath.Elem, e)
				np = append(np, expPath)
			}
		}
	}

	return np, nil
}

// resolvedPathElem takes an input QueryStep and resolves it into a slice of gNMI
// PathElems that can be exactly matched. The input kv keyQuery is used to resolve
// any variable names that require substitution.
//
// For example, if the QueryStep provided specifies:
// {
//   name: "interface"
//   key_name: "%%interface%%"
// }
//
// The values of kv["%%interface%%"] will be appended to a gNMI PathElem with the
// name "interface" and returned. If kv["%%interface%%"] =
// []map[string]string{{"name": "eth0"}} then the value returned is:
//
// {
//	name: "interface"
//	key {
//		key: "name"
//		value: "eth0"
//	}
// }
func resolvedPathElem(p *tpb.DataTreePaths_QueryStep, kv keyQuery) ([]*gpb.PathElem, error) {
	if p.GetKeyName() == "" {
		return []*gpb.PathElem{{Name: p.Name, Key: p.Key}}, nil
	}

	v, ok := kv[p.GetKeyName()]
	if !ok {
		return nil, fmt.Errorf("could not substitute for key name %s, no specified values", p.GetKeyName())
	}

	elems := []*gpb.PathElem{}
	for _, keys := range v {
		elems = append(elems, &gpb.PathElem{Name: p.Name, Key: keys})
	}

	return elems, nil
}
