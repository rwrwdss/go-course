package main

import (
	"reflect"
	"testing"
)

// ⚠️ Это автотесты пайплайна. Не меняй их — меняй код в users.go.

func ids(users []User) []int {
	out := []int{}
	for _, u := range users {
		out = append(out, u.ID)
	}
	return out
}

func TestActiveUsers(t *testing.T) {
	got := ids(ActiveUsers(mockUsers()))
	want := []int{1, 2, 4, 6, 7}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ActiveUsers вернул ID %v, ожидалось %v", got, want)
	}
}

func TestFindByID_ReturnsPointerIntoSlice(t *testing.T) {
	users := mockUsers()
	u, ok := FindByID(users, 4)
	if !ok || u == nil {
		t.Fatalf("пользователь с ID=4 не найден")
	}
	u.Email = "changed@example.com"
	if users[3].Email != "changed@example.com" {
		t.Errorf("изменение через указатель не попало в слайс: users[3].Email = %q", users[3].Email)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	u, ok := FindByID(mockUsers(), 999)
	if ok || u != nil {
		t.Errorf("для несуществующего ID ожидалось (nil, false), получено (%v, %v)", u, ok)
	}
}

func TestApplyBonus(t *testing.T) {
	users := mockUsers()
	ApplyBonus(users, 100)
	want := map[int]int{1: 1600, 2: 400, 3: 0, 4: 9900, 5: 50, 6: 820, 7: 4200}
	for _, u := range users {
		if u.Balance != want[u.ID] {
			t.Errorf("баланс пользователя #%d (%s) = %d, ожидалось %d", u.ID, u.Name, u.Balance, want[u.ID])
		}
	}
}

func TestAdultNames(t *testing.T) {
	got := AdultNames(mockUsers())
	want := []string{"Анна", "Виктор", "Галина", "Елена", "Жанна"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AdultNames = %v, ожидалось %v", got, want)
	}
}

func TestPaginate(t *testing.T) {
	tests := []struct {
		name string
		page int
		size int
		want []int
	}{
		{"первая страница", 1, 3, []int{1, 2, 3}},
		{"вторая страница", 2, 3, []int{4, 5, 6}},
		{"последняя неполная страница", 3, 3, []int{7}},
		{"страница за пределами", 4, 3, []int{}},
		{"размер больше данных", 1, 100, []int{1, 2, 3, 4, 5, 6, 7}},
		{"нулевая страница", 0, 3, []int{}},
		{"отрицательный размер", 1, -1, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Paginate(mockUsers(), tt.page, tt.size)
			if res == nil {
				t.Fatalf("Paginate вернул nil — нужно вернуть пустой слайс []User{} (в JSON это [] вместо null)")
			}
			if got := ids(res); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Paginate(page=%d, size=%d) вернул ID %v, ожидалось %v", tt.page, tt.size, got, tt.want)
			}
		})
	}
}
