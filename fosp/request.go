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
	"fmt"
	"io"
	"net/textproto"
	"net/url"
)

// Request represents a FOSP request message.
type Request struct {
	Method string
	Header textproto.MIMEHeader
	Body   io.Reader

	URL *url.URL
}

// NewRequest creates a new request.
func NewRequest(method string, url *url.URL) *Request {
	return &Request{Method: method, Header: make(map[string][]string), Body: nil, URL: url}
}

func (r *Request) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.URL)
}

func (r *Request) nop() {}
