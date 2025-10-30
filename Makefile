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

.PHONY: clean-l2
clean-l2:
	docker compose -f internal/l2/infra/docker/docker-compose.yml down -v
	docker ps -aq --filter "label=${L2_LABEL}" | xargs -r docker rm -f
	docker volume ls -q | grep -E "(rollup-a|rollup-b)" | xargs -r docker volume rm
	rm -rf ./.localnet/state ./.localnet/networks ./.localnet/compiled-contracts ./.localnet/docker-compose.yml ./.cache

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

######

### Observability ###
OBSERVABILITY_LABEL=stack=localnet-observability

.PHONY: run-observability
run-observability: build
	${BINARY_PATH} observability

.PHONY: show-observability
show-observability:
	docker ps -a --filter "label=${OBSERVABILITY_LABEL}"

.PHONY: clean-observability
clean-observability:
	docker ps -aq --filter "label=${OBSERVABILITY_LABEL}" | xargs -r docker rm -f
######
