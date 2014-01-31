package main

type Message interface {
	SetHead(string, string)
	Head(string) (string, bool)
	Headers() map[string]string
	SetBody(string)
	Body() string
	Bytes() []byte
	String() string
}

type BasicMessage struct {
	headers map[string]string
	body    string
}

func (bm *BasicMessage) SetHead(k, v string) {
	bm.headers[k] = v
}

func (bm *BasicMessage) DeleteHead(k string) {
	delete(bm.headers, k)
}

func (bm *BasicMessage) Head(k string) (string, bool) {
	head, ok := bm.headers[k]
	return head, ok
}

func (bm *BasicMessage) Headers() map[string]string {
	return bm.headers
}

func (bm *BasicMessage) SetBody(b string) {
	bm.body = b
}

func (bm *BasicMessage) Body() string {
	return bm.body
}

func (bm *BasicMessage) String() string {
	return "NOT_IMPLEMENTED"
}

func (bm *BasicMessage) Bytes() []byte {
	return []byte(bm.String())
}
