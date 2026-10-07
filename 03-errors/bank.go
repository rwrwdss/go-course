package main

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel-ошибки: их сравнивают через errors.Is.
var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// ValidationError — ошибка валидации входных данных. Её «достают» через errors.As.
type ValidationError struct {
	Field   string
	Message string
}

// Error реализует интерфейс error.
// Формат: "validation failed: <field>: <message>", например
// "validation failed: amount: must be positive".
// ✍️ ЗАДАНИЕ 1: допиши метод.
func (e *ValidationError) Error() string {
	// TODO
	return ""
}

// Account — банковский счёт. Balance в копейках (никаких float для денег!).
type Account struct {
	ID      string
	Owner   string
	Balance int64
}

// Bank хранит счета.
type Bank struct {
	accounts map[string]*Account
}

// NewBank создаёт банк из списка счетов.
func NewBank(accounts []Account) *Bank {
	b := &Bank{accounts: make(map[string]*Account, len(accounts))}
	for _, a := range accounts {
		acc := a
		b.accounts[a.ID] = &acc
	}
	return b
}

// Balance возвращает баланс счёта.
func (b *Bank) Balance(id string) (int64, error) {
	acc, ok := b.accounts[id]
	if !ok {
		return 0, fmt.Errorf("balance %s: %w", id, ErrAccountNotFound)
	}
	return acc.Balance, nil
}

// Withdraw списывает amount со счёта id.
// Требования:
//   - amount <= 0           → *ValidationError{Field: "amount", Message: "must be positive"}
//   - счёта нет             → ошибка, для которой errors.Is(err, ErrAccountNotFound) == true
//   - денег не хватает      → ошибка, для которой errors.Is(err, ErrInsufficientFunds) == true,
//     а текст содержит id счёта, например "withdraw acc-2: insufficient funds"
//
// 🐞 ЗАДАНИЕ 2: сейчас ошибки создаются «с нуля», и errors.Is / errors.As их не распознают.
// Валидации amount нет вообще.
func (b *Bank) Withdraw(id string, amount int64) error {
	acc, ok := b.accounts[id]
	if !ok {
		return errors.New("account not found")
	}
	if acc.Balance < amount {
		return fmt.Errorf("withdraw %s: insufficient funds", id)
	}
	acc.Balance -= amount
	return nil
}

// Deposit зачисляет amount на счёт id.
func (b *Bank) Deposit(id string, amount int64) error {
	if amount <= 0 {
		return &ValidationError{Field: "amount", Message: "must be positive"}
	}
	acc, ok := b.accounts[id]
	if !ok {
		return fmt.Errorf("deposit %s: %w", id, ErrAccountNotFound)
	}
	acc.Balance += amount
	return nil
}

// Transfer переводит amount со счёта from на счёт to.
// Если перевод невозможен — НИ ОДИН баланс не должен измениться, а ошибка должна вернуться наверх.
// Перевод самому себе запрещён → *ValidationError{Field: "to", Message: "cannot transfer to the same account"}.
//
// 🐞 ЗАДАНИЕ 3: здесь сразу три бага:
//   - ошибка Withdraw игнорируется;
//   - деньги списываются ДО проверки, что получатель существует (деньги «сгорают»);
//   - нет проверки перевода самому себе.
func (b *Bank) Transfer(from, to string, amount int64) error {
	_ = b.Withdraw(from, amount)
	if err := b.Deposit(to, amount); err != nil {
		return err
	}
	return nil
}

// HTTPStatus переводит ошибку бизнес-логики в HTTP-статус для ответа API:
//
//	nil                    → 200 OK
//	*ValidationError       → 400 Bad Request
//	ErrAccountNotFound     → 404 Not Found
//	ErrInsufficientFunds   → 409 Conflict
//	любая другая ошибка    → 500 Internal Server Error
//
// Ошибки могут быть обёрнуты (fmt.Errorf("...: %w", err)) несколько раз!
// ✍️ ЗАДАНИЕ 4: допиши функцию через errors.Is и errors.As.
func HTTPStatus(err error) int {
	// TODO
	return http.StatusOK
}
