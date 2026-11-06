package opensearch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDisabledClient(t *testing.T) {
	client := New(Config{})
	if client.Enabled() {
		t.Fatalf("empty address reports enabled")
	}
	if err := client.IndexPost(context.Background(), "id", PostDocument{}); err == nil {
		t.Fatalf("disabled index accepted")
	}
}

func TestEnsureIndicesCreatesMissing(t *testing.T) {
	var created []string
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.Method == http.MethodHead:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut:
			created = append(created, r.URL.Path)
			bodies = append(bodies, string(body))
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	client := New(Config{Address: server.URL, IndexPrefix: "treeblog"})
	if err := client.EnsureIndices(context.Background()); err != nil {
		t.Fatalf("EnsureIndices: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("created = %v", created)
	}
	joined := strings.Join(bodies, "\n")
	for _, want := range []string{"nori_tokenizer", "treeblog_korean", "content_plain", "published_at"} {
		if !strings.Contains(joined, want) {
			t.Errorf("index body missing %q", want)
		}
	}
}

func TestIndexAndDeletePost(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(server.Close)

	client := New(Config{Address: server.URL})
	ctx := context.Background()
	if err := client.IndexPost(ctx, "post-1", PostDocument{Title: "안녕"}); err != nil {
		t.Fatalf("IndexPost: %v", err)
	}
	if err := client.DeletePost(ctx, "post-1"); err != nil {
		t.Fatalf("DeletePost: %v", err)
	}
	want := []string{"PUT /treeblog-posts/_doc/post-1", "DELETE /treeblog-posts/_doc/post-1"}
	if len(calls) != 2 || calls[0] != want[0] || calls[1] != want[1] {
		t.Fatalf("calls = %v", calls)
	}
}

func TestSearchPostsParsesRanking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hits":{"total":{"value":2},"hits":[
			{"_id":"a","_score":3.5},{"_id":"b","_score":1.2}]}}`))
	}))
	t.Cleanup(server.Close)

	client := New(Config{Address: server.URL})
	ranked, total, err := client.SearchPosts(context.Background(), "트리", "", 10, 0)
	if err != nil {
		t.Fatalf("SearchPosts: %v", err)
	}
	if total != 2 || len(ranked) != 2 || ranked[0].ID != "a" || ranked[0].Score != 3.5 {
		t.Fatalf("ranked = %+v total %d", ranked, total)
	}
}

func TestSearchPostsSurfacesErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("cluster_block_exception"))
	}))
	t.Cleanup(server.Close)

	client := New(Config{Address: server.URL})
	if _, _, err := client.SearchPosts(context.Background(), "x", "", 10, 0); err == nil {
		t.Fatalf("5xx accepted")
	}
}
