package main

import "fmt"

func main() {
	bank := mockBank()

	try := func(title string, err error) {
		fmt.Printf("%-40s → err=%v, http=%d\n", title, err, HTTPStatus(err))
	}

	try("Withdraw acc-1 1000", bank.Withdraw("acc-1", 1000))
	try("Withdraw acc-2 999999", bank.Withdraw("acc-2", 999_999))
	try("Withdraw acc-404 10", bank.Withdraw("acc-404", 10))
	try("Withdraw acc-1 -5", bank.Withdraw("acc-1", -5))
	try("Transfer acc-1 → acc-404 500", bank.Transfer("acc-1", "acc-404", 500))
	try("Transfer acc-3 → acc-1 500 (пустой счёт)", bank.Transfer("acc-3", "acc-1", 500))

	for _, id := range []string{"acc-1", "acc-2", "acc-3"} {
		bal, _ := bank.Balance(id)
		fmt.Printf("Баланс %s: %d коп.\n", id, bal)
	}
}
