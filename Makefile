.PHONY: all
all: help

.PHONY: help
help:
	@echo "You can use sub-command, Usage: make <sub-command>"
	@echo "\n---------------- sub-command list ----------------"
	@cat Makefile | grep -e '^\.PHONY: .*$$' | grep -v -e "all" -e "help" | sed -e 's/^\.PHONY: //g' | sed -e 's/^/- /g' | sort

.PHONY: build
build:
	docker build -t risken-review:latest .

.PHONY: sh
sh: build
	docker run -it --rm -v $(CURDIR):/tmp/code --entrypoint /bin/sh risken-review:latest

.PHONY: run
run: build
	docker run \
		--rm \
		--env-file=.env \
		-v $(CURDIR):/tmp/code \
		risken-review:latest
