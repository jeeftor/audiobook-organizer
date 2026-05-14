// internal/abs/provider_test.go
// Tests for ABS metadata provider

package abs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockProviderServer creates a test server with library items
func mockProviderServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch {
		case r.URL.Path == "/api/libraries":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"libraries": []Library{
					{ID: "lib_main", Name: "Test Library", MediaType: "book"},
				},
			})

		case strings.HasPrefix(r.URL.Path, "/api/libraries/lib_main/items"):
			json.NewEncoder(w).Encode(LibraryItemsResponse{
				Results: []LibraryItem{
					{
						ID:        "li_001",
						LibraryID: "lib_main",
						Path:      "/audiobooks/Brandon Sanderson/The Final Empire",
						RelPath:   "Brandon Sanderson/The Final Empire",
						Media: Media{
							Metadata: Metadata{
								Title: "The Final Empire",
								Authors: []Author{
									{Name: "Brandon Sanderson"},
								},
								Series: []Series{
									{Name: "Mistborn"},
								},
							},
						},
					},
					{
						ID:        "li_002",
						LibraryID: "lib_main",
						Path:      "/audiobooks/Brandon Sanderson/The Well of Ascension",
						RelPath:   "Brandon Sanderson/The Well of Ascension",
						Media: Media{
							Metadata: Metadata{
								Title: "The Well of Ascension",
								Authors: []Author{
									{Name: "Brandon Sanderson"},
								},
								Series: []Series{
									{Name: "Mistborn"},
								},
							},
						},
					},
				},
				Total: 2,
			})

		case r.URL.Path == "/api/libraries/lib_main/scan":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Scan started"})
		}
	}))
}

func TestMetadataProvider_APIOnly(t *testing.T) {
	server := mockProviderServer()
	defer server.Close()

	// Create provider with manual path mapping
	mappings := []PathMapping{
		{ABSPrefix: "/audiobooks", LocalPrefix: "/mnt/media/audiobooks"},
	}
	provider := NewMetadataProvider(server.URL, "test-token", "lib_main", mappings)

	// Load items
	if err := provider.LoadAllItems(); err != nil {
		t.Fatalf("LoadAllItems failed: %v", err)
	}

	// Get all items
	items, err := provider.GetAllItems()
	if err != nil {
		t.Fatalf("GetAllItems failed: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Check first item
	if items[0].Title != "The Final Empire" {
		t.Errorf("Expected title 'The Final Empire', got %s", items[0].Title)
	}

	if len(items[0].Authors) != 1 || items[0].Authors[0] != "Brandon Sanderson" {
		t.Errorf("Expected author 'Brandon Sanderson', got %v", items[0].Authors)
	}

	if len(items[0].Series) != 1 || items[0].Series[0] != "Mistborn" {
		t.Errorf("Expected series 'Mistborn', got %v", items[0].Series)
	}
}

func TestMetadataProvider_FindItemByPath(t *testing.T) {
	server := mockProviderServer()
	defer server.Close()

	mappings := []PathMapping{
		{ABSPrefix: "/audiobooks", LocalPrefix: "/mnt/media/audiobooks"},
	}
	provider := NewMetadataProvider(server.URL, "test-token", "lib_main", mappings)

	if err := provider.LoadAllItems(); err != nil {
		t.Fatalf("LoadAllItems failed: %v", err)
	}

	// Test finding item by local path
	localPath := "/mnt/media/audiobooks/Brandon Sanderson/The Final Empire"
	item, err := provider.FindItemByPath(localPath)
	if err != nil {
		t.Fatalf("FindItemByPath failed: %v", err)
	}

	if item.Media.Metadata.Title != "The Final Empire" {
		t.Errorf("Expected 'The Final Empire', got %s", item.Media.Metadata.Title)
	}
}

func TestMetadataProvider_GetMetadata(t *testing.T) {
	server := mockProviderServer()
	defer server.Close()

	mappings := []PathMapping{
		{ABSPrefix: "/audiobooks", LocalPrefix: "/mnt/media/audiobooks"},
	}
	provider := NewMetadataProvider(server.URL, "test-token", "lib_main", mappings)

	localPath := "/mnt/media/audiobooks/Brandon Sanderson/The Final Empire"
	meta, err := provider.GetMetadata(localPath)
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if meta.Title != "The Final Empire" {
		t.Errorf("Expected title 'The Final Empire', got %s", meta.Title)
	}

	if meta.SourceType != "abs" {
		t.Errorf("Expected source type 'abs', got %s", meta.SourceType)
	}
}

func TestMetadataProvider_PathMappings(t *testing.T) {
	server := mockProviderServer()
	defer server.Close()

	mappings := []PathMapping{
		{ABSPrefix: "/audiobooks", LocalPrefix: "/mnt/media/audiobooks"},
		{ABSPrefix: "/podcasts", LocalPrefix: "/mnt/media/podcasts"},
	}
	provider := NewMetadataProvider(server.URL, "test-token", "lib_main", mappings)

	returnedMappings := provider.GetPathMappings()
	if len(returnedMappings) != 2 {
		t.Errorf("Expected 2 mappings, got %d", len(returnedMappings))
	}
}

func TestMetadataProvider_ScanLibrary(t *testing.T) {
	server := mockProviderServer()
	defer server.Close()

	mappings := []PathMapping{
		{ABSPrefix: "/audiobooks", LocalPrefix: "/mnt/media/audiobooks"},
	}
	provider := NewMetadataProvider(server.URL, "test-token", "lib_main", mappings)

	if err := provider.ScanLibrary(); err != nil {
		t.Fatalf("ScanLibrary failed: %v", err)
	}
}
