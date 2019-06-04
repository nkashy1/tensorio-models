GOPATH:=$(shell go env GOPATH)
GRPC_LIST:=$(shell go list -m -f "{{.Dir}}" github.com/grpc-ecosystem/grpc-gateway)
GRPC_GATEWAY_PROTO_DIR:="${GRPC_LIST}/third_party/googleapis"
TIMESTAMP:=$(shell date -u +%s)
RUN_ARGS=-backend memory
CWD:=$(shell pwd)
GRPC_PYTHON_PLUGIN_PATH=$(shell which grpc_python_plugin)
PYTHON_CLIENT_PATH="clients/python/tensorio_models"

default: fmt build

fmt:
	gofmt -w -s .

docker-models:
	docker build -t docai/tensorio-models -f dockerfiles/Dockerfile.repository .

docker-flea:
	docker build -t docai/tensorio-flea -f dockerfiles/Dockerfile.flea .

run-models: docker-models
	docker run -v $(CWD)/common/fixtures/AuthTokens.txt:/tmp/AuthTokens.txt -e AUTH_TOKENS_FILE=/tmp/AuthTokens.txt -p 8080:8080 -p 8081:8081 docai/tensorio-models ${RUN_ARGS}

run-flea: docker-flea
	docker run -v $(CWD)/common/fixtures/AuthTokens.txt:/tmp/AuthTokens.txt -e MODELS_URI=http://example.com:8081/v1/repository -e AUTH_TOKENS_FILE=/tmp/AuthTokens.txt -p 8082:8082 -p 8083:8083 docai/tensorio-flea ${RUN_ARGS}

api/repository.pb.go: api/repository.proto
	cd api && protoc -I . repository.proto --go_out=plugins=grpc:. --proto_path=${GOPATH}/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

api/repository.pb.gw.go: api/repository.proto
	cd api && protoc -I . repository.proto --grpc-gateway_out=logtostderr=true:. --proto_path=$(GOPATH)/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

api/repository.swagger.json: api/repository.proto
	cd api && protoc -I . repository.proto --swagger_out=logtostderr=true:. --proto_path=$(GOPATH)/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

api/flea.pb.go: api/flea.proto api/repository.proto
	cd api && protoc -I . flea.proto --go_out=plugins=grpc:. --proto_path=${GOPATH}/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

api/flea.pb.gw.go: api/flea.proto api/repository.proto
	cd api && protoc -I . flea.proto --grpc-gateway_out=logtostderr=true:. --proto_path=$(GOPATH)/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

api/flea.swagger.json: api/flea.proto  api/repository.proto
	cd api && protoc -I . flea.proto --swagger_out=logtostderr=true:. --proto_path=$(GOPATH)/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR)

grpc-stub-python:
	protoc --plugin=protoc-gen-grpc_python=$(GRPC_PYTHON_PLUGIN_PATH) -I api/ --proto_path=${GOPATH}/src --proto_path=$(GOPATH)/pkg/mod --proto_path=$(GRPC_GATEWAY_PROTO_DIR) api/*.proto --python_out $(PYTHON_CLIENT_PATH) --grpc_python_out $(PYTHON_CLIENT_PATH)
	find $(PYTHON_CLIENT_PATH) -type f -name "*.py" | xargs sed -i.bak 's/^import repository_pb2/from \. import repository_pb2/g'
	find $(PYTHON_CLIENT_PATH) -type f -name "*.py" | xargs sed -i.bak 's/^import flea_pb2/from \. import flea_pb2/g'

build: api/repository.pb.go api/repository.pb.gw.go api/repository.swagger.json api/flea.pb.go api/flea.pb.gw.go api/flea.swagger.json
	go test ./... -cover
	go build ./...

test-helm:
	./test-helm.sh

coverage: api/repository.pb.go api/repository.pb.gw.go api/repository.swagger.json api/flea.pb.go api/flea.pb.gw.go api/flea.swagger.json
	go test -coverprofile=test.out ./...
	go tool cover -html=test.out -o coverage-$(TIMESTAMP).html
	echo "Coverage report: coverage-$(TIMESTAMP).html"

coverage-cleanup:
	rm -f test.out coverage-*.html

clean: coverage-cleanup
	rm -f api/repository.pb.go api/repository.pb.gw.go api/repository.swagger.json api/flea.pb.go api/flea.pb.gw.go api/flea.swagger.json
