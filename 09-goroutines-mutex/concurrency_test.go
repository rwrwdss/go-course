package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ⚠️ Автотесты пайплайна. Запускаются с флагом -race (детектор гонок). Не меняй их.

func TestViewCounter_Concurrent(t *testing.T) {
	c := NewViewCounter()
	pages := []string{"/", "/about", "/blog"}
	var wg sync.WaitGroup
	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 300; i++ {
				c.Inc(pages[i%3])
				_ = c.Get("/")
			}
		}()
	}
	wg.Wait()
	for _, p := range pages {
		if got := c.Get(p); got != 5000 {
			t.Errorf("просмотров %s = %d, ожидалось 5000", p, got)
		}
	}
}

func TestViewCounter_SnapshotIsCopy(t *testing.T) {
	c := NewViewCounter()
	c.Inc("/")
	snap := c.Snapshot()
	snap["/"] = 999
	snap["/hack"] = 1
	if c.Get("/") != 1 || c.Get("/hack") != 0 {
		t.Errorf("изменение снапшота испортило счётчик: / = %d, /hack = %d", c.Get("/"), c.Get("/hack"))
	}
}

func TestFetchAll_OrderAndCompleteness(t *testing.T) {
	urls := make([]string, 20)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://svc-%d", i)
	}
	errDown := errors.New("service down")
	results := FetchAll(urls, func(url string) (int, error) {
		time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
		if url == "https://svc-7" {
			return 0, errDown
		}
		return 200, nil
	})
	if len(results) != len(urls) {
		t.Fatalf("получено %d результатов, ожидалось %d (ждёшь ли ты завершения горутин?)", len(results), len(urls))
	}
	for i, r := range results {
		if r.URL != urls[i] {
			t.Fatalf("results[%d].URL = %q, ожидалось %q — порядок должен совпадать с входным", i, r.URL, urls[i])
		}
	}
	if !errors.Is(results[7].Err, errDown) || results[0].Status != 200 {
		t.Errorf("ошибки и статусы должны сохраниться: %+v, %+v", results[0], results[7])
	}
}

func TestFetchAll_IsParallel(t *testing.T) {
	urls := make([]string, 10)
	for i := range urls {
		urls[i] = fmt.Sprintf("u%d", i)
	}
	start := time.Now()
	results := FetchAll(urls, func(string) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 200, nil
	})
	elapsed := time.Since(start)
	if len(results) != 10 {
		t.Fatalf("получено %d результатов, ожидалось 10", len(results))
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("10 запросов по 100мс заняли %v — похоже, они выполняются последовательно", elapsed)
	}
}

func TestSumParallel(t *testing.T) {
	nums := make([]int, 100_000)
	for i := range nums {
		nums[i] = i + 1
	}
	want := 100_000 * 100_001 / 2
	for _, workers := range []int{1, 3, 4, 7, 16, 0, -1, 200_000} {
		t.Run(fmt.Sprintf("workers=%d", workers), func(t *testing.T) {
			if got := SumParallel(nums, workers); got != want {
				t.Errorf("SumParallel(workers=%d) = %d, ожидалось %d", workers, got, want)
			}
		})
	}
	t.Run("пустой слайс", func(t *testing.T) {
		if got := SumParallel(nil, 4); got != 0 {
			t.Errorf("SumParallel(nil) = %d, ожидалось 0", got)
		}
	})
}

func TestConfigProvider_LoadsOnce(t *testing.T) {
	var calls atomic.Int32
	p := NewConfigProvider(func() Config {
		calls.Add(1)
		time.Sleep(20 * time.Millisecond)
		return Config{DSN: "postgres://localhost/app", LogLevel: "info"}
	})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cfg := p.Get(); cfg.DSN != "postgres://localhost/app" {
				t.Errorf("получен пустой/неполный конфиг: %+v", cfg)
			}
		}()
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Errorf("load вызван %d раз, ожидался ровно 1", n)
	}
}
