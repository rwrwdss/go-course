package main

import "fmt"

func main() {
	orders := mockOrders()

	totals := TotalByCustomer(orders)
	fmt.Println("== Сумма оплаченных заказов по клиентам ==")
	fmt.Println(" ", totals)

	fmt.Println("== Заказы по статусам ==")
	fmt.Println(" ", CountByStatus(orders))

	fmt.Println("== Категории ==")
	fmt.Println(" ", UniqueCategories(orders))

	fmt.Println("== Топ-3 клиента ==")
	fmt.Println(" ", TopCustomers(totals, 3))

	fmt.Println("== Есть ли клиент zed? ==")
	fmt.Println(" ", HasCustomer(map[string]int{"zed": 0}, "zed"))
}
