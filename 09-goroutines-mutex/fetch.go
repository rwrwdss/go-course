package main

// Result — результат опроса одного сервиса.
type Result struct {
	URL    string
	Status int
	Err    error
}

// FetchAll опрашивает все urls ПАРАЛЛЕЛЬНО и возвращает результаты
// В ТОМ ЖЕ ПОРЯДКЕ, что и urls. fetch — функция похода в сеть (в тестах — mock с задержкой).
//
// 🐞 ЗАДАНИЕ 3: функция возвращает пустой/неполный результат, порядок не тот, есть гонка.
// Нужны sync.WaitGroup и запись по индексу.
func FetchAll(urls []string, fetch func(url string) (int, error)) []Result {
	results := make([]Result, 0, len(urls))
	for _, u := range urls {
		go func() {
			status, err := fetch(u)
			results = append(results, Result{URL: u, Status: status, Err: err})
		}()
	}
	return results
}
