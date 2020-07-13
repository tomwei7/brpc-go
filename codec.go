// Notes:
// 1. 12-byte header [PRPC][body_size][meta_size]
// 2. body_size and meta_size are in network byte order
// 3. Use service->full_name() + method_name to specify the method to call
// 4. `attachment_size' is set iff request/response has attachment
// 5. Not supported: chunk_info
package brpc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	brpcpb "github.com/tomwei7/brpc-go/genproto/brpc"
)

var (
	magicHead = []byte("PRPC")
)

var (
	ErrInvalidHead           = errors.New("Invalid RPC head")
	ErrCompressTypeUnsupport = errors.New("CompressType is unsupport yet")
)

type encodec struct {
	bw *bufio.Writer
}

func newEncodec(w io.Writer) *encodec {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	return &encodec{bw: bw}
}

func (ec *encodec) Encode(f *frame) error {
	meta, err := proto.Marshal(&f.RPCMeta)
	if err != nil {
		return err
	}
	metaSize := uint32(len(meta))

	if err = compressBody(f); err != nil {
		return err
	}
	bodySize := metaSize + uint32(len(f.Body))

	var rpcHead [12]byte
	copy(rpcHead[:], magicHead)
	binary.BigEndian.PutUint32(rpcHead[len(magicHead):], bodySize)
	binary.BigEndian.PutUint32(rpcHead[len(magicHead)+4:], metaSize)

	for _, b := range [][]byte{rpcHead[:], meta, f.Body} {
		_, err := ec.bw.Write(b)
		if err != nil {
			return err
		}
	}
	return ec.bw.Flush()
}

type decodec struct {
	br *bufio.Reader
}

func newDecodec(r io.Reader) *decodec {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &decodec{br: br}
}

func (d *decodec) Decode(f *frame) error {
	var rpcHead [12]byte
	_, err := io.ReadFull(d.br, rpcHead[:])
	if err != nil {
		return err
	}

	if !bytes.Equal(rpcHead[:len(magicHead)], magicHead) {
		return fmt.Errorf("%w receive: %v", ErrInvalidHead, rpcHead[:len(magicHead)])
	}

	bodySize := binary.BigEndian.Uint32(rpcHead[len(magicHead):])
	metaSize := binary.BigEndian.Uint32(rpcHead[len(magicHead)+4:])
	if bodySize <= metaSize {
		return fmt.Errorf("%w body size less than meta size %d < %d", ErrInvalidHead, bodySize, metaSize)
	}

	// TODO: limit max body size.
	body := make([]byte, bodySize)
	if _, err = io.ReadFull(d.br, body); err != nil {
		return err
	}

	if err = proto.Unmarshal(body[:metaSize], &f.RPCMeta); err != nil {
		return err
	}
	return unCompressBody(f, body[metaSize:])
}

func unCompressBody(f *frame, payload []byte) error {
	var uncompressBody []byte
	compressType := f.CompressType()
	switch compressType {
	case brpcpb.CompressType_COMPRESS_TYPE_NONE:
		uncompressBody = payload
	default:
		return fmt.Errorf("%w, compress type: %s", ErrCompressTypeUnsupport, compressType)
	}
	f.Body = uncompressBody
	return nil
}

func compressBody(f *frame) error {
	var compressBody []byte
	compressType := f.CompressType()
	switch compressType {
	case brpcpb.CompressType_COMPRESS_TYPE_NONE:
		compressBody = f.Body
	default:
		return fmt.Errorf("%w, compress type: %s", ErrCompressTypeUnsupport, compressType)
	}
	f.Body = compressBody
	return nil
}
