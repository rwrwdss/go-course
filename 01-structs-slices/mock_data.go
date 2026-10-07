package main

// mockUsers — тестовые данные, как будто пришли из базы.
func mockUsers() []User {
	return []User{
		{ID: 1, Name: "Анна", Email: "anna@example.com", Age: 28, IsActive: true, Balance: 1500},
		{ID: 2, Name: "Борис", Email: "boris@example.com", Age: 17, IsActive: true, Balance: 300},
		{ID: 3, Name: "Виктор", Email: "victor@example.com", Age: 35, IsActive: false, Balance: 0},
		{ID: 4, Name: "Галина", Email: "galina@example.com", Age: 42, IsActive: true, Balance: 9800},
		{ID: 5, Name: "Дмитрий", Email: "dmitry@example.com", Age: 16, IsActive: false, Balance: 50},
		{ID: 6, Name: "Елена", Email: "elena@example.com", Age: 23, IsActive: true, Balance: 720},
		{ID: 7, Name: "Жанна", Email: "zhanna@example.com", Age: 31, IsActive: true, Balance: 4100},
	}
}
