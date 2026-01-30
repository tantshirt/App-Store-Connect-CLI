package asc

import (
	"context"
	"testing"
)

func TestPaginateAllWithObserver_CallsObserverPerPage(t *testing.T) {
	first := &AppsResponse{
		Links: Links{Next: "next-2"},
		Data: []Resource[AppAttributes]{
			{ID: "1"},
		},
	}

	second := &AppsResponse{
		Links: Links{Next: ""},
		Data: []Resource[AppAttributes]{
			{ID: "2"},
		},
	}

	var pages []int
	var nexts []string

	resp, err := PaginateAllWithObserver(context.Background(), first, func(ctx context.Context, nextURL string) (PaginatedResponse, error) {
		if nextURL != "next-2" {
			t.Fatalf("expected nextURL %q, got %q", "next-2", nextURL)
		}
		return second, nil
	}, func(page int, nextURL string) {
		pages = append(pages, page)
		nexts = append(nexts, nextURL)
	})
	if err != nil {
		t.Fatalf("PaginateAllWithObserver() error: %v", err)
	}

	got, ok := resp.(*AppsResponse)
	if !ok {
		t.Fatalf("expected *AppsResponse, got %T", resp)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 aggregated items, got %d", len(got.Data))
	}

	if len(pages) != 2 {
		t.Fatalf("expected observer to be called twice, got %d", len(pages))
	}
	if pages[0] != 1 || pages[1] != 2 {
		t.Fatalf("unexpected pages: %v", pages)
	}
	if nexts[0] != "next-2" || nexts[1] != "" {
		t.Fatalf("unexpected next URLs: %v", nexts)
	}
}

