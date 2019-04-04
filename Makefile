GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOTOOL=$(GOCMD) tool

IMAGE := andreymgn/rsoi-user

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

dep:
	dep ensure --vendor-only

build: fmt dep
	$(GOBUILD) ./cmd/...

build-scratch: fmt dep
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o RSOI-user ./cmd/...

image: build-scratch
	docker build -t $(IMAGE) .

push-image:
	docker push $(IMAGE)
