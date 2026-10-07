package main

// SumParallel считает сумму nums, разбивая слайс на workers частей,
// каждую часть суммирует в отдельной горутине, затем складывает частичные суммы.
// workers < 1 считать как 1. workers больше len(nums) — допустимо.
//
// ✍️ ЗАДАНИЕ 4: реализуй. Тесты запускаются с -race — общая переменная без защиты будет поймана.
func SumParallel(nums []int, workers int) int {
	// TODO
	return 0
}
