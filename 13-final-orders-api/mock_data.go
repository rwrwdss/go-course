package main

import "gocourse/13-final-orders-api/internal/domain"

// mockCatalog — каталог товаров. Цены в копейках.
func mockCatalog() []domain.Product {
	return []domain.Product{
		{ID: 1, Name: "Кофе в зёрнах 1 кг", Price: 159_000},
		{ID: 2, Name: "Кофемолка", Price: 499_000},
		{ID: 3, Name: "Фильтры для пуровера", Price: 45_000},
		{ID: 4, Name: "Кружка", Price: 79_900},
	}
}
