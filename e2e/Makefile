include ../Makefile.versions

PROJECT_DIR := $(CURDIR)/../

# CLI tools are managed by aqua. See aqua.yaml.
# They are executed via `aqua exec` with the versions in aqua.yaml, so PATH does not need to be configured.
KIND := aqua exec -- kind
KUBECTL := aqua exec -- kubectl
KUSTOMIZE := aqua exec -- kustomize

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
setup:
	$(MAKE) -C $(PROJECT_DIR) setup

.PHONY: start
start:
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME) --config=$(KIND_CONFIG) --image=kindest/node:$(KIND_NODE_VERSION) --wait 1m
	$(MAKE) -C ../ docker-build
	$(KIND) load docker-image necoperf-daemon:dev --name=$(KIND_CLUSTER_NAME)
	$(KIND) load docker-image necoperf-cli:dev --name=$(KIND_CLUSTER_NAME)
	$(KUSTOMIZE) build ../config/namespace | $(KUBECTL) apply -f -
	$(KUSTOMIZE) build ../config/rbac | $(KUBECTL) apply -f -

# `aqua which` only resolves the path and does not install the package, so kubectl is executed once to install it.
.PHONY: test
test:
	$(KUBECTL) version --client > /dev/null
	env RUN_E2E=1 KUBECTL="$$(aqua which kubectl)" \
		go test -v -race . -ginkgo.v -ginkgo.fail-fast

.PHONY: stop
stop:
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)
	-docker image rm necoperf-daemon:dev
	-docker image rm necoperf-cli:dev
	-docker image prune -f

.PHONY: clean
clean:
	rm -rf $(PROJECT_DIR)/bin

.PHONY: show-log
show-log:
	$(KUBECTL) logs -n necoperf daemonset/necoperf-daemon
