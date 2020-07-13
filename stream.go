package brpc

import (
	"context"

	"google.golang.org/grpc"
)

type stream struct {
	pipe *pipeline
}

func (s *stream) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return nil
}
