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

// MessageType denotes the type of a FOSP message.
type MessageType uint

const (
	// Text denotes a text only message.
	Text MessageType = 1 << iota
	// Binary denotes a message with binary content, e.g., a READ request.
	Binary
)

// Message is the common interface of all FOSP messag objects.
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

// BasicMessage is the core object common for all FOSP message objects.
type BasicMessage struct {
	headers     map[string]string
	body        []byte
	messageType MessageType
}

// SetHead adds or overrides a head field of a message.
func (bm *BasicMessage) SetHead(k, v string) {
	bm.headers[k] = v
}

// DeleteHead removes a head field from a message.
func (bm *BasicMessage) DeleteHead(k string) {
	delete(bm.headers, k)
}

// Head returns a head field from a message.
func (bm *BasicMessage) Head(k string) (string, bool) {
	head, ok := bm.headers[k]
	return head, ok
}

// Headers returns all head fields from a message.
func (bm *BasicMessage) Headers() map[string]string {
	return bm.headers
}

// SetBody sets the body content of a message.
func (bm *BasicMessage) SetBody(b []byte) {
	bm.body = b
}

// SetBodyString works like SetBody but accepts a string.
func (bm *BasicMessage) SetBodyString(b string) {
	bm.body = []byte(b)
}

// Body returns the body content of a message.
func (bm *BasicMessage) Body() []byte {
	return bm.body
}

// BodyString returns the body content of a message as a string.
func (bm *BasicMessage) BodyString() string {
	return string(bm.body)
}

// Type returns the message Type.
func (bm *BasicMessage) Type() MessageType {
	return bm.messageType
}

// SetType sets the MessageType.
func (bm *BasicMessage) SetType(mt MessageType) {
	bm.messageType = mt
}
