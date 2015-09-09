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
	"testing"
)

func TestParser(t *testing.T) {
	raw := []byte("OPTIONS * 1\r\n")
	buffer := bytes.NewBuffer(raw)
	msg, seq, err := parseMessage(buffer)
	if err != nil {
		t.Errorf("Parsing of message went wrong: %s", err)
	}
	if seq != 1 {
		t.Errorf("Wrong sequence number was returned: %d", seq)
	}
	if _, ok := msg.(*fosp.Request); !ok {
		t.Errorf("Returned message is not a request: %#v", msg)
	}
}
