# Модуль 10 · Каналы: pipeline, fan-in, worker pool, select и таймауты

## 🎯 Цель
Освоить каналы, главный инструмент конкурентности в Go: конвейеры обработки, слияние потоков (fan-in), пул воркеров с ограничением параллельности, `select` с таймаутом. Плюс научиться не допускать двух типичных проблем: deadlock и утечку горутин.

## 📚 Теория

**Канал** — труба, по которой горутины передают значения:
```go
ch := make(chan int)      // небуферизованный: отправитель ждёт получателя
ch := make(chan int, 10)  // буферизованный: отправка блокируется, только когда буфер полон
ch <- 5                   // отправить
v := <-ch                 // получить
v, ok := <-ch             // ok == false → канал закрыт и пуст
close(ch)                 // закрыть: «значений больше не будет»
for v := range ch { }     // читает, пока канал не ЗАКРОЮТ
```

**Правила:**
- Закрывает канал **отправитель**, а не получатель. Удобнее всего через `defer close(out)` в горутине-производителе.
- Если канал не закрыть, `range` по нему будет ждать вечно (deadlock или зависание).
- Отправка в закрытый канал вызывает **panic**.
- Направления `<-chan int` (только чтение) и `chan<- int` (только запись) делают API безопаснее.

**Pipeline** — цепочка стадий, где каждая читает из входного канала и пишет в выходной:
```go
for v := range Square(Generate(1, 2, 3)) { ... }
```

**Fan-in (Merge)** — несколько каналов сливаются в один. На каждый вход запускается горутина, а отдельная горутина делает `wg.Wait()` и `close(out)`.

**Worker pool** — N воркеров читают задачи из общего канала. Так ограничивают параллельность: нельзя открыть 10 000 соединений к БД одновременно.
```
jobs chan ──► worker 1 ──┐
          ──► worker 2 ──┼──► results chan
          ──► worker N ──┘
```

**`select`** ждёт сразу несколько операций и выполняет ту, что готова первой:
```go
select {
case resp := <-ch:
    return resp, nil
case <-time.After(timeout):
    return "", ErrTimeout
}
```

**Утечка горутин.** Если горутина пытается отправить в небуферизованный канал, который уже никто не читает, она висит **вечно** и держит память. Решения: буфер на всех отправителей (`make(chan T, len(replicas))`) или отмена через context (модуль 11).

## 🗂 Файлы
| Файл | Задания |
|---|---|
| `pipeline.go` | 1–3: Generate, Square, Merge |
| `pool.go` | 4: worker pool |
| `replicas.go` | 5: первая ответившая реплика + таймаут |
| `channels_test.go` | Автотесты (с защитой от зависаний). **Не меняй** |

## 📝 Задания

| # | Функция | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `Generate` | 🐞 | `range` по каналу завершается (канал закрыт) |
| 2 | `Square` | ✍️ | `Square(Generate(1,2,3,4))` → `1 4 9 16`, канал закрывается |
| 3 | `Merge` | ✍️ | все значения из всех каналов, выход закрывается после закрытия всех входов |
| 4 | `RunPool` | 🐞 | 24 задачи по 30 мс на 4 воркерах ≈ 180 мс; одновременно не больше 4 задач; результаты отсортированы по `JobID` |
| 5 | `FirstResponse` | ✍️ | возвращает самый быстрый ответ, не дожидаясь остальных; таймаут → `ErrTimeout`; без утечки горутин |

Ожидаемый вывод `go run ./10-channels-worker-pool`:
```
Квадраты: 1 4 9 16 25
Сумма после Merge: 136
Обработано 12 картинок за 300ms (последовательно было бы 1.2s)
Самая быстрая реплика: replica-ru <nil>
```

## ▶️ Как запустить
```bash
go run ./tools/ci 10                  # 🚦 пайплайн (или make ci M=10)
go run ./10-channels-worker-pool
```

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

Не хватает одной строки: `defer close(out)` в начале горутины.
</details>

<details><summary>Подсказка к заданию 2</summary>

Та же схема, что у Generate: горутина, `defer close(out)`, `for v := range in { out <- v*v }`. Убери `close(out)` из заготовки, иначе горутина будет писать в закрытый канал.
</details>

<details><summary>Подсказка к заданию 3</summary>

Для каждого входного канала запусти горутину, которая перекладывает значения в `out` и в конце вызывает `wg.Done()`. Ещё одна горутина: `wg.Wait(); close(out)`. Почему `Wait` нельзя вызвать прямо в `Merge` перед `return`?
</details>

<details><summary>Подсказка к заданию 4</summary>

1) `jobsCh := make(chan Job)`, `resultsCh := make(chan JobResult, len(jobs))`. 2) Запусти `workers` горутин с циклом `for j := range jobsCh`. 3) Отправь все задачи и закрой `jobsCh`. 4) `wg.Wait()`, закрой `resultsCh`, собери результаты. 5) `sort.Slice` по `JobID`.
</details>

<details><summary>Подсказка к заданию 5</summary>

Канал с буфером `len(replicas)`, по горутине на реплику, затем `select` между `<-ch` и `<-time.After(timeout)`.
</details>

## ✅ Решения

<details><summary>🔒 Решение заданий 1–3 — pipeline</summary>

```go
func Generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func Square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

func Merge(chans ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, ch := range chans {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()  // если ждать в самой Merge, получим deadlock:
		close(out) // никто ещё не читает out, а горутины на нём заблокированы
	}()
	return out
}
```
</details>

<details><summary>🔒 Решение задания 4 — worker pool</summary>

```go
func RunPool(jobs []Job, workers int, process func(Job) JobResult) []JobResult {
	if workers < 1 {
		workers = 1
	}
	jobsCh := make(chan Job)
	resultsCh := make(chan JobResult, len(jobs)) // буфер: воркеры не блокируются на отправке

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobsCh {
				resultsCh <- process(j)
			}
		}()
	}

	for _, j := range jobs {
		jobsCh <- j
	}
	close(jobsCh) // воркеры выйдут из range
	wg.Wait()
	close(resultsCh)

	results := make([]JobResult, 0, len(jobs))
	for r := range resultsCh {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].JobID < results[j].JobID })
	return results
}
```
</details>

<details><summary>🔒 Решение задания 5 — FirstResponse</summary>

```go
func FirstResponse(replicas []func() string, timeout time.Duration) (string, error) {
	ch := make(chan string, len(replicas)) // буфер = число отправителей → никто не зависнет
	for _, replica := range replicas {
		go func() {
			ch <- replica()
		}()
	}
	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		return "", ErrTimeout
	}
}
```
С `make(chan string)` без буфера медленные реплики навсегда застрянут на `ch <-`. Это и проверяет тест `NoGoroutineLeak`.
</details>

## 🧠 Вопросы для самопроверки
1. Что произойдёт при чтении из закрытого канала? При записи в закрытый? При чтении из nil-канала?
2. Кто должен закрывать канал и почему?
3. Чем буферизованный канал отличается от небуферизованного с точки зрения синхронизации?
4. Как ограничить число одновременно выполняемых горутин без пула? (Подсказка: семафор на `chan struct{}`.)
5. Когда лучше мьютекс, а когда каналы?
