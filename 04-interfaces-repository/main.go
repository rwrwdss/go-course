package main

import "fmt"

func main() {
	repo := NewInMemoryProductRepo(mockProducts())
	svc := NewProductService(repo)

	fmt.Println("== Каталог ==")
	for _, p := range repo.List() {
		fmt.Println(" ", p)
	}

	a := repo.Create(Product{Name: "Веб-камера", Price: 4200, Stock: 5})
	b := repo.Create(Product{Name: "Коврик", Price: 600, Stock: 20})
	fmt.Printf("== Созданы товары с ID %d и %d ==\n", a.ID, b.ID)

	if p, err := svc.Buy(1, 3); err != nil {
		fmt.Println("Ошибка покупки:", err)
	} else {
		fmt.Println("Куплено! Остаток:", p.Stock)
	}
	_, err := svc.Buy(2, 1)
	fmt.Println("Покупка мыши:", err)

	fmt.Println("Стоимость склада:", svc.TotalStockValue())
}
