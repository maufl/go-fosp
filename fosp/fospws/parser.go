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

package fospws

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/maufl/go-fosp/fosp"
	"io"
	"net/textproto"
	"net/url"
	"path"
	"strconv"
)

type nestedError struct {
	Message string
	Nested  error
}

func (e *nestedError) Error() string {
	return fmt.Sprintf("%s, with nested error: %s", e.Message, e.Nested)
}

func newNestedError(msg string, err error) *nestedError {
	return &nestedError{Message: msg, Nested: err}
}

func parseMessage(in io.Reader) (msg fosp.Message, seq uint64, err error) {
	var (
		firstLine []byte
		rawurl    string
		isPrefix  bool
		readerErr error
		fragments [][]byte
		code      int
		msgURL    *url.URL
		reader    *bufio.Reader
		ok        bool
	)
	err = errors.New("Failed to parse message, unknown error")
	if reader, ok = in.(*bufio.Reader); !ok {
		reader = bufio.NewReader(in)
	}
	if firstLine, isPrefix, readerErr = reader.ReadLine(); isPrefix {
		err = errors.New("First line of message is too long")
		return
	} else if readerErr != nil && readerErr != io.EOF {
		err = newNestedError("Reader error", readerErr)
		return
	}
	if fragments = bytes.Split(firstLine, []byte(" ")); len(fragments) < 2 {
		err = errors.New("First line does not consist of at least 2 parts")
		return
	}
	identifier := string(fragments[0])
	switch identifier {
	case fosp.OPTIONS, fosp.AUTH, fosp.GET, fosp.LIST, fosp.CREATE, fosp.PATCH, fosp.DELETE, fosp.READ, fosp.WRITE:
		if len(fragments) != 3 {
			err = errors.New("Request line does not consist of 3 parts")
			return
		}
		rawurl = string(fragments[1])
		if rawurl != "*" {
			rawurl = "fosp://" + string(fragments[1])
			if msgURL, err = url.Parse(rawurl); err != nil {
				err = errors.New("Invalid request URL")
				return
			}
			msgURL.Path = path.Clean(msgURL.Path)
			if msgURL.Path == "." {
				msgURL.Path = "/"
			}
		}
		if seq, err = strconv.ParseUint(string(fragments[2]), 10, 64); err != nil || seq < 1 {
			err = newNestedError("The request sequence number is not valid", err)
			return
		}
		req := fosp.NewRequest(identifier, msgURL)
		if req.Header, err = textproto.NewReader(reader).ReadMIMEHeader(); err != nil && err != io.EOF {
			err = newNestedError("The request header is not valid", err)
			return
		}
		req.Body = reader
		return req, seq, nil
	case fosp.SUCCEEDED, fosp.FAILED:
		if len(fragments) != 3 {
			err = errors.New("Response line does not consist of 3 parts")
			return
		}
		if code, err = strconv.Atoi(string(fragments[1])); err != nil {
			err = newNestedError("Status code is invalid", err)
			return
		}
		if seq, err = strconv.ParseUint(string(fragments[2]), 10, 64); err != nil || seq < 1 {
			err = newNestedError("The response sequence number is not valid", err)
			if seq < 1 {
				err = errors.New("The sequence number is 0")
			}
			return
		}
		resp := fosp.NewResponse(identifier, uint(code))
		if resp.Header, err = textproto.NewReader(reader).ReadMIMEHeader(); err != nil && err != io.EOF {
			err = newNestedError("The response header is not valid", err)
			return
		}
		resp.Body = reader
		return resp, seq, nil
	case fosp.CREATED, fosp.UPDATED, fosp.DELETED:
		if len(fragments) != 2 {
			err = errors.New("Notification line does not consist of 2 parts")
			return
		}
		rawurl = string(fragments[1])
		if rawurl != "*" {
			rawurl = "fosp://" + string(fragments[1])
			if msgURL, err = url.Parse(rawurl); err != nil {
				err = errors.New("Invalid request URL")
				return
			}
			msgURL.Path = path.Clean(msgURL.Path)
			if msgURL.Path == "." {
				msgURL.Path = "/"
			}
		}
		evt := fosp.NewNotification(identifier, msgURL)
		if evt.Header, err = textproto.NewReader(reader).ReadMIMEHeader(); err != nil && err != io.EOF {
			err = newNestedError("The notification header is not valid", err)
			return
		}
		evt.Body = reader
		return evt, seq, nil
	default:
		err = errors.New("Unrecognized identifier " + identifier)
		return
	}
}

func serializeMessage(msg fosp.Message, seq uint64) []byte {
	buffer := bytes.NewBuffer([]byte{})
	var (
		u      string
		header textproto.MIMEHeader
		body   io.Reader
	)
	switch msg := msg.(type) {
	case *fosp.Request:
		if msg.URL == nil {
			u = "*"
		} else {
			u = msg.URL.String()
		}
		buffer.WriteString(fmt.Sprintf("%s %s %d\r\n", msg.Method, u, seq))
		header = msg.Header
		body = msg.Body
	case *fosp.Response:
		buffer.WriteString(fmt.Sprintf("%s %d %d\r\n", msg.Status, msg.Code, seq))
		header = msg.Header
		body = msg.Body
	case *fosp.Notification:
		if msg.URL == nil {
			u = "*"
		} else {
			u = msg.URL.String()
		}
		buffer.WriteString(fmt.Sprintf("%s %s\r\n", msg.Event, u))
		header = msg.Header
		body = msg.Body
	default:
		panic("Only valid FOSP messages can be serialized")
	}
	for key, values := range header {
		for _, value := range values {
			buffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	if body != nil {
		buffer.WriteString("\r\n")
		if _, err := buffer.ReadFrom(body); err != nil && err != io.EOF {
			panic(err.Error())
		}
	}
	return buffer.Bytes()
}
