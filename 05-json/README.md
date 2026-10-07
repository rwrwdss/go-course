# Модуль 05 · JSON: теги, DTO, конверт ответа, строгий парсинг

## 🎯 Цель
Освоить `encoding/json` так, как он используется на бэкенде: теги структур, `omitempty`, скрытие полей, разделение входного и выходного формата (DTO), единый конверт ответа, строгий разбор тела запроса.

## 📚 Теория

**Кодирование и декодирование:**
```go
raw, err := json.Marshal(v)      // структура → []byte
err := json.Unmarshal(raw, &v)   // []byte → структура (нужен УКАЗАТЕЛЬ)
```

**Только экспортируемые поля** (с заглавной буквы) участвуют в JSON. Поле `year int` пакет `json` не видит вообще. `go vet` об этом предупреждает: *struct field year has json tag but is not exported*.

**Теги:**
```go
type Movie struct {
    Rating float64  `json:"rating"`           // имя ключа
    Genres []string `json:"genres,omitempty"` // не писать, если пусто (nil, len 0, 0, "", false)
    Secret string   `json:"-"`                // никогда не сериализовать
}
```

**DTO (Data Transfer Object).** Внешнее API отдаёт `imdb_rating`, а наш API должен отдавать `rating`. У одного поля один тег, поэтому заводят две структуры: «как приходит» и «как отдаём», и перекладывают поля. Это нормальная практика: внешний формат не протекает в твою модель.

**Конверт ответа** — единая форма всех ответов API:
```json
{"success": true, "data": [...], "meta": {"total": 3}}
{"success": false, "error": "not found"}
```

**nil-слайс → `null`, пустой слайс → `[]`.** Для фронтенда это разные вещи.

**Строгий разбор запроса:**
```go
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields() // {"min_ratng": 9} → ошибка, а не тихо проигнорированное поле
err := dec.Decode(&f)       // пустое тело → io.EOF
```

**`go:embed`** встраивает файл в бинарник на этапе компиляции. Удобно для mock-данных, шаблонов и миграций.

**Ошибку `Unmarshal` нельзя игнорировать**: невалидный JSON от клиента должен превратиться в 400 Bad Request, а не в пустой результат.

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `movies.go` | **Работаешь здесь** |
| `data/movies.json` | Mock-ответ «внешнего API» (5 фильмов, ключ `imdb_rating`) |
| `main.go` | Демо: печатает JSON-ответ |
| `movies_test.go` | Автотесты. **Не меняй** |

## 📝 Задания

| # | Что | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `Movie` | 🐞 | `year` читается; на вход `imdb_rating`, на выход `rating`; `InternalNote` скрыт; пустые `genres` не выводятся |
| 2 | `LoadMovies` | 🐞 | невалидный JSON → ошибка |
| 3 | `BuildResponse` | ✍️ | конверт `{"success":true,"data":[...],"meta":{"total":N}}`, пустой результат `"data":[]` |
| 4 | `ParseFilter` | ✍️ | неизвестные поля, неверные типы и пустое тело → ошибка |

> 🚦 Сейчас пайплайн падает уже на стадии **vet**: линтер нашёл неэкспортируемое поле с тегом. В реальном CI до тестов дело бы тоже не дошло.

Ожидаемый вывод `go run ./05-json` после решения:
```json
{"success":true,"data":[{"id":1,"title":"Inception","year":2010,"rating":8.8,"genres":["sci-fi","thriller"],"director":{"name":"Christopher Nolan","country":"UK"}},{"id":4,"title":"Interstellar","year":2014,"rating":8.7,"genres":["sci-fi","drama"],"director":{"name":"Christopher Nolan","country":"UK"}},{"id":5,"title":"Ирония судьбы","year":1975,"rating":8,"genres":["comedy"],"director":{"name":"Эльдар Рязанов","country":"RU"}}],"meta":{"total":3}}
Фильтр с опечаткой: parse filter: json: unknown field "min_ratng"
```

## ▶️ Как запустить
```bash
go run ./05-json
go run ./tools/ci 05       # или: make ci M=05
```

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

- `year` → `Year`.
- Заведи `type externalMovie struct { ... Rating float64 \`json:"imdb_rating"\` ... }`, распарси JSON в неё, а потом переложи поля в `Movie` с тегом `json:"rating"`.
- `json:"-"` и `json:"genres,omitempty"`.
</details>

<details><summary>Подсказка к заданию 3</summary>

Начни с `filtered := []Movie{}` (не `var filtered []Movie`!), а дальше `json.Marshal(APIResponse{Success: true, Data: filtered, Meta: &Meta{Total: len(filtered)}})`.
</details>

<details><summary>Подсказка к заданию 4</summary>

`json.NewDecoder(bytes.NewReader(body))`, затем `dec.DisallowUnknownFields()`, затем `dec.Decode(&f)`.
</details>

## ✅ Решения

<details><summary>🔒 Решение заданий 1 и 2 — структуры и LoadMovies</summary>

```go
type Movie struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Year         int      `json:"year"`
	Rating       float64  `json:"rating"`
	Genres       []string `json:"genres,omitempty"`
	Director     Director `json:"director"`
	InternalNote string   `json:"-"`
}

// externalMovie — формат внешнего API (DTO на вход).
type externalMovie struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Year     int      `json:"year"`
	Rating   float64  `json:"imdb_rating"`
	Genres   []string `json:"genres"`
	Director Director `json:"director"`
}

func LoadMovies(data []byte) ([]Movie, error) {
	var external []externalMovie
	if err := json.Unmarshal(data, &external); err != nil {
		return nil, fmt.Errorf("decode movies: %w", err)
	}
	movies := make([]Movie, 0, len(external))
	for _, e := range external {
		movies = append(movies, Movie{
			ID:       e.ID,
			Title:    e.Title,
			Year:     e.Year,
			Rating:   e.Rating,
			Genres:   e.Genres,
			Director: e.Director,
		})
	}
	return movies, nil
}
```
</details>

<details><summary>🔒 Решение задания 3 — BuildResponse</summary>

```go
func BuildResponse(movies []Movie, minRating float64) ([]byte, error) {
	filtered := []Movie{} // не nil → в JSON будет [], а не null
	for _, m := range movies {
		if m.Rating >= minRating {
			filtered = append(filtered, m)
		}
	}
	return json.Marshal(APIResponse{
		Success: true,
		Data:    filtered,
		Meta:    &Meta{Total: len(filtered)},
	})
}
```
</details>

<details><summary>🔒 Решение задания 4 — ParseFilter</summary>

```go
import "bytes"

func ParseFilter(body []byte) (Filter, error) {
	var f Filter
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return Filter{}, fmt.Errorf("parse filter: %w", err)
	}
	return f, nil
}
```
</details>

## 🧠 Вопросы для самопроверки
1. Почему неэкспортируемые поля не попадают в JSON?
2. Что делает `omitempty` с полем `int`, у которого значение 0? Когда это проблема и как её решить через `*int`?
3. В какой Go-тип раскодируется JSON-число при разборе в `map[string]any`?
4. Зачем разделять DTO и доменную модель?
