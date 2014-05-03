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

var test, host string

func main() {
	logging.SetFormatter(logging.MustStringFormatter("[%{time:2006-01-02T15:04} | %{level:.3s} | %{module}]  %{message}"))
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logBackend.Color = true
	logging.SetBackend(logBackend)
	logging.SetLevel(logging.NOTICE, "")

	flag.StringVar(&test, "test", "", "The test that should be run. Possible values are: sanity-check")
	flag.StringVar(&host, "host", "localhost.localdomain", "The domain of the host to connect to")
	flag.Parse()

	success := false
	switch test {
	case "sanity-check":
		success = testSanityCheck()
	default:
		flag.PrintDefaults()
	}
	if success {
		println("Test succeeded")
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
