
all: get skywalker

skywalker:
	GOPATH=`pwd` go build -o bin/skywalker src/skywalker/skywalker.go

get:
	GOPATH=`pwd` go get --fix skywalker

fmt:
	find . -name '*.go'|xargs gofmt -w

clean:
	rm -rf bin/skywalker
