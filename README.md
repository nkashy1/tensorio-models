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

## Running server

The server can be run with `make run`, but there is a caveat -- `make run` requires a
`RUN_ARGS` argument.

For example:
```
RUN_ARGS="-backend memory" make run
```

Without this argument, `make run` fails -- simply because the `repository` binary has required
arguments.


### Running server against GCS for testing:

First, make sure you have service account credentials available locally for a service account that
has access to the GCS bucket(s) you would like to use in your tests. Then, expose the JSON file
containing those service account credentials in your shells as follows:
```
export GOOGLE_APPLICATION_CREDENTIALS=$HOME/secrets/<creds-file>.json
```

Run the tensorio-models backend:
```
make docker-models

docker run \
    -v $GOOGLE_APPLICATION_CREDENTIALS:/etc/sacred.json \
    -e GOOGLE_APPLICATION_CREDENTIALS=/etc/sacred.json \
    -e REPOSITORY_GCS_BUCKET=tensorio-models-backend-dev \
    -p 8080:8080 \
    -p 8081:8081 \
    docai/tensorio-models \
    -backend gcs
```

Run the flea backend:
```
make docker-flea

# Check your environmemt PRIVATE_PEM_KEY by doing:
go test github.com/doc-ai/tensorio-models/signed_url -test.v -count=1

docker run \
    -v $GOOGLE_APPLICATION_CREDENTIALS:/etc/sacred.json \
    -e GOOGLE_APPLICATION_CREDENTIALS=/etc/sacred.json \
    -e FLEA_GCS_BUCKET=tensorio-models-backend-dev \
    -e FLEA_UPLOAD_GCS_BUCKET=tensorio-models-backend-dev \
    -e GOOGLE_ACCESS_ID="$GOOGLE_ACCESS_ID" \
    -e PRIVATE_PEM_KEY="$PRIVATE_PEM_KEY" \
    -e MODELS_URI=localhost:8081/v1/repository \
    -p 8082:8082 \
    -p 8083:8083 \
    docai/tensorio-flea \
    -backend gcs
```

Run the tests in the [`e2e/`](./e2e/) directory:

```
# Tests tensorio-models backend
./e2e/setup.sh
```

```
# Tests flea backend
./e2e/create-sample-tasks.sh
```
