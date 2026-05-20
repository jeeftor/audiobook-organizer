package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeeftor/audiobook-organizer/internal/app"
)

const testToken = "test-session-token"

func TestHealthDoesNotRequireAuthentication(t *testing.T) {
	handler := newTestHandler(t)

	rec := performRequest(handler, http.MethodGet, "/api/health", nil, "")

	assertStatus(t, rec, http.StatusOK)
	assertJSONField(t, rec, "status", "ok")
}

func TestProtectedEndpointsRequireSessionToken(t *testing.T) {
	handler := newTestHandler(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "initial config", path: "/api/config/initial"},
		{name: "options", path: "/api/config/options"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRequest(handler, http.MethodGet, tt.path, nil, "")
			assertStatus(t, rec, http.StatusUnauthorized)
			assertJSONField(t, rec, "error", "invalid or missing web session token")
		})
	}
}

func TestProtectedEndpointsAcceptSupportedTokenTransports(t *testing.T) {
	handler := newTestHandler(t)

	t.Run("custom header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/config/initial", nil)
		req.Header.Set("X-Audiobook-Organizer-Token", testToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		assertJSONField(t, rec, "host", "127.0.0.1")
	})

	t.Run("authorization bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/config/options", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assertStatus(t, rec, http.StatusOK)
		assertJSONContainsOption(t, rec, "layouts", "author-series-title-number")
	})

	t.Run("query string", func(t *testing.T) {
		rec := performRequest(
			handler,
			http.MethodGet,
			"/api/config/options?token="+testToken,
			nil,
			"",
		)

		assertStatus(t, rec, http.StatusOK)
		assertJSONContainsOption(t, rec, "scan_modes", "abs")
	})
}

func TestInitialConfigRedactsSecrets(t *testing.T) {
	handler := newTestHandler(t)

	rec := performRequest(handler, http.MethodGet, "/api/config/initial", nil, testToken)

	assertStatus(t, rec, http.StatusOK)
	assertJSONField(t, rec, "abs.token", "redacted")
	assertJSONField(t, rec, "abs.headers.0.value", "redacted")
}

func TestStaticRoutesServeIndexAndSPAFallback(t *testing.T) {
	handler := newTestHandler(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "root", path: "/"},
		{name: "spa fallback", path: "/library/not-a-real-static-file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRequest(handler, http.MethodGet, tt.path, nil, "")

			assertStatus(t, rec, http.StatusOK)
			if !strings.Contains(rec.Body.String(), `<div id="app"></div>`) {
				t.Fatalf("expected embedded app shell, got body:\n%s", rec.Body.String())
			}
			if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(
				contentType,
				"text/html",
			) {
				t.Fatalf("expected text/html content type, got %q", contentType)
			}
		})
	}
}

func TestStaticAssetPathRejectsTraversal(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "root", path: "/", want: "index.html"},
		{name: "asset", path: "/assets/index.js", want: "assets/index.js"},
		{
			name: "spa route",
			path: "/library/not-a-real-static-file",
			want: "library/not-a-real-static-file",
		},
		{name: "parent traversal", path: "/../secret", want: "index.html"},
		{name: "nested parent traversal", path: "/assets/../../secret", want: "index.html"},
		{name: "current directory segment", path: "/assets/./index.js", want: "index.html"},
		{name: "windows separator", path: `/assets\..\secret`, want: "index.html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := staticAssetPath(tt.path)
			if got != tt.want {
				t.Fatalf("staticAssetPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestStaticRoutesFallBackForTraversalRequests(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/%2e%2e/secret", nil)
	rec := httptest.NewRecorder()
	srv.handleStatic(rec, req)

	assertStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), `<div id="app"></div>`) {
		t.Fatalf("expected embedded app shell, got body:\n%s", rec.Body.String())
	}
}

func TestPostEndpointsRejectWrongMethodAndInvalidJSON(t *testing.T) {
	handler := newTestHandler(t)

	tests := []string{
		"/api/organize/preview",
		"/api/organize/run",
		"/api/rename/preview",
		"/api/rename/run",
		"/api/abs/libraries",
		"/api/abs/test-paths",
		"/api/abs/items",
		"/api/abs/library-state",
		"/api/abs/scan-trigger",
		"/api/abs/clean-missing",
	}

	for _, path := range tests {
		t.Run(path+" method", func(t *testing.T) {
			rec := performRequest(handler, http.MethodGet, path, nil, testToken)
			assertStatus(t, rec, http.StatusMethodNotAllowed)
			assertJSONField(t, rec, "error", "method not allowed")
		})

		t.Run(path+" invalid json", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("{"))
			req.Header.Set("X-Audiobook-Organizer-Token", testToken)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assertStatus(t, rec, http.StatusBadRequest)
		})
	}
}

func TestOrganizePreviewEndpointReturnsDryRunSummary(t *testing.T) {
	handler := newTestHandler(t)
	inputDir, outputDir := createOrganizerFixture(t)

	body := map[string]any{
		"config": map[string]any{
			"base_dir":   inputDir,
			"output_dir": outputDir,
			"dry_run":    false,
			"layout":     "author-series-title",
			"field_mapping": map[string]any{
				"title_field":   "title",
				"series_field":  "series",
				"author_fields": []string{"authors"},
			},
		},
	}

	rec := performRequest(handler, http.MethodPost, "/api/organize/preview", body, testToken)

	assertStatus(t, rec, http.StatusOK)
	assertJSONArrayLength(t, rec, "summary.MetadataFound", 1)
	assertJSONArrayLength(t, rec, "summary.Moves", 1)
}

func TestOrganizeRunEndpointMovesFiles(t *testing.T) {
	handler := newTestHandler(t)
	inputDir, outputDir := createOrganizerFixture(t)

	body := map[string]any{
		"config": map[string]any{
			"base_dir":   inputDir,
			"output_dir": outputDir,
			"dry_run":    true,
			"layout":     "author-title",
			"field_mapping": map[string]any{
				"title_field":   "title",
				"series_field":  "series",
				"author_fields": []string{"authors"},
			},
		},
	}

	rec := performRequest(handler, http.MethodPost, "/api/organize/run", body, testToken)

	assertStatus(t, rec, http.StatusOK)
	assertJSONArrayLength(t, rec, "summary.MetadataFound", 1)
	assertJSONArrayLength(t, rec, "summary.Moves", 1)
	resolvedOutputDir, err := filepath.EvalSymlinks(outputDir)
	if err != nil {
		t.Fatalf("resolve output dir: %v", err)
	}
	assertJSONField(t, rec, "log_path", filepath.Join(resolvedOutputDir, ".abook-org.log"))
	assertFileExists(
		t,
		filepath.Join(outputDir, "REST Author", "REST Test Book", "audio.mp3"),
	)
}

func TestRenamePreviewEndpointReturnsCandidates(t *testing.T) {
	handler := newTestHandler(t)
	inputDir := createRenameFixture(t)

	body := map[string]any{
		"config": map[string]any{
			"base_dir":              inputDir,
			"template":              "{author} - {title}",
			"dry_run":               false,
			"author_format":         "first-last",
			"recursive":             true,
			"preserve_path":         true,
			"use_embedded_metadata": false,
			"field_mapping": map[string]any{
				"title_field":   "title",
				"series_field":  "series",
				"author_fields": []string{"authors"},
			},
		},
	}

	rec := performRequest(handler, http.MethodPost, "/api/rename/preview", body, testToken)

	assertStatus(t, rec, http.StatusOK)
	assertJSONField(t, rec, "summary.FilesScanned", float64(1))
	assertJSONArrayLength(t, rec, "candidates", 1)
}

func TestRenameRunEndpointRenamesFiles(t *testing.T) {
	handler := newTestHandler(t)
	inputDir := createRenameFixture(t)

	body := map[string]any{
		"config": map[string]any{
			"base_dir":              inputDir,
			"template":              "{author} - {title}",
			"dry_run":               true,
			"author_format":         "first-last",
			"recursive":             true,
			"preserve_path":         true,
			"use_embedded_metadata": false,
			"field_mapping": map[string]any{
				"title_field":   "title",
				"series_field":  "series",
				"author_fields": []string{"authors"},
			},
		},
	}

	rec := performRequest(handler, http.MethodPost, "/api/rename/run", body, testToken)

	assertStatus(t, rec, http.StatusOK)
	assertJSONField(t, rec, "summary.FilesScanned", float64(1))
	assertJSONField(t, rec, "summary.FilesRenamed", float64(1))
	assertJSONArrayLength(t, rec, "candidates", 1)
	assertJSONField(t, rec, "log_path", filepath.Join(inputDir, ".abook-rename.log"))
	assertFileExists(t, filepath.Join(inputDir, ".abook-rename.log"))
	assertFileExists(
		t,
		filepath.Join(inputDir, "rename_book", "Rename Author - Rename REST Book.mp3"),
	)
	assertFileMissing(t, filepath.Join(inputDir, "rename_book", "audio.mp3"))
}

func TestABSTestPathsEndpointWorksWithoutDocker(t *testing.T) {
	handler := newTestHandler(t)

	body := map[string]any{
		"input_dir": "/host/audiobooks",
		"config": map[string]any{
			"path_mappings": []map[string]string{
				{"abs_prefix": "/audiobooks", "local_prefix": "/host/audiobooks"},
			},
		},
	}

	rec := performRequest(handler, http.MethodPost, "/api/abs/test-paths", body, testToken)

	assertStatus(t, rec, http.StatusOK)
	assertJSONField(t, rec, "mappings.0.abs_prefix", "/audiobooks")
	assertJSONField(t, rec, "mappings.0.local_prefix", "/host/audiobooks")
}

func TestABSEndpointsReturnValidationErrorsBeforeDockerIsNeeded(t *testing.T) {
	handler := newTestHandler(t)

	tests := []struct {
		name          string
		path          string
		body          map[string]any
		expectedError string
	}{
		{
			name:          "libraries require url",
			path:          "/api/abs/libraries",
			body:          map[string]any{},
			expectedError: "abs url is required",
		},
		{
			name: "items require path mappings",
			path: "/api/abs/items",
			body: map[string]any{
				"config": map[string]any{"url": "http://127.0.0.1", "token": "token"},
			},
			expectedError: "path mappings are required",
		},
		{
			name:          "library state requires url",
			path:          "/api/abs/library-state",
			body:          map[string]any{"config": map[string]any{"token": "token"}},
			expectedError: "abs url is required",
		},
		{
			name:          "scan trigger requires url",
			path:          "/api/abs/scan-trigger",
			body:          map[string]any{"config": map[string]any{"token": "token"}},
			expectedError: "abs url is required",
		},
		{
			name:          "clean missing requires url",
			path:          "/api/abs/clean-missing",
			body:          map[string]any{"config": map[string]any{"token": "token"}},
			expectedError: "abs url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRequest(handler, http.MethodPost, tt.path, tt.body, testToken)
			assertStatus(t, rec, http.StatusBadRequest)
			assertJSONField(t, rec, "error", tt.expectedError)
		})
	}
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	return newTestServer(t).routes()
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := app.DefaultWebConfig("127.0.0.1", 0, false, "/input", "/output")
	cfg.ABS = app.ABSConfigDTO{
		URL:       "http://abs.local",
		Token:     "secret-token",
		LibraryID: "main",
		Headers: []app.HeaderDTO{
			{Name: "X-Secret", Value: "secret-header"},
		},
	}
	service := app.NewService(cfg)
	srv, err := New(Config{Token: testToken}, service)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	return srv
}

func performRequest(
	handler http.Handler,
	method, path string,
	body any,
	token string,
) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("X-Audiobook-Organizer-Token", token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func createOrganizerFixture(t *testing.T) (string, string) {
	t.Helper()

	root := t.TempDir()
	inputDir := filepath.Join(root, "input")
	outputDir := filepath.Join(root, "output")
	bookDir := filepath.Join(inputDir, "test_book")

	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		t.Fatalf("create book dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	writeFile(
		t,
		filepath.Join(bookDir, "metadata.json"),
		`{"title":"REST Test Book","authors":["REST Author"],"series":["REST Series #1"]}`,
	)
	writeFile(t, filepath.Join(bookDir, "audio.mp3"), "fake audio")

	return inputDir, outputDir
}

func createRenameFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	bookDir := filepath.Join(root, "rename_book")
	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		t.Fatalf("create rename book dir: %v", err)
	}
	writeFile(
		t,
		filepath.Join(bookDir, "metadata.json"),
		`{"title":"Rename REST Book","authors":["Rename Author"]}`,
	)
	copyFile(
		t,
		filepath.Join("..", "..", "testdata", "mp3flat", "charlesdexterward_01_lovecraft_64kb.mp3"),
		filepath.Join(bookDir, "audio.mp3"),
	)

	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func copyFile(t *testing.T, source, target string) {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read %s: %v", source, err)
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", target, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s\nstat error: %v", path, err)
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be missing: %s\nstat error: %v", path, err)
	}
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body:\n%s", rec.Code, want, rec.Body.String())
	}
}

func assertJSONField(t *testing.T, rec *httptest.ResponseRecorder, path string, want any) {
	t.Helper()
	got := jsonField(t, rec, path)
	if got != want {
		t.Fatalf("%s = %#v, want %#v; body:\n%s", path, got, want, rec.Body.String())
	}
}

func assertJSONContainsOption(t *testing.T, rec *httptest.ResponseRecorder, path, value string) {
	t.Helper()
	items, ok := jsonField(t, rec, path).([]any)
	if !ok {
		t.Fatalf("%s is not an array; body:\n%s", path, rec.Body.String())
	}
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if ok && obj["value"] == value {
			return
		}
	}
	t.Fatalf("%s does not contain option %q; body:\n%s", path, value, rec.Body.String())
}

func assertJSONArrayLength(t *testing.T, rec *httptest.ResponseRecorder, path string, want int) {
	t.Helper()
	items, ok := jsonField(t, rec, path).([]any)
	if !ok {
		t.Fatalf("%s is not an array; body:\n%s", path, rec.Body.String())
	}
	if len(items) != want {
		t.Fatalf("%s length = %d, want %d; body:\n%s", path, len(items), want, rec.Body.String())
	}
}

func jsonField(t *testing.T, rec *httptest.ResponseRecorder, path string) any {
	t.Helper()

	var payload any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v; body:\n%s", err, rec.Body.String())
	}

	current := payload
	for _, segment := range strings.Split(path, ".") {
		switch typed := current.(type) {
		case map[string]any:
			current = typed[segment]
		case []any:
			if segment != "0" {
				t.Fatalf("unsupported array path segment %q in %q", segment, path)
			}
			if len(typed) == 0 {
				t.Fatalf("empty array while resolving %q", path)
			}
			current = typed[0]
		default:
			t.Fatalf("cannot resolve %q through %#v", segment, current)
		}
	}
	return current
}
