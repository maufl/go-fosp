package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"bitbucket.org/maufl/go-fosp/fosp"
)

type config struct {
	Localdomain string `json:"localdomain"`
	Database    string `json:"database"`
	BasePath    string `json:"basepath"`
}

func main() {
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
	fmt.Println("%v+", conf)
	driver := fosp.NewPostgresqlDriver(conf.Database, conf.BasePath)
	server := fosp.NewServer(driver, conf.Localdomain)
	http.HandleFunc("/", server.RequestHandler)
	if err := http.ListenAndServe(":1337", nil); err != nil {
		log.Fatal("Failed to listen on address :8080 :: ", err)
	}
}
