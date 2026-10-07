package main

// mockBooks — стартовые данные библиотеки.
func mockBooks() []Book {
	return []Book{
		{ID: 1, Title: "Властелин колец", Author: "Дж. Р. Р. Толкин", Year: 1954},
		{ID: 2, Title: "Хоббит", Author: "Дж. Р. Р. Толкин", Year: 1937},
		{ID: 3, Title: "Мастер и Маргарита", Author: "Михаил Булгаков", Year: 1967},
		{ID: 4, Title: "The Go Programming Language", Author: "Alan Donovan", Year: 2015},
		{ID: 5, Title: "Собачье сердце", Author: "Михаил Булгаков", Year: 1925},
	}
}
