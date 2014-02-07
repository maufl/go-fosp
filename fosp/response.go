package fosp

import (
	"errors"
	"fmt"
)

type ResponseType uint

const (
	Succeeded ResponseType = 1 << iota
	Failed
)

func (rt ResponseType) String() string {
	switch rt {
	case Succeeded:
		return "SUCCEEDED"
	case Failed:
		return "FAILED"
	default:
		return "NA_RESPONSE_TYPE"
	}
}

func ParseResponseType(s string) (ResponseType, error) {
	switch s {
	case "SUCCEEDED":
		return Succeeded, nil
	case "FAILED":
		return Failed, nil
	default:
		return 0, errors.New("Not a valid response type")
	}
}

type Response struct {
	BasicMessage

	response ResponseType
	status   uint
	seq      int
}

func NewResponse(rt ResponseType, status uint, seq int, headers map[string]string, body []byte) *Response {
	return &Response{BasicMessage{headers, body, Text}, rt, status, seq}
}

func (r *Response) String() string {
	result := fmt.Sprintf("%s %d %d\r\n", r.response, r.status, r.seq)
	for k, v := range r.headers {
		result += k + ": " + v + "\r\n"
	}
	if string(r.body) != "" {
		result += "\r\n" + string(r.body)
	}
	return result
}

func (r *Response) Bytes() []byte {
	return []byte(r.String())
}
