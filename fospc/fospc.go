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

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/maufl/go-fosp/fosp"
	"github.com/maufl/go-fosp/fosp/fospws"
	"github.com/op/go-logging"
	"github.com/shavac/readline"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
)

var state struct {
	Remote   string
	Username string
	Password string
	Cwd      string
}
var prompt = state.Username + " @ " + state.Cwd + " >"
var connection *fospws.Connection

type emptyMessageHandler struct{}

func (e emptyMessageHandler) HandleMessage(msg *fospws.NumberedMessage) {}

var e = emptyMessageHandler{}

func main() {
	logging.SetFormatter(logging.MustStringFormatter("[%{time:2006-01-02T15:04} | %{level:.3s} | %{module}]  %{message}"))
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logBackend.Color = true
	logging.SetBackend(logBackend)
	logging.SetLevel(logging.NOTICE, "")

	flag.StringVar(&state.Remote, "h", "", "The host to which to connect on startup.")
	flag.StringVar(&state.Username, "u", "", "The username which to use.")
	flag.StringVar(&state.Password, "p", "", "The passwort of the user.")
	flag.Parse()

	if state.Remote != "" {
		open(state.Remote)
		if state.Username != "" && state.Password != "" {
			auth(state.Username + " " + state.Password)
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

func determinURL(path string) (*url.URL, error) {
	url, err := url.Parse(path)
	if err != nil {
		url, err = url.Parse(state.Cwd + "/" + path)
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
	case "quit", "exit":
		quit(args)
	case "open":
		open(args)
	case "auth":
		auth(args)
	case "get":
		get(args)
	case "list":
		list(args)
	case "create":
		create(args)
	case "patch":
		patch(args)
	case "delete":
		del(args)
	case "read":
		read(args)
	case "write":
		write(args)
	default:
		println("Unknown command " + cmd)
	}
}

func open(args string) {
	var err error
	if connection, err = fospws.OpenConnection(args); err != nil {
		println(err.Error())
	} else {
		state.Remote = args
		connection.RegisterMessageHandler(e)
	}
}

func quit(args string) {
	os.Exit(0)
}

func auth(args string) {
	parts := strings.Split(args, " ")
	if len(parts) != 2 {
		println("Not enough arguments for authenticate")
	}
	authenticationId := parts[0]
	password := parts[1]
	content := map[string]map[string]string{
		"sasl": map[string]string{
			"mechanism":        "PLAIN",
			"initial-response": strings.Join([]string{"", authenticationId, password}, "\x00"),
		},
	}
	encoded, err := json.Marshal(content)
	if err != nil {
		println("Error while building request " + err.Error())
		return
	}
	req := fosp.NewRequest(fosp.AUTH, nil)
	req.Body = bytes.NewBuffer(encoded)
	if resp, err := connection.SendRequest(req); err == nil && resp.Status == fosp.SUCCEEDED {
		state.Username = parts[0]
		state.Password = parts[0]
		state.Cwd = state.Username + "@" + state.Remote
		buildPrompt()
		println("Authentication succeeded")
	} else {
		println("Authentication failed")
	}
}

func get(args string) {
	url, err := determinURL(args)
	if err != nil {
		println(args + " is not a valid path")
		return
	}
	req := fosp.NewRequest(fosp.GET, url)
	if resp, err := connection.SendRequest(req); err == nil {
		bytes, _ := ioutil.ReadAll(resp.Body)
		println(resp.String())
		println(prettyJSON(bytes))
	} else {
		println("Get failed: " + err.Error())
	}
}

func list(args string) {
	url, err := determinURL(args)
	if err != nil {
		println(args + " is not a valid path")
		return
	}
	req := fosp.NewRequest(fosp.LIST, url)
	if resp, err := connection.SendRequest(req); err == nil {
		bytes, _ := ioutil.ReadAll(resp.Body)
		println(prettyJSON(bytes))
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
	req := fosp.NewRequest(fosp.CREATE, url)
	if content != "" {
		req.Body = bytes.NewBufferString(content)
	}
	if _, err := connection.SendRequest(req); err == nil {
		println("Create succeeded")
	} else {
		println("Create failed: " + err.Error())
	}
}

func patch(args string) {
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
	req := fosp.NewRequest(fosp.PATCH, url)
	if content != "" {
		req.Body = bytes.NewBufferString(content)
	}
	if _, err := connection.SendRequest(req); err == nil {
		println("Patch succeeded")
	} else {
		println("Patch failed: " + err.Error())
	}
}

func del(args string) {
	url, err := determinURL(args)
	if err != nil {
		println(args + " is not a valid path")
		return
	}
	req := fosp.NewRequest(fosp.DELETE, url)
	if _, err := connection.SendRequest(req); err == nil {
		println("Delete succeeded")
	} else {
		println("Delete failed: " + err.Error())
	}
}

func read(args string) {
	tokens := strings.SplitN(args, " ", 2)
	path := tokens[0]
	if len(tokens) != 2 {
		println("A destination filename is required")
		return
	}
	filename := tokens[1]
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		println("Could not create file " + filename)
	}
	url, err := determinURL(path)
	if err != nil {
		println(path + " is not a valid path")
		return
	}
	req := fosp.NewRequest(fosp.READ, url)
	if resp, err := connection.SendRequest(req); err == nil && resp.Status == fosp.SUCCEEDED {
		if _, err = io.Copy(file, resp.Body); err == nil {
			println("Read succeeded")
		} else {
			println("Error when saving file " + err.Error())
		}
	} else {
		if err != nil {
			println("Read failed: " + err.Error())
		} else {
			println("Read failed, received FAILED response")
		}
	}
}

func write(args string) {
	tokens := strings.SplitN(args, " ", 2)
	path := tokens[0]
	if len(tokens) != 2 {
		println("A source filename is required")
		return
	}
	filename := tokens[1]
	file, err := os.Open(filename)
	if err != nil {
		println("Could not read file " + filename)
		return
	}
	url, err := determinURL(path)
	if err != nil {
		println(path + " is not a valid path")
		return
	}
	req := fosp.NewRequest(fosp.WRITE, url)
	req.Body = file
	if _, err := connection.SendRequest(req); err == nil {
		println("Write succeeded")
	} else {
		println("Write failed: " + err.Error())
	}
}

func prettyJSON(in []byte) string {
	var tmp interface{}
	err := json.Unmarshal(in, &tmp)
	if err != nil {
		return string(in)
	}
	var pretty []byte
	pretty, err = json.MarshalIndent(tmp, "", "  ")
	if err != nil {
		return string(in)
	}
	return string(pretty)
}
