package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathValidationRequest requests non-mutating local path validation.
type PathValidationRequest struct {
	Paths []PathValidationItem `json:"paths"`
}

// PathValidationItem identifies one local path to validate.
type PathValidationItem struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Kind string `json:"kind"`
}

// PathValidationResponse contains validation results keyed by request item ID.
type PathValidationResponse struct {
	Results []PathValidationResult `json:"results"`
}

// PathValidationResult reports whether a local path is usable.
type PathValidationResult struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// DirectoryBrowseResponse contains a permitted server-side directory and its children.
type DirectoryBrowseResponse struct {
	Path        string           `json:"path"`
	Parent      string           `json:"parent,omitempty"`
	Directories []DirectoryEntry `json:"directories"`
}

// DirectoryEntry identifies a selectable server-side directory.
type DirectoryEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// BrowseDirectories lists directories beneath configured browse roots.
func (s *Service) BrowseDirectories(
	ctx context.Context,
	path string,
) (*DirectoryBrowseResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(s.config.BrowseRoots) == 0 {
		return nil, fmt.Errorf("server folder browsing is not configured")
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, fmt.Errorf("resolve directory path: %w", err)
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("access directory path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}
	if !isWithinBrowseRoots(resolvedPath, s.config.BrowseRoots) {
		return nil, fmt.Errorf("directory is outside configured browse roots")
	}

	entries, err := os.ReadDir(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}
	response := &DirectoryBrowseResponse{Path: resolvedPath}
	parent := filepath.Dir(resolvedPath)
	if isWithinBrowseRoots(parent, s.config.BrowseRoots) {
		response.Parent = parent
	}
	for _, entry := range entries {
		childPath := filepath.Join(resolvedPath, entry.Name())
		childInfo, statErr := os.Stat(childPath)
		if statErr != nil || !childInfo.IsDir() {
			continue
		}
		resolvedChild, resolveErr := filepath.EvalSymlinks(childPath)
		if resolveErr != nil || !isWithinBrowseRoots(resolvedChild, s.config.BrowseRoots) {
			continue
		}
		response.Directories = append(
			response.Directories,
			DirectoryEntry{Name: entry.Name(), Path: resolvedChild},
		)
	}
	return response, nil
}

func isWithinBrowseRoots(path string, roots []string) bool {
	for _, root := range roots {
		resolvedRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		relative, err := filepath.Rel(resolvedRoot, path)
		if err == nil && relative != ".." &&
			!strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// ValidatePaths checks local workflow paths without creating directories.
func (s *Service) ValidatePaths(
	ctx context.Context,
	req PathValidationRequest,
) (*PathValidationResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	results := make([]PathValidationResult, 0, len(req.Paths))
	for _, path := range req.Paths {
		results = append(results, validatePath(path))
	}
	return &PathValidationResponse{Results: results}, nil
}

func validatePath(item PathValidationItem) PathValidationResult {
	result := PathValidationResult{
		ID:   item.ID,
		Path: strings.TrimSpace(item.Path),
	}
	if result.Path == "" {
		result.Error = "Path is required."
		return result
	}

	switch item.Kind {
	case "output-directory":
		result.Error = validateOutputDirectory(result.Path)
	default:
		result.Error = validateExistingDirectory(result.Path)
	}
	result.Valid = result.Error == ""
	return result
}

func validateExistingDirectory(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("Directory does not exist: %s", path)
		}
		if os.IsPermission(err) {
			return fmt.Sprintf("Permission denied accessing directory: %s", path)
		}
		return fmt.Sprintf("Cannot access directory %s: %v", path, err)
	}
	if !info.IsDir() {
		return fmt.Sprintf("Path is not a directory: %s", path)
	}
	return ""
}

func validateOutputDirectory(path string) string {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Sprintf("Output path is not a directory: %s", path)
		}
		return ""
	}
	if os.IsPermission(err) {
		return fmt.Sprintf("Permission denied accessing output directory: %s", path)
	}
	if !os.IsNotExist(err) {
		return fmt.Sprintf("Cannot access output directory %s: %v", path, err)
	}

	parent := filepath.Dir(path)
	parentInfo, parentErr := os.Stat(parent)
	if parentErr != nil {
		if os.IsNotExist(parentErr) {
			return fmt.Sprintf("Output parent directory does not exist: %s", parent)
		}
		if os.IsPermission(parentErr) {
			return fmt.Sprintf("Permission denied accessing output parent directory: %s", parent)
		}
		return fmt.Sprintf("Cannot access output parent directory %s: %v", parent, parentErr)
	}
	if !parentInfo.IsDir() {
		return fmt.Sprintf("Output parent path is not a directory: %s", parent)
	}
	return ""
}
