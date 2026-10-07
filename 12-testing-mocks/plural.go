// Package notify — сервис уведомлений пользователей интернет-магазина.
package notify

import "fmt"

// Pluralize возвращает число со словом в правильной форме (русский язык):
//
//	1 заказ, 2 заказа, 5 заказов, 11 заказов, 21 заказ, 22 заказа, 111 заказов, 0 заказов
//
// one — форма для 1 (заказ), few — для 2-4 (заказа), many — для 5-20 и 0 (заказов).
// 🐞 ЗАДАНИЕ 3: функция ошибается на некоторых числах. Тесты подскажут, на каких.
func Pluralize(n int, one, few, many string) string {
	word := many
	switch n % 10 {
	case 1:
		word = one
	case 2, 3, 4:
		word = few
	}
	return fmt.Sprintf("%d %s", n, word)
}
