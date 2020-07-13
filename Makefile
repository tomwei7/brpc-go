BRPC_PROTOS=$(shell find proto -name "*.proto" -type f)
genproto: $(BRPC_PROTOS)
	@mkdir -p genproto
	protoc -Iproto --go_opt=paths=source_relative --go_out=genproto $?
	
.PHONY: genproto
