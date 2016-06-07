
all: skywalker

skywalker:
	GOPATH=`pwd` go build -o bin/luker src/skywalker/skywalker.go

fmt:
	find . -name '*.go'|xargs gofmt -w

clean:
	rm -rf bin/luker
