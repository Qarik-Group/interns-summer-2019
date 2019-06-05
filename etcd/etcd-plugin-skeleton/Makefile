all: test build

init:
	go get "github.com/jhunt/go-ansi"
	go get github.com/starkandwayne/shield/plugin

build: 
	go build -o etcdPluginBase -v

test: 
	go test -v ./...

clean: 
	go clean
	rm -f etcdPluginBase

run:
	go build -o etcdPluginBase -v ./...
	./etcdPluginBaseß --help

