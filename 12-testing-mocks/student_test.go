package notify

import "testing"

// ✍️ ЗАДАНИЕ 4: это ТВОЙ файл. Допиши тесты, чтобы покрыть оставшиеся ветки SendDigest
// и поднять покрытие до порога из ci.cfg (90%).
// Используй тестовых двойников из notifier_test.go: at(...), mockUsers(), &spySender{...}.
// Когда тест готов — удали строку t.Skip.

func TestSendDigest_Unsubscribed(t *testing.T) {
	t.Skip("TODO: пользователь 3 отписан → ErrUnsubscribed, писем 0")
}

func TestSendDigest_UserNotFound(t *testing.T) {
	t.Skip("TODO: пользователь 404 → errors.Is(err, errUserNotFound)")
}

func TestSendDigest_SenderError(t *testing.T) {
	t.Skip("TODO: spySender{err: ...} → ошибка обёрнута, errors.Is находит исходную")
}

func TestSendDigest_PluralInBody(t *testing.T) {
	t.Skip("TODO: пользователь 2 (21 заказ) → в теле письма \"У вас 21 заказ.\"")
}

func TestRealClock(t *testing.T) {
	t.Skip("TODO: RealClock{}.Now() близко к time.Now() (разница < 1 секунды)")
}
