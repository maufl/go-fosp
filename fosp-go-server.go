package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type config struct {
	Localdomain string `json:"localdomain"`
	Database    string `json:"database"`
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
	driver := NewPostgresqlDriver("postgres://fosp:fosp@localhost/fosp?sslmode=disable")
	server := NewServer(driver, conf.Localdomain)
	http.HandleFunc("/", server.requestHandler)
	if err := http.ListenAndServe(":1337", nil); err != nil {
		log.Fatal("Failed to listen on address :8080 :: ", err)
	}
}
