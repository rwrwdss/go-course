package main

// Статусы заказа.
const (
	StatusPaid      = "paid"
	StatusPending   = "pending"
	StatusCancelled = "cancelled"
)

// Order — заказ. Amount хранится в рублях целым числом.
type Order struct {
	ID       int
	Customer string
	Category string
	Amount   int
	Status   string
}

// TotalByCustomer считает сумму ОПЛАЧЕННЫХ (StatusPaid) заказов по каждому клиенту.
// 🐞 ЗАДАНИЕ 1: функция падает с паникой. Найди причину и исправь.
func TotalByCustomer(orders []Order) map[string]int {
	var totals map[string]int
	for _, o := range orders {
		if o.Status != StatusPaid {
			continue
		}
		totals[o.Customer] += o.Amount
	}
	return totals
}

// CountByStatus возвращает количество заказов в каждом статусе.
// ✍️ ЗАДАНИЕ 2: допиши функцию.
func CountByStatus(orders []Order) map[string]int {
	// TODO
	return nil
}

// UniqueCategories возвращает список категорий БЕЗ повторов, отсортированный по алфавиту.
// 🐞 ЗАДАНИЕ 3: в результате есть дубликаты. Используй map как множество (set).
func UniqueCategories(orders []Order) []string {
	var result []string
	for _, o := range orders {
		result = append(result, o.Category)
	}
	return result
}

// TopCustomers возвращает имена n клиентов с наибольшей суммой.
// Порядок: по убыванию суммы; при равной сумме — по имени по алфавиту.
// Если n больше числа клиентов — вернуть всех.
// ✍️ ЗАДАНИЕ 4: допиши функцию. Помни: порядок обхода map в Go СЛУЧАЙНЫЙ.
func TopCustomers(totals map[string]int, n int) []string {
	// TODO
	return nil
}

// HasCustomer сообщает, есть ли клиент в мапе totals.
// 🐞 ЗАДАНИЕ 5: клиент с нулевой суммой считается «отсутствующим». Используй форму «comma ok».
func HasCustomer(totals map[string]int, name string) bool {
	return totals[name] != 0
}
