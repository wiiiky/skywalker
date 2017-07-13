ENV=GOPATH=`pwd`

all: get skywalker forctl forgtl

skywalker: src/skywalker/skywalker.go
	$(ENV) go build -o bin/skywalker $?

forctl: src/forctl/forctl.go
	$(ENV) go build -o bin/forctl $?

forgtl: src/forgtl/forgtl.go
	$(ENV) go build -o bin/forgtl $?

get:
	$(ENV) go get --fix skywalker
	$(ENV) go get --fix forctl
	$(ENV) go get --fix forgtl

fmt:
	find ./src/forctl ./src/skywalker ./src/forgtl -name '*.go'|xargs gofmt -w

proto: proto-gen
	PATH="$(PATH):./bin" $(ENV) protoc --go_out=./ ./src/skywalker/rpc/*.proto

proto-gen:
	GOPATH=`pwd` go get github.com/golang/protobuf/protoc-gen-go

clean: bin/skywalker bin/forctl bin/forgtl
	rm -rf $?
