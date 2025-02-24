package main

import (
	"github.com/primalivet/mail/internal/server"
	"log"
)

func main() {
	server := server.New("localhost", 2525)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
