
all: get skywalker

skywalker:
	GOPATH=`pwd` go build -o bin/skywalker src/skywalker/skywalker.go

get:
	GOPATH=`pwd` go get --fix skywalker

fmt:
	find . -name '*.go'|xargs gofmt -w

proto:
	PATH="$(PATH):./bin" GOPATH=`pwd` protoc --go_out=./ ./src/skywalker/core/message/*.proto

proto-gen:
	GOPATH=`pwd` go get -u github.com/golang/protobuf/protoc-gen-go

clean:
	rm -rf bin/skywalker
