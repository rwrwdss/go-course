package main

import "context"

// ProcessOrders обрабатывает заказы по очереди. Если ctx отменён (сервер выключается,
// клиент ушёл, истёк дедлайн) — НЕМЕДЛЕННО прекращает работу и возвращает
// количество уже обработанных заказов и ctx.Err().
// Если process вернул ошибку — тоже остановиться и вернуть её.
//
// ✍️ ЗАДАНИЕ 4: сейчас контекст игнорируется, и ошибки process тоже.
func ProcessOrders(ctx context.Context, orderIDs []int, process func(ctx context.Context, id int) error) (int, error) {
	processed := 0
	for _, id := range orderIDs {
		_ = process(ctx, id)
		processed++
	}
	return processed, nil
}
