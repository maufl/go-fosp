package main

import (
  "net/http"
  "log"
)

func main() {
  var driver = new(PostgresqlDriver)
  var database = new(Database)
  driver.open()
  database.driver = driver
  var server = server{driver, database, make(map[string][]*connection), "localhost.localdomain"}
  database.server = &server
  http.HandleFunc("/", server.requestHandler)
  if err := http.ListenAndServe(":1337", nil); err != nil {
    log.Fatal("Failed to listen on address :8080 :: ", err)
  }
}
