package main

import (
	"bytes"
	"context"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/tomwei7/brpc-go"

	genproto "echo/genproto"
)

func main() {
	target := os.Args[1]
	cc, err := brpc.DialContext(context.Background(), target)
	if err != nil {
		log.Fatal(err)
	}

	buf := &bytes.Buffer{}
	client := genproto.NewEchoServiceClient(cc)
	resp, err := client.Echo(context.Background(),
		&genproto.EchoRequest{Message: proto.String("hello brpc-go")},
		brpc.WithAttachment([]byte("this is attachment")),
		brpc.ReceiveAttachmentTo(buf),
	)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("resp: %s attachment: %s", resp, buf.Bytes())
	}
}
