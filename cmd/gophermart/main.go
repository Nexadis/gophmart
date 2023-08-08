package main

import (
	"log"

	"github.com/Nexadis/gophmart/internal/server"
)

func main() {
	config := server.NewConfig()
	config.Parse()
	s := server.New(config)
	log.Fatal(s.Run())
}
