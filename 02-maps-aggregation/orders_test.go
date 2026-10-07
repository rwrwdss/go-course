package main

import (
	"reflect"
	"testing"
)

// ⚠️ Автотесты пайплайна. Не меняй их — меняй код в orders.go.

func TestTotalByCustomer(t *testing.T) {
	got := TotalByCustomer(mockOrders())
	want := map[string]int{"alice": 3500, "bob": 45000, "carol": 12800, "eve": 3500, "dave": 3500}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TotalByCustomer = %v, ожидалось %v", got, want)
	}
}

func TestTotalByCustomer_EmptyInput(t *testing.T) {
	got := TotalByCustomer(nil)
	if got == nil || len(got) != 0 {
		t.Errorf("для пустого списка ожидалась пустая (не nil) мапа, получено %#v", got)
	}
}

func TestCountByStatus(t *testing.T) {
	got := CountByStatus(mockOrders())
	want := map[string]int{"paid": 7, "pending": 1, "cancelled": 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CountByStatus = %v, ожидалось %v", got, want)
	}
}

func TestUniqueCategories(t *testing.T) {
	got := UniqueCategories(mockOrders())
	want := []string{"books", "electronics", "food"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UniqueCategories = %v, ожидалось %v", got, want)
	}
}

func TestTopCustomers(t *testing.T) {
	totals := map[string]int{"alice": 3500, "bob": 45000, "carol": 12800, "eve": 3500, "dave": 3500}
	tests := []struct {
		name string
		n    int
		want []string
	}{
		{"топ-1", 1, []string{"bob"}},
		{"топ-3 с равными суммами", 3, []string{"bob", "carol", "alice"}},
		{"все клиенты", 5, []string{"bob", "carol", "alice", "dave", "eve"}},
		{"n больше числа клиентов", 50, []string{"bob", "carol", "alice", "dave", "eve"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Гоняем несколько раз: порядок обхода map случайный, и баг может проявиться не сразу.
			for i := 0; i < 20; i++ {
				got := TopCustomers(totals, tt.n)
				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("TopCustomers(n=%d) = %v, ожидалось %v", tt.n, got, tt.want)
				}
			}
		})
	}
}

func TestHasCustomer(t *testing.T) {
	totals := map[string]int{"zed": 0, "amy": 100}
	if !HasCustomer(totals, "zed") {
		t.Errorf("клиент zed есть в мапе (с суммой 0), а HasCustomer вернул false")
	}
	if !HasCustomer(totals, "amy") {
		t.Errorf("клиент amy есть в мапе, а HasCustomer вернул false")
	}
	if HasCustomer(totals, "nobody") {
		t.Errorf("клиента nobody нет, а HasCustomer вернул true")
	}
}
