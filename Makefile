GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOTOOL=$(GOCMD) tool


all: build

test:
	GOCACHE=off $(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -rf $(BIN_DIR)

fmt:
	$(GOFMT) ./...

cover:
	$(GOTEST) -coverprofile cp.out ./...
	$(GOTOOL) cover -html=cp.out

proto:
	for f in pkg/**/proto/*.proto; do \
		protoc --go_out=plugins=grpc:. $$f; \
		echo compiled: $$f; \
	done

build:
	$(GOBUILD) ./cmd/...

