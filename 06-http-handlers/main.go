package main

import (
	"log"
	"net/http"
)

func main() {
	addr := ":8080"
	log.Printf("📚 Library API слушает http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, NewServer(mockBooks())))
}
