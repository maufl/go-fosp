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
	"flag"
	"github.com/maufl/go-fosp/fosp"
	"github.com/op/go-logging"
	"github.com/shavac/readline"
	"os"
	"strings"
)

type stateStruct struct {
	Remote   string
	Username string
	Password string
	Cwd      string
}

var state = stateStruct{"", "", "", ""}
var prompt = state.Username + " @ " + state.Cwd + " >"
var client = fosp.Client{}

func main() {
	logging.SetFormatter(logging.MustStringFormatter("[%{time:2006-01-02T15:04} | %{level:.3s} | %{module}]  %{message}"))
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logBackend.Color = true
	logging.SetBackend(logBackend)

	flag.StringVar(&state.Remote, "h", "", "The host to which to connect on startup.")
	flag.StringVar(&state.Username, "u", "", "The username which to use.")
	flag.StringVar(&state.Password, "p", "", "The passwort of the user.")
	flag.Parse()

	if state.Remote != "" {
		open(state.Remote)
		connect("")
		if state.Username != "" && state.Password != "" {
			authenticate(state.Username + " " + state.Password)
		}
	}

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
			parseCommand(*result)
			readline.AddHistory(*result) //allow user to recall this line
		}
	}
}

func buildPrompt() {
	prompt = state.Username + " @ " + state.Cwd + " >"
}

func determinURL(path string) (*fosp.URL, error) {
	url, err := fosp.ParseURL(path)
	if err != nil {
		url, err = fosp.ParseURL(state.Cwd + "/" + path)
	}
	return url, err
}

func parseCommand(input string) {
	input = strings.TrimSpace(input)
	tokens := strings.SplitN(input, " ", 2)
	cmd := tokens[0]
	args := ""
	if len(tokens) == 2 {
		args = tokens[1]
	}

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
	case "list":
		list(args)
	case "create":
		create(args)
	default:
		println("Unknown command " + cmd)
	}
}

func open(args string) {
	if err := client.OpenConnection(args); err != nil {
		println(err.Error())
	} else {
		state.Remote = args
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
	if resp, err := client.Authenticate(parts[0], parts[1]); err == nil && resp.ResponseType() == fosp.Succeeded {
		state.Username = parts[0]
		state.Password = parts[0]
		state.Cwd = state.Username + "@" + state.Remote
		buildPrompt()
	}
}

func selekt(args string) {
	url, err := determinURL(args)
	if err != nil {
		println(args + " is not a valid path")
		return
	}
	if resp, err := client.Select(url); err == nil {
		println(resp.BodyString())
	} else {
		println("Select failed: " + err.Error())
	}
}

func list(args string) {
	url, err := determinURL(args)
	if err != nil {
		println(args + " is not a valid path")
		return
	}
	if resp, err := client.List(url); err == nil {
		println(resp.BodyString())
	} else {
		println("Select failed: " + err.Error())
	}
}

func create(args string) {
	tokens := strings.SplitN(args, " ", 2)
	path := tokens[0]
	content := ""
	if len(tokens) == 2 {
		content = tokens[1]
	}
	url, err := determinURL(path)
	if err != nil {
		println(path + " is not a valid path")
		return
	}
	obj, err := fosp.Unmarshal(content)
	if err != nil {
		println(content + " is not a valid FOSP object")
		return
	}
	if _, err := client.Create(url, obj); err == nil {
		println("Create succeeded")
	} else {
		println("Create failed: " + err.Error())
	}
}
