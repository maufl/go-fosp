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
	"encoding/json"
	"flag"
	"github.com/op/go-logging"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
)

var lg = logging.MustGetLogger("go-fosp/fospd")

type config struct {
	Localdomain  string            `json:"localdomain"`
	Listen       string            `json:"listen"`
	ListenSecure string            `json:"listensecure"`
	Database     string            `json:"database"`
	BasePath     string            `json:"basepath"`
	Logging      map[string]string `json:"logging"`
	Key          string            `json:"keyfile"`
	Certificate  string            `json:"certfile"`
}

func main() {
	logging.SetFormatter(logging.MustStringFormatter("[%{time:2006-01-02T15:04} | %{level:.3s} | %{shortfile}]  %{message}"))
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logBackend.Color = true
	logging.SetBackend(logBackend)
	logging.SetLevel(logging.DEBUG, "")
	configFile := flag.String("c", "config.json", "A configuration file in json format")
	cpuprofile := flag.String("cpuprofile", "", "Write cpu profile to file")
	flag.Parse()
	file, err := os.Open(*configFile)
	if err != nil {
		lg.Fatalf("Config file not found: %s", *configFile)
	}
	decoder := json.NewDecoder(file)
	conf := &config{}
	err = decoder.Decode(conf)
	if err != nil {
		lg.Fatalf("Failed to read config file: %s", err.Error())
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			lg.Fatal("Error opening cpuprofile file :: %s", err)
		}
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				if sig == os.Interrupt {
					pprof.StopCPUProfile()
					f.Close()
					os.Exit(1)
				}
			}
		}()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		defer f.Close()
	}
	for module, level := range conf.Logging {
		if iLevel, err := logging.LogLevel(level); err == nil {
			lg.Info("Setting log level of module %s to %s", module, level)
			logging.SetLevel(iLevel, module)
		} else {
			lg.Warning("Unrecognized log level %s", level)
		}
	}
	lg.Debug("Configuration %v+", conf)

	driver := NewPostgresqlDriver(conf.Database, conf.BasePath)
	server := NewServer(driver, conf.Localdomain)
	http.HandleFunc("/", server.RequestHandler)
	lg.Info("Serving domain %s", conf.Localdomain)
	ch := make(chan bool)
	go func() {
		if conf.Listen != "" {
			lg.Info("Listening with http on %s", conf.Listen)
			err = http.ListenAndServe(conf.Listen, nil)
			if err != nil {
				lg.Fatalf("Failed to listen on address %s: %s", conf.Listen, err)
			}
		}
		ch <- true
	}()
	go func() {
		if conf.Key != "" && conf.Certificate != "" && conf.ListenSecure != "" {
			lg.Info("Listening with https on %s", conf.ListenSecure)
			err = http.ListenAndServeTLS(conf.ListenSecure, conf.Certificate, conf.Key, nil)
			if err != nil {
				lg.Fatalf("Failed to listen on address %s: %s", conf.ListenSecure, err)
			}
		}
		ch <- true
	}()
	<-ch
	<-ch
}
