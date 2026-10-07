package main

// ViewCounter считает просмотры страниц. Inc вызывается из сотен горутин одновременно
// (каждый HTTP-запрос — своя горутина).
//
// 🐞 ЗАДАНИЕ 1: при конкурентном доступе — data race / "fatal error: concurrent map writes".
// Защити данные мьютексом (sync.Mutex или sync.RWMutex).
type ViewCounter struct {
	views map[string]int
}

func NewViewCounter() *ViewCounter {
	return &ViewCounter{views: make(map[string]int)}
}

func (c *ViewCounter) Inc(page string) {
	c.views[page]++
}

func (c *ViewCounter) Get(page string) int {
	return c.views[page]
}

// Snapshot возвращает КОПИЮ текущих счётчиков (для отчёта).
// 🐞 ЗАДАНИЕ 2: сейчас возвращается внутренняя мапа — вызывающий код может её испортить,
// и читать её без блокировки небезопасно.
func (c *ViewCounter) Snapshot() map[string]int {
	return c.views
}
