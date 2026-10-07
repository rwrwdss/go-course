package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"gocourse/13-final-orders-api/internal/httpapi"
	"gocourse/13-final-orders-api/internal/repo"
	"gocourse/13-final-orders-api/internal/service"
)

// shutdownTimeout — сколько ждём завершения активных запросов при остановке.
const shutdownTimeout = 5 * time.Second

// newApp собирает приложение из слоёв (composition root): repo → service → http.
func newApp() http.Handler {
	catalog := repo.NewCatalog(mockCatalog())
	orders := repo.NewOrderRepo()
	svc := service.NewOrderService(catalog, orders)
	return httpapi.NewHandler(svc)
}

// run запускает HTTP-сервер и блокируется, пока ctx не будет отменён (Ctrl+C / SIGTERM).
// После отмены — graceful shutdown: перестаём принимать новые соединения, ждём
// завершения текущих запросов (но не дольше shutdownTimeout) и возвращаем nil.
// Если сервер не смог запуститься (например, порт занят) — вернуть ошибку.
//
// 🐞 ЗАДАНИЕ 5: сейчас run игнорирует ctx и никогда не возвращается — сервер нельзя
// корректно остановить, а при деплое активные запросы обрываются на полуслове.
func run(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: newApp()}
	log.Printf("🛒 Orders API слушает %s", addr)
	return srv.ListenAndServe()
}
