package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// WeatherHandler — GET /weather?city=moscow
type WeatherHandler struct {
	client  WeatherClient
	timeout time.Duration // сколько мы готовы ждать внешний API
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ServeHTTP обрабатывает запрос погоды.
//
//	нет ?city=                      → 400 {"error":"city is required"}
//	ErrCityNotFound                 → 404 {"error":"city not found"}
//	внешний API не уложился в timeout → 504 {"error":"upstream timeout"}
//	прочие ошибки                   → 500 {"error":"internal error"}
//	успех                           → 200 {"city":"moscow","temperature":-3.5}
//
// 🐞 ЗАДАНИЕ 1: пайплайн падает на стадии vet (lostcancel). Разберись, почему cancel обязателен.
// 🐞 ЗАДАНИЕ 2: если клиент оборвал соединение, мы продолжаем ждать внешний API впустую.
// Контекст должен наследоваться от контекста ЗАПРОСА, а не от context.Background().
// 🐞 ЗАДАНИЕ 3: все ошибки превращаются в 500, а нет city — вообще не проверяется.
func (h *WeatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")

	ctx, _ := context.WithTimeout(context.Background(), h.timeout)
	temp, err := h.client.Temperature(ctx, city)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"city": city, "temperature": temp})
}
