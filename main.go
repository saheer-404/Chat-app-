package main

import (
	"log"
	"net/http"

	"github.com/calvincolton/gorilla-sockets/handlers"
)

func main() {
	mux := routes()

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Gorilla Sockets! Starting web server on port 8080")

	_ = http.ListenAndServe(":8080", mux)
}
