package brpc

import (
	"context"

	"google.golang.org/grpc"
)

var _ grpc.ClientConnInterface = &ClientConn{}

type ClientConn struct{}

// Invoke performs a unary RPC and returns after the response is received
// into reply.
func (c *ClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return nil
}

func (c *ClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}
