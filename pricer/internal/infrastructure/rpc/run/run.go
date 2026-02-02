package run

import (
	"github.com/shortlink-org/go-sdk/grpc"
)

type Response struct{}

func Run(runRPCServer *grpc.Server) (*Response, error) {
	if runRPCServer != nil {
		go runRPCServer.Run()
	}

	return nil, nil
}
