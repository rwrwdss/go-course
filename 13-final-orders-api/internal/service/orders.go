// Package service — бизнес-логика заказов. Не знает ни про HTTP, ни про конкретную БД.
package service

import (
	"gocourse/13-final-orders-api/internal/domain"
)

// ProductCatalog — что сервису нужно от каталога.
type ProductCatalog interface {
	Product(id int) (domain.Product, error)
}

// OrderStore — что сервису нужно от хранилища заказов.
type OrderStore interface {
	Create(o domain.Order) domain.Order
	Get(id int) (domain.Order, error)
	Update(o domain.Order) error
	List() []domain.Order
}

// OrderService — бизнес-логика заказов.
type OrderService struct {
	catalog ProductCatalog
	store   OrderStore
}

func NewOrderService(catalog ProductCatalog, store OrderStore) *OrderService {
	return &OrderService{catalog: catalog, store: store}
}

// ItemInput — позиция из запроса клиента.
type ItemInput struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// CreateOrder создаёт заказ:
//   - customer пустой                    → ошибка, errors.Is(err, domain.ErrValidation)
//   - items пустой                       → ErrValidation
//   - quantity <= 0                      → ErrValidation
//   - товара нет в каталоге              → errors.Is(err, domain.ErrNotFound)
//   - Price позиции = цена из каталога; Total = Σ Price*Quantity; Status = new.
//
// ✍️ ЗАДАНИЕ 2: реализуй.
func (s *OrderService) CreateOrder(customer string, items []ItemInput) (domain.Order, error) {
	// TODO
	return domain.Order{}, nil
}

// allowedTransitions — граф переходов статусов:
//
//	new ──► paid ──► shipped
//	 └────► cancelled ◄── (из paid тоже можно отменить)
var allowedTransitions = map[domain.Status][]domain.Status{
	domain.StatusNew:  {domain.StatusPaid, domain.StatusCancelled},
	domain.StatusPaid: {domain.StatusShipped, domain.StatusCancelled},
}

// ChangeStatus переводит заказ в новый статус. Недопустимый переход → domain.ErrInvalidTransition.
// 🐞 ЗАДАНИЕ 3: сейчас разрешён любой переход (shipped → new, cancelled → paid...).
func (s *OrderService) ChangeStatus(id int, to domain.Status) (domain.Order, error) {
	order, err := s.store.Get(id)
	if err != nil {
		return domain.Order{}, err
	}
	order.Status = to
	if err := s.store.Update(order); err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

// Get возвращает заказ по ID.
func (s *OrderService) Get(id int) (domain.Order, error) {
	return s.store.Get(id)
}

// List возвращает все заказы.
func (s *OrderService) List() []domain.Order {
	return s.store.List()
}
