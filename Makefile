GOPATH:=$(shell go env GOPATH)
GRPC_LIST:=$(shell go list -m -f "{{.Dir}}" github.com/grpc-ecosystem/grpc-gateway)
GRPC_GATEWAY_PROTO_DIR:="${GRPC_LIST}/third_party/googleapis"
TIMESTAMP:=$(shell date -u +%s)

default: fmt build

fmt:
	gofmt -w -s .

docker-build:
	docker build -t tensorio/repository -f dockerfiles/Dockerfile.repository .

run: docker-build
	docker run -p 8080:8080 -p 8081:8081 tensorio/repository ${RUN_ARGS}

api/repository.pb.go: api/repository.proto
	cd api && protoc -I . repository.proto --go_out=plugins=grpc:. --proto_path=${GOPATH}/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

api/repository.pb.gw.go: api/repository.proto
	cd api && protoc -I . repository.proto --grpc-gateway_out=logtostderr=true:. --proto_path=$(GOPATH)/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

api/repository.swagger.json: api/repository.proto
	cd api && protoc -I . repository.proto --swagger_out=logtostderr=true:. --proto_path=$(GOPATH)/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

build: api/repository.pb.go api/repository.pb.gw.go api/repository.swagger.json
	go test ./... -cover
	go build ./...

coverage: api/repository.pb.go api/repository.pb.gw.go api/repository.swagger.json
	go test -coverprofile=test.out ./...
	go tool cover -html=test.out -o coverage-$(TIMESTAMP).html
	echo "Coverage report: coverage-$(TIMESTAMP).html"

coverage-cleanup:
	rm test.out coverage-*.html
