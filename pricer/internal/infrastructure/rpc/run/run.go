package run

import (
	"github.com/shortlink-org/go-sdk/grpc"
)

type Response struct{}

func Run(runRPCServer *grpc.Server) (*Response, error) {
	if runRPCServer != nil {
		go runRPCServer.Run()
	}

	// Run() blocks until shutdown; return value is only for compatibility.
	return nil, nil //nolint:nilnil // Run is fire-and-forget, caller ignores result
}
