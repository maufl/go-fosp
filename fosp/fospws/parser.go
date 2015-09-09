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
	"strconv"
	_ "strings"
)

// ErrInvalidMessageFormat is returned when a message is incorrectly formatted.
var ErrInvalidMessageFormat = errors.New("invalid formatted message")

// ErrInvalidHeaderFormat is returned when the headers of a messag are incorrectly formatted.
var ErrInvalidHeaderFormat = errors.New("invalid formatted header")

var ErrLineTooLong = errors.New("line of message was too long")

var ErrReaderError = errors.New("reader returned an error")

func parseMessage(in io.Reader) (msg fosp.Message, seq int, err error) {
	var (
		firstLine []byte
		rawurl    string
		isPrefix  bool
		fragments [][]byte
		code      int
		url       *url.URL
		reader    *bufio.Reader
		ok        bool
	)
	seq = -1
	err = ErrInvalidMessageFormat
	if reader, ok = in.(*bufio.Reader); !ok {
		reader = bufio.NewReader(in)
	}
	if firstLine, isPrefix, err = reader.ReadLine(); isPrefix {
		err = ErrLineTooLong
		return
	} else if err != nil {
		err = ErrReaderError
		return
	}
	if fragments = bytes.Split(firstLine, []byte(" ")); len(fragments) < 2 {
		return
	}
	identifier := string(fragments[0])
	switch identifier {
	case fosp.OPTIONS, fosp.AUTH, fosp.GET, fosp.LIST, fosp.CREATE, fosp.PATCH, fosp.DELETE, fosp.READ, fosp.WRITE:
		if len(fragments) != 3 {
			return
		}
		rawurl = string(fragments[1])
		if url, err = url.Parse(rawurl); rawurl != "*" && err != nil {
			return
		}
		if seq, err = strconv.Atoi(string(fragments[2])); err != nil {
			return
		}
		req := fosp.NewRequest(identifier, url)
		if req.Header, err = textproto.NewReader(reader).ReadMIMEHeader(); err != nil {
			return
		}
		req.Body = reader
		msg = req
		return
	case fosp.SUCCEEDED, fosp.FAILED:
		if len(fragments) != 3 {
			return
		}
		if code, err = strconv.Atoi(string(fragments[1])); err != nil {
			return
		}
		if seq, err = strconv.Atoi(string(fragments[2])); err != nil {
			return
		}
		resp := fosp.NewResponse(identifier, uint(code))
		if resp.Header, err = textproto.NewReader(reader).ReadMIMEHeader(); err != nil {
			return
		}
		resp.Body = reader
		msg = resp
		return
	case fosp.CREATED, fosp.UPDATED, fosp.DELETED:
		if len(fragments) != 2 {
			return
		}
		rawurl = string(fragments[1])
		if url, err = url.Parse(rawurl); err != nil {
			return
		}
		evt := fosp.NewNotification(identifier, url)
		if evt.Header, err = textproto.NewReader(reader).ReadMIMEHeader(); err != nil {
			return
		}
		evt.Body = reader
		msg = evt
		return
	default:
		return
	}
}

func serializeMessage(msg fosp.Message, seq int) []byte {
	buffer := bytes.NewBuffer([]byte{})
	var (
		header textproto.MIMEHeader
		body io.Reader
	)
	switch msg := msg.(type) {
	case *fosp.Request:
		buffer.WriteString(fmt.Sprintf("%s %s %d\r\n", msg.Method, msg.URL, seq))
		header = msg.Header
		body = msg.Body
	case *fosp.Response:
		buffer.WriteString(fmt.Sprintf("%s %d %d\r\n", msg.Status, msg.Code, seq))
		header = msg.Header
		body = msg.Body
	case * fosp.Notification:
		buffer.WriteString(fmt.Sprintf("%s %s\r\n", msg.Event, msg.URL))
		header = msg.Header
		body = msg.Body
	default:
		panic("Only valid FOSP messages can be serialized")
	}
	for key, values := range(header) {
		for _, value := range(values) {
			buffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	buffer.WriteString("\r\n")
	if _, err := buffer.ReadFrom(body); err != nil {
		panic(err.Error())
	}
	return buffer.Bytes()
}
