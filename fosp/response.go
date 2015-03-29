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

// ErrInvalidResponseTyp is returned when a string is parsed that does not represent an valid ResponseType.
var ErrInvalidResponseType = errors.New("invalid response type")

// ResponseType represents the type of a FOSP response message.
type ResponseType uint

const (
	// Succeeded denotes a SUCCEEDED response.
	Succeeded ResponseType = 1 << iota
	// Failed dentoes a FAILED response.
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

// ParseResponseType parses a string and returns the corresponding ResponseType or an error.
func ParseResponseType(s string) (ResponseType, error) {
	switch s {
	case "SUCCEEDED":
		return Succeeded, nil
	case "FAILED":
		return Failed, nil
	default:
		return 0, ErrInvalidResponseType
	}
}

// Response represents a FOSP response message.
type Response struct {
	BasicMessage

	response ResponseType
	status   uint
	seq      int
}

// NewResponse creates a new response message.
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

// Bytes returns the string representation of the Response as byte array.
func (r *Response) Bytes() []byte {
	return []byte(r.String())
}

func (r *Response) Status() uint {
	return r.status
}

func (r *Response) Seq() int {
	return r.seq
}

// ResponseType returns the ResponseType of this response.
func (r *Response) ResponseType() ResponseType {
	return r.response
}
