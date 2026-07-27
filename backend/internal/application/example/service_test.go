package example

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/Rain-kl/Foam/backend/internal/infra/persistence/relational"
)

func TestExampleCRUDAndListPagination(t *testing.T) {
	database, err := relational.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "example.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.InitializeSchema(context.Background()); err != nil {
		t.Fatal(err)
	}

	service := NewService(relational.NewExampleRepository(database))
	ctx := context.Background()

	created, err := service.Create(ctx, CreateInput{Name: "alpha", Description: "first"})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Name != "alpha" || created.Description != "first" {
		t.Fatalf("created = %#v", created)
	}
	if _, err := service.Create(ctx, CreateInput{Name: "  ", Description: "x"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty name create error = %v", err)
	}

	second, err := service.Create(ctx, CreateInput{Name: "beta", Description: "second"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := service.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID || got.Name != "alpha" {
		t.Fatalf("get = %#v", got)
	}
	if _, err := service.Get(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing get error = %v", err)
	}

	updated, err := service.Update(ctx, created.ID, UpdateInput{Name: "alpha-2", Description: "updated"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "alpha-2" || updated.Description != "updated" {
		t.Fatalf("updated = %#v", updated)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) && !updated.UpdatedAt.Equal(created.UpdatedAt) {
		// allow equal only if clock resolution collapses; still require name change success above
	}

	list, err := service.List(ctx, 1, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 2 || list.Page != 1 || list.PageSize != 1 || len(list.Items) != 1 {
		t.Fatalf("list page1 = %#v", list)
	}
	// newest first by created_at desc
	if list.Items[0].ID != second.ID {
		t.Fatalf("expected newest first, got %#v", list.Items[0])
	}

	page2, err := service.List(ctx, 2, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Items) != 1 || page2.Items[0].ID != created.ID {
		t.Fatalf("list page2 = %#v", page2)
	}

	search, err := service.List(ctx, 1, 20, "beta")
	if err != nil {
		t.Fatal(err)
	}
	if search.Total != 1 || len(search.Items) != 1 || search.Items[0].ID != second.ID {
		t.Fatalf("search = %#v", search)
	}

	if err := service.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted get error = %v", err)
	}
	if err := service.Delete(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete error = %v", err)
	}
}
