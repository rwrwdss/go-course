package main

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

// collect читает канал до закрытия, но не дольше 1 секунды — чтобы тест не висел вечно.
func collect(t *testing.T, ch <-chan int) []int {
	t.Helper()
	out := []int{}
	deadline := time.After(time.Second)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, v)
		case <-deadline:
			t.Fatalf("канал не закрылся за 1с (получено %v) — кто-то забыл close() или горутина зависла", out)
			return nil
		}
	}
}

func TestGenerate_ClosesChannel(t *testing.T) {
	if got := collect(t, Generate(1, 2, 3)); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Generate = %v, ожидалось [1 2 3]", got)
	}
}

func TestSquare(t *testing.T) {
	if got := collect(t, Square(Generate(1, 2, 3, 4))); !slices.Equal(got, []int{1, 4, 9, 16}) {
		t.Errorf("Square = %v, ожидалось [1 4 9 16]", got)
	}
}

func TestMerge(t *testing.T) {
	got := collect(t, Merge(Generate(1, 2, 3), Generate(10, 20), Generate(), Generate(100)))
	slices.Sort(got)
	if want := []int{1, 2, 3, 10, 20, 100}; !slices.Equal(got, want) {
		t.Errorf("Merge = %v, ожидалось (в любом порядке) %v", got, want)
	}
}

func TestRunPool(t *testing.T) {
	const workers = 4
	jobs := make([]Job, 24)
	for i := range jobs {
		jobs[i] = Job{ID: len(jobs) - i, Payload: fmt.Sprintf("job-%d", len(jobs)-i)} // ID в обратном порядке
	}

	var running, maxRunning atomic.Int32
	start := time.Now()
	results := RunPool(jobs, workers, func(j Job) JobResult {
		now := running.Add(1)
		for {
			prev := maxRunning.Load()
			if now <= prev || maxRunning.CompareAndSwap(prev, now) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		running.Add(-1)
		if j.ID == 5 {
			return JobResult{JobID: j.ID, Err: errors.New("corrupted file")}
		}
		return JobResult{JobID: j.ID, Output: "done:" + j.Payload}
	})
	elapsed := time.Since(start)

	if len(results) != len(jobs) {
		t.Fatalf("получено %d результатов, ожидалось %d", len(results), len(jobs))
	}
	for i, r := range results {
		if r.JobID != i+1 {
			t.Fatalf("results[%d].JobID = %d — результаты должны быть отсортированы по JobID", i, r.JobID)
		}
	}
	if results[4].Err == nil || results[0].Output != "done:job-1" {
		t.Errorf("результаты обработки потеряны: %+v, %+v", results[0], results[4])
	}
	if m := maxRunning.Load(); m > workers {
		t.Errorf("одновременно выполнялось %d задач, а воркеров всего %d", m, workers)
	} else if m < 2 {
		t.Errorf("задачи выполнялись строго по одной (max=%d) — пул не параллелит работу", m)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("24 задачи по 30мс на 4 воркерах заняли %v (ожидалось ~180мс)", elapsed)
	}
}

func sleeper(d time.Duration, answer string) func() string {
	return func() string {
		time.Sleep(d)
		return answer
	}
}

func TestFirstResponse_Fastest(t *testing.T) {
	start := time.Now()
	got, err := FirstResponse([]func() string{
		sleeper(300*time.Millisecond, "slow"),
		sleeper(30*time.Millisecond, "fast"),
		sleeper(150*time.Millisecond, "medium"),
	}, time.Second)
	if err != nil || got != "fast" {
		t.Fatalf("FirstResponse = (%q, %v), ожидалось (\"fast\", nil)", got, err)
	}
	if d := time.Since(start); d > 150*time.Millisecond {
		t.Errorf("ответ получен через %v — нужно возвращать первый ответ, не дожидаясь остальных", d)
	}
}

func TestFirstResponse_Timeout(t *testing.T) {
	start := time.Now()
	_, err := FirstResponse([]func() string{
		sleeper(400*time.Millisecond, "a"),
		sleeper(400*time.Millisecond, "b"),
	}, 50*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("ожидалась ErrTimeout, получено %v", err)
	}
	if d := time.Since(start); d > 200*time.Millisecond {
		t.Errorf("таймаут 50мс, а функция вернулась через %v", d)
	}
}

func TestFirstResponse_NoGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	_, _ = FirstResponse([]func() string{
		sleeper(10*time.Millisecond, "a"),
		sleeper(60*time.Millisecond, "b"),
		sleeper(60*time.Millisecond, "c"),
	}, time.Second)
	time.Sleep(200 * time.Millisecond) // все реплики уже точно ответили
	if after := runtime.NumGoroutine(); after > before+1 {
		t.Errorf("утечка горутин: было %d, стало %d — медленные реплики заблокированы на отправке в канал", before, after)
	}
}
