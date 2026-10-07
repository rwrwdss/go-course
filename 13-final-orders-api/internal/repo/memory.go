// Package repo — хранилище заказов и каталог товаров (in-memory вместо БД).
package repo

import (
	"fmt"
	"sort"

	"gocourse/13-final-orders-api/internal/domain"
)

// Catalog — каталог товаров (только чтение, поэтому без мьютекса).
type Catalog struct {
	products map[int]domain.Product
}

func NewCatalog(products []domain.Product) *Catalog {
	c := &Catalog{products: make(map[int]domain.Product, len(products))}
	for _, p := range products {
		c.products[p.ID] = p
	}
	return c
}

func (c *Catalog) Product(id int) (domain.Product, error) {
	p, ok := c.products[id]
	if !ok {
		return domain.Product{}, fmt.Errorf("product %d: %w", id, domain.ErrNotFound)
	}
	return p, nil
}

// OrderRepo хранит заказы. Им пользуются одновременно все HTTP-запросы.
//
// 🐞 ЗАДАНИЕ 1: репозиторий не потокобезопасен — нагрузочный тест с -race это ловит.
type OrderRepo struct {
	orders map[int]domain.Order
	nextID int
}

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{orders: make(map[int]domain.Order), nextID: 1}
}

func (r *OrderRepo) Create(o domain.Order) domain.Order {
	o.ID = r.nextID
	r.nextID++
	r.orders[o.ID] = o
	return o
}

func (r *OrderRepo) Get(id int) (domain.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return domain.Order{}, fmt.Errorf("order %d: %w", id, domain.ErrNotFound)
	}
	return o, nil
}

func (r *OrderRepo) Update(o domain.Order) error {
	if _, ok := r.orders[o.ID]; !ok {
		return fmt.Errorf("order %d: %w", o.ID, domain.ErrNotFound)
	}
	r.orders[o.ID] = o
	return nil
}

func (r *OrderRepo) List() []domain.Order {
	out := make([]domain.Order, 0, len(r.orders))
	for _, o := range r.orders {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
