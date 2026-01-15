package main

import (
	"log"
	"net/http"
	"url-shortener/shortener"
)

func main() {
	s := shortener.New()
	http.HandleFunc("/", s.Handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
