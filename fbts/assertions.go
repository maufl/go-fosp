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

package main

import (
	"github.com/maufl/go-fosp/fosp"
)

// E represents the result that is expected of a Response-error pair
type E struct {
	Error    bool
	Failed   bool
	Code     uint
	Body     string
	JSONBody string
}

var expectE = expect(E{})
var expectFailed = expect(E{Failed: true})

func expect(e E) func(*fosp.Response, error) (*fosp.Response, error) {
	return func(resp *fosp.Response, err error) (*fosp.Response, error) {
		if !e.Error && err != nil {
			panic("Expected error to be nil but was not")
		}
		if e.Error && err == nil {
			panic("Expected error to not be nil but was nil")
		}
		if !e.Failed && resp.ResponseType() == fosp.Failed {
			panic("Expected the request to succeed but response is Failed: " + resp.BodyString())
		}
		if e.Failed && resp.ResponseType() == fosp.Succeeded {
			panic("Expected the request to fail but response is Succeeded")
		}
		// TODO: Add check for Code, requires change to Response struct
		if e.Body != "" && e.Body != resp.BodyString() {
			panic("Expected the body to be " + e.Body + " but is " + resp.BodyString())
		}
		// TODO: Add check for body content by comparing JSON data
		return resp, err
	}
}
