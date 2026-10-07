package service

import (
	"errors"
	"testing"

	"gocourse/13-final-orders-api/internal/domain"
	"gocourse/13-final-orders-api/internal/repo"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

func newService() *OrderService {
	catalog := repo.NewCatalog([]domain.Product{
		{ID: 1, Name: "Кофе", Price: 159_000},
		{ID: 2, Name: "Кофемолка", Price: 499_000},
	})
	return NewOrderService(catalog, repo.NewOrderRepo())
}

func TestCreateOrder_OK(t *testing.T) {
	order, err := newService().CreateOrder("anna", []ItemInput{{ProductID: 1, Quantity: 2}, {ProductID: 2, Quantity: 1}})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if order.ID != 1 || order.Customer != "anna" || order.Status != domain.StatusNew {
		t.Errorf("заказ = %+v", order)
	}
	if order.Total != 2*159_000+499_000 {
		t.Errorf("Total = %d, ожидалось %d", order.Total, 2*159_000+499_000)
	}
	if len(order.Items) != 2 || order.Items[0].Price != 159_000 || order.Items[1].Price != 499_000 {
		t.Errorf("в позициях должна быть зафиксирована цена из каталога: %+v", order.Items)
	}
}

func TestCreateOrder_Errors(t *testing.T) {
	tests := []struct {
		name     string
		customer string
		items    []ItemInput
		wantErr  error
	}{
		{"пустой покупатель", "  ", []ItemInput{{ProductID: 1, Quantity: 1}}, domain.ErrValidation},
		{"нет позиций", "anna", nil, domain.ErrValidation},
		{"нулевое количество", "anna", []ItemInput{{ProductID: 1, Quantity: 0}}, domain.ErrValidation},
		{"отрицательное количество", "anna", []ItemInput{{ProductID: 1, Quantity: -3}}, domain.ErrValidation},
		{"неизвестный товар", "anna", []ItemInput{{ProductID: 1, Quantity: 1}, {ProductID: 99, Quantity: 1}}, domain.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService()
			_, err := svc.CreateOrder(tt.customer, tt.items)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ошибка = %v, ожидалось %v", err, tt.wantErr)
			}
			if n := len(svc.List()); n != 0 {
				t.Errorf("при ошибке заказ не должен сохраняться, а в хранилище %d заказ(ов)", n)
			}
		})
	}
}

func TestChangeStatus(t *testing.T) {
	tests := []struct {
		path    []domain.Status // последовательность переходов от new
		wantErr bool            // ошибка на ПОСЛЕДНЕМ переходе
	}{
		{[]domain.Status{"paid"}, false},
		{[]domain.Status{"cancelled"}, false},
		{[]domain.Status{"paid", "shipped"}, false},
		{[]domain.Status{"paid", "cancelled"}, false},
		{[]domain.Status{"shipped"}, true},
		{[]domain.Status{"new"}, true},
		{[]domain.Status{"paid", "shipped", "cancelled"}, true},
		{[]domain.Status{"cancelled", "paid"}, true},
		{[]domain.Status{"paid", "shipped", "new"}, true},
		{[]domain.Status{"delivered"}, true},
	}
	for _, tt := range tests {
		name := "new"
		for _, s := range tt.path {
			name += "→" + string(s)
		}
		t.Run(name, func(t *testing.T) {
			svc := newService()
			order, err := svc.CreateOrder("anna", []ItemInput{{ProductID: 1, Quantity: 1}})
			if err != nil {
				t.Fatalf("CreateOrder: %v", err)
			}
			for i, to := range tt.path {
				before, _ := svc.Get(order.ID)
				updated, err := svc.ChangeStatus(order.ID, to)
				last := i == len(tt.path)-1
				if !last {
					if err != nil {
						t.Fatalf("промежуточный переход в %s: %v", to, err)
					}
					continue
				}
				if tt.wantErr {
					if !errors.Is(err, domain.ErrInvalidTransition) {
						t.Errorf("ожидалась ErrInvalidTransition, получено %v", err)
					}
					if got, _ := svc.Get(order.ID); got.Status != before.Status {
						t.Errorf("недопустимый переход изменил статус: был %s, стал %s", before.Status, got.Status)
					}
				} else if err != nil || updated.Status != to {
					t.Errorf("переход в %s: статус %s, ошибка %v", to, updated.Status, err)
				}
			}
		})
	}
}

func TestChangeStatus_NotFound(t *testing.T) {
	if _, err := newService().ChangeStatus(42, domain.StatusPaid); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ожидалась ErrNotFound, получено %v", err)
	}
}
