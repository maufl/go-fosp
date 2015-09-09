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
	"bytes"
	"github.com/maufl/go-fosp/fosp"
	"net/url"
	"reflect"
	"testing"
)

type Expectation struct {
	Error error
	Message fosp.Message
	Method string
	Status string
	Event string
	URL *url.URL
	Seq uint
}

type ParserTestCase struct {
	RawMessage string
	Expect Expectation
}

var testCases []ParserTestCase = []ParserTestCase{
	{
		RawMessage: "AUTH * 1\r\n",
		Expect: Expectation{
			Message: &fosp.Request{},
			Method: fosp.AUTH,
		},
	},
	{
		RawMessage: "UPDATED felix@maufl.de/social/me\r\n",
		Expect: Expectation{
			Message: &fosp.Notification{},
			Event: fosp.UPDATED,
		},
	},
}

type SerializerTestCase struct {
	Message fosp.Message
	Method string
	Status string
	Event string
	RawURL string
	Code uint
	Seq uint
	Expect []byte
}

var serializerTestCases []SerializerTestCase = []SerializerTestCase{
	{
		Message: &fosp.Request{},
		Method: fosp.READ,
		RawURL: "felix@maufl.de/social/me",
		Seq: 1,
		Expect: []byte("READ felix@maufl.de/social/me 1\r\n"),
	},
}

func TestParser(t *testing.T) {
	for _, testCase := range(testCases) {
		raw := []byte(testCase.RawMessage)
		buffer := bytes.NewBuffer(raw)
		msg, seq, err := parseMessage(buffer)
		if err != nil {
			t.Errorf("Parsing of message went wrong: %s", err)
		}
		if testCase.Expect.Seq != 0 && uint(seq) != testCase.Expect.Seq {
			t.Errorf("Wrong sequence number was returned: expected %d got %d", testCase.Expect.Seq, seq)
		}
		if reflect.TypeOf(msg) != reflect.TypeOf(testCase.Expect.Message) {
			t.Errorf("Returned message is not a request: %#v", msg)
		}
	}
}

func TestSerializer(t *testing.T) {
	for _, testCase := range(serializerTestCases) {
		msg := testCase.Message
		url, err := url.Parse(testCase.RawURL)
		if err != nil {
			t.Errorf("Test case contains invalid URL %s", testCase.RawURL)
			continue
		}
		switch m := msg.(type) {
		case *fosp.Request:
			m.Method = testCase.Method
			m.URL = url
		case *fosp.Response:
			m.Status = testCase.Status
			m.Code = testCase.Code
		case *fosp.Notification:
			m.Event = testCase.Event
			m.URL = url
		default:
			t.Errorf("Testcase containts invalid FOSP message type")
			continue
		}
		raw := serializeMessage(msg, testCase.Seq)
		if bytes.Compare(raw, testCase.Expect) != 0 {
			t.Errorf("Serialized message differs from expected serialization: expected %s got %s", raw, testCase.Expect)
		}
	}
}
