package main

import (
	"errors"
	"fmt"
)

// ErrNotFound — товар не найден.
var ErrNotFound = errors.New("product not found")

// Product — товар каталога.
type Product struct {
	ID    int
	Name  string
	Price int
	Stock int
}

// ProductRepository — контракт хранилища товаров.
// Бизнес-логика (ProductService) зависит ТОЛЬКО от этого интерфейса,
// поэтому завтра in-memory можно заменить на PostgreSQL, не трогая сервис.
type ProductRepository interface {
	GetByID(id int) (Product, error)
	List() []Product
	Create(p Product) Product
	Update(p Product) error
	Delete(id int) error
}

// Проверка на этапе компиляции: InMemoryProductRepo обязан реализовывать ProductRepository.
// 🐞 ЗАДАНИЕ 1: из-за этой строки код не компилируется. Прочитай ошибку компилятора и допиши недостающее.
var _ ProductRepository = (*InMemoryProductRepo)(nil)

// InMemoryProductRepo хранит товары в памяти.
type InMemoryProductRepo struct {
	items  map[int]Product
	nextID int
}

// NewInMemoryProductRepo создаёт репозиторий и заполняет его начальными данными.
func NewInMemoryProductRepo(seed []Product) *InMemoryProductRepo {
	r := &InMemoryProductRepo{items: make(map[int]Product), nextID: 1}
	for _, p := range seed {
		r.items[p.ID] = p
		if p.ID >= r.nextID {
			r.nextID = p.ID + 1
		}
	}
	return r
}

// GetByID возвращает товар по ID или ErrNotFound (обёрнутую).
func (r *InMemoryProductRepo) GetByID(id int) (Product, error) {
	p, ok := r.items[id]
	if !ok {
		return Product{}, fmt.Errorf("get product %d: %w", id, ErrNotFound)
	}
	return p, nil
}

// List возвращает все товары, отсортированные по ID по возрастанию.
// 🐞 ЗАДАНИЕ 2: порядок «скачет» от запуска к запуску. Почему? Исправь.
func (r *InMemoryProductRepo) List() []Product {
	result := make([]Product, 0, len(r.items))
	for _, p := range r.items {
		result = append(result, p)
	}
	return result
}

// Create присваивает товару новый уникальный ID, сохраняет его и возвращает.
// 🐞 ЗАДАНИЕ 3: все новые товары получают ОДИНАКОВЫЙ ID и затирают друг друга. Почему?
// Подсказка: посмотри внимательно на получателя (receiver) метода.
func (r InMemoryProductRepo) Create(p Product) Product {
	p.ID = r.nextID
	r.nextID++
	r.items[p.ID] = p
	return p
}

// Update обновляет существующий товар или возвращает ErrNotFound.
func (r *InMemoryProductRepo) Update(p Product) error {
	if _, ok := r.items[p.ID]; !ok {
		return fmt.Errorf("update product %d: %w", p.ID, ErrNotFound)
	}
	r.items[p.ID] = p
	return nil
}
