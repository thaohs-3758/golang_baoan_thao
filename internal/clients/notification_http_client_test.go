package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

func TestNotificationHTTPClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/notifications" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"notifications": []map[string]any{{"id": "n1", "title": "Hello"}},
			"pagination": map[string]any{
				"total": 1,
			},
		})
	}))
	defer server.Close()

	client := NewNotificationHTTPClient(server.URL, server.Client())
	items, total, err := client.List("user-1", repositories.NotificationFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one notification, got total=%d len=%d", total, len(items))
	}
}
