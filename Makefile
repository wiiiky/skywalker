ENV=GOPATH=`pwd`

all: get skywalker forctl 

skywalker: src/skywalker/skywalker.go
	$(ENV) go build -o bin/skywalker $?

forctl: src/forctl/forctl.go
	$(ENV) go build -o bin/forctl $?

get:
	$(ENV) go get --fix skywalker
	$(ENV) go get --fix forctl

fmt:
	find ./src/forctl ./src/skywalker -name '*.go'|xargs gofmt -w

proto: proto-gen
	PATH="$(PATH):./bin" $(ENV) protoc --go_out=./ ./src/skywalker/rpc/*.proto

proto-gen:
	GOPATH=`pwd` go get github.com/golang/protobuf/protoc-gen-go

clean: bin/skywalker bin/forctl 
	rm -rf $?
