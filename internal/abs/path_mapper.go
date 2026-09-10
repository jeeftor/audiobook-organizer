// internal/abs/path_mapper.go
// Path mapping between ABS paths and local filesystem paths

package abs

import (
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // Pure Go SQLite driver (no CGO, supports cross-compilation)
)

// PathMapping represents a mapping between ABS and local paths
type PathMapping struct {
	ABSPrefix   string // What ABS sees (e.g., "/audiobooks")
	LocalPrefix string // Local path (e.g., "/mnt/media/audiobooks")
}

// PathMapper handles path translation
type PathMapper struct {
	Mappings []PathMapping
}

// NewPathMapper creates a mapper with manual mappings (API-only mode)
func NewPathMapper(mappings []PathMapping) *PathMapper {
	return &PathMapper{Mappings: mappings}
}

// NewPathMapperFromSQLite discovers mappings from ABS SQLite database
func NewPathMapperFromSQLite(dbPath string, userInputPath string) (*PathMapper, error) {
	// Open in read-only mode
	db, err := sql.Open(
		"sqlite",
		(&url.URL{Scheme: "file", Path: dbPath, RawQuery: "mode=ro"}).String(),
	)
	if err != nil {
		return nil, fmt.Errorf("opening ABS database: %w", err)
	}
	defer db.Close()

	// ABS stores the container-visible absolute path directly in libraryFolders.
	rows, err := db.Query(`SELECT id, path FROM libraryFolders`)
	if err != nil {
		return nil, fmt.Errorf("querying library folders: %w", err)
	}
	defer rows.Close()

	var mappings []PathMapping
	for rows.Next() {
		var folderID, folderPath string
		if err := rows.Scan(&folderID, &folderPath); err != nil {
			return nil, fmt.Errorf("reading library folder: %w", err)
		}

		// Check if user input path matches this folder
		if pathPrefix(userInputPath, folderPath) {
			mappings = append(mappings, PathMapping{
				ABSPrefix:   folderPath,
				LocalPrefix: folderPath,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading library folders: %w", err)
	}
	if len(mappings) == 0 {
		return nil, fmt.Errorf(
			"no ABS library folder matches %s; use --abs-path-map when the local mount differs from the ABS path",
			userInputPath,
		)
	}

	return &PathMapper{Mappings: mappings}, nil
}

// ToLocal converts an ABS path to a local path
func (pm *PathMapper) ToLocal(absPath string) string {
	best := -1
	length := -1
	for i, m := range pm.Mappings {
		if m.ABSPrefix == "" {
			if best < 0 {
				best = i
				length = 0
			}
		} else if pathPrefix(absPath, m.ABSPrefix) && len(filepath.Clean(m.ABSPrefix)) > length {
			best = i
			length = len(filepath.Clean(m.ABSPrefix))
		}
	}
	if best < 0 {
		return absPath
	}
	m := pm.Mappings[best]
	if m.ABSPrefix == "" {
		if pathPrefix(absPath, m.LocalPrefix) {
			return absPath
		}
		return filepath.Join(m.LocalPrefix, absPath)
	}
	suffix, _ := filepath.Rel(filepath.Clean(m.ABSPrefix), filepath.Clean(absPath))
	return filepath.Join(m.LocalPrefix, suffix)
}

// ToABS converts a local path to an ABS path
func (pm *PathMapper) ToABS(localPath string) string {
	best := -1
	for i, m := range pm.Mappings {
		if pathPrefix(localPath, m.LocalPrefix) &&
			(best < 0 || len(filepath.Clean(m.LocalPrefix)) > len(filepath.Clean(pm.Mappings[best].LocalPrefix))) {
			best = i
		}
	}
	if best < 0 {
		return localPath
	}
	m := pm.Mappings[best]
	suffix, _ := filepath.Rel(filepath.Clean(m.LocalPrefix), filepath.Clean(localPath))
	if m.ABSPrefix == "" {
		return "/" + filepath.ToSlash(suffix)
	}
	return filepath.ToSlash(filepath.Join(m.ABSPrefix, suffix))
}

// pathPrefix requires a whole path component, not a textual prefix.
func pathPrefix(value, prefix string) bool {
	relative, err := filepath.Rel(filepath.Clean(prefix), filepath.Clean(value))
	return err == nil && relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

// ParsePathMapping parses a path mapping from CLI format: "/abs:/local"
func ParsePathMapping(s string) (PathMapping, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return PathMapping{}, fmt.Errorf(
			"invalid path mapping format: %s (expected '/abs:/local')",
			s,
		)
	}
	return PathMapping{
		ABSPrefix:   parts[0],
		LocalPrefix: parts[1],
	}, nil
}

// ListLibraries returns all library paths from SQLite (for debugging)
func ListLibraries(dbPath string) ([]Folder, error) {
	db, err := sql.Open(
		"sqlite",
		(&url.URL{Scheme: "file", Path: dbPath, RawQuery: "mode=ro"}).String(),
	)
	if err != nil {
		return nil, fmt.Errorf("opening ABS database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, path, path, libraryId FROM libraryFolders
	`)
	if err != nil {
		return nil, fmt.Errorf("querying folders: %w", err)
	}
	defer rows.Close()

	var folders []Folder
	for rows.Next() {
		var f Folder
		if err := rows.Scan(&f.ID, &f.Path, &f.FullPath, &f.LibraryID); err != nil {
			continue
		}
		folders = append(folders, f)
	}

	return folders, rows.Err()
}
