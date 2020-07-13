package brpc

import (
	"context"
	"net"
	"time"

	brpcpb "github.com/tomwei7/brpc-go/genproto/brpc"
	policypb "github.com/tomwei7/brpc-go/genproto/brpc/policy"
)

type frame struct {
	RPCMeta policypb.RpcMeta
	Body    []byte
}

func (f *frame) AttachmentSize() (size int) {
	if f.RPCMeta.AttachmentSize != nil {
		size = int(*f.RPCMeta.AttachmentSize)
	}
	return
}

func (f *frame) CompressType() brpcpb.CompressType {
	compressType := brpcpb.CompressType_COMPRESS_TYPE_NONE
	if f.RPCMeta.CompressType != nil {
		compressType = brpcpb.CompressType(*f.RPCMeta.CompressType)
	}
	return compressType
}

type pipeline struct {
	enc *encodec
	dec *decodec

	conn net.Conn
}

func newPipeline(conn net.Conn) *pipeline {
	enc := newEncodec(conn)
	dec := newDecodec(conn)
	return &pipeline{enc: enc, dec: dec, conn: conn}
}

func (p *pipeline) SendRequest(ctx context.Context, req *frame) error {
	deadline, ok := ctx.Deadline()
	if ok {
		if time.Now().After(deadline) {
			return context.DeadlineExceeded
		}
		if err := p.conn.SetWriteDeadline(deadline); err != nil {
			return err
		}
	}
	return p.enc.Encode(req)
}

func (p *pipeline) ReceiveResponse(ctx context.Context) (*frame, error) {
	deadline, ok := ctx.Deadline()
	if ok {
		if time.Now().After(deadline) {
			return nil, context.DeadlineExceeded
		}
		if err := p.conn.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
	}
	f := new(frame)
	return f, p.dec.Decode(f)
}
