package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// ⚠️ Автотесты пайплайна. Не меняй их — меняй код в bank.go.

func balanceOf(t *testing.T, b *Bank, id string) int64 {
	t.Helper()
	v, err := b.Balance(id)
	if err != nil {
		t.Fatalf("Balance(%s): %v", id, err)
	}
	return v
}

func TestValidationError_Message(t *testing.T) {
	err := &ValidationError{Field: "amount", Message: "must be positive"}
	want := "validation failed: amount: must be positive"
	if err.Error() != want {
		t.Errorf("Error() = %q, ожидалось %q", err.Error(), want)
	}
}

func TestWithdraw_OK(t *testing.T) {
	b := mockBank()
	if err := b.Withdraw("acc-1", 30_000); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if got := balanceOf(t, b, "acc-1"); got != 70_000 {
		t.Errorf("баланс = %d, ожидалось 70000", got)
	}
}

func TestWithdraw_NotFound_IsSentinel(t *testing.T) {
	err := mockBank().Withdraw("acc-404", 100)
	if !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("errors.Is(err, ErrAccountNotFound) = false, err = %v", err)
	}
}

func TestWithdraw_InsufficientFunds_IsWrapped(t *testing.T) {
	err := mockBank().Withdraw("acc-2", 999_999)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("errors.Is(err, ErrInsufficientFunds) = false, err = %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "acc-2") {
		t.Errorf("текст ошибки должен содержать id счёта для отладки, получено %q", err.Error())
	}
}

func TestWithdraw_NonPositiveAmount_IsValidationError(t *testing.T) {
	for _, amount := range []int64{0, -100} {
		b := mockBank()
		err := b.Withdraw("acc-1", amount)
		var vErr *ValidationError
		if !errors.As(err, &vErr) {
			t.Errorf("amount=%d: ожидалась *ValidationError, получено %v", amount, err)
			continue
		}
		if vErr.Field != "amount" {
			t.Errorf("amount=%d: Field = %q, ожидалось \"amount\"", amount, vErr.Field)
		}
		if got := balanceOf(t, b, "acc-1"); got != 100_000 {
			t.Errorf("amount=%d: баланс изменился: %d", amount, got)
		}
	}
}

func TestTransfer_OK(t *testing.T) {
	b := mockBank()
	if err := b.Transfer("acc-1", "acc-3", 25_000); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if a, c := balanceOf(t, b, "acc-1"), balanceOf(t, b, "acc-3"); a != 75_000 || c != 25_000 {
		t.Errorf("балансы acc-1=%d acc-3=%d, ожидалось 75000 и 25000", a, c)
	}
}

func TestTransfer_InsufficientFunds_NoMoneyCreated(t *testing.T) {
	b := mockBank()
	err := b.Transfer("acc-3", "acc-1", 500)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("ожидалась ErrInsufficientFunds, получено %v", err)
	}
	if got := balanceOf(t, b, "acc-1"); got != 100_000 {
		t.Errorf("деньги появились из воздуха: acc-1 = %d, ожидалось 100000", got)
	}
}

func TestTransfer_UnknownRecipient_NoMoneyLost(t *testing.T) {
	b := mockBank()
	err := b.Transfer("acc-1", "acc-404", 500)
	if !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("ожидалась ErrAccountNotFound, получено %v", err)
	}
	if got := balanceOf(t, b, "acc-1"); got != 100_000 {
		t.Errorf("деньги сгорели: acc-1 = %d, ожидалось 100000", got)
	}
}

func TestTransfer_SameAccount(t *testing.T) {
	err := mockBank().Transfer("acc-1", "acc-1", 100)
	var vErr *ValidationError
	if !errors.As(err, &vErr) || vErr.Field != "to" {
		t.Errorf("ожидалась *ValidationError с Field=\"to\", получено %v", err)
	}
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 200},
		{"валидация", &ValidationError{Field: "amount", Message: "x"}, 400},
		{"валидация обёрнутая", fmt.Errorf("handler: %w", &ValidationError{Field: "a", Message: "b"}), 400},
		{"не найден", ErrAccountNotFound, 404},
		{"не найден дважды обёрнутый", fmt.Errorf("api: %w", fmt.Errorf("repo: %w", ErrAccountNotFound)), 404},
		{"нет денег", fmt.Errorf("withdraw: %w", ErrInsufficientFunds), 409},
		{"неизвестная", errors.New("db is down"), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatus(tt.err); got != tt.want {
				t.Errorf("HTTPStatus(%v) = %d, ожидалось %d", tt.err, got, tt.want)
			}
		})
	}
}
