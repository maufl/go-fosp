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
	"bytes"
	"errors"
	"github.com/op/go-logging"
	"strconv"
	"strings"
)

func parseMessage(b []byte) (Message, error) {
	lg := logging.MustGetLogger("go-fosp/fosp/parser")
	lines := bytes.Split(b, []byte("\r\n"))
	scalp := strings.Split(string(lines[0]), " ")
	if len(scalp) < 2 {
		return nil, errors.New("Invalid formatted message")
	}
	var msg Message
	if reqType, e := ParseRequestType(scalp[0]); e == nil {
		if len(scalp) != 3 {
			return nil, errors.New("Invalid formatted message")
		}
		var url *Url
		if scalp[1] != "*" {
			var err error
			url, err = parseUrl(scalp[1])
			if err != nil {
				return nil, err
			}
		}
		seq, _ := strconv.Atoi(scalp[2])
		msg = NewRequest(reqType, url, seq, make(map[string]string), []byte(""))
	} else if respType, e := ParseResponseType(scalp[0]); e == nil {
		if len(scalp) != 3 {
			return nil, errors.New("Invalid formatted message")
		}
		status, _ := strconv.Atoi(scalp[1])
		seq, _ := strconv.Atoi(scalp[2])
		msg = NewResponse(respType, uint(status), seq, map[string]string{}, []byte(""))
	} else if event, e := ParseEvent(scalp[0]); e == nil {
		if len(scalp) != 2 {
			return nil, errors.New("Invalid formatted notification")
		}
		url, err := parseUrl(scalp[1])
		if err != nil {
			return nil, err
		}
		msg = NewNotification(event, url, map[string]string{}, "")
	} else {
		return nil, errors.New("Invalid formated message")
	}
	// First line was already processed
	lines = lines[1:]
	for {
		// Break if there are no more lines
		if len(lines) == 0 {
			break
		}
		line := string(lines[0])
		// Break if it is an empty line
		if strings.TrimSpace(line) == "" {
			// Discard the empty line
			lines = lines[1:]
			break
		}
		head := strings.Split(line, ": ")
		if len(head) != 2 {
			return nil, errors.New("Invalid header :: " + line)
		}
		msg.SetHead(head[0], head[1])
		// Discard the processed line
		lines = lines[1:]
	}

	lg.Debug("Number of lines for body is %d", len(lines))

	body := bytes.Join(lines, []byte("\r\n"))
	msg.SetBody(body)
	return msg, nil
}
