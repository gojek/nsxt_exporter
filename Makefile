ENVVAR=CGO_ENABLED=0 GO111MODULE=on
GOOS?=linux
PKGS=./...

GO?=go
GOFMT?=$(GO) fmt

ifdef LDFLAGS
  LDFLAGS_FLAG=--ldflags "${LDFLAGS}"
else
  LDFLAGS_FLAG=
endif

.PHONY: build
build:
	$(ENVVAR) GOOS=$(GOOS) go build -o nsxt-exporter ${LDFLAGS_FLAG}

.PHONY: fmt
fmt:
	$(GOFMT) $(PKGS)
