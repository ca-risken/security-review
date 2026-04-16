TAG ?= latest
DEADCODE_VERSION ?= v0.43.0
GOLANGCI_LINT_VERSION ?= v2.8.0
GO_BIN_DIR ?= $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)

.PHONY: all
all: help

.PHONY: help
help:
	@echo "You can use sub-command, Usage: make <sub-command>"
	@echo "\n---------------- sub-command list ----------------"
	@cat Makefile | grep -e '^\.PHONY: .*$$' | grep -v -e "all" -e "help" | sed -e 's/^\.PHONY: //g' | sed -e 's/^/- /g' | sort

.PHONY: install
install:
	GOBIN=$(GO_BIN_DIR) go install golang.org/x/tools/cmd/deadcode@$(DEADCODE_VERSION)
	GOBIN=$(GO_BIN_DIR) go install github.com/vektra/mockery/v2@v2.36.0
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_BIN_DIR) $(GOLANGCI_LINT_VERSION)

.PHONY: generate-mock
generate-mock:
	# for dir in $$(ls -d pkg/**); do \
	# 	pushd $$dir && mockery --all && popd; \
	# done
	cd pkg && mockery --all

.PHONY: lint
lint:
	go run golang.org/x/tools/cmd/deadcode@$(DEADCODE_VERSION) ./...
	GO111MODULE=on GOFLAGS=-buildvcs=false golangci-lint run --timeout 5m

.PHONY: build
build:
	docker build -t ssgca/risken-review:$(TAG) .

.PHONY: sh
sh: build
	docker run -it --rm -v $(CURDIR):/tmp/workspace --entrypoint /bin/sh ssgca/risken-review:$(TAG)

.PHONY: run
run: build
	docker run \
		--rm \
		--env-file=.env \
		-v $(CURDIR):/tmp/workspace \
		ssgca/risken-review:$(TAG)

.PHONY: run-options
run-options: build
	docker run \
		--rm \
		--env-file=.env \
		-v $(CURDIR):/tmp/workspace \
		ssgca/risken-review:$(TAG) --error --no-pr-comment

.PHONY: login
login:
	docker login

.PHONY: start-buildx
start-buildx:
	docker buildx create --name mybuilder --use
	docker buildx inspect --bootstrap

# start-buildxが実行されていること
.PHONY: push
push:
	docker buildx build --platform linux/amd64,linux/arm64 -t ssgca/security-review:$(TAG) . --push
