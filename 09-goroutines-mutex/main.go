package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	counter := NewViewCounter()
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc([]string{"/", "/about", "/blog"}[i%3])
		}()
	}
	wg.Wait()
	fmt.Println("Просмотры:", counter.Snapshot())

	start := time.Now()
	urls := []string{"https://api.users", "https://api.orders", "https://api.payments", "https://api.search"}
	results := FetchAll(urls, func(url string) (int, error) {
		time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
		return 200, nil
	})
	fmt.Printf("Опрошено %d сервисов за %v\n", len(results), time.Since(start).Round(time.Millisecond))

	nums := make([]int, 1_000_000)
	for i := range nums {
		nums[i] = i + 1
	}
	fmt.Println("Сумма 1..1_000_000 =", SumParallel(nums, 8))
}
