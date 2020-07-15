package brpc

import (
	"context"
	"net"
	"time"

	"google.golang.org/protobuf/proto"

	brpcpb "github.com/tomwei7/brpc-go/genproto/brpc"
	policypb "github.com/tomwei7/brpc-go/genproto/brpc/policy"
)

type frame struct {
	RPCMeta policypb.RpcMeta
	Body    []byte
}

func newReqFrame(correlationID int64, reqMeta *policypb.RpcRequestMeta) *frame {
	return &frame{
		RPCMeta: policypb.RpcMeta{
			Request:       reqMeta,
			CompressType:  proto.Int32(int32(brpcpb.CompressType_COMPRESS_TYPE_NONE)),
			CorrelationId: proto.Int64(correlationID),
		},
	}

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

func (f *frame) AppendMessage(message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	f.Body = append(f.Body, data...)
	return nil
}

func (f *frame) AppendAttachment(attachment []byte) {
	if len(attachment) == 0 {
		return
	}
	f.RPCMeta.AttachmentSize = proto.Int32(int32(len(attachment)))
	f.Body = append(f.Body, attachment...)
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

func (p *pipeline) Close() error {
	return p.conn.Close()
}
