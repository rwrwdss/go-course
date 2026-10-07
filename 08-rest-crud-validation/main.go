package main

import (
	"log"
	"net/http"
)

func main() {
	addr := ":8080"
	log.Printf("✅ Task API слушает http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, NewAPI(NewStore(mockTasks()))))
}
