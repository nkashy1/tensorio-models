#!/bin/sh +e
GRPC_GATEWAY_PROTO_DIR="$(go list -m -f "{{.Dir}}" github.com/grpc-ecosystem/grpc-gateway)/third_party/googleapis"

protoc -I . repository.proto --go_out=plugins=grpc:. --proto_path=$GOPATH/src --proto_path=$GOPATH/pkg/mod --proto_path=$GRPC_GATEWAY_PROTO_DIR
protoc -I . repository.proto --grpc-gateway_out=logtostderr=true:. --proto_path=$GOPATH/src --proto_path=$GOPATH/pkg/mod --proto_path=$GRPC_GATEWAY_PROTO_DIR
protoc -I . repository.proto --swagger_out=logtostderr=true:. --proto_path=$GOPATH/src --proto_path=$GOPATH/pkg/mod --proto_path=$GRPC_GATEWAY_PROTO_DIR
