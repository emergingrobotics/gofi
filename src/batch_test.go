package gofi

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBatchGet_Success(t *testing.T) {
	ids := []string{"id1", "id2", "id3"}

	getter := func(ctx context.Context, id string) (*string, error) {
		result := "item-" + id
		return &result, nil
	}

	results := BatchGet(context.Background(), ids, getter)

	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("results[%d].Error = %v, want nil", i, result.Error)
		}

		if result.Item == nil {
			t.Errorf("results[%d].Item = nil, want non-nil", i)
			continue
		}

		expected := "item-" + ids[i]
		if *result.Item != expected {
			t.Errorf("results[%d].Item = %s, want %s", i, *result.Item, expected)
		}

		if result.Index != i {
			t.Errorf("results[%d].Index = %d, want %d", i, result.Index, i)
		}
	}
}

func TestBatchGet_WithErrors(t *testing.T) {
	ids := []string{"id1", "id2", "id3"}

	getter := func(ctx context.Context, id string) (*string, error) {
		if id == "id2" {
			return nil, errors.New("not found")
		}
		result := "item-" + id
		return &result, nil
	}

	results := BatchGet(context.Background(), ids, getter)

	// First should succeed
	if results[0].Error != nil {
		t.Errorf("results[0].Error = %v, want nil", results[0].Error)
	}

	// Second should fail
	if results[1].Error == nil {
		t.Error("results[1].Error = nil, want error")
	}

	if results[1].Item != nil {
		t.Error("results[1].Item should be nil when error occurs")
	}

	// Third should succeed
	if results[2].Error != nil {
		t.Errorf("results[2].Error = %v, want nil", results[2].Error)
	}
}

func TestBatchCreate_Success(t *testing.T) {
	items := []*string{
		stringPtr("item1"),
		stringPtr("item2"),
		stringPtr("item3"),
	}

	creator := func(ctx context.Context, item *string) (*string, error) {
		created := "created-" + *item
		return &created, nil
	}

	results := BatchCreate(context.Background(), items, creator)

	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("results[%d].Error = %v, want nil", i, result.Error)
		}

		if result.Item == nil {
			t.Errorf("results[%d].Item = nil, want non-nil", i)
			continue
		}

		expected := "created-" + *items[i]
		if *result.Item != expected {
			t.Errorf("results[%d].Item = %s, want %s", i, *result.Item, expected)
		}
	}
}

func TestBatchDelete_Success(t *testing.T) {
	ids := []string{"id1", "id2", "id3"}

	deleter := func(ctx context.Context, id string) error {
		return nil
	}

	errors := BatchDelete(context.Background(), ids, deleter)

	if len(errors) != 3 {
		t.Fatalf("len(errors) = %d, want 3", len(errors))
	}

	for i, err := range errors {
		if err != nil {
			t.Errorf("errors[%d] = %v, want nil", i, err)
		}
	}
}

func TestBatchDelete_WithErrors(t *testing.T) {
	ids := []string{"id1", "id2", "id3"}

	deleter := func(ctx context.Context, id string) error {
		if id == "id2" {
			return errors.New("delete failed")
		}
		return nil
	}

	errs := BatchDelete(context.Background(), ids, deleter)

	// First should succeed
	if errs[0] != nil {
		t.Errorf("errs[0] = %v, want nil", errs[0])
	}

	// Second should fail
	if errs[1] == nil {
		t.Error("errs[1] = nil, want error")
	}

	// Third should succeed
	if errs[2] != nil {
		t.Errorf("errs[2] = %v, want nil", errs[2])
	}
}

func TestBatchUpdate_Success(t *testing.T) {
	items := []*string{
		stringPtr("item1"),
		stringPtr("item2"),
	}

	updater := func(ctx context.Context, item *string) (*string, error) {
		updated := "updated-" + *item
		return &updated, nil
	}

	results := BatchUpdate(context.Background(), items, updater)

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("results[%d].Error = %v, want nil", i, result.Error)
		}

		if result.Item == nil {
			t.Fatalf("results[%d].Item = nil, want non-nil", i)
		}

		expected := "updated-" + *items[i]
		if *result.Item != expected {
			t.Errorf("results[%d].Item = %s, want %s", i, *result.Item, expected)
		}
	}
}

func TestBatch_ContextCancellation(t *testing.T) {
	ids := []string{"id1", "id2", "id3"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	getter := func(ctx context.Context, id string) (*string, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			result := "item-" + id
			return &result, nil
		}
	}

	results := BatchGet(ctx, ids, getter)

	// All should have context errors
	for i, result := range results {
		if result.Error == nil {
			t.Errorf("results[%d].Error = nil, want context error", i)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
