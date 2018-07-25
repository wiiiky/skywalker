ENV=GOPATH=`pwd`

all: get skywalker forctl 

skywalker: src/skywalker/skywalker.go
	$(ENV) go build -o bin/skywalker $?

forctl: src/forctl/forctl.go
	$(ENV) go build -o bin/forctl $?

install: all
	mv bin/skywalker /usr/local/bin/
	mv bin/forctl /usr/local/bin/
	mv -n example/ss.yml /etc/skywalker.yml
	mv -n script/systemd/skywalker.service /etc/systemd/system

get:
	$(ENV) go get --fix skywalker
	$(ENV) go get --fix forctl

fmt:
	find ./src/forctl ./src/skywalker -name '*.go'|xargs gofmt -w

proto: proto-gen
	PATH="$(PATH):./bin" $(ENV) protoc --go_out=./ ./src/skywalker/rpc/*.proto
	PATH="$(PATH):./bin" $(ENV) protoc --go_out=./ ./src/skywalker/agent/walker/*.proto

proto-gen:
	$(ENV) go get github.com/golang/protobuf/protoc-gen-go

clean: bin/skywalker bin/forctl 
	rm -rf $?
