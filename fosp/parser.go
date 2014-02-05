package fosp

import (
	"errors"
	"strconv"
	"strings"
)

func parseMessage(b string) (Message, error) {
	lines := strings.Split(b, "\r\n")
	scalp := strings.Split(lines[0], " ")
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
		msg = NewRequest(reqType, url, seq, make(map[string]string), "")
	} else if respType, e := ParseResponseType(scalp[0]); e == nil {
		if len(scalp) != 3 {
			return nil, errors.New("Invalid formatted message")
		}
		status, _ := strconv.Atoi(scalp[1])
		seq, _ := strconv.Atoi(scalp[2])
		msg = NewResponse(respType, uint(status), seq, map[string]string{}, "")
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

	lines = lines[1:]
	for {
		if len(lines) == 0 {
			break
		}
		line := lines[0]

		if strings.TrimSpace(line) == "" {
			break
		}
		head := strings.Split(line, ": ")
		if len(head) != 2 {
			return nil, errors.New("Invalid header :: " + line)
		}
		msg.SetHead(head[0], head[1])

		lines = lines[1:]
	}

	body := strings.Join(lines, "")
	msg.SetBody(body)
	return msg, nil
}