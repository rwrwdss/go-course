package main

// Job — задача на обработку (например, сжать картинку, отправить письмо).
type Job struct {
	ID      int
	Payload string
}

// JobResult — результат обработки.
type JobResult struct {
	JobID  int
	Output string
	Err    error
}

// RunPool обрабатывает jobs с помощью ровно workers воркеров (горутин), которые
// читают задачи из общего канала. Одновременно выполняется НЕ БОЛЬШЕ workers задач.
// Результаты возвращаются отсортированными по JobID.
//
// 🐞 ЗАДАНИЕ 4: сейчас задачи обрабатываются ПОСЛЕДОВАТЕЛЬНО (в 4 раза медленнее, чем нужно),
// а результаты не отсортированы. Перепиши на worker pool:
// канал jobs → workers горутин → канал results → сортировка.
func RunPool(jobs []Job, workers int, process func(Job) JobResult) []JobResult {
	results := make([]JobResult, 0, len(jobs))
	for _, j := range jobs {
		results = append(results, process(j))
	}
	return results
}
