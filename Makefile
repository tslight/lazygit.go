.DEFAULT_GOAL := all
VERSION := $(shell git describe --tags --abbrev=0)

# https://www.forkingbytes.com/blog/dynamic-versioning-your-go-application/
FLAGS := "-ldflags=-s -w -X main.Version=$(VERSION)"

# Not used here, but this is fascinating:
# https://stackoverflow.com/a/12110773/11133327
OPERATING_SYSTEMS = darwin linux windows freebsd openbsd
$(OPERATING_SYSTEMS):
	GOARCH=$(ARCH) GOOS=$(@) go build $(FLAGS) -o ./$(CMD)-$(@)-$(ARCH) ./cmd/$(CMD)

ARCHITECTURES = amd64 arm64
$(ARCHITECTURES):; @CMD=$(CMD) ARCH=$(@) $(MAKE) $(OPERATING_SYSTEMS)

CMDS = gitlab github
$(CMDS):; @CMD=$(@) $(MAKE) -j $(ARCHITECTURES)

all: $(CMDS)

clean:; @rm -fv ./git*-*-*

test:
	@go vet ./...
	@go test ./... -covermode=count -coverprofile=c.out
	@go tool cover -func=c.out

install:
	go build $(FLAGS) -o ./gitlab ./cmd/gitlab
	go build $(FLAGS) -o ./github ./cmd/github
	mkdir -p $(GOPATH)/bin
	install -m 0755 ./gitlab $(GOPATH)/bin/gitlab
	install -m 0755 ./github $(GOPATH)/bin/github
