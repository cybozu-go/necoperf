include ../Makefile.versions

ARCH ?= amd64
OS ?= linux

PROJECT_DIR := $(CURDIR)/../
BIN_DIR := $(PROJECT_DIR)/bin

CURL := curl -sSLf
KUBECTL := $(BIN_DIR)/kubectl
KUSTOMIZE := $(BIN_DIR)/kustomize

KIND := $(BIN_DIR)/kind
KIND_CLUSTER_NAME := necoperf
KIND_CONFIG := kind-config.yaml

export KUBECONFIG

.PHONY: help
help:
	@echo "Choose one of the following target"
	@echo
	@echo "setup      Setup tools"
	@echo "start      Start kind cluster and install accurate"
	@echo "test       Run e2e tests"
	@echo "logs       Save logs as logs.tar.gz"
	@echo "stop       Stop the kind cluster"

.PHONY: setup
setup: kubectl kustomize kind

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --config=$(KIND_CONFIG) --image=kindest/node:v$(E2ETEST_K8S_VERSION) --wait 1m
	$(MAKE) -C ../ docker-build
	$(KIND) load docker-image necoperf-daemon:dev --name=$(KIND_CLUSTER_NAME)
	$(KIND) load docker-image necoperf-cli:dev --name=$(KIND_CLUSTER_NAME)
	$(KUSTOMIZE) build ../config/namespace | $(KUBECTL) apply -f -
	$(KUSTOMIZE) build ../config/rbac | $(KUBECTL) apply -f -

.PHONY: test
test:
	env RUN_E2E=1 \
		go test -v -race . -ginkgo.v -ginkgo.fail-fast

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)
	-docker image rm necoperf-daemon:dev
	-docker image rm necoperf-cli:dev
	-docker image prune -f

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(KUSTOMIZE)-$(KUSTOMIZE_VERSION)
	ln -sf $(notdir $<) $@

$(KUSTOMIZE)-$(KUSTOMIZE_VERSION):
	mkdir -p $(dir $@)
	curl -fsL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv$(KUSTOMIZE_VERSION)/kustomize_v$(KUSTOMIZE_VERSION)_linux_amd64.tar.gz | \
	tar -xzf - -O > $@
	chmod a+x $@

.PHONY: kind
kind: $(KIND)
$(KIND): $(KIND)-$(KIND_VERSION)
	ln -sf $(notdir $<) $@

$(KIND)-$(KIND_VERSION):
	mkdir -p $(dir $@)
	$(CURL) -o $@ https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-$(OS)-$(ARCH)
	chmod a+x $@

.PHONY: kubectl
kubectl: $(KUBECTL)
$(KUBECTL): $(KUBECTL)-$(E2ETEST_K8S_VERSION)
	ln -sf $(notdir $<) $@

$(KUBECTL)-$(E2ETEST_K8S_VERSION):
	mkdir -p $(dir $@)
	$(CURL) -o $@ https://storage.googleapis.com/kubernetes-release/release/v$(E2ETEST_K8S_VERSION)/bin/$(OS)/$(ARCH)/kubectl
	chmod a+x $@

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
