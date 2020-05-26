ENVVAR=CGO_ENABLED=0 GO111MODULE=on
PKGS=./...

GO?=go
GOTEST?=$(GO) test
GOFMT?=$(GO) fmt
GOOS?=$(shell $(GO) env GOHOSTOS)
GOARCH?=$(shell $(GO) env GOHOSTARCH)

DOCKER_IMAGE_NAME?=nsxt-exporter
DOCKER_IMAGE_TAG?=$(shell git rev-parse --short HEAD)
DOCKERFILE_PATH?=./Dockerfile
DOCKERBUILD_CONTEXT?=./
DOCKER_REPO?=cloudnativeid
DOCKER_ARCHS?=amd64

BUILD_DOCKER_ARCHS = $(addprefix docker-,$(DOCKER_ARCHS))
PUBLISH_DOCKER_ARCHS = $(addprefix docker-publish-,$(DOCKER_ARCHS))
TAG_DOCKER_ARCHS = $(addprefix docker-tag-latest-,$(DOCKER_ARCHS))

ifdef LDFLAGS
  LDFLAGS_FLAG=--ldflags "${LDFLAGS}"
else
  LDFLAGS_FLAG=
endif

.PHONY: build
build:
	mkdir -p .build/${GOOS}-${GOARCH}
	$(ENVVAR) GOOS=$(GOOS) go build -o .build/${GOOS}-${GOARCH}/nsxt_exporter ${LDFLAGS_FLAG}

.PHONY: test
test:
	$(GOTEST) $(PKGS)

.PHONY: fmt
fmt:
	$(GOFMT) $(PKGS)

.PHONY: docker $(BUILD_DOCKER_ARCHS)
docker: $(BUILD_DOCKER_ARCHS)
$(BUILD_DOCKER_ARCHS): docker-%:
	docker build -t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" \
		-f $(DOCKERFILE_PATH) \
		--build-arg ARCH="$*" \
		--build-arg OS="linux" \
		$(DOCKERBUILD_CONTEXT)

.PHONY: docker-publish $(PUBLISH_DOCKER_ARCHS)
docker-publish: $(PUBLISH_DOCKER_ARCHS)
$(PUBLISH_DOCKER_ARCHS): docker-publish-%:
	docker push "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"

.PHONY: docker-tag-latest $(TAG_DOCKER_ARCHS)
docker-tag-latest: $(TAG_DOCKER_ARCHS)
$(TAG_DOCKER_ARCHS): docker-tag-latest-%:
	docker tag "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):latest"
