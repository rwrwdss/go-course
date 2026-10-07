package main

// mockProducts — стартовый каталог товаров. Цены в рублях.
func mockProducts() []Product {
	return []Product{
		{ID: 1, Name: "Клавиатура", Price: 3500, Stock: 10},
		{ID: 2, Name: "Мышь", Price: 1200, Stock: 0},
		{ID: 3, Name: "Монитор", Price: 18000, Stock: 3},
		{ID: 4, Name: "Наушники", Price: 5600, Stock: 7},
	}
}
