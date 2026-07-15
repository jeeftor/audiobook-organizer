// internal/abs/provider.go
// ABS metadata provider for audiobook-organizer

package abs

import (
	"fmt"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// MetadataProvider provides metadata from ABS
type MetadataProvider struct {
	client       *Client
	mapper       *PathMapper
	libraryID    string
	itemsCache   []LibraryItem // Cache of all items for path lookup
	allLibraries bool          // If true, scan all libraries not just one
}

// SetClient allows replacing the client (useful for adding custom headers)
func (p *MetadataProvider) SetClient(client *Client) {
	p.client = client
}

// NewMetadataProvider creates an ABS metadata provider
// Supports both API-only (with manual mappings) and SQLite+API modes
func NewMetadataProvider(
	apiURL, apiToken, libraryID string,
	pathMappings []PathMapping,
) *MetadataProvider {
	client := NewClient(apiURL, apiToken)
	mapper := NewPathMapper(pathMappings)

	return &MetadataProvider{
		client:       client,
		mapper:       mapper,
		libraryID:    libraryID,
		itemsCache:   nil,
		allLibraries: false,
	}
}

// NewMetadataProviderAllLibraries creates provider that scans ALL libraries
func NewMetadataProviderAllLibraries(
	apiURL, apiToken string,
	pathMappings []PathMapping,
) *MetadataProvider {
	client := NewClient(apiURL, apiToken)
	mapper := NewPathMapper(pathMappings)

	return &MetadataProvider{
		client:       client,
		mapper:       mapper,
		libraryID:    "", // Empty = all libraries
		itemsCache:   nil,
		allLibraries: true,
	}
}

// NewMetadataProviderWithSQLite creates provider with auto path discovery from SQLite
func NewMetadataProviderWithSQLite(
	apiURL, apiToken, libraryID, sqlitePath, userInputPath string,
) (*MetadataProvider, error) {
	// Discover path mappings from SQLite
	mapper, err := NewPathMapperFromSQLite(sqlitePath, userInputPath)
	if err != nil {
		return nil, fmt.Errorf("path discovery failed: %w", err)
	}

	client := NewClient(apiURL, apiToken)

	return &MetadataProvider{
		client:     client,
		mapper:     mapper,
		libraryID:  libraryID,
		itemsCache: nil,
	}, nil
}

// LoadAllItems fetches all library items from ABS (for path matching)
func (p *MetadataProvider) LoadAllItems() error {
	if p.allLibraries {
		return p.loadAllLibraries()
	}

	items, err := p.client.GetAllLibraryItems(p.libraryID)
	if err != nil {
		return fmt.Errorf("fetching library items: %w", err)
	}
	p.itemsCache = items
	return nil
}

// loadAllLibraries fetches items from ALL libraries
func (p *MetadataProvider) loadAllLibraries() error {
	libraries, err := p.client.GetLibraries()
	if err != nil {
		return fmt.Errorf("fetching libraries: %w", err)
	}

	var allItems []LibraryItem
	for _, lib := range libraries {
		items, err := p.client.GetAllLibraryItems(lib.ID)
		if err != nil {
			// Log warning but continue with other libraries
			continue
		}
		// Tag each item with its library info for later
		for i := range items {
			items[i].LibraryID = lib.ID
		}
		allItems = append(allItems, items...)
	}

	p.itemsCache = allItems
	return nil
}

// FindItemByPath finds an ABS item matching the local file path
// Works across all libraries if allLibraries mode is enabled
func (p *MetadataProvider) FindItemByPath(localPath string) (*LibraryItem, error) {
	if p.itemsCache == nil {
		if err := p.LoadAllItems(); err != nil {
			return nil, err
		}
	}

	// Convert local path to ABS path
	absPath := p.mapper.ToABS(localPath)

	// Find matching item
	for _, item := range p.itemsCache {
		if item.Path == absPath {
			return &item, nil
		}
		// Try matching on relative path
		if item.RelPath == absPath {
			return &item, nil
		}
		// Try matching with trailing slash variations
		if strings.TrimSuffix(item.Path, "/") == strings.TrimSuffix(absPath, "/") {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("no ABS item found for path: %s", localPath)
}

// FindItemsByLibrary returns items grouped by library ID
func (p *MetadataProvider) FindItemsByLibrary() map[string][]LibraryItem {
	byLib := make(map[string][]LibraryItem)
	for _, item := range p.itemsCache {
		byLib[item.LibraryID] = append(byLib[item.LibraryID], item)
	}
	return byLib
}

// GetMetadata returns metadata for a local audiobook path
// This implements the organizer.MetadataProvider interface
func (p *MetadataProvider) GetMetadata(localPath string) (organizer.Metadata, error) {
	// Find the ABS item for this path
	item, err := p.FindItemByPath(localPath)
	if err != nil {
		return organizer.NewMetadata(), err
	}

	return p.convertToOrganizerMetadata(item), nil
}

// convertToOrganizerMetadata converts ABS LibraryItem to organizer.Metadata
func (p *MetadataProvider) convertToOrganizerMetadata(item *LibraryItem) organizer.Metadata {
	meta := organizer.NewMetadata()
	meta.SourceType = "abs"
	meta.SourcePath = p.mapper.ToLocal(item.Path)

	absMedia := item.Media.Metadata

	// Title
	meta.Title = absMedia.Title

	// Authors - handle both array format and flattened string format
	for _, author := range absMedia.Authors {
		if author.Name != "" {
			meta.Authors = append(meta.Authors, author.Name)
		}
	}
	// Also check flattened authorName if array is empty
	if len(meta.Authors) == 0 && absMedia.AuthorName != "" {
		meta.Authors = append(meta.Authors, absMedia.AuthorName)
	}
	if len(meta.Authors) == 0 && item.AuthorNamesFirstLast != "" {
		meta.Authors = append(meta.Authors, splitABSNames(item.AuthorNamesFirstLast)...)
	}
	if len(meta.Authors) == 0 && item.AuthorNamesLastFirst != "" {
		meta.Authors = append(meta.Authors, splitABSNames(item.AuthorNamesLastFirst)...)
	}

	// Series
	for _, series := range absMedia.Series {
		if series.Name != "" {
			meta.Series = append(meta.Series, series.Name)
		}
	}
	if len(meta.Series) == 0 && absMedia.SeriesName != "" {
		meta.Series = append(meta.Series, absMedia.SeriesName)
	}

	// Store ABS-specific data in RawData for advanced use
	meta.RawData["title"] = meta.Title
	meta.RawData["authors"] = strings.Join(meta.Authors, ", ")
	meta.RawData["series"] = strings.Join(meta.Series, ", ")
	meta.RawData["series_number"] = absMedia.SeriesSequence
	meta.RawData["narrator"] = absMedia.NarratorName
	meta.RawData["publisher"] = absMedia.Publisher
	meta.RawData["published_year"] = absMedia.PublishedYear
	meta.RawData["published_date"] = absMedia.PublishedDate
	meta.RawData["language"] = absMedia.Language
	meta.RawData["genres"] = strings.Join(absMedia.Genres, ", ")
	meta.RawData["tags"] = strings.Join(absMedia.Tags, ", ")
	meta.RawData["source_path"] = meta.SourcePath
	meta.RawData["authorNamesFirstLast"] = item.AuthorNamesFirstLast
	meta.RawData["authorNamesLastFirst"] = item.AuthorNamesLastFirst
	meta.RawData["abs_item_id"] = item.ID
	meta.RawData["abs_library_id"] = item.LibraryID
	meta.RawData["abs_path"] = item.Path // Original ABS path before mapping
	meta.RawData["abs_relpath"] = item.RelPath
	meta.RawData["abs_duration"] = item.Media.Duration
	meta.RawData["abs_narrator"] = absMedia.NarratorName
	meta.RawData["abs_asin"] = absMedia.ASIN
	meta.RawData["abs_isbn"] = absMedia.ISBN
	meta.RawData["abs_published_year"] = absMedia.PublishedYear
	meta.RawData["abs_explicit"] = absMedia.Explicit

	return meta
}

// GetAllItems returns all library items as organizer metadata
func (p *MetadataProvider) GetAllItems() ([]organizer.Metadata, error) {
	if p.itemsCache == nil {
		if err := p.LoadAllItems(); err != nil {
			return nil, err
		}
	}

	var results []organizer.Metadata
	for _, item := range p.itemsCache {
		meta := p.convertToOrganizerMetadata(&item)
		results = append(results, meta)
	}

	return results, nil
}

func splitABSNames(value string) []string {
	var names []string
	for _, name := range strings.Split(value, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

// ScanLibrary triggers an ABS library scan
func (p *MetadataProvider) ScanLibrary() error {
	return p.client.ScanLibrary(p.libraryID)
}

// GetPathMappings returns the current path mappings (for display/debugging)
func (p *MetadataProvider) GetPathMappings() []PathMapping {
	return p.mapper.Mappings
}

// Mapper returns the path mapper (for path conversion)
func (p *MetadataProvider) Mapper() *PathMapper {
	return p.mapper
}

// Client returns the underlying ABS API client (for advanced use)
func (p *MetadataProvider) Client() *Client {
	return p.client
}
