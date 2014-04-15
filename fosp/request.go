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

import (
	"errors"
	"fmt"
)

// RequestType is the type of a FOSP request.
type RequestType uint

// One constant for each type of FOSP requests.
const (
	Connect RequestType = 1 << iota
	Register
	Authenticate
	Create
	Select
	Update
	Delete
	List
	Read
	Write
)

func (rt RequestType) String() string {
	switch rt {
	case Connect:
		return "CONNECT"
	case Register:
		return "REGISTER"
	case Authenticate:
		return "AUTHENTICATE"
	case Select:
		return "SELECT"
	case Create:
		return "CREATE"
	case Update:
		return "UPDATE"
	case Delete:
		return "DELETE"
	case List:
		return "LIST"
	case Read:
		return "READ"
	case Write:
		return "WRITE"
	default:
		return "NA_REQUEST_TYPE"
	}
}

// ParseRequestType parses a string and returns the corresponding RequestType.
func ParseRequestType(s string) (RequestType, error) {
	switch s {
	case "CONNECT":
		return Connect, nil
	case "REGISTER":
		return Register, nil
	case "AUTHENTICATE":
		return Authenticate, nil
	case "SELECT":
		return Select, nil
	case "CREATE":
		return Create, nil
	case "UPDATE":
		return Update, nil
	case "DELETE":
		return Delete, nil
	case "LIST":
		return List, nil
	case "READ":
		return Read, nil
	case "WRITE":
		return Write, nil
	default:
		return 0, errors.New("not a valid request type")
	}
}

// Request represents a FOSP request message.
type Request struct {
	BasicMessage

	request RequestType
	url     *URL
	seq     int
}

// NewRequest creates a new request.
func NewRequest(rt RequestType, url *URL, seq int, headers map[string]string, body []byte) *Request {
	return &Request{BasicMessage{headers, body, Text}, rt, url, seq}
}

// URL returns the URL of the Request.
func (r *Request) URL() *URL {
	return r.url
}

func (r Request) String() string {
	result := fmt.Sprintf("%s %s %d\r\n", r.request, r.url, r.seq)
	for k, v := range r.headers {
		result += k + ": " + v + "\r\n"
	}
	if string(r.body) != "" {
		result += "\r\n" + string(r.body)
	}
	return result
}

// Bytes returns the string representation of the Request as byte array.
func (r *Request) Bytes() []byte {
	return []byte(r.String())
}

// Failed returns a Response of the Failed type with the same sequence number.
func (r Request) Failed(status uint, body string) *Response {
	resp := NewResponse(Failed, status, r.seq, make(map[string]string), []byte(body))
	if user, ok := r.headers["User"]; ok {
		resp.SetHead("User", user)
	}
	return resp
}

// SucceededWithBody returns a Response of the Succeeded type with the same sequence number and a given body.
func (r Request) SucceededWithBody(status uint, body []byte) *Response {
	resp := NewResponse(Succeeded, status, r.seq, make(map[string]string), body)
	if user, ok := r.headers["User"]; ok {
		resp.SetHead("User", user)
	}
	return resp
}

// Succeeded returns a Response of the Succeded type with the same sequence number.
func (r *Request) Succeeded(status uint) *Response {
	return r.SucceededWithBody(status, []byte(""))
}

// BodyObject returns the Object representation of the body content or an error.
func (r Request) BodyObject() (*Object, error) {
	o, err := Unmarshal(string(r.body))
	return o, err
}

// AuthenticationObject represents the information sent in an AUTHENTICATE request.
type AuthenticationObject struct {
	Name     string
	Password string
	Type     string
	Domain   string
}

// ConnectionNegotiationObject represents the information sent in a CONNECT request.
type ConnectionNegotiationObject struct {
	Version string
}
