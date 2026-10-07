package main

import (
	_ "embed"
	"encoding/json"
)

// moviesJSON — mock-ответ внешнего API, встроенный в бинарник через go:embed.
//
//go:embed data/movies.json
var moviesJSON []byte

// Director — режиссёр.
type Director struct {
	Name    string `json:"name"`
	Country string `json:"country"`
}

// Movie — фильм.
//
// 🐞 ЗАДАНИЕ 1: поправь структуру, чтобы:
//   - поле year из JSON попадало в структуру (сейчас всегда 0);
//   - рейтинг читался из ключа "imdb_rating", а в наш ответ уходил под ключом "rating";
//   - InternalNote НИКОГДА не попадал в JSON-ответ;
//   - пустой список жанров не попадал в ответ (ключ "genres" отсутствует, если жанров нет).
//
// Подсказка по рейтингу: используй ДВЕ структуры — для входа (externalMovie) и для выхода (Movie).
type Movie struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	year         int      `json:"year"`
	Rating       float64  `json:"imdb_rating"`
	Genres       []string `json:"genres"`
	Director     Director `json:"director"`
	InternalNote string
}

// LoadMovies разбирает JSON внешнего API в []Movie.
// 🐞 ЗАДАНИЕ 2: при битом JSON функция «молча» возвращает пустой список. Нужно вернуть ошибку.
func LoadMovies(data []byte) ([]Movie, error) {
	var movies []Movie
	json.Unmarshal(data, &movies)
	return movies, nil
}

// Meta — метаданные ответа.
type Meta struct {
	Total int `json:"total"`
}

// APIResponse — единый конверт ответа нашего API.
type APIResponse struct {
	Success bool    `json:"success"`
	Data    []Movie `json:"data"`
	Error   string  `json:"error,omitempty"`
	Meta    *Meta   `json:"meta,omitempty"`
}

// BuildResponse отбирает фильмы с рейтингом >= minRating и упаковывает их в конверт:
//
//	{"success":true,"data":[...],"meta":{"total":N}}
//
// Если ничего не нашлось, "data" должен быть [] (а не null), а total равен 0.
// ✍️ ЗАДАНИЕ 3: допиши функцию (используй json.Marshal).
func BuildResponse(movies []Movie, minRating float64) ([]byte, error) {
	// TODO
	return nil, nil
}

// ParseFilter разбирает тело запроса вида {"min_rating": 8.5, "director": "Christopher Nolan"}.
// Неизвестные поля — ошибка (защита от опечаток клиента: "min_ratng").
// Пустое тело — тоже ошибка.
// ✍️ ЗАДАНИЕ 4: допиши функцию (json.NewDecoder + DisallowUnknownFields).
type Filter struct {
	MinRating float64 `json:"min_rating"`
	Director  string  `json:"director"`
}

func ParseFilter(body []byte) (Filter, error) {
	// TODO
	return Filter{}, nil
}
