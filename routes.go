package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/calvincolton/gorilla-sockets/handlers"
)

func routes() http.Handler {
	mux := pat.New()
	fileServer := http.FileServer(http.Dir("./static/"))

	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return mux
}
