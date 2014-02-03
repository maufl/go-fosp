package main

type MessageType uint

const (
	Text MessageType = 1 << iota
	Binary
)

type Message interface {
	SetHead(string, string)
	Head(string) (string, bool)
	Headers() map[string]string
	SetBody(string)
	Body() string
	Bytes() []byte
	String() string
	Type() MessageType
	SetType(MessageType) 
}

type BasicMessage struct {
	headers map[string]string
	body    string
	message_type MessageType
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

func (bm *BasicMessage) Type() MessageType {
	return bm.message_type
}

func (bm *BasicMessage) SetType(mt MessageType) {
	bm.message_type = mt
}
