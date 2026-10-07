package main

import (
	"errors"
	"reflect"
	"testing"
)

// stubRepo — подменная реализация ProductRepository для тестов.
// Сервису всё равно, что за хранилище: главное, чтобы оно реализовывало интерфейс.
type stubRepo struct {
	products map[int]Product
	updated  []Product
}

func (s *stubRepo) GetByID(id int) (Product, error) {
	p, ok := s.products[id]
	if !ok {
		return Product{}, ErrNotFound
	}
	return p, nil
}
func (s *stubRepo) List() []Product {
	return []Product{s.products[1], s.products[2]}
}
func (s *stubRepo) Create(p Product) Product { return p }
func (s *stubRepo) Update(p Product) error {
	s.updated = append(s.updated, p)
	s.products[p.ID] = p
	return nil
}
func (s *stubRepo) Delete(int) error { return nil }

func newStub() *stubRepo {
	return &stubRepo{products: map[int]Product{
		1: {ID: 1, Name: "Клавиатура", Price: 3500, Stock: 10},
		2: {ID: 2, Name: "Мышь", Price: 1200, Stock: 0},
	}}
}

func TestService_Buy_OK(t *testing.T) {
	repo := newStub()
	p, err := NewProductService(repo).Buy(1, 3)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if p.Stock != 7 {
		t.Errorf("остаток = %d, ожидалось 7", p.Stock)
	}
	if len(repo.updated) != 1 || repo.updated[0].Stock != 7 {
		t.Errorf("сервис должен сохранить товар через repo.Update со Stock=7, вызовы Update: %v", repo.updated)
	}
}

func TestService_Buy_Errors(t *testing.T) {
	tests := []struct {
		name    string
		id, qty int
		wantErr error
	}{
		{"нулевое количество", 1, 0, ErrInvalidQuantity},
		{"отрицательное количество", 1, -2, ErrInvalidQuantity},
		{"нет на складе", 2, 1, ErrOutOfStock},
		{"больше чем есть", 1, 11, ErrOutOfStock},
		{"несуществующий товар", 42, 1, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newStub()
			_, err := NewProductService(repo).Buy(tt.id, tt.qty)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Buy(%d, %d) ошибка = %v, ожидалось %v", tt.id, tt.qty, err, tt.wantErr)
			}
			if len(repo.updated) != 0 {
				t.Errorf("при ошибке Update вызываться не должен, а был вызван: %v", repo.updated)
			}
		})
	}
}

func TestService_TotalStockValue(t *testing.T) {
	if got := NewProductService(newStub()).TotalStockValue(); got != 35000 {
		t.Errorf("TotalStockValue = %d, ожидалось 35000", got)
	}
}

func TestProduct_Stringer(t *testing.T) {
	got := Describe(Product{Name: "Клавиатура", Price: 3500, Stock: 10}, Product{Name: "Мышь", Price: 1200})
	want := []string{"Клавиатура (3500 ₽, остаток 10)", "Мышь (1200 ₽, остаток 0)"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("String() = %q, ожидалось %q", got, want)
	}
}
