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
