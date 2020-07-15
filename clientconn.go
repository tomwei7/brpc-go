package brpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

const (
	defaultConcurrency = 128
)

var _ grpc.ClientConnInterface = &ClientConn{}

type ClientConn struct {
	s *stream
}

// Invoke performs a unary RPC and returns after the response is received
// into reply.
func (c *ClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return c.s.Invoke(ctx, method, args, reply, opts...)
}

func (c *ClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func DialContext(ctx context.Context, address string) (*ClientConn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	s := newStream(conn, defaultConcurrency)
	return &ClientConn{s: s}, nil
}
