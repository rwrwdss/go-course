package main

import (
	"fmt"
	"log"
)

func main() {
	movies, err := LoadMovies(moviesJSON)
	if err != nil {
		log.Fatalf("не удалось загрузить фильмы: %v", err)
	}
	for i := range movies {
		movies[i].InternalNote = "секретная заметка редактора"
	}

	resp, err := BuildResponse(movies, 8.0)
	if err != nil {
		log.Fatalf("не удалось собрать ответ: %v", err)
	}
	fmt.Println(string(resp))

	_, err = ParseFilter([]byte(`{"min_ratng": 9}`))
	fmt.Println("Фильтр с опечаткой:", err)
}
