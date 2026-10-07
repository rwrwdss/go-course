package main

import (
	"errors"
	"time"
)

// ErrTimeout — ни одна реплика не ответила вовремя.
var ErrTimeout = errors.New("timeout: no replica responded")

// FirstResponse опрашивает ВСЕ реплики одновременно и возвращает ответ самой быстрой.
// Если за timeout никто не ответил — ErrTimeout.
// Важно: горутины медленных реплик не должны «зависнуть навсегда» (утечка горутин).
//
// ✍️ ЗАДАНИЕ 5: реализуй через select + time.After. Подумай, каким должен быть буфер канала.
func FirstResponse(replicas []func() string, timeout time.Duration) (string, error) {
	// TODO
	return "", nil
}
