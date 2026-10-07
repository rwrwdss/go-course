package main

// mockBank — банк с тремя счетами. Балансы в копейках.
func mockBank() *Bank {
	return NewBank([]Account{
		{ID: "acc-1", Owner: "Анна", Balance: 100_000},
		{ID: "acc-2", Owner: "Борис", Balance: 5_000},
		{ID: "acc-3", Owner: "Виктор", Balance: 0},
	})
}
