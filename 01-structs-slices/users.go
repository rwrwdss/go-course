package main

import "slices"

// User — пользователь нашего сервиса.
type User struct {
	ID       int
	Name     string
	Email    string
	Age      int
	IsActive bool
	Balance  int // в рублях
}

// ActiveUsers возвращает только активных пользователей (IsActive == true).
// 🐞 ЗАДАНИЕ 1: функция возвращает не тех пользователей. Найди и исправь баг.
func ActiveUsers(users []User) []User {
	var result []User
	for _, u := range users {
		if u.IsActive == true {
			result = append(result, u)
		}
	}
	return result
}

// FindByID ищет пользователя по ID и возвращает УКАЗАТЕЛЬ на элемент слайса,
// чтобы вызывающий код мог изменить пользователя прямо в слайсе.
// 🐞 ЗАДАНИЕ 2: изменения через возвращённый указатель не попадают в слайс. Почему?
func FindByID(users []User, id int) (*User, bool) {
	for i := range users {
		if users[i].ID == id {
			return &users[i], true
		}
	}
	return nil, false
}

// ApplyBonus начисляет бонус на баланс всем АКТИВНЫМ пользователям (меняет слайс на месте).
// 🐞 ЗАДАНИЕ 3: после вызова балансы не меняются. Исправь.
func ApplyBonus(users []User, bonus int) {
	for i := range users {
		if users[i].IsActive == true {
			users[i].Balance += bonus
		}
	}
}

// AdultNames возвращает имена совершеннолетних (Age >= 18), отсортированные по алфавиту.
// ✍️ ЗАДАНИЕ 4: допиши функцию.
func AdultNames(users []User) []string {
	// TODO: собери имена, где Age >= 18, и отсортируй их (пакет sort или slices)
	var tempUsers []string
	for i := range users {
		if users[i].Age >= 18 {
			tempUsers = append(tempUsers, users[i].Name)

		}
	}
	slices.Sort(tempUsers)
	return tempUsers
}

// Paginate возвращает страницу пользователей.
// page начинается с 1, size — размер страницы.
// Если page < 1, size < 1 или страница за пределами данных — вернуть пустой (НЕ nil) слайс []User{}.
// ✍️ ЗАДАНИЕ 5: допиши функцию. Не забудь про последнюю неполную страницу!
func Paginate(users []User, page, size int) []User {
	if page < 1 || size < 1 {
		return []User{}
	}
	start := (page - 1) * size

	if start >= len(users) {
		return []User{}
	}

	end := start + size

	if end > len(users) {
		end = len(users)
	}
	return users[start:end]
}
