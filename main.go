package main

import (
	"github.com/taalhach/restcron/server"
	"log"
)

func main() {
	s, err := server.NewServer("postgres", "postgres")
	if err != nil {
		log.Fatal(err)
	}
	s.RunServer()
}
