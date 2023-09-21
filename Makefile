BIN_DIR := $(shell pwd)/bin

# Tool versions
MDBOOK_VERSION = 0.4.27
PROTOC_VERSION = 24.2
PROTOC_GEN_GO_VERSION = 1.31.0
PROTOC_GEN_GO_GRPC_VERSION = 1.3.0
PROTOC_GEN_DOC_VERSION = 1.5.1
MDBOOK := $(BIN_DIR)/mdbook

# Test tools
PROTOC := PATH=$(PWD)/bin:'$(PATH)' $(PWD)/bin/protoc -I=$(PWD)/include:.
PROTOC_OUTPUTS = pkg/rpc/necoperf.pb.go pkg/rpc/necoperf_grpc.pb.go docs/necoperf-grpc.md
STATICCHECK = $(BIN_DIR)/staticcheck

.PHONY: all
all: test

.PHONY: book
book: $(MDBOOK)
	rm -rf docs/book
	cd docs; $(MDBOOK) build


.PHONY: test
test:
	if find . -name go.mod | grep -q go.mod; then \
		$(MAKE) test-go; \
	fi

.PHONY: test-go
test-go: test-tools
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	$(STATICCHECK) ./...
	go install ./...
	go test -race -v ./...
	go vet ./...

.PHONY: generate
generate:
	$(MAKE) $(PROTOC_OUTPUTS)

pkg/rpc/necoperf.pb.go: pkg/rpc/necoperf.proto
	$(PROTOC) --go_out=module=github.com/cybozu-go/necoperf:. $<

pkg/rpc/necoperf_grpc.pb.go: pkg/rpc/necoperf.proto
	$(PROTOC) --go-grpc_out=module=github.com/cybozu-go/necoperf:. $<

docs/necoperf-grpc.md: pkg/rpc/necoperf.proto
	$(PROTOC) --doc_out=docs --doc_opt=markdown,$@ $<

##@ Tools

.PHONY: setup
setup:
	curl -sfL -o protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-x86_64.zip
	unzip -o protoc.zip bin/protoc 'include/*'
	rm -f protoc.zip
	GOBIN=$(PWD)/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC_GEN_GO_VERSION)
	GOBIN=$(PWD)/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)
	GOBIN=$(PWD)/bin go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v$(PROTOC_GEN_DOC_VERSION)

$(MDBOOK):
	mkdir -p bin
	curl -fsL https://github.com/rust-lang/mdBook/releases/download/v$(MDBOOK_VERSION)/mdbook-v$(MDBOOK_VERSION)-x86_64-unknown-linux-gnu.tar.gz | tar -C bin -xzf -

.PHONY: test-tools
test-tools: $(STATICCHECK)

$(STATICCHECK):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install honnef.co/go/tools/cmd/staticcheck@latest
