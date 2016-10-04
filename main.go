package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting journlr...")

	r := mux.Router()

	r.Handle("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Hack Sonoma"))
	}).Method("GET")

	http.ListenAndServe(":6891", handlers.LoggingHandler(os.Stdout, r))

}
