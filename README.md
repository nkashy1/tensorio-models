# tensorio-models
TensorIO model repositories (backend and client libraries)

## Requirements:

The following need to be on your PATH:

* protoc
* protoc-gen-go
* protoc-gen-swagger
* protoc-gen-grpc-gateway

### Installing protoc

[Please see the documentation from proto-lens](https://google.github.io/proto-lens/installing-protoc.html)

### Installing proto generators

```sh
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go install github.com/golang/protobuf/protoc-gen-go
```
