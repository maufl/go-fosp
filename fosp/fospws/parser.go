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

func serializeMessage(msg fosp.Message) []byte {
	return []byte{}
}
