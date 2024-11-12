DOCKER_REGISTRY := registry.wantia.app
DOCKER_IMAGE := mwantia/valkey-snapshot
DOCKER_VERSION := v1.0.2
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: all setup test release cleanup

all: cleanup setup release

setup:
	docker buildx create --use --name multi-arch-builder || true

test:
	go run cmd/server/main.go --config test/config.yml

release: setup
	docker buildx build --push --platform ${DOCKER_PLATFORMS} -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_VERSION) -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest .

cleanup:
	rm -rf tmp/*