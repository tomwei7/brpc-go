module echo

go 1.14

replace github.com/tomwei7/brpc-go => ../../

require (
	github.com/golang/protobuf v1.4.2
	github.com/tomwei7/brpc-go v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
)
