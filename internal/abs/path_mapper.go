// internal/abs/path_mapper.go
// Path mapping between ABS paths and local filesystem paths

package abs

import (
	"database/sql"
	"fmt"
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
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return nil, fmt.Errorf("opening ABS database: %w", err)
	}
	defer db.Close()

	// Query library folders - tables: libraries, libraryFolders, folders
	rows, err := db.Query(`
		SELECT
			l.id as library_id,
			l.name as library_name,
			f.path as folder_path,
			f.fullPath as folder_full_path
		FROM libraries l
		JOIN libraryFolders lf ON l.id = lf.libraryId
		JOIN folders f ON lf.folderId = f.id
	`)
	if err != nil {
		return nil, fmt.Errorf("querying library folders: %w", err)
	}
	defer rows.Close()

	var mappings []PathMapping
	for rows.Next() {
		var libID, libName, folderPath, folderFullPath string
		if err := rows.Scan(&libID, &libName, &folderPath, &folderFullPath); err != nil {
			continue
		}

		// Check if user input path matches this folder
		if strings.HasPrefix(userInputPath, folderFullPath) {
			mappings = append(mappings, PathMapping{
				ABSPrefix:   folderPath,
				LocalPrefix: folderFullPath,
			})
		}
	}

	if len(mappings) == 0 {
		return nil, fmt.Errorf("no ABS library folder matches %s", userInputPath)
	}

	return &PathMapper{Mappings: mappings}, nil
}

// ToLocal converts an ABS path to a local path
func (pm *PathMapper) ToLocal(absPath string) string {
	for _, m := range pm.Mappings {
		// Handle empty ABSPrefix - means ABS uses full local paths
		if m.ABSPrefix == "" {
			// If path already starts with local prefix, it's already local
			if strings.HasPrefix(absPath, m.LocalPrefix) {
				return absPath
			}
			// Otherwise, prepend the local prefix
			return filepath.Join(m.LocalPrefix, absPath)
		}

		// Normal case: replace ABSPrefix with LocalPrefix
		if strings.HasPrefix(absPath, m.ABSPrefix) {
			return strings.Replace(absPath, m.ABSPrefix, m.LocalPrefix, 1)
		}
	}
	// Return as-is if no mapping found
	return absPath
}

// ToABS converts a local path to an ABS path
func (pm *PathMapper) ToABS(localPath string) string {
	for _, m := range pm.Mappings {
		if strings.HasPrefix(localPath, m.LocalPrefix) {
			return strings.Replace(localPath, m.LocalPrefix, m.ABSPrefix, 1)
		}
	}
	return localPath
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
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return nil, fmt.Errorf("opening ABS database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT f.id, f.path, f.fullPath
		FROM folders f
		JOIN libraryFolders lf ON f.id = lf.folderId
	`)
	if err != nil {
		return nil, fmt.Errorf("querying folders: %w", err)
	}
	defer rows.Close()

	var folders []Folder
	for rows.Next() {
		var f Folder
		if err := rows.Scan(&f.ID, &f.Path, &f.FullPath); err != nil {
			continue
		}
		folders = append(folders, f)
	}

	return folders, nil
}
