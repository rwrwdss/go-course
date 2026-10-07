package repo

import (
	"sync"
	"testing"

	"gocourse/13-final-orders-api/internal/domain"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

func TestOrderRepo_ConcurrentAccess(t *testing.T) {
	r := NewOrderRepo()
	const workers, perWorker = 20, 50

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				o := r.Create(domain.Order{Customer: "load-test", Status: domain.StatusNew})
				o.Status = domain.StatusPaid
				_ = r.Update(o)
				_, _ = r.Get(o.ID)
				_ = r.List()
			}
		}()
	}
	wg.Wait()

	orders := r.List()
	if len(orders) != workers*perWorker {
		t.Fatalf("создано %d заказов, ожидалось %d", len(orders), workers*perWorker)
	}
	seen := map[int]bool{}
	for _, o := range orders {
		if seen[o.ID] {
			t.Fatalf("дубликат ID %d", o.ID)
		}
		seen[o.ID] = true
	}
}
