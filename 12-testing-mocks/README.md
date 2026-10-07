# Модуль 12 · Тестирование: DI, моки, table-driven тесты, покрытие

## 🎯 Цель
Научиться писать **тестируемый** код: зависимости внедряются через интерфейсы (Dependency Injection), в тестах подменяются фейками, шпионами и стабами. Плюс научиться писать table-driven тесты и пройти **порог покрытия** в CI, как в настоящих командах.

## 📚 Теория

**Почему код бывает нетестируемым.** Функция вызывает `time.Now()`, ходит в SMTP или создаёт подключение к БД внутри себя. Такой тест либо зависит от времени суток, либо требует живой инфраструктуры.

**Dependency Injection** — зависимости передаются снаружи через конструктор, и тип у них интерфейсный:
```go
type Clock interface { Now() time.Time }

func NewNotifier(users UserStore, sender Sender, clock Clock) *Notifier
// прод:   NewNotifier(pgStore, smtpSender, RealClock{})
// тесты:  NewNotifier(stubUsers{...}, &spySender{}, fakeClock{...})
```

**Виды тестовых двойников:**
| Двойник | Что делает | Пример в модуле |
|---|---|---|
| **Stub** | возвращает заранее заданные данные | `stubUsers` |
| **Fake** | упрощённая рабочая реализация | `fakeClock` |
| **Spy** | запоминает, как его вызывали | `spySender.sent` |
| **Mock** | проверяет ожидания вызовов (часто генерируется: `gomock`, `mockery`) | — |

В Go чаще всего хватает ручных двойников на 5–10 строк, внешние библиотеки не нужны.

**Table-driven тесты** — идиома Go:
```go
tests := []struct{ name string; n int; want string }{
    {"единица", 1, "1 заказ"},
    {"исключение 11", 11, "11 заказов"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        if got := Pluralize(tt.n, ...); got != tt.want {
            t.Errorf("got %q, want %q", got, tt.want)
        }
    })
}
```

**`t.Errorf` и `t.Fatalf`:** первый отмечает тест как упавший и продолжает выполнение, второй сразу останавливает тест. **`t.Helper()`** заставляет показывать в ошибке строку вызывающего кода. **`t.Skip`** пропускает тест.

**Покрытие (coverage)** — доля строк кода, которые выполнились во время тестов:
```bash
go test -cover ./12-testing-mocks/
go test -coverprofile=cover.out ./12-testing-mocks/ && go tool cover -html=cover.out  # подсветка в браузере
```
100% покрытия не гарантирует отсутствие багов, но непокрытые ветки — это **точно** непроверенный код. В этом модуле CI требует ≥ 90% (см. `ci.cfg`).

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `notifier.go` | **Работаешь здесь**: сервис уведомлений |
| `plural.go` | **И здесь**: склонение числительных |
| `notifier_test.go` | Автотесты и тестовые двойники. **Не меняй** |
| `student_test.go` | ✍️ **Твои тесты**: заготовки с `t.Skip` |
| `ci.cfg` | Порог покрытия для пайплайна (`coverage_min=90`) |

## 📝 Задания

| # | Что | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `Notifier` / `NewNotifier` | 🐞 DI | сигнатура `NewNotifier(users UserStore, sender Sender, clock Clock)`; стадия **vet** зелёная (тесты компилируются) |
| 2 | `SendDigest` | 🐞 | время берётся из `n.clock`; в 23:00–07:59 → `ErrQuietHours`, в 08:00 и 22:59 отправка разрешена |
| 3 | `Pluralize` | 🐞 | `11/12/14/111 заказов`, `21/101 заказ`, `22 заказа` |
| 4 | `student_test.go` | ✍️ тесты | 5 тестов без `t.Skip`; стадия **coverage** ≥ 90% |

Пайплайн этого модуля состоит из **5 стадий**, последняя `📊 coverage`. Код может быть идеальным, но без тестов пайплайн останется красным.

## ▶️ Как запустить
```bash
go run ./tools/ci 12                       # 🚦 пайплайн (или make ci M=12)
go test -v -run TestPluralize ./12-testing-mocks/
go test -cover ./12-testing-mocks/
```
У модуля нет `main`: это библиотечный пакет, и «запускается» он только тестами.

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

В структуре замени `sender *SMTPSender` на `sender Sender` и добавь поле `clock Clock`. Конструктор принимает три аргумента. Интерфейсы `Sender` и `Clock` уже объявлены в файле.
</details>

<details><summary>Подсказка к заданию 3</summary>

Числа 11–14 — исключение: у них всегда форма many. Проверь `n%100` до того, как смотреть на `n%10`.
</details>

<details><summary>Подсказка к заданию 4</summary>

Шаблон теста:
```go
sender := &spySender{}
err := NewNotifier(mockUsers(), sender, at(12, 0)).SendDigest(3)
if !errors.Is(err, ErrUnsubscribed) { t.Errorf(...) }
if len(sender.sent) != 0 { t.Errorf(...) }
```
Для ошибки отправки: `&spySender{err: errors.New("smtp is down")}`.
</details>

## ✅ Решения

<details><summary>🔒 Решение заданий 1–2 — DI</summary>

```go
type Notifier struct {
	users  UserStore
	sender Sender
	clock  Clock
}

func NewNotifier(users UserStore, sender Sender, clock Clock) *Notifier {
	return &Notifier{users: users, sender: sender, clock: clock}
}

func (n *Notifier) SendDigest(userID int) error {
	hour := n.clock.Now().Hour() // вместо time.Now()
	// ... остальное без изменений
}
```
</details>

<details><summary>🔒 Решение задания 3 — Pluralize</summary>

```go
func Pluralize(n int, one, few, many string) string {
	word := many
	if n%100 < 11 || n%100 > 14 {
		switch n % 10 {
		case 1:
			word = one
		case 2, 3, 4:
			word = few
		}
	}
	return fmt.Sprintf("%d %s", n, word)
}
```
</details>

<details><summary>🔒 Решение задания 4 — student_test.go</summary>

```go
package notify

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSendDigest_Unsubscribed(t *testing.T) {
	sender := &spySender{}
	err := NewNotifier(mockUsers(), sender, at(12, 0)).SendDigest(3)
	if !errors.Is(err, ErrUnsubscribed) {
		t.Errorf("ожидалась ErrUnsubscribed, получено %v", err)
	}
	if len(sender.sent) != 0 {
		t.Errorf("отписанному пользователю ушло %d писем", len(sender.sent))
	}
}

func TestSendDigest_UserNotFound(t *testing.T) {
	err := NewNotifier(mockUsers(), &spySender{}, at(12, 0)).SendDigest(404)
	if !errors.Is(err, errUserNotFound) {
		t.Errorf("ожидалась errUserNotFound, получено %v", err)
	}
}

func TestSendDigest_SenderError(t *testing.T) {
	errSMTP := errors.New("smtp is down")
	err := NewNotifier(mockUsers(), &spySender{err: errSMTP}, at(12, 0)).SendDigest(1)
	if !errors.Is(err, errSMTP) {
		t.Errorf("ожидалась обёрнутая errSMTP, получено %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "anna@example.com") {
		t.Errorf("в ошибке должен быть адрес получателя: %v", err)
	}
}

func TestSendDigest_PluralInBody(t *testing.T) {
	sender := &spySender{}
	if err := NewNotifier(mockUsers(), sender, at(12, 0)).SendDigest(2); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(sender.sent) != 1 || !strings.Contains(sender.sent[0].body, "У вас 21 заказ.") {
		t.Errorf("письма: %+v", sender.sent)
	}
}

func TestRealClock(t *testing.T) {
	if d := time.Since(RealClock{}.Now()); d < 0 || d > time.Second {
		t.Errorf("RealClock отстаёт/спешит на %v", d)
	}
}
```
Итоговое покрытие около 95%. Непокрытым остаётся только `SMTPSender.Send`: он ходит в сеть, и его проверяют интеграционными тестами.
</details>

## 🧠 Вопросы для самопроверки
1. Чем stub отличается от spy и от mock?
2. Почему интерфейс лучше объявлять там, где его **используют** (в пакете сервиса), а не там, где реализуют?
3. Почему 100% покрытия не означает «багов нет»?
4. Как запустить один конкретный подтест? (`go test -run 'TestPluralize/единица'`)
