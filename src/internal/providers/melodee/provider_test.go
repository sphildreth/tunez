package melodee

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tunez/tunez/internal/provider"
)

func TestProvider_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/authenticate" {
			json.NewEncoder(w).Encode(map[string]string{"accessToken": "fake-token"})
			return
		}
		if r.URL.Path == "/api/v1/search/songs" {
			q := r.URL.Query().Get("q")
			if q == "test" {
				json.NewEncoder(w).Encode(map[string]any{
					"items": []map[string]any{
						{
							"id":    "1",
							"title": "Test Song",
						},
					},
					"total":   1,
					"hasMore": false,
				})
				return
			}
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	p := New()
	cfg := map[string]any{
		"base_url": server.URL,
		"username": "user",
		"password": "pw",
	}

	if err := p.Initialize(context.Background(), cfg); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	res, err := p.Search(context.Background(), "test", provider.ListReq{})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(res.Tracks.Items) != 1 {
		t.Errorf("Expected 1 track, got %d", len(res.Tracks.Items))
	}
	if res.Tracks.Items[0].Title != "Test Song" {
		t.Errorf("Expected 'Test Song', got %s", res.Tracks.Items[0].Title)
	}
}
