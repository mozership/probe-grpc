package main

import (
	"context"
	"errors"

	"github.com/hashicorp/go-hclog"
	"github.com/mozership/probe/sdk"
)

// Action makes the gRPC or Connect call a step gives.
type Action struct {
	log hclog.Logger
}

// Run makes the call in with, under no guard.
func (a *Action) Run(with map[string]any) (map[string]any, error) {
	ret, _, err := a.RunStep(sdk.Call{With: with})
	return ret, err
}

// RunStep makes the call for a step, under the guard of the run, which
// refuses a host it does not allow and, under --read-only, a method its
// definitions do not declare free of side effects.
func (a *Action) RunStep(call sdk.Call) (map[string]any, map[string]any, error) {
	with := call.With
	// Validate that required parameters are provided
	if len(with) == 0 {
		return map[string]any{}, nil, errors.New("grpc action requires parameters in 'with' section. Please specify request details like addr, service, method")
	}

	sdk.LogParams(a.log, "received grpc request parameters", with)

	before := WithBefore(func(ctx context.Context, service, method string) {
		a.log.Trace("grpc request prepared", "service", service, "method", method)
	})
	after := WithAfter(func(res *Res) {
		a.log.Trace("grpc response received", "status", res.StatusCode)
	})
	ret, err := RequestStep(call, before, after)

	sdk.LogOutcome(a.log, "grpc request", ret, err)

	return ret, nil, err
}
