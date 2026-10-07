package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()
	return addr
}

func waitUntilUp(t *testing.T, addr string) {
	t.Helper()
	for i := 0; i < 50; i++ {
		resp, err := http.Get("http://" + addr + "/health")
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("сервер не поднялся на %s", addr)
}

func TestRun_GracefulShutdown(t *testing.T) {
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, addr) }()

	waitUntilUp(t, addr)
	cancel() // имитируем Ctrl+C / SIGTERM

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("при штатной остановке run должен вернуть nil, получено %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("run не завершился через 2с после отмены контекста — нет graceful shutdown")
	}

	if _, err := http.Get("http://" + addr + "/health"); err == nil {
		t.Errorf("после остановки сервер всё ещё принимает запросы")
	}
}

func TestRun_PortBusy(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	done := make(chan error, 1)
	go func() { done <- run(context.Background(), l.Addr().String()) }()
	select {
	case err := <-done:
		if err == nil {
			t.Errorf("порт занят — run должен вернуть ошибку")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("порт занят, а run не вернул ошибку за 2с")
	}
}
