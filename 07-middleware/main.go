package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "[http] ", log.LstdFlags)
	addr := ":8080"
	logger.Printf("🛡  Middleware demo слушает http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, NewApp(logger)))
}
