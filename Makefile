BINARY_NAME=sentry-nomad

build:
	go mod download && CGO_ENABLED=0 go build -o ${BINARY_NAME}
.PHONY: build

run: build
	./${BINARY_NAME}
.PHONY: run

build-docker:
	docker build -t sentry-nomad .
.PHONY: build-docker

clean:
	go clean
	rm -f ${BINARY_NAME}
.PHONY: clean

# Runs nomad locally in dev mode
nomad-run:
	sudo nomad agent -dev -bind 127.0.0.1 -log-level INFO
.PHONY: nomad-run

lint:
	go mod tidy
	go vet
	go fmt
.PHONY: lint
