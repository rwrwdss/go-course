package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /weather", &WeatherHandler{client: &mockWeatherAPI{}, timeout: 500 * time.Millisecond})

	addr := ":8080"
	log.Printf("🌦  Weather API слушает http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
