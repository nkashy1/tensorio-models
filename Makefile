default: fmt generate docker-build

fmt:
	gofmt -w -s .

docker-build:
	docker build -t tensorio/repository -f dockerfiles/Dockerfile.repository .

run: docker-build
	docker run -p 8080:8080 -p 8081:8081 tensorio/repository

generate-clean:
	rm api/*pb* ; true

generate: generate-clean
	go generate ./...
