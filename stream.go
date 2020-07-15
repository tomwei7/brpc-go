package brpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	leagcyproto "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	policypb "github.com/tomwei7/brpc-go/genproto/brpc/policy"
)

var ErrUnavailable = fmt.Errorf("stream unavailable")
var ErrReachMaxConcurrency = fmt.Errorf("stream reach max concurrency")

type resp struct {
	err   error
	frame *frame
}

type eventHub struct {
	mx  sync.Mutex
	chs map[int64]chan *resp
}

func (e *eventHub) register(correlationID int64) (<-chan *resp, context.CancelFunc) {
	ch := make(chan *resp, 1)
	e.mx.Lock()
	defer e.mx.Unlock()
	e.chs[correlationID] = ch
	return ch, func() { e.unRegister(correlationID) }
}

func (e *eventHub) unRegister(correlationID int64) {
	e.mx.Lock()
	defer e.mx.Unlock()
	delete(e.chs, correlationID)
}

func (e *eventHub) emitResp(correlationID int64, f *frame, err error) {
	resp := &resp{err: err, frame: f}
	e.mx.Lock()
	defer e.mx.Unlock()
	ch, ok := e.chs[correlationID]
	if !ok {
		return
	}
	select {
	case ch <- resp:
	default:
	}
}

func (e *eventHub) emitError(err error) {
	e.mx.Lock()
	defer e.mx.Unlock()
	for _, ch := range e.chs {
		select {
		case ch <- &resp{err: err}:
		default:
		}
	}
}

type stream struct {
	correlationId int64
	unavailable   bool

	pipe *pipeline

	reqCh    chan struct{}
	eventHub *eventHub
}

func newStream(conn net.Conn, maxConcurrency int) *stream {
	pipe := newPipeline(conn)
	s := &stream{
		pipe:     pipe,
		reqCh:    make(chan struct{}, maxConcurrency),
		eventHub: &eventHub{chs: make(map[int64]chan *resp)},
	}
	go s.ioloop()
	return s
}

func (s *stream) nextCorrelationId() int64 {
	return atomic.AddInt64(&s.correlationId, 1)
}

func splitServiceNameandMethod(method string) (string, string, error) {
	seqs := strings.SplitN(method, "/", 3)
	if len(seqs) != 3 {
		return "", "", fmt.Errorf("invalid method: %s", method)
	}
	return seqs[1], seqs[2], nil
}

func (s *stream) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	if s.unavailable {
		return ErrUnavailable
	}

	correlationID := s.nextCorrelationId()
	rpcOpts := newRPCOptions(opts...)

	ch, cancal := s.eventHub.register(correlationID)
	defer cancal()

	if err := s.writeRequest(ctx, correlationID, method, args, rpcOpts); err != nil {
		return err
	}

	select {
	case rp := <-ch:
		return s.processResp(rp, rpcOpts, reply)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *stream) processResp(rp *resp, rpcOpts *rpcOptions, reply interface{}) error {
	if rp.err != nil {
		return rp.err
	}

	rpcResp := rp.frame.RPCMeta.Response
	if rpcResp == nil {
		return fmt.Errorf("miss RPCMeta Response")
	}

	if rpcResp.GetErrorCode() != 0 {
		return &RPCError{Code: rpcResp.GetErrorCode(), Message: rpcResp.GetErrorText()}
	}

	attachmentSize := rp.frame.AttachmentSize()
	offset := len(rp.frame.Body) - attachmentSize
	if err := proto.Unmarshal(rp.frame.Body[:offset], leagcyproto.MessageV2(reply)); err != nil {
		return nil
	}
	if rpcOpts.AttachmentWriter != nil {
		rpcOpts.AttachmentWriter.Write(rp.frame.Body[offset:])
	}
	return nil
}

func (s *stream) writeRequest(ctx context.Context, correlationID int64, method string, args interface{}, rpcOpts *rpcOptions) error {
	serviceName, methodName, err := splitServiceNameandMethod(method)
	if err != nil {
		return err
	}
	reqMeta := &policypb.RpcRequestMeta{
		ServiceName:  proto.String(serviceName),
		MethodName:   proto.String(methodName),
		LogId:        rpcOpts.LogID,
		TraceId:      rpcOpts.TraceID,
		SpanId:       rpcOpts.SpanID,
		ParentSpanId: rpcOpts.ParentSpanID,
	}
	f := newReqFrame(correlationID, reqMeta)
	if err = f.AppendMessage(leagcyproto.MessageV2(args)); err != nil {
		return err
	}
	f.AppendAttachment(rpcOpts.Attachment)

	select {
	case s.reqCh <- struct{}{}:
	default:
		return ErrReachMaxConcurrency
	}
	return s.pipe.SendRequest(ctx, f)
}

func (s *stream) ioloop() {
	for range s.reqCh {
		if s.unavailable {
			s.eventHub.emitError(ErrUnavailable)
		}

		f, err := s.pipe.ReceiveResponse(context.Background())
		if err != nil {
			s.eventHub.emitError(err)
			continue
		}
		s.eventHub.emitResp(f.RPCMeta.GetCorrelationId(), f, nil)
	}
}
