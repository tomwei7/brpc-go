package brpc

import (
	"fmt"
)

type RPCError struct {
	Code    int32
	Message string
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPCError Code: %d Message: %s", e.Code, e.Message)
}
