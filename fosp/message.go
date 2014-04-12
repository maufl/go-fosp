// Copyright (C) 2014 Felix Maurer
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package fosp

type MessageType uint

const (
	Text MessageType = 1 << iota
	Binary
)

type Message interface {
	SetHead(string, string)
	Head(string) (string, bool)
	Headers() map[string]string
	SetBody([]byte)
	SetBodyString(string)
	Body() []byte
	BodyString() string
	Bytes() []byte
	String() string
	Type() MessageType
	SetType(MessageType)
}

type BasicMessage struct {
	headers      map[string]string
	body         []byte
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

func (bm *BasicMessage) SetBody(b []byte) {
	bm.body = b
}

func (bm *BasicMessage) SetBodyString(b string) {
	bm.body = []byte(b)
}

func (bm *BasicMessage) Body() []byte {
	return bm.body
}

func (bm *BasicMessage) BodyString() string {
	return string(bm.body)
}

func (bm *BasicMessage) Type() MessageType {
	return bm.message_type
}

func (bm *BasicMessage) SetType(mt MessageType) {
	bm.message_type = mt
}
