package main

import (
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/mozership/probe/sdk"
)

// params are the keys the action takes in with: those of Req, and proto,
// which it reads apart from Req.
func params() []string {
	return append(sdk.ParamNames(Req{}), "proto")
}

// resFields are the fields of res: those of Res, and those the action adds
// for a call checked against .proto files, one whose responses stream, and a
// listing.
func resFields() []string {
	return append(sdk.ParamNames(Res{}), "complete", "contract", "messages", "methods", "services", "types", "violations")
}

func readManifest(t *testing.T, path string) *sdk.Manifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m, err := sdk.ParseManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// TestManifest checks that action.yml declares what the action does, and
// that the action.yml of the e2e workflows declares the same.
func TestManifest(t *testing.T) {
	m := readManifest(t, "action.yml")
	if want := []string{sdk.KindReadOnly, sdk.KindAllowHost}; !slices.Equal(m.Guard, want) {
		t.Errorf("guard = %v, want %v", m.Guard, want)
	}
	if !slices.Equal(m.Params, params()) {
		t.Errorf("params = %v, want %v", m.Params, params())
	}
	if !slices.Equal(m.Res, resFields()) {
		t.Errorf("res = %v, want %v", m.Res, resFields())
	}
	if !slices.Equal(m.Contracts, []string{"proto"}) {
		t.Errorf("contracts = %v, want [proto]", m.Contracts)
	}
	choices := map[string][]string{"protocol": {"grpc", "connect"}, "codec": {"json", "proto"}}
	if len(m.Choices) != len(choices) {
		t.Errorf("choices = %v, want %v", m.Choices, choices)
	}
	for key, values := range choices {
		if !slices.Equal(m.Choices[key].Values, values) {
			t.Errorf("choices.%s = %v, want %v", key, m.Choices[key].Values, values)
		}
	}

	local := readManifest(t, "e2e/action/action.yml")
	if !slices.Equal(local.Guard, m.Guard) || !slices.Equal(local.Params, m.Params) || !slices.Equal(local.Res, m.Res) || !slices.Equal(local.Contracts, m.Contracts) || !maps.EqualFunc(local.Choices, m.Choices, func(a, b sdk.Choice) bool { return slices.Equal(a.Values, b.Values) && a.IgnoreCase == b.IgnoreCase }) {
		t.Error("e2e/action/action.yml declares other than action.yml does")
	}
}

// TestResFields runs calls of each kind, and checks that res holds only the
// fields action.yml declares.
func TestResFields(t *testing.T) {
	grpcAddr := startUserServer(t)
	s := startConnectServer(t, "")
	files := protoOf(writeProto(t, userProto(t)), false)
	a := &Action{log: hclog.NewNullLogger()}

	var calls []map[string]any
	for _, protocol := range []map[string]any{
		{"addr": grpcAddr, "metadata": map[string]any{"x-probe": "1"}, "timeout": "5s"},
		{"protocol": "connect", "addr": s.url, "proto": files},
		{"protocol": "connect", "addr": s.url, "codec": "proto", "proto": files},
	} {
		for _, call := range []map[string]any{
			{"method": "GetUser", "body": `{"user_id": "1"}`},
			{"method": "WatchUsers", "body": `{"count": 2}`, "max_messages": 5},
			{"method": "GetUser", "body": `{"user_id": "1"}`, "proto": files},
		} {
			with := map[string]any{"service": "UserService"}
			maps.Copy(with, protocol)
			maps.Copy(with, call)
			calls = append(calls, with)
		}
	}
	calls = append(calls,
		map[string]any{"addr": grpcAddr, "list": true},
		map[string]any{"addr": grpcAddr, "list": true, "service": "UserService"},
	)
	for _, with := range calls {
		ret, _, err := a.RunStep(sdk.Call{With: with, Contract: &sdk.Contract{}})
		if err != nil {
			t.Fatalf("%v: %v", with, err)
		}
		res, ok := ret["res"].(map[string]any)
		if !ok {
			t.Fatalf("%v: res = %#v", with, ret["res"])
		}
		for f := range res {
			if !slices.Contains(resFields(), f) {
				t.Errorf("%v: res has %s, which action.yml does not declare", with, f)
			}
		}
	}
}

func TestRequired(t *testing.T) {
	a := &Action{log: hclog.NewNullLogger()}
	base := map[string]any{"addr": "127.0.0.1:1", "service": "UserService", "method": "GetUser", "timeout": "1s"}
	for _, key := range []string{"addr", "service", "method"} {
		with := maps.Clone(base)
		delete(with, key)
		if _, err := a.Run(with); err == nil {
			t.Errorf("without %s: Run succeeded, want an error", key)
		}
	}
}
