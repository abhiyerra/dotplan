package main

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func startHTTP() {
	r := mux.Router()

	r.Handle("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write("Hello Hack Sonoma")
	}).Method("GET")

	http.ListenAndServe(":6891", handlers.LoggingHandler(os.Stdout, r))
}
