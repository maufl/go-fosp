package main

import (
  "fmt"
  "errors"
)

type ResponseType uint

const (
  _ = iota
  Succeeded ResponseType = iota
  Failed
)

var stringToResponseType = map[string]ResponseType {
  "SUCCEEDED": Succeeded,
  "FAILED": Failed,
}

var responseTypeToString = map[ResponseType]string {
  Succeeded: "SUCCEEDED",
  Failed: "FAILED",
}

func (rt ResponseType) String() string {
  if t, ok := responseTypeToString[rt]; ok {
    return t
  } else {
    return "NA_RESPONSE_TYPE"
  }
}

func GetResponseType(s string) (ResponseType, error) {
  if t := stringToResponseType[s]; t == 0 {
    return 0, errors.New("Not a valid response type")
  } else {
    return t, nil
  }
}

type Response struct {
  headers map[string]string
  body string

  response ResponseType
  status uint
  seq int
}

func (r *Response) SetHead(k, v string) {
  r.headers[k] = v
}

func (r Response) GetHead(k string) string {
  return r.headers[k]
}

func (r *Response) DeleteHead(k string) {
  delete(r.headers, k)
}

func (r *Response) SetBody(b string) {
  r.body = b
}

func (r Response) GetBody() string {
  return r.body
}

func (r Response) String() string {
  result := fmt.Sprintf("%s %d %d\r\n", r.response, r.status, r.seq)
  for k,v := range r.headers {
    result += k + ": " + v + "\r\n"
  }
  if r.body != "" {
    result += "\r\n" + r.body
  }
  return result
}

func (r Response) Bytes() []byte {
  return []byte(r.String())
}
