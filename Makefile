# If the USE_SUDO_FOR_DOCKER env var is set, prefix docker commands with 'sudo'
ifdef USE_SUDO_FOR_DOCKER
	SUDO_CMD = sudo
endif

IMAGE ?= quay.io/oracle-osb/servicebroker
TAG ?= $(shell git describe --tags --always)
PULL ?= IfNotPresent

build: ## Builds the oracle service broker
	go build -i github.com/bpradipt/oracle-osb/cmd/servicebroker

test: ## Runs the tests
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

linux: ## Builds a Linux executable
	GOOS=linux CGO_ENABLED=1 \
	go build -o servicebroker-linux --ldflags="-s" github.com/bpradipt/oracle-osb/cmd/servicebroker

image: linux ## Builds a Linux based image
	cp servicebroker-linux image/servicebroker
	$(SUDO_CMD) docker build image/ -t "$(IMAGE):$(TAG)"

clean: ## Cleans up build artifacts
	rm -f servicebroker
	rm -f servicebroker-linux
	rm -f image/servicebroker

push: image ## Pushes the image to dockerhub, REQUIRES SPECIAL PERMISSION
	$(SUDO_CMD) docker push "$(IMAGE):$(TAG)"

deploy-helm: image ## Deploys image with helm
	helm upgrade --tls --install oracle-osb --namespace oracle-broker \
	charts/servicebroker \
	--set image="$(IMAGE):$(TAG)",imagePullPolicy="$(PULL)"

create-ns: ## Cleans up the namespaces
	kubectl create ns test-ns

provision: create-ns ## Provisions a service instance
	kubectl apply -f manifests/service-instance.yaml

bind: ## Creates a binding
	kubectl apply -f manifests/service-binding.yaml

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: build test linux image clean push deploy-helm create-ns provision bind help
