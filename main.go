package main

import (
	"github.com/taalhach/restcron/server"
	"log"
)

const (
	addr  = ":5432" //postgres db url
	user = "postgres" // db username
	password = "postgres" // db password
)
func main() {
	s, err := server.NewServer(addr,user, password)
	if err != nil {
		log.Fatal(err)
	}
	s.RunServer()
}
