package main

// mockTasks — 12 задач для таск-трекера.
func mockTasks() []Task {
	return []Task{
		{ID: 1, Title: "Настроить CI", Priority: "high", Done: true},
		{ID: 2, Title: "Написать README", Priority: "low", Done: false},
		{ID: 3, Title: "Добавить пагинацию", Priority: "medium", Done: false},
		{ID: 4, Title: "Покрыть тестами сервис", Priority: "high", Done: false},
		{ID: 5, Title: "Обновить зависимости", Priority: "low", Done: true},
		{ID: 6, Title: "Починить утечку горутин", Priority: "high", Done: false},
		{ID: 7, Title: "Добавить метрики", Priority: "medium", Done: false},
		{ID: 8, Title: "Ревью PR #42", Priority: "medium", Done: true},
		{ID: 9, Title: "Миграция БД", Priority: "high", Done: false},
		{ID: 10, Title: "Graceful shutdown", Priority: "medium", Done: false},
		{ID: 11, Title: "Rate limiting", Priority: "low", Done: false},
		{ID: 12, Title: "Документация API", Priority: "low", Done: true},
	}
}
