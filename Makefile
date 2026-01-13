EXEC_DIRECTORY=./cmd/localnet
BINARY_NAME=localnet
BINARY_PATH=${EXEC_DIRECTORY}/bin/${BINARY_NAME}
BINARY_DIR=${EXEC_DIRECTORY}/bin
ENCLAVE_NAME=localnet

.PHONY: default
default: run

### Go ###
.PHONY: build
build:
	go build -o ${BINARY_PATH} ${EXEC_DIRECTORY}
	@if [ -f configs/config.yaml ]; then \
		cp configs/config.yaml ${BINARY_DIR}/config.yaml; \
		echo "Copied config.yaml to ${BINARY_DIR}"; \
	else \
		echo "Warning: configs/config.yaml not found, skipping copy"; \
	fi

.PHONY: run
run: build
	${BINARY_PATH}

.PHONY: clean
clean: clean-observability clean-l2 clean-l1

.PHONY: stop
stop: stop-observability stop-l2 stop-l1
	
.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run -v ./...
######

### L1 ###
.PHONY: run-l1
run-l1: build
	${BINARY_PATH} l1

.PHONY: show-l1
show-l1:
	kurtosis enclave inspect ${ENCLAVE_NAME}

.PHONY: stop-l1
stop-l1:
	kurtosis enclave stop ${ENCLAVE_NAME} || true

.PHONY: clean-l1
clean-l1:
	kurtosis clean -a

SSV_NODE_COUNT?=4
.PHONY: restart-ssv-nodes
restart-ssv-nodes:
	@echo "Updating SSV Node services. Count: $(SSV_NODE_COUNT) ..."
	@for i in $(shell seq 0 $(shell expr $(SSV_NODE_COUNT) - 1)); do \
		echo "Updating service: ssv-node-$$i"; \
		kurtosis service update $(ENCLAVE_NAME) ssv-node-$$i; \
	done
######

### L2 ###
L2_LABEL=stack=localnet-l2

.PHONY: run-l2
run-l2: build
	${BINARY_PATH} l2

.PHONY: show-l2
show-l2:
	docker ps -a --filter "label=${L2_LABEL}"

.PHONY: stop-l2
stop-l2:
	docker compose -f .localnet/docker-compose.yml down || true

.PHONY: clean-l2
clean-l2:
	docker compose -f internal/l2/infra/docker/docker-compose.yml down -v
	docker ps -aq --filter "label=${L2_LABEL}" | xargs -r docker rm -f
	docker volume ls -q | grep -E "(rollup-a|rollup-b|blockscout|op-rbuilder)" | xargs -r docker volume rm
	rm -rf ./.localnet/state ./.localnet/networks ./.localnet/compiled-contracts ./.localnet/docker-compose.yml ./.localnet/docker-compose.blockscout.yml ./.localnet/.tmp ./.localnet/registry ./.cache

.PHONY: clean-l2-full
clean-l2-full: clean-l2
	rm -rf ./.localnet/services
	docker images -q "local/publisher" | xargs -r docker rmi -f
	docker images -q "local/op-geth" | xargs -r docker rmi -f
	docker images -q "us-docker.pkg.dev/oplabs-tools-artifacts/images/op-node" | xargs -r docker rmi -f
	docker images -q "us-docker.pkg.dev/oplabs-tools-artifacts/images/op-batcher" | xargs -r docker rmi -f
	docker images -q "us-docker.pkg.dev/oplabs-tools-artifacts/images/op-proposer" | xargs -r docker rmi -f
	docker images -q "us-docker.pkg.dev/oplabs-tools-artifacts/images/op-deployer" | xargs -r docker rmi -f

## Compile L2 contracts ##
.PHONY: run-l2-compile
run-l2-compile: build
	${BINARY_PATH} l2 compile

## Deploy L2 services for local development ##
# Usage: make run-l2-deploy SERVICE=op-geth
SERVICE?=all
.PHONY: run-l2-deploy
run-l2-deploy: build
	${BINARY_PATH} l2 deploy $(SERVICE)

######

### Observability ###
OBSERVABILITY_LABEL=stack=localnet-observability

.PHONY: run-observability
run-observability: build
	${BINARY_PATH} observability

.PHONY: show-observability
show-observability:
	docker ps -a --filter "label=${OBSERVABILITY_LABEL}"

.PHONY: stop-observability
stop-observability:
	docker ps -aq --filter "label=${OBSERVABILITY_LABEL}" | xargs -r docker stop

.PHONY: clean-observability
clean-observability:
	docker ps -aq --filter "label=${OBSERVABILITY_LABEL}" | xargs -r docker rm -f
######

### Docker ###
DOCKER_IMAGE_NAME?=compose-network/local-testnet
DOCKER_IMAGE_TAG?=latest

.PHONY: docker-build
docker-build:
	docker build -f build/Dockerfile -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .

.PHONY: docker-run-l2
docker-run-l2:
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PWD):/workspace \
		-w /workspace \
		-e HOST_PROJECT_PATH=$(PWD) \
		${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} l2 $(ARGS)
######
