package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	fmt.Print("Квадраты: ")
	for v := range Square(Generate(1, 2, 3, 4, 5)) {
		fmt.Print(v, " ")
	}
	fmt.Println()

	sum := 0
	for v := range Merge(Generate(1, 2, 3), Generate(10, 20), Generate(100)) {
		sum += v
	}
	fmt.Println("Сумма после Merge:", sum)

	jobs := make([]Job, 12)
	for i := range jobs {
		jobs[i] = Job{ID: i + 1, Payload: fmt.Sprintf("image-%02d.png", i+1)}
	}
	start := time.Now()
	results := RunPool(jobs, 4, func(j Job) JobResult {
		time.Sleep(100 * time.Millisecond) // «тяжёлая» обработка
		return JobResult{JobID: j.ID, Output: strings.Replace(j.Payload, ".png", ".webp", 1)}
	})
	fmt.Printf("Обработано %d картинок за %v (последовательно было бы 1.2s)\n", len(results), time.Since(start).Round(10*time.Millisecond))

	resp, err := FirstResponse([]func() string{
		func() string { time.Sleep(300 * time.Millisecond); return "replica-eu" },
		func() string { time.Sleep(50 * time.Millisecond); return "replica-ru" },
		func() string { time.Sleep(200 * time.Millisecond); return "replica-us" },
	}, time.Second)
	fmt.Println("Самая быстрая реплика:", resp, err)
}
