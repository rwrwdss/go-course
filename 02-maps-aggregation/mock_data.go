package main

// mockOrders — выгрузка заказов интернет-магазина.
func mockOrders() []Order {
	return []Order{
		{ID: 101, Customer: "alice", Category: "books", Amount: 1200, Status: StatusPaid},
		{ID: 102, Customer: "bob", Category: "electronics", Amount: 45000, Status: StatusPaid},
		{ID: 103, Customer: "alice", Category: "electronics", Amount: 3000, Status: StatusCancelled},
		{ID: 104, Customer: "carol", Category: "books", Amount: 800, Status: StatusPaid},
		{ID: 105, Customer: "dave", Category: "food", Amount: 1500, Status: StatusPending},
		{ID: 106, Customer: "alice", Category: "food", Amount: 2300, Status: StatusPaid},
		{ID: 107, Customer: "carol", Category: "electronics", Amount: 12000, Status: StatusPaid},
		{ID: 108, Customer: "eve", Category: "books", Amount: 3500, Status: StatusPaid},
		{ID: 109, Customer: "bob", Category: "food", Amount: 700, Status: StatusCancelled},
		{ID: 110, Customer: "dave", Category: "books", Amount: 3500, Status: StatusPaid},
	}
}
