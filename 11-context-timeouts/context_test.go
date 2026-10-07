package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

func TestWeather_Responses(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantCode int
		wantBody string
	}{
		{"успех", "/weather?city=moscow", 200, `{"city":"moscow","temperature":-3.5}`},
		{"нет параметра city", "/weather", 400, `{"error":"city is required"}`},
		{"неизвестный город", "/weather?city=atlantis", 404, `{"error":"city not found"}`},
		{"внешний API тормозит", "/weather?city=tokyo", 504, `{"error":"upstream timeout"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &WeatherHandler{client: &mockWeatherAPI{}, timeout: 100 * time.Millisecond}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest("GET", tt.url, nil))
			if rec.Code != tt.wantCode {
				t.Errorf("статус = %d, ожидалось %d", rec.Code, tt.wantCode)
			}
			var got, want any
			_ = json.Unmarshal(rec.Body.Bytes(), &got)
			_ = json.Unmarshal([]byte(tt.wantBody), &want)
			if gb, _ := json.Marshal(got); string(gb) != mustMarshal(want) {
				t.Errorf("тело = %s, ожидалось %s", strings.TrimSpace(rec.Body.String()), tt.wantBody)
			}
		})
	}
}

func mustMarshal(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestWeather_TimeoutIsRespected(t *testing.T) {
	h := &WeatherHandler{client: &mockWeatherAPI{}, timeout: 100 * time.Millisecond}
	start := time.Now()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/weather?city=tokyo", nil))
	if d := time.Since(start); d > 400*time.Millisecond {
		t.Errorf("таймаут 100мс, а хендлер работал %v", d)
	}
}

func TestWeather_ClientDisconnectCancelsUpstream(t *testing.T) {
	api := &mockWeatherAPI{}
	h := &WeatherHandler{client: api, timeout: 5 * time.Second}

	// Клиент «уходит» через 30мс (закрыл вкладку / оборвал соединение).
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest("GET", "/weather?city=tokyo", nil).WithContext(ctx)

	start := time.Now()
	h.ServeHTTP(httptest.NewRecorder(), req)
	if d := time.Since(start); d > 500*time.Millisecond {
		t.Errorf("клиент ушёл через 30мс, а хендлер продолжал ждать внешний API %v — используй r.Context()", d)
	}
	if api.aborted.Load() != 1 {
		t.Errorf("запрос к внешнему API должен быть прерван отменой контекста")
	}
}

func TestProcessOrders_AllProcessed(t *testing.T) {
	var seen []int
	n, err := ProcessOrders(context.Background(), []int{1, 2, 3}, func(ctx context.Context, id int) error {
		seen = append(seen, id)
		return nil
	})
	if n != 3 || err != nil || len(seen) != 3 {
		t.Errorf("ProcessOrders = (%d, %v), обработаны %v; ожидалось (3, nil)", n, err, seen)
	}
}

func TestProcessOrders_StopsOnCancel(t *testing.T) {
	ids := make([]int, 100)
	for i := range ids {
		ids[i] = i + 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(55*time.Millisecond, cancel)

	start := time.Now()
	n, err := ProcessOrders(ctx, ids, func(ctx context.Context, id int) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("ожидалась ошибка context.Canceled, получено %v", err)
	}
	if n < 3 || n > 8 {
		t.Errorf("обработано %d заказов, ожидалось около 5 (остановка через ~55мс)", n)
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("после отмены обработка продолжалась: всего %v", elapsed)
	}
}

func TestProcessOrders_StopsOnError(t *testing.T) {
	errPayment := errors.New("payment declined")
	n, err := ProcessOrders(context.Background(), []int{1, 2, 3, 4}, func(ctx context.Context, id int) error {
		if id == 3 {
			return errPayment
		}
		return nil
	})
	if !errors.Is(err, errPayment) || n != 2 {
		t.Errorf("ProcessOrders = (%d, %v), ожидалось (2, payment declined)", n, err)
	}
}
