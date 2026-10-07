package main

import (
	"errors"
	"reflect"
	"testing"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

func TestRepo_List_SortedByID(t *testing.T) {
	repo := NewInMemoryProductRepo(mockProducts())
	for i := 0; i < 20; i++ {
		var got []int
		for _, p := range repo.List() {
			got = append(got, p.ID)
		}
		if want := []int{1, 2, 3, 4}; !reflect.DeepEqual(got, want) {
			t.Fatalf("List() вернул ID в порядке %v, ожидалось %v", got, want)
		}
	}
}

func TestRepo_Create_AssignsUniqueIDs(t *testing.T) {
	repo := NewInMemoryProductRepo(mockProducts())
	a := repo.Create(Product{Name: "A"})
	b := repo.Create(Product{Name: "B"})
	if a.ID != 5 || b.ID != 6 {
		t.Errorf("ожидались ID 5 и 6, получено %d и %d", a.ID, b.ID)
	}
	if n := len(repo.List()); n != 6 {
		t.Errorf("в репозитории %d товаров, ожидалось 6", n)
	}
}

func TestRepo_Delete(t *testing.T) {
	repo := NewInMemoryProductRepo(mockProducts())
	if err := repo.Delete(3); err != nil {
		t.Fatalf("Delete(3) вернул ошибку: %v", err)
	}
	if _, err := repo.GetByID(3); !errors.Is(err, ErrNotFound) {
		t.Errorf("после удаления GetByID(3) должен вернуть ErrNotFound, получено %v", err)
	}
	if err := repo.Delete(3); !errors.Is(err, ErrNotFound) {
		t.Errorf("повторный Delete(3) должен вернуть ErrNotFound, получено %v", err)
	}
}
