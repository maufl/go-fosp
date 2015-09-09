// Copyright (C) 2015 Felix Maurer
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
	"bytes"
	"fmt"
	"io"
	"net/textproto"
)

const (
	StatusOK        uint = 200
	StatusCreated        = 201
	StatusNoContent      = 204

	StatusMovedPermanently = 301
	StatusNotModified      = 304

	StatusBadRequest            = 400
	StatusUnauthorized          = 401
	StatusForbidden             = 403
	StatusNotFound              = 404
	StatusMethodNotAllowed      = 405
	StatusConflict              = 409
	StatusPreconditionFailed    = 412
	StatusRequestEntityTooLarge = 413

	StatusInternalServerError = 500
	StatusNotImplemented      = 501
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
	StatusGatewayTimeout      = 504
)

// Response represents a FOSP response message.
type Response struct {
	Status string
	Header textproto.MIMEHeader
	Body   io.Reader

	Code uint
}

// NewResponse creates a new response message.
func NewResponse(status string, code uint) *Response {
	return &Response{Status: status, Header: make(map[string][]string), Body: &bytes.Buffer{}, Code: code}
}

func (r *Response) String() string {
	return fmt.Sprintf("%s %d", r.Status, r.Code)
}

func (r *Response) nop() {}
