package main

import (
	"fmt"

	"github.com/mozership/probe/sdk"
)

// contractFlag reads the bool key of the contract option name, o, taking
// def when it is left out.
func contractFlag(o map[string]any, name, key string, def bool) (bool, error) {
	v, ok := o[key]
	if !ok {
		return def, nil
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("%s.%s must be true or false", name, key)
	}
	return b, nil
}

// violation is a way a request or a response breaks the contract, in the
// shape res.violations holds and sdk.ContractCheckOf reads.
func violation(in, message, reason, field string) map[string]any {
	v := map[string]any{"in": in, "message": message}
	if reason != "" {
		v["reason"] = reason
	}
	if field != "" {
		v["field"] = field
	}
	return v
}

// reportContract tells the runner how the step kept to its contract, and
// puts the violations and what was matched in res. A step not checked
// against a contract, whose violations are nil, reports nothing.
func reportContract(c *sdk.Contract, ret map[string]any, violations []any, matched map[string]any) {
	if violations != nil {
		c.Report(sdk.ContractCheckOf(violations, matched))
	}
	res, ok := ret["res"].(map[string]any)
	if !ok {
		return
	}
	if violations != nil {
		res["violations"] = violations
	}
	if matched != nil {
		res["contract"] = matched
	}
}
