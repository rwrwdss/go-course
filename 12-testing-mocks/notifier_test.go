package notify

import (
	"errors"
	"testing"
	"time"
)

// ⚠️ Автотесты пайплайна. Не меняй их. Свои тесты пиши в student_test.go.

// ---------- Тестовые двойники (test doubles) ----------

// fakeClock — часы, которые всегда показывают заданное время.
type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

func at(hour, minute int) fakeClock {
	return fakeClock{now: time.Date(2026, 3, 15, hour, minute, 0, 0, time.UTC)}
}

// stubUsers — заглушка хранилища пользователей.
type stubUsers map[int]User

var errUserNotFound = errors.New("user not found")

func (s stubUsers) GetUser(id int) (User, error) {
	u, ok := s[id]
	if !ok {
		return User{}, errUserNotFound
	}
	return u, nil
}

// spySender — «шпион»: запоминает все отправленные письма, может вернуть заданную ошибку.
type spySender struct {
	sent []sentMail
	err  error
}

type sentMail struct{ to, subject, body string }

func (s *spySender) Send(to, subject, body string) error {
	if s.err != nil {
		return s.err
	}
	s.sent = append(s.sent, sentMail{to, subject, body})
	return nil
}

func mockUsers() stubUsers {
	return stubUsers{
		1: {ID: 1, Name: "Анна", Email: "anna@example.com", OrdersCount: 5, Subscribed: true},
		2: {ID: 2, Name: "Борис", Email: "boris@example.com", OrdersCount: 21, Subscribed: true},
		3: {ID: 3, Name: "Вера", Email: "vera@example.com", OrdersCount: 3, Subscribed: false},
	}
}

// ---------- Тесты ----------

func TestPluralize(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0 заказов"}, {1, "1 заказ"}, {2, "2 заказа"}, {4, "4 заказа"}, {5, "5 заказов"},
		{11, "11 заказов"}, {12, "12 заказов"}, {14, "14 заказов"}, {19, "19 заказов"},
		{21, "21 заказ"}, {22, "22 заказа"}, {101, "101 заказ"}, {111, "111 заказов"}, {112, "112 заказов"},
	}
	for _, tt := range tests {
		if got := Pluralize(tt.n, "заказ", "заказа", "заказов"); got != tt.want {
			t.Errorf("Pluralize(%d) = %q, ожидалось %q", tt.n, got, tt.want)
		}
	}
}

func TestSendDigest_HappyPath(t *testing.T) {
	sender := &spySender{}
	n := NewNotifier(mockUsers(), sender, at(12, 0))

	if err := n.SendDigest(1); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("ожидалось 1 письмо, отправлено %d", len(sender.sent))
	}
	want := sentMail{"anna@example.com", "Ваши заказы", "Привет, Анна! У вас 5 заказов."}
	if sender.sent[0] != want {
		t.Errorf("письмо = %+v, ожидалось %+v", sender.sent[0], want)
	}
}

func TestSendDigest_QuietHours(t *testing.T) {
	for _, clock := range []fakeClock{at(23, 0), at(23, 59), at(3, 0), at(7, 59)} {
		sender := &spySender{}
		err := NewNotifier(mockUsers(), sender, clock).SendDigest(1)
		if !errors.Is(err, ErrQuietHours) {
			t.Errorf("в %s ожидалась ErrQuietHours, получено %v", clock.now.Format("15:04"), err)
		}
		if len(sender.sent) != 0 {
			t.Errorf("в %s письмо не должно уходить", clock.now.Format("15:04"))
		}
	}
	for _, clock := range []fakeClock{at(8, 0), at(22, 59)} {
		if err := NewNotifier(mockUsers(), &spySender{}, clock).SendDigest(1); err != nil {
			t.Errorf("в %s отправка разрешена, получена ошибка %v", clock.now.Format("15:04"), err)
		}
	}
}
