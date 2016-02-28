package test

type InTest struct {
    header bool
}

type OutTest struct {
}

func (t *InTest) Name() string {
    return "Test In"
}

func (t *InTest) Start(cfg interface {}) bool {
    return true
}

func (t *InTest) Read(data []byte) (interface{}, interface{}, error) {
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

func (t *OutTest) Start(cfg interface{}) bool {
    return true
}

func (t *OutTest) GetRemoteAddress(addr string, port string) (string, string) {
    return "www.baidu.com", port
}

func (t *OutTest) Read(data []byte) (interface{}, interface{}, error) {
    return data, nil, nil
}

func (t *OutTest) Close() {
}
