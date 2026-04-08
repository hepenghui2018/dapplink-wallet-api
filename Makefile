GITCOMMIT := $(shell git rev-parse HEAD)
GITDATE := $(shell git show -s --format='%ct')

LDFLAGSSTRING +=-X main.GitCommit=$(GITCOMMIT)
LDFLAGSSTRING +=-X main.GitDate=$(GITDATE)
LDFLAGS := -ldflags "$(LDFLAGSSTRING)"


wallet-api:
	env GO111MODULE=on go build -v $(LDFLAGS) ./cmd/wallet-api

clean:
	rm wallet-api

protoc:
	sh ./bin/compile.sh

test:
	go test -v ./...

lint:
	golangci-lint run ./...


.PHONY: \
	wallet-api \
	clean \
	protoc \
	test \
	lint
