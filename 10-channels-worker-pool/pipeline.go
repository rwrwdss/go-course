package main

// Generate отправляет числа в канал и ЗАКРЫВАЕТ его, когда числа закончились.
// 🐞 ЗАДАНИЕ 1: потребитель `for v := range Generate(...)` зависает навсегда. Почему?
func Generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// Square — стадия конвейера: читает числа из in, отправляет их квадраты в выходной канал.
// Выходной канал закрывается, когда in закрыт и всё обработано.
// ✍️ ЗАДАНИЕ 2: реализуй.
func Square(in <-chan int) <-chan int {
	out := make(chan int)
	// TODO
	close(out)
	return out
}

// Merge (fan-in) сливает несколько каналов в один. Выходной канал закрывается,
// когда закрыты ВСЕ входные каналы. Порядок значений не важен.
// ✍️ ЗАДАНИЕ 3: реализуй (горутина на каждый вход + sync.WaitGroup + закрытие out после Wait).
func Merge(chans ...<-chan int) <-chan int {
	out := make(chan int)
	// TODO
	close(out)
	return out
}
