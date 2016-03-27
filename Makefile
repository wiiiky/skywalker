
all: skywalker

skywalker:
	GOPATH=`pwd` go build -o bin/luker src/skywalker/skywalker.go
