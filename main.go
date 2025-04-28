package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func pong(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "pong")
}

func main() {
	http.HandleFunc("/", pong)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("⇨  listening on :%s /", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
