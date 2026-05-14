// internal/abs/client_test.go
// Tests for ABS API client

package abs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockABSServer creates a test server that simulates ABS API
func mockABSServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
			return
		}

		switch r.URL.Path {
		case "/api/libraries":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"libraries": []Library{
					{
						ID:        "lib_main",
						Name:      "Audiobooks",
						MediaType: "book",
						Folders: []Folder{
							{
								ID:       "folder_1",
								Path:     "/audiobooks",
								FullPath: "/mnt/media/audiobooks",
							},
						},
					},
				},
			})

		case "/api/libraries/lib_main/items":
			limit := 100
			offset := 0
			// Parse query params (simplified)
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
				},
				Total:  1,
				Limit:  limit,
				Offset: offset,
			})

		case "/api/items/li_001":
			json.NewEncoder(w).Encode(LibraryItem{
				ID:        "li_001",
				LibraryID: "lib_main",
				Path:      "/audiobooks/Brandon Sanderson/The Final Empire",
				Media: Media{
					Metadata: Metadata{
						Title: "The Final Empire",
						Authors: []Author{
							{Name: "Brandon Sanderson"},
						},
					},
				},
			})

		case "/api/libraries/lib_main/scan":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Scan started"})

		case "/api/libraries/lib_main/issues":
			if r.Method != http.MethodDelete {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"success": true})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestClient_GetLibraries(t *testing.T) {
	server := mockABSServer()
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	libs, err := client.GetLibraries()
	if err != nil {
		t.Fatalf("GetLibraries failed: %v", err)
	}

	if len(libs) != 1 {
		t.Errorf("Expected 1 library, got %d", len(libs))
	}

	if libs[0].Name != "Audiobooks" {
		t.Errorf("Expected library name 'Audiobooks', got %s", libs[0].Name)
	}
}

func TestClient_GetLibraries_AuthError(t *testing.T) {
	server := mockABSServer()
	defer server.Close()

	client := NewClient(server.URL, "wrong-token")

	_, err := client.GetLibraries()
	if err == nil {
		t.Fatal("Expected auth error, got nil")
	}

	if err.Error() != "authentication failed: invalid token" {
		t.Errorf("Expected auth error, got: %v", err)
	}
}

func TestClient_GetAllLibraryItems(t *testing.T) {
	server := mockABSServer()
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	items, err := client.GetAllLibraryItems("lib_main")
	if err != nil {
		t.Fatalf("GetAllLibraryItems failed: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}

	if items[0].Media.Metadata.Title != "The Final Empire" {
		t.Errorf("Expected title 'The Final Empire', got %s", items[0].Media.Metadata.Title)
	}
}

func TestClient_GetLibraryItem(t *testing.T) {
	server := mockABSServer()
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	item, err := client.GetLibraryItem("li_001")
	if err != nil {
		t.Fatalf("GetLibraryItem failed: %v", err)
	}

	if item.ID != "li_001" {
		t.Errorf("Expected ID 'li_001', got %s", item.ID)
	}
}

func TestClient_ScanLibrary(t *testing.T) {
	server := mockABSServer()
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	err := client.ScanLibrary("lib_main")
	if err != nil {
		t.Fatalf("ScanLibrary failed: %v", err)
	}
}

func TestClient_ScanLibraryForce(t *testing.T) {
	var rawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/libraries/lib_main/scan" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rawQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]string{"message": "Scan started"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	if err := client.ScanLibraryForce("lib_main"); err != nil {
		t.Fatalf("ScanLibraryForce failed: %v", err)
	}
	if rawQuery != "force=1" {
		t.Fatalf("ScanLibraryForce query = %q, want force=1", rawQuery)
	}
}

func TestClient_RemoveLibraryItemsWithIssues(t *testing.T) {
	server := mockABSServer()
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	if err := client.RemoveLibraryItemsWithIssues("lib_main"); err != nil {
		t.Fatalf("RemoveLibraryItemsWithIssues failed: %v", err)
	}
}

func TestClient_Integration(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would test against a real ABS instance
	// Requires: ABS_TEST_URL and ABS_TEST_TOKEN env vars
	url := "http://localhost:13378"
	token := "test-token"

	client := NewClient(url, token)
	client.SetHTTPClient(&http.Client{Timeout: 5 * time.Second})

	// Try to connect (will fail if no ABS running, which is expected in CI)
	_, err := client.GetLibraries()
	if err != nil {
		t.Logf("Integration test skipped: %v", err)
		t.Skip("No ABS instance available for integration test")
	}
}
