package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/app"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/config/initial", s.withAuth(s.handleInitialConfig))
	mux.HandleFunc("/api/config/options", s.withAuth(s.handleOptions))
	mux.HandleFunc("/api/organize/preview", s.withAuth(s.handleOrganizePreview))
	mux.HandleFunc("/api/organize/run", s.withAuth(s.handleOrganizeRun))
	mux.HandleFunc("/api/rename/preview", s.withAuth(s.handleRenamePreview))
	mux.HandleFunc("/api/abs/libraries", s.withAuth(s.handleABSLibraries))
	mux.HandleFunc("/api/abs/test-paths", s.withAuth(s.handleABSTestPaths))
	mux.HandleFunc("/api/abs/items", s.withAuth(s.handleABSItems))
	mux.HandleFunc("/api/abs/library-state", s.withAuth(s.handleABSLibraryState))
	mux.HandleFunc("/api/abs/scan-trigger", s.withAuth(s.handleABSScanTrigger))
	mux.HandleFunc("/api/abs/clean-missing", s.withAuth(s.handleABSCleanMissing))
	mux.HandleFunc("/", s.handleStatic)

	return mux
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.config.Token == "" || requestToken(r) == s.config.Token {
			next(w, r)
			return
		}
		writeError(w, http.StatusUnauthorized, errors.New("invalid or missing web session token"))
	}
}

func requestToken(r *http.Request) string {
	if token := r.Header.Get("X-Audiobook-Organizer-Token"); token != "" {
		return token
	}
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return r.URL.Query().Get("token")
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleInitialConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.app.Config())
}

func (s *Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.app.Options(r.Context()))
}

func (s *Server) handleOrganizePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req app.OrganizeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.app.PreviewOrganize(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleOrganizeRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req app.OrganizeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.app.RunOrganize(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRenamePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req app.RenameRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.app.PreviewRename(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleABSCleanMissing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req app.ABSCleanMissingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.app.CleanABSMissing(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleABSLibraries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var cfg app.ABSConfigDTO
	if !decodeJSON(w, r, &cfg) {
		return
	}
	libraries, err := s.app.ListABSLibraries(r.Context(), cfg)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"libraries": libraries})
}

func (s *Server) handleABSTestPaths(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req app.ABSPathMappingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.app.TestABSPathMappings(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleABSItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req app.ABSItemsRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.app.LoadABSItems(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleABSLibraryState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req app.ABSLibraryStateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.app.LoadABSLibraryState(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleABSScanTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req app.ABSScanTriggerRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.app.TriggerABSScan(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	if _, err := fs.Stat(s.static, path); err != nil {
		path = "index.html"
	}
	file, err := s.static.Open(path)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if stat.IsDir() {
		path = strings.TrimSuffix(path, "/") + "/index.html"
		file, err = s.static.Open(path)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		defer file.Close()
		stat, err = file.Stat()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if contentType := mime.TypeByExtension(filepath.Ext(path)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	http.ServeContent(w, r, path, stat.ModTime(), bytes.NewReader(data))
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
