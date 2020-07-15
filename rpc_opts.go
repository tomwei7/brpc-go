package brpc

import (
	"io"

	"google.golang.org/grpc"
)

type rpcOptions struct {
	Attachment       []byte
	LogID            *int64
	TraceID          *int64
	SpanID           *int64
	ParentSpanID     *int64
	AttachmentWriter io.Writer
}

type rpcOption interface {
	apply(rpcOpts *rpcOptions)
}

var _ rpcOption = &attachmentOption{}

type attachmentOption struct {
	grpc.EmptyCallOption
	Attachment []byte
}

func (a *attachmentOption) apply(rpcOpts *rpcOptions) {
	rpcOpts.Attachment = a.Attachment
}

func WithAttachment(attachment []byte) grpc.CallOption {
	return &attachmentOption{Attachment: attachment}
}

var _ rpcOption = &logIDOption{}

type logIDOption struct {
	grpc.EmptyCallOption
	LogID int64
}

func (l *logIDOption) apply(rpcOpts *rpcOptions) {
	rpcOpts.LogID = &l.LogID
}

func WithLogID(logID int64) grpc.CallOption {
	return &logIDOption{LogID: logID}
}

var _ rpcOption = &traceOption{}

type traceOption struct {
	grpc.EmptyCallOption
	TraceID      int64
	SpanID       int64
	ParentSpanID int64
}

func (t *traceOption) apply(rpcOpts *rpcOptions) {
	rpcOpts.TraceID = &t.TraceID
	rpcOpts.SpanID = &t.SpanID
	rpcOpts.ParentSpanID = &t.ParentSpanID
}

func WithTrace(traceID, spanID, parentSpanID int64) grpc.CallOption {
	return &traceOption{TraceID: traceID, SpanID: spanID, ParentSpanID: parentSpanID}
}

func newRPCOptions(opts ...grpc.CallOption) *rpcOptions {
	rpcOpts := new(rpcOptions)
	for _, opt := range opts {
		rpcOpt, ok := opt.(rpcOption)
		if ok {
			rpcOpt.apply(rpcOpts)
		}
	}
	return rpcOpts
}

type receiveAttachment struct {
	grpc.EmptyCallOption
	AttachmentWriter io.Writer
}

func (t *receiveAttachment) apply(rpcOpts *rpcOptions) {
	rpcOpts.AttachmentWriter = t.AttachmentWriter
}

func ReceiveAttachmentTo(w io.Writer) grpc.CallOption {
	return &receiveAttachment{AttachmentWriter: w}
}
