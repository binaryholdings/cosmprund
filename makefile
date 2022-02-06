VERSION := $(shell git describe --tags)
COMMIT  := $(shell git log -1 --format='%H')

all: install

LD_FLAGS = -X github.com/strangelove-ventures/lens/cmd.Version=$(VERSION) \
	-X github.com/strangelove-ventures/lens/cmd.Commit=$(COMMIT) \

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

build:
	@echo "Building Pruning"
	@go build -mod readonly $(BUILD_FLAGS) -o build/cosmprund main.go

install:
	@echo "Installing Lens"
	@go install -mod readonly $(BUILD_FLAGS) ./...

clean:
	rm -rf build

.PHONY: all lint test race msan tools clean build
