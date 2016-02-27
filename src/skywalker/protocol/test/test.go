package test

import (
    "skywalker/protocol"
)

type InTest struct {
    header bool
}

type OutTest struct {
}

func (t *InTest) Name() string {
    return "Test In"
}

func (t *InTest) Start(inbound bool, cfg interface {}) bool {
    if inbound {
        return true
    }else {
        return false
    }
}

func (t *InTest) Read(data []byte) (interface{}, interface{}, protocol.ProtocolError) {
    if t.header == false {
        t.header = true
        return [][]byte{[]byte("www.baidu.com:80"), data}, nil, nil
    }
    return data, nil, nil
}

func (t *InTest) Close() {
}

func (t *OutTest) Name() string {
    return "Test Out"
}

func (t *OutTest) Start(inbound bool, cfg interface{}) bool {
    if ! inbound {
        return true
    } else {
        return false
    }
}

func (t *OutTest) Read(data []byte) (interface{}, interface{}, protocol.ProtocolError) {
    return data, nil, nil
}

func (t *OutTest) Close() {
}
