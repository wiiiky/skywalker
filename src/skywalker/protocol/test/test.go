package test

import (
    "skywalker/shell"
)

type InTest struct {
    header bool
}

type OutTest struct {
}

func (t *InTest) Start(opts *shell.Options) bool {
    return true
}

func (t *InTest) Read(data []byte) interface{} {
    if t.header == false {
        t.header = true
        return [][]byte{[]byte("www.baidu.com:80"), data}
    }
    return data
}

func (t *InTest) Close() {
}

func (t *OutTest) Start(opts *shell.Options) bool {
    return true
}

func (t *OutTest) Read(data []byte) interface{} {
    return data
}

func (t *OutTest) Close() {
}
