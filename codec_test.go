package brpc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	policypb "github.com/tomwei7/brpc-go/genproto/brpc/policy"
)

func TestCodece(t *testing.T) {
	body := []byte("ALL IS WELL")
	input := &frame{
		RPCMeta: policypb.RpcMeta{
			CorrelationId:  proto.Int64(42),
			AttachmentSize: proto.Int32(int32(len(body))),
		},
		Body: body,
	}

	buf := &bytes.Buffer{}

	err := newEncodec(buf).Encode(input)
	if assert.Nil(t, err) {
		ouput := &frame{}
		err = newDecodec(buf).Decode(ouput)
		if assert.Nil(t, err) {
			assert.True(t, proto.Equal(&input.RPCMeta, &ouput.RPCMeta))
		}
	}
}
