BINARY_NAME=sentry-nomad

build:
	go mod download && CGO_ENABLED=0 go build -o ${BINARY_NAME}

run: build
	./${BINARY_NAME}

build-docker:
	docker build -t sentry-nomad .

clean:
	go clean
	rm -f ${BINARY_NAME}
