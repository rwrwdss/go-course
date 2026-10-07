package main

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// ErrCityNotFound — внешний API не знает такого города.
var ErrCityNotFound = errors.New("city not found")

// WeatherClient — клиент внешнего погодного API.
type WeatherClient interface {
	Temperature(ctx context.Context, city string) (float64, error)
}

// mockWeatherAPI имитирует внешний сервис: у каждого города своя задержка ответа.
// Он «честно» уважает контекст: если ctx отменён — прекращает работу и возвращает ctx.Err().
type mockWeatherAPI struct {
	aborted atomic.Int32 // сколько запросов было прервано контекстом (для тестов)
}

var cityData = map[string]struct {
	temp  float64
	delay time.Duration
}{
	"moscow": {temp: -3.5, delay: 20 * time.Millisecond},
	"london": {temp: 11.2, delay: 50 * time.Millisecond},
	"tokyo":  {temp: 18.9, delay: 2 * time.Second}, // очень медленный регион
}

func (m *mockWeatherAPI) Temperature(ctx context.Context, city string) (float64, error) {
	data, ok := cityData[city]
	if !ok {
		return 0, ErrCityNotFound
	}
	select {
	case <-time.After(data.delay):
		return data.temp, nil
	case <-ctx.Done():
		m.aborted.Add(1)
		return 0, ctx.Err()
	}
}
