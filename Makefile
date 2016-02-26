
all: skywalker

skywalker:
	GOPATH=`pwd` go build -o bin/skwer src/skywalker/skywalker.go
