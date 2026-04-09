include Makefile.versions

BIN_DIR := $(PWD)/bin
ENVTEST := $(BIN_DIR)/setup-envtest
MDBOOK := $(BIN_DIR)/mdbook
STATICCHECK := $(BIN_DIR)/staticcheck

PROTOC := $(BIN_DIR)/protoc -I=$(PWD)/include:.
PROTOC_OUTPUTS = internal/rpc/necoperf.pb.go internal/rpc/necoperf_grpc.pb.go docs/necoperf-grpc.md

.PHONY: all
all: build

.PHONY: check-generated
check-generated:
	go mod tidy
	go fmt ./...
	$(MAKE) generate
	$(MAKE) book
	git diff --exit-code

.PHONY: book
book: $(MDBOOK)
	rm -rf docs/book
	cd docs; $(MDBOOK) build

.PHONY: lint
lint:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	$(STATICCHECK) ./...
	go vet ./...

.PHONY: test
test:
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(BIN_DIR) -p path)" go test ./... -coverprofile cover.out -v

.PHONY: build
build:
	mkdir -p bin
	GOBIN=$(BIN_DIR) go install ./cmd/...

.PHONY: generate
generate:
	$(MAKE) $(PROTOC_OUTPUTS)

internal/rpc/necoperf.pb.go: internal/rpc/necoperf.proto
	$(PROTOC) --go_out=module=github.com/cybozu-go/necoperf:. $<

internal/rpc/necoperf_grpc.pb.go: internal/rpc/necoperf.proto
	$(PROTOC) --go-grpc_out=module=github.com/cybozu-go/necoperf:. $<

docs/necoperf-grpc.md: internal/rpc/necoperf.proto
	$(PROTOC) --doc_out=docs --doc_opt=markdown,$@ $<

.PHONY: docker-build
docker-build:
	docker build -t necoperf-daemon:dev --build-arg="FLATCAR_VERSION=$(FLATCAR_VERSION)" -f Dockerfile.daemon .
	docker build -t necoperf-cli:dev -f Dockerfile.cli .

.PHONY: e2e
e2e:
	$(MAKE) -C e2e

.PHONY: run-perf
run-perf:
	uname -r
	docker run --name perf --rm --entrypoint perf ghcr.io/flatcar/flatcar-sdk-amd64:$(FLATCAR_VERSION) version
	docker run --name perf --rm --entrypoint perf --privileged ghcr.io/flatcar/flatcar-sdk-amd64:$(FLATCAR_VERSION) stat sleep 1
	docker run --name perf --rm --entrypoint perf --privileged -v $(PWD):/out ghcr.io/flatcar/flatcar-sdk-amd64:$(FLATCAR_VERSION) record -ag -F 99 --call-graph dwarf -o /out/perf.data sleep 1
	docker run --name perf --rm --entrypoint perf --privileged -v $(PWD):/out ghcr.io/flatcar/flatcar-sdk-amd64:$(FLATCAR_VERSION) script -F event -i /out/perf.data | sort -u

##@ Tools

.PHONY: setup
setup: grpctools $(ENVTEST) $(MDBOOK) $(STATICCHECK)

.PHONY: grpctools
grpctools:
	mkdir -p $(BIN_DIR)
	curl -sfL -o protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-x86_64.zip
	unzip -o protoc.zip bin/protoc 'include/*'
	rm -f protoc.zip
	GOBIN=$(BIN_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC_GEN_GO_VERSION)
	GOBIN=$(BIN_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)
	GOBIN=$(BIN_DIR) go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v$(PROTOC_GEN_DOC_VERSION)

$(ENVTEST):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_VERSION)

$(MDBOOK):
	mkdir -p $(BIN_DIR)
	curl -fsL https://github.com/rust-lang/mdBook/releases/download/v$(MDBOOK_VERSION)/mdbook-v$(MDBOOK_VERSION)-x86_64-unknown-linux-gnu.tar.gz | tar -C bin -xzf -

$(STATICCHECK):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install honnef.co/go/tools/cmd/staticcheck@v$(STATICCHECK_VERSION)
