package abs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestReviewPaginationContract(t *testing.T) {
	items := make([]LibraryItem, 101)
	for i := range items {
		items[i].ID = fmt.Sprint(i)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		start := page * limit
		end := min(start+limit, len(items))
		t.Logf("request=%s", r.URL.String())
		json.NewEncoder(w).
			Encode(map[string]any{"results": items[start:end], "total": len(items), "limit": limit, "page": page})
	}))
	defer srv.Close()
	got, err := NewClient(srv.URL, "fixture-token").GetAllLibraryItems("library")
	if err != nil {
		t.Fatal(err)
	}
	unique := map[string]bool{}
	for _, item := range got {
		unique[item.ID] = true
	}
	if len(got) != 101 || len(unique) != 101 {
		t.Fatalf(
			"want 101 distinct books; got %d entries and %d distinct books",
			len(got),
			len(unique),
		)
	}
}

func TestReviewPathMappingBoundary(t *testing.T) {
	mapper := NewPathMapper([]PathMapping{{ABSPrefix: "/books", LocalPrefix: "/local"}})
	if got := mapper.ToLocal("/books-other/book"); got != "/books-other/book" {
		t.Errorf("unrelated library remapped to %s", got)
	}
	nested := NewPathMapper(
		[]PathMapping{
			{ABSPrefix: "/books", LocalPrefix: "/local"},
			{ABSPrefix: "/books/special", LocalPrefix: "/special"},
		},
	)
	if got := nested.ToLocal("/books/special/book"); got != "/special/book" {
		t.Errorf("specific mapping ignored: %s", got)
	}
	if got := mapper.ToABS("/local-other/book"); got != "/local-other/book" {
		t.Errorf("unrelated local prefix remapped: %s", got)
	}
	reverse := NewPathMapper(
		[]PathMapping{
			{ABSPrefix: "/books", LocalPrefix: "/local/"},
			{ABSPrefix: "/special", LocalPrefix: "/local/nested/"},
		},
	)
	if got := reverse.ToABS("/local/nested/book"); got != "/special/book" {
		t.Errorf("specific reverse mapping ignored: %s", got)
	}
	if got := reverse.ToABS("/local"); got != "/books" {
		t.Errorf("exact root mapping failed: %s", got)
	}
}

func TestReviewPaginationStopsOnIncompleteResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(LibraryItemsResponse{Total: 1})
	}))
	defer server.Close()
	if _, err := NewClient(server.URL, "token").GetAllLibraryItems("library"); err == nil {
		t.Fatal("silently accepted incomplete pagination")
	}
}
