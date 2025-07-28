package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Gamer server is running")
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
