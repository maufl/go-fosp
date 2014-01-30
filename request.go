package main

import (
  "fmt"
  "errors"
)

type RequestType uint

const (
  _ = iota
  Connect RequestType = iota
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

var stringToRequestType = map[string]RequestType {
  "CONNECT": Connect,
  "REGISTER": Register,
  "AUTHENTICATE": Authenticate,
  "SELECT": Select,
  "CREATE": Create,
  "UPDATE": Update,
  "DELETE": Delete,
  "LIST": List,
  "READ": Read,
  "WRITE": Write,
}

var requestTypeToString = map[RequestType]string {
  Connect: "CONNECT",
  Register: "REGISTER",
  Authenticate: "AUTHENTICATE",
  Select: "SELECT",
  Create: "CREATE",
  Update: "UPDATE",
  Delete: "DELETE",
  List: "LIST",
  Read: "READ",
  Write: "WRITE",
}

func (rt RequestType) String() string {
  if t, ok := requestTypeToString[rt]; ok {
    return t
  } else {
    return "NA_REQUEST_TYPE"
  }
}

func GetRequestType(s string) (RequestType, error) {
  if t := stringToRequestType[s]; t == 0 {
    return 0, errors.New("Not a valid request type")
  } else {
    return t, nil
  }
}

type Request struct {
  headers map[string]string
  body string

  request RequestType
  url *Url
  seq int
}

func NewRequest(rt RequestType, url *Url, seq int, headers map[string]string, body string) *Request {
  req := &Request{headers, body, rt, url, seq}
  return req
}

func (r *Request) SetHead(k, v string) {
  r.headers[k] = v
}

func (r Request) GetHead(k string) (string, bool) {
  head, ok := r.headers[k]
  return head, ok
}

func (r *Request) Headers() map[string]string {
  return r.headers
}

func (r *Request) SetBody(b string) {
  r.body = b
}

func (r Request) GetBody() string {
  return r.body
}

func (r Request) String() string {
  result := fmt.Sprintf("%s %s %d\r\n", r.request, r.url, r.seq)
  for k,v := range r.headers {
    result += k + ": " + v + "\r\n"
  }
  if r.body != "" {
    result += "\r\n" + r.body
  }
  return result
}

func (r Request) Bytes() []byte {
  return []byte(r.String())
}

func (r Request) Failed(status uint, body string) *Response {
  resp := &Response{response: Failed, status: status, seq: r.seq, body: body}
  if user, ok := r.headers["User"]; ok {
    resp.SetHead("User", user)
  }
  return resp
}

func (r Request) Succeeded(status uint, body string) *Response {
  resp := &Response{response: Succeeded, status: status, seq: r.seq, body: body}
  if user, ok := r.headers["User"]; ok {
    resp.SetHead("User", user)
  }
  return resp
}

func (r Request) GetBodyObject() (*Object, error) {
  o, err := Unmarshal(r.body)
  return o, err
}

type AuthenticationObject struct {
  Name string
  Password string
  Type string
  Domain string
}

type ConnectionNegotiationObject struct {
  Version string
}
