package main

import (
	"errors"
	"fmt"
)

// Ошибки бизнес-логики.
var (
	ErrInvalidQuantity = errors.New("quantity must be positive")
	ErrOutOfStock      = errors.New("not enough items in stock")
)

// ProductService — бизнес-логика каталога. Хранилище внедряется через интерфейс.
type ProductService struct {
	repo ProductRepository
}

// NewProductService создаёт сервис (это и есть Dependency Injection через конструктор).
func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// Buy покупает qty штук товара id:
//   - qty <= 0                 → ErrInvalidQuantity
//   - товара нет               → ошибка, для которой errors.Is(err, ErrNotFound)
//   - Stock < qty              → ErrOutOfStock (можно обернуть)
//   - иначе уменьшает Stock через repo.Update и возвращает обновлённый товар.
//
// ✍️ ЗАДАНИЕ 5: реализуй метод.
func (s *ProductService) Buy(id, qty int) (Product, error) {
	// TODO
	return Product{}, nil
}

// TotalStockValue возвращает суммарную стоимость всех товаров на складе: Σ Price*Stock.
func (s *ProductService) TotalStockValue() int {
	total := 0
	for _, p := range s.repo.List() {
		total += p.Price * p.Stock
	}
	return total
}

// Describe — пример полиморфизма: любой тип с методом String() — это fmt.Stringer.
func Describe(items ...fmt.Stringer) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.String())
	}
	return out
}

// String делает Product реализацией fmt.Stringer.
// Формат: "Клавиатура (3500 ₽, остаток 10)".
// ✍️ ЗАДАНИЕ 6: допиши метод.
func (p Product) String() string {
	// TODO
	return ""
}
