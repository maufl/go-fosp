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

type RequestType uint

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
		return 0, errors.New("Not a valid request type")
	}
}

type Request struct {
	BasicMessage

	request RequestType
	url     *Url
	seq     int
}

func NewRequest(rt RequestType, url *Url, seq int, headers map[string]string, body []byte) *Request {
	return &Request{BasicMessage{headers, body, Text}, rt, url, seq}
}

func (r *Request) Url() *Url {
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

func (r *Request) Bytes() []byte {
	return []byte(r.String())
}

func (r Request) Failed(status uint, body string) *Response {
	resp := NewResponse(Failed, status, r.seq, make(map[string]string), []byte(body))
	if user, ok := r.headers["User"]; ok {
		resp.SetHead("User", user)
	}
	return resp
}

func (r Request) SucceededWithBody(status uint, body []byte) *Response {
	resp := NewResponse(Succeeded, status, r.seq, make(map[string]string), body)
	if user, ok := r.headers["User"]; ok {
		resp.SetHead("User", user)
	}
	return resp
}

func (r *Request) Succeeded(status uint) *Response {
	return r.SucceededWithBody(status, []byte(""))
}

func (r Request) BodyObject() (*Object, error) {
	o, err := Unmarshal(string(r.body))
	return o, err
}

type AuthenticationObject struct {
	Name     string
	Password string
	Type     string
	Domain   string
}

type ConnectionNegotiationObject struct {
	Version string
}
