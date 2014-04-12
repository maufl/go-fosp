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
	"bitbucket.org/maufl/go-fosp/fosp"
	"encoding/json"
	"flag"
	"github.com/op/go-logging"
	"net/http"
	"os"
)

var lg = logging.MustGetLogger("go-fosp/fospd")

type config struct {
	Localdomain string            `json:"localdomain"`
	Listen      string            `json:"listen"`
	Database    string            `json:"database"`
	BasePath    string            `json:"basepath"`
	Logging     map[string]string `json:"logging"`
}

func main() {
	logging.SetFormatter(logging.MustStringFormatter("[%{time:2006-01-02T15:04} | %{level:.3s} | %{module}]  %{message}"))
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logBackend.Color = true
	logging.SetBackend(logBackend)
	configFile := flag.String("c", "config.json", "A configuration file in json format")
	flag.Parse()
	file, err := os.Open(*configFile)
	if err != nil {
		println("Config file not found")
		return
	}
	decoder := json.NewDecoder(file)
	conf := &config{}
	err = decoder.Decode(conf)
	if err != nil {
		println("Failed to read config file: " + err.Error())
		return
	}
	for module, level := range conf.Logging {
		if iLevel, err := logging.LogLevel(level); err == nil {
			logging.SetLevel(iLevel, module)
		} else {
			println("Unrecognized log level " + level)
		}
	}
	lg.Debug("Configuration %v+", conf)

	driver := fosp.NewPostgresqlDriver(conf.Database, conf.BasePath)
	server := fosp.NewServer(driver, conf.Localdomain)
	http.HandleFunc("/", server.RequestHandler)
	lg.Info("Listening on address %s", conf.Listen)
	if err := http.ListenAndServe(conf.Listen, nil); err != nil {
		lg.Fatalf("Failed to listen on address %s: %s", conf.Listen, err)
	}
}
