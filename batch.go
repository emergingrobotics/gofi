package gofi

import (
	"context"
	"sync"
)

// BatchResult represents the result of a batch operation.
type BatchResult[T any] struct {
	// Item is the result item (nil if error occurred).
	Item *T

	// Error is the error that occurred (nil if successful).
	Error error

	// Index is the index of this item in the original batch.
	Index int
}

// BatchGet performs concurrent Get operations and returns results.
// The getter function is called for each ID concurrently.
func BatchGet[T any](
	ctx context.Context,
	ids []string,
	getter func(ctx context.Context, id string) (*T, error),
) []BatchResult[T] {
	results := make([]BatchResult[T], len(ids))
	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)
		go func(idx int, itemID string) {
			defer wg.Done()

			item, err := getter(ctx, itemID)
			results[idx] = BatchResult[T]{
				Item:  item,
				Error: err,
				Index: idx,
			}
		}(i, id)
	}

	wg.Wait()
	return results
}

// BatchCreate performs concurrent Create operations and returns results.
// The creator function is called for each item concurrently.
func BatchCreate[T any](
	ctx context.Context,
	items []*T,
	creator func(ctx context.Context, item *T) (*T, error),
) []BatchResult[T] {
	results := make([]BatchResult[T], len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(idx int, createItem *T) {
			defer wg.Done()

			created, err := creator(ctx, createItem)
			results[idx] = BatchResult[T]{
				Item:  created,
				Error: err,
				Index: idx,
			}
		}(i, item)
	}

	wg.Wait()
	return results
}

// BatchDelete performs concurrent Delete operations and returns errors.
// The deleter function is called for each ID concurrently.
func BatchDelete(
	ctx context.Context,
	ids []string,
	deleter func(ctx context.Context, id string) error,
) []error {
	errors := make([]error, len(ids))
	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)
		go func(idx int, itemID string) {
			defer wg.Done()
			errors[idx] = deleter(ctx, itemID)
		}(i, id)
	}

	wg.Wait()
	return errors
}

// BatchUpdate performs concurrent Update operations and returns results.
func BatchUpdate[T any](
	ctx context.Context,
	items []*T,
	updater func(ctx context.Context, item *T) (*T, error),
) []BatchResult[T] {
	results := make([]BatchResult[T], len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(idx int, updateItem *T) {
			defer wg.Done()

			updated, err := updater(ctx, updateItem)
			results[idx] = BatchResult[T]{
				Item:  updated,
				Error: err,
				Index: idx,
			}
		}(i, item)
	}

	wg.Wait()
	return results
}
