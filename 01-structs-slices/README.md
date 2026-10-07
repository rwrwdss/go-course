# Модуль 01 · Структуры, слайсы, указатели

## 🎯 Цель
Разобраться, как устроены структуры и слайсы в Go и **когда работаешь с копией, а когда с оригиналом**. На этом спотыкается почти каждый новичок, и в продакшене из-за этого «молча» теряются изменения.

## 📚 Теория

**Структура** — набор именованных полей. При присваивании и передаче в функцию структура **копируется целиком**.

```go
a := User{Name: "Анна"}
b := a          // b — независимая копия
b.Name = "Борис"
fmt.Println(a.Name) // Анна
```

**Слайс** — это маленькая структура `{указатель на массив, len, cap}`. Когда слайс передают в функцию, копируется только этот «заголовок», а массив под ним общий. Поэтому функция **может менять элементы** слайса, переданного ей в аргументе.

**`for _, v := range slice` — `v` КОПИЯ элемента.** Изменения `v` не попадают в слайс. Чтобы менять элементы на месте, нужен индекс:

```go
for i := range users {
    users[i].Balance += 100 // меняем оригинал
}
```

**Указатель** `*User` хранит адрес значения. `&users[i]` указывает на элемент внутри массива слайса. А `&v`, где `v` взят из range, указывает на копию.

**nil-слайс и пустой слайс.** `var s []User` — это nil, а `[]User{}` — пустой, но не nil. Для `len`, `append` и `range` разницы нет. Но в JSON nil-слайс превращается в `null`, а пустой в `[]`. Фронтенд обычно ждёт `[]`.

**Подслайс** `s[a:b]` берёт элементы с a до b-1. Если индекс больше `len`, будет паника `index out of range`, поэтому границы нужно проверять.

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `users.go` | **Здесь ты работаешь**: функции с багами и заготовками |
| `mock_data.go` | Mock-данные: 7 пользователей |
| `main.go` | Демо-запуск, печатает результаты |
| `users_test.go` | Автотесты пайплайна. **Не меняй их** |

## 📝 Задания

| # | Функция | Тип | Что должно получиться на mock-данных |
|---|---|---|---|
| 1 | `ActiveUsers` | 🐞 баг | ID `[1 2 4 6 7]` |
| 2 | `FindByID` | 🐞 баг | изменение `u.Email` через указатель видно в `users[3]` |
| 3 | `ApplyBonus` | 🐞 баг | бонус 100 только активным: Анна 1600, Борис 400, Виктор 0, Галина 9900, Дмитрий 50, Елена 820, Жанна 4200 |
| 4 | `AdultNames` | ✍️ дописать | `[Анна Виктор Галина Елена Жанна]` |
| 5 | `Paginate` | ✍️ дописать | `page=2,size=3` → ID `[4 5 6]`; `page=3,size=3` → `[7]`; за пределами → `[]` (не nil) |

## ▶️ Как запустить
Все команды выполняются из **корня курса** (папка `mock`):

```bash
go run ./01-structs-slices        # демо-вывод
go run ./tools/ci 01              # 🚦 CI-пайплайн: gofmt → build → vet → tests
make ci M=01                      # то же самое через make
go test -v ./01-structs-slices    # «сырые» тесты, если нужен подробный вывод
```

Пока задания не сделаны, пайплайн красный: `❌ PIPELINE FAILED на стадии tests`. Модуль считается выполненным, когда видишь `✅ PIPELINE PASSED`.

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

Перечитай условие `if`. Символ `!` означает отрицание.
</details>

<details><summary>Подсказка к заданиям 2 и 3</summary>

`u` в `for _, u := range users` — это **копия**. Используй индекс: `for i := range users` и обращайся к `users[i]`.
</details>

<details><summary>Подсказка к заданию 4</summary>

`sort.Strings(names)` сортирует слайс строк на месте. Можно и `slices.Sort(names)` из пакета `slices` (Go 1.21+).
</details>

<details><summary>Подсказка к заданию 5</summary>

`start := (page-1)*size`, `end := start+size`. Если `start >= len(users)`, возвращай пустой слайс. Если `end > len(users)`, обрежь `end` до `len(users)`.
</details>

## ✅ Решения

<details><summary>🔒 Решение задания 1 — ActiveUsers</summary>

```go
func ActiveUsers(users []User) []User {
	var result []User
	for _, u := range users {
		if u.IsActive { // было !u.IsActive
			result = append(result, u)
		}
	}
	return result
}
```
</details>

<details><summary>🔒 Решение задания 2 — FindByID</summary>

```go
func FindByID(users []User, id int) (*User, bool) {
	for i := range users {
		if users[i].ID == id {
			return &users[i], true // адрес элемента слайса, а не копии
		}
	}
	return nil, false
}
```
Почему так: `&u` указывал на переменную цикла, то есть на копию. Изменения уходили в копию, а слайс оставался прежним.
</details>

<details><summary>🔒 Решение задания 3 — ApplyBonus</summary>

```go
func ApplyBonus(users []User, bonus int) {
	for i := range users {
		if users[i].IsActive {
			users[i].Balance += bonus
		}
	}
}
```
Можно и так: `u := &users[i]; u.Balance += bonus`.
</details>

<details><summary>🔒 Решение задания 4 — AdultNames</summary>

```go
import "sort"

func AdultNames(users []User) []string {
	names := []string{}
	for _, u := range users {
		if u.Age >= 18 {
			names = append(names, u.Name)
		}
	}
	sort.Strings(names)
	return names
}
```
</details>

<details><summary>🔒 Решение задания 5 — Paginate</summary>

```go
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
```
⚠️ Подслайс `users[start:end]` использует **тот же массив**, что и исходный слайс. Если вызывающий код поменяет результат, изменится и исходный слайс. Когда это нежелательно, делай копию: `slices.Clone(users[start:end])`.
</details>

## 🧠 Вопросы для самопроверки (часто спрашивают на собеседовании)
1. Чем отличаются `len` и `cap` у слайса? Что делает `append`, когда `cap` закончился?
2. Почему функция может изменить элементы переданного слайса, но не может изменить его длину для вызывающего кода?
3. Чем nil-слайс отличается от пустого? Как каждый из них выглядит в JSON?
4. Когда передавать структуру по значению, а когда по указателю?
