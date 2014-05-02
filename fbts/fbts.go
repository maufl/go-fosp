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
	"github.com/op/go-logging"
	"os"
)

var test, user, password, host string

func main() {
	logging.SetFormatter(logging.MustStringFormatter("[%{time:2006-01-02T15:04} | %{level:.3s} | %{module}]  %{message}"))
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logBackend.Color = true
	logging.SetBackend(logBackend)
	logging.SetLevel(logging.NOTICE, "")

	flag.StringVar(&test, "test", "sanity-check", "Simple sanity check of the server")
	flag.StringVar(&user, "user", "test", "The user name that should be used")
	flag.StringVar(&password, "password", "test", "The password that should be uses")
	flag.StringVar(&host, "host", "localhost.localdomain", "The domain of the host to connect to")
	flag.Parse()

	switch test {
	case "sanity-check":
		testSanityCheck()
	}
}
