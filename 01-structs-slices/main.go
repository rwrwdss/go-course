package main

import "fmt"

func main() {
	users := mockUsers()

	fmt.Println("== Активные пользователи ==")
	for _, u := range ActiveUsers(users) {
		fmt.Printf("  #%d %s\n", u.ID, u.Name)
	}

	if u, ok := FindByID(users, 4); ok {
		u.Email = "galina.new@example.com"
	}
	fmt.Println("== Email пользователя #4 после изменения ==")
	fmt.Println(" ", users[3].Email)

	ApplyBonus(users, 100)
	fmt.Println("== Балансы после бонуса 100 ==")
	for _, u := range users {
		fmt.Printf("  %-8s %5d\n", u.Name, u.Balance)
	}

	fmt.Println("== Совершеннолетние ==")
	fmt.Println(" ", AdultNames(users))

	fmt.Println("== Страница 2 (по 3 на странице) ==")
	for _, u := range Paginate(users, 2, 3) {
		fmt.Printf("  #%d %s\n", u.ID, u.Name)
	}
}
