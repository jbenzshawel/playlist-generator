default: help

.PHONY: help
help: # show help for each of the Makefile command
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done


.PHONY: build
build: # build: Compile the application binary
	@go build .

.PHONY: test
test: # test: Runs all tests in the project
	@go test -race -v ./...


.PHONY: fmt
fmt: # fmt: Formats code and sorts imports
	@gosimports -w -local github.com/jbenzshawel/playlist-generator .

.PHONY: mocks
mocks: # mocks: Regenerates mocks
	@mockery && make fmt

