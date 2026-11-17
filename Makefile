include Makefile.versions

BIN_DIR := $(PWD)/bin

# CLI tools are managed by aqua. See aqua.yaml.
# They are executed via `aqua exec` with the versions in aqua.yaml, so PATH does not need to be configured.
SETUP_ENVTEST := aqua exec -- setup-envtest
MDBOOK := aqua exec -- mdbook
STATICCHECK := aqua exec -- staticcheck

# protoc looks up its plugins (protoc-gen-*) from PATH, so the aqua bin directory is added to PATH.
PROTOC := PATH="$$(aqua root-dir)/bin:$$PATH" aqua exec -- protoc -I=.
PROTOC_OUTPUTS = internal/rpc/necoperf.pb.go internal/rpc/necoperf_grpc.pb.go docs/necoperf-grpc.md

DOCKER_BUILD_ARGS := --build-arg GOPROXY=$(shell go env GOPROXY)
ifneq ($(wildcard $(HOME)/.netrc),)
# Pass Takumi Guard credentials to the build when they exist.
DOCKER_BUILD_ARGS += --secret id=netrc,src=$(HOME)/.netrc
endif

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
book:
	rm -rf docs/book
	cd docs; $(MDBOOK) build

.PHONY: lint
lint:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	$(STATICCHECK) ./...
	go vet ./...

.PHONY: test
test:
	KUBEBUILDER_ASSETS="$$($(SETUP_ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(BIN_DIR) -p path)" go test ./... -coverprofile cover.out -v

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
	docker build $(DOCKER_BUILD_ARGS) -t necoperf-daemon:dev --build-arg="FLATCAR_VERSION=$(FLATCAR_VERSION)" -f Dockerfile.daemon .
	docker build $(DOCKER_BUILD_ARGS) -t necoperf-cli:dev -f Dockerfile.cli .

.PHONY: e2e
e2e:
	$(MAKE) -C e2e

.PHONY: run-perf
run-perf:
	uname -r
	docker run --name perf --rm --entrypoint perf ghcr.io/flatcar/flatcar-sdk-amd64:$(FLATCAR_VERSION) version
	docker run --name perf --rm --entrypoint perf --privileged ghcr.io/flatcar/flatcar-sdk-amd64:$(FLATCAR_VERSION) list hw sw
	docker run --name perf --rm --entrypoint perf --privileged -v $(PWD):/out ghcr.io/flatcar/flatcar-sdk-amd64:$(FLATCAR_VERSION) record -v -e cycles -ag -F 99 --call-graph dwarf -o /out/perf.data sleep 1
	docker run --name perf --rm --entrypoint perf --privileged -v $(PWD):/out ghcr.io/flatcar/flatcar-sdk-amd64:$(FLATCAR_VERSION) script -F event -i /out/perf.data | sort -u

##@ Tools

.PHONY: setup
setup: ## Install tools with aqua.
	aqua i -l
