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
	"github.com/shavac/readline"
	"github.com/op/go-logging"
	"strings"
	"os"
)

type stateStruct struct {
	Username string
	Password string
	Cwd string
}

var state = stateStruct{"<anonymous>", "", "<nowhere>"}
var prompt = state.Username + " @ " + state.Cwd + " >"
var client = fosp.Client{}

func main() {
	logging.SetFormatter(logging.MustStringFormatter("[%{time:2006-01-02T15:04} | %{level:.3s} | %{module}]  %{message}"))
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logBackend.Color = true
	logging.SetBackend(logBackend)

	loop()
	return
}

func loop() {
	for {
		switch result := readline.ReadLine(&prompt); true {
		case result == nil:
			println()
			return
		case *result != "": //ignore blank lines
			parseCommand(*result);
			readline.AddHistory(*result); //allow user to recall this line
		}
	}
}

func parseCommand(input string) {
	input = strings.TrimSpace(input)
	tokens := strings.SplitN(input, " ", 2)
	cmd := tokens[0]
	args := ""
	if len(tokens) == 2 { args = tokens[1] }

	switch cmd {
	case "quit":
		quit(args)
	case "open":
		open(args)
	case "connect":
		connect(args)
	case "authenticate":
		authenticate(args)
	case "select":
		selekt(args)
	default:
		println("Unknown command " + cmd)
	}
}

func open(args string) {
	if err := client.OpenConnection(args); err != nil {
		println(err.Error())
	}
}

func quit(args string) {
	os.Exit(0)
}

func connect(args string) {
	_, err := client.Connect()
	if err != nil {
		println(err.Error())
	}
}

func authenticate(args string) {
	parts := strings.Split(args, " ")
	if len(parts) != 2 {
		println("Not enough arguments for authenticate")
	}
	client.Authenticate(parts[0], parts[1])
}

func selekt(args string) {
	url, err := fosp.ParseURL(args)
	if err != nil {
		println(args + " is not a valid URL")
		return
	}
	if resp, err := client.Select(url); err == nil {
		println(resp.BodyString())
	} else {
		println("Select failed: " + err.Error())
	}
}
