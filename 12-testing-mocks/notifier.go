package notify

import (
	"errors"
	"fmt"
	"time"
)

// Ошибки сервиса.
var (
	ErrQuietHours   = errors.New("quiet hours: notifications are not sent from 23:00 to 08:00")
	ErrUnsubscribed = errors.New("user unsubscribed")
)

const (
	quietFrom = 23 // с 23:00
	quietTo   = 8  // до 08:00
)

// User — получатель уведомлений.
type User struct {
	ID          int
	Name        string
	Email       string
	OrdersCount int
	Subscribed  bool
}

// UserStore — откуда берём пользователей (в проде — БД).
type UserStore interface {
	GetUser(id int) (User, error)
}

// Sender — как отправляем письма (в проде — SMTP / внешний API).
type Sender interface {
	Send(to, subject, body string) error
}

// Clock — источник текущего времени. Нужен, чтобы в тестах можно было «подкрутить часы».
type Clock interface {
	Now() time.Time
}

// RealClock — настоящие часы.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// SMTPSender — «настоящая» отправка писем. В тестах сети нет, поэтому она всегда падает.
type SMTPSender struct {
	Host string
}

func (s *SMTPSender) Send(to, subject, body string) error {
	return fmt.Errorf("smtp %s: dial tcp: connection refused", s.Host)
}

// Notifier рассылает дайджесты.
//
// 🐞 ЗАДАНИЕ 1: Notifier жёстко привязан к *SMTPSender и к time.Now() — протестировать его
// без реального SMTP-сервера и без ожидания 23:00 невозможно. Тесты не компилируются (стадия vet).
// Сделай зависимости интерфейсами (Sender, Clock) и принимай их в конструкторе:
//
//	NewNotifier(users UserStore, sender Sender, clock Clock) *Notifier
type Notifier struct {
	users  UserStore
	sender *SMTPSender
}

func NewNotifier(users UserStore, sender *SMTPSender) *Notifier {
	return &Notifier{users: users, sender: sender}
}

// SendDigest отправляет пользователю письмо «Ваши заказы»:
//   - в тихие часы (23:00–08:00)       → ErrQuietHours, письмо не отправляется;
//   - пользователь не найден           → ошибка стора, обёрнутая через %w;
//   - пользователь отписался           → ErrUnsubscribed;
//   - ошибка отправки                  → обёрнутая ошибка sender;
//   - тема: "Ваши заказы", текст: "Привет, Анна! У вас 5 заказов."
//
// 🐞 ЗАДАНИЕ 2: используй часы из зависимости, а не time.Now().
func (n *Notifier) SendDigest(userID int) error {
	hour := time.Now().Hour()
	if hour >= quietFrom || hour < quietTo {
		return ErrQuietHours
	}

	user, err := n.users.GetUser(userID)
	if err != nil {
		return fmt.Errorf("send digest to user %d: %w", userID, err)
	}
	if !user.Subscribed {
		return ErrUnsubscribed
	}

	body := fmt.Sprintf("Привет, %s! У вас %s.", user.Name, Pluralize(user.OrdersCount, "заказ", "заказа", "заказов"))
	if err := n.sender.Send(user.Email, "Ваши заказы", body); err != nil {
		return fmt.Errorf("send digest to %s: %w", user.Email, err)
	}
	return nil
}
