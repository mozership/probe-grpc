// Command probe-grpc serves a Probe action that calls gRPC and Connect
// services, and checks the calls against .proto files. Probe starts it for a
// step that uses github.com/mozership/probe-grpc@<commit>.
package main

import (
	"github.com/hashicorp/go-hclog"
	"github.com/mozership/probe/sdk"
)

func main() {
	sdk.Serve(func(log hclog.Logger) sdk.Action {
		return &Action{log: log}
	})
}
