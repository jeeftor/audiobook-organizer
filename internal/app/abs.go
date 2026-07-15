package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeeftor/audiobook-organizer/internal/abs"
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// ABSPathMappingRequest requests path mapping validation.
type ABSPathMappingRequest struct {
	Config   ABSConfigDTO `json:"config"`
	InputDir string       `json:"input_dir"`
}

// ABSPathMappingResponse returns resolved mappings.
type ABSPathMappingResponse struct {
	Mappings []PathMappingDTO `json:"mappings"`
}

// ABSItemsRequest requests ABS metadata loading.
type ABSItemsRequest struct {
	Config ABSConfigDTO `json:"config"`
}

// ABSItemsResponse contains ABS items as organizer metadata.
type ABSItemsResponse struct {
	Items []organizer.Metadata `json:"items"`
}

// ABSLibraryStateRequest requests raw ABS library item state.
type ABSLibraryStateRequest struct {
	Config ABSConfigDTO `json:"config"`
}

// ABSLibraryStateResponse contains raw ABS item paths and missing status.
type ABSLibraryStateResponse struct {
	LibraryID string              `json:"library_id"`
	Items     []ABSLibraryItemDTO `json:"items"`
}

// ABSLibraryItemDTO is the REST-safe subset needed by tests and the UI.
type ABSLibraryItemDTO struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	RelPath   string `json:"rel_path"`
	IsMissing bool   `json:"is_missing"`
	IsInvalid bool   `json:"is_invalid"`
	MediaType string `json:"media_type"`
	Title     string `json:"title,omitempty"`
}

// ABSScanTriggerRequest requests an ABS library scan.
type ABSScanTriggerRequest struct {
	Config ABSConfigDTO `json:"config"`
}

// ABSScanTriggerResponse reports scan trigger status.
type ABSScanTriggerResponse struct {
	Triggered bool   `json:"triggered"`
	LibraryID string `json:"library_id"`
}

// ABSCleanMissingRequest requests cleanup of missing ABS library items.
type ABSCleanMissingRequest struct {
	Config ABSConfigDTO `json:"config"`
}

// ABSCleanMissingResponse reports missing-item cleanup status.
type ABSCleanMissingResponse struct {
	Cleaned   bool   `json:"cleaned"`
	LibraryID string `json:"library_id"`
}

// NewABSClient creates an ABS API client with custom headers.
func (s *Service) NewABSClient(cfg ABSConfigDTO) (*abs.Client, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("abs url is required")
	}
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, fmt.Errorf("abs token is required")
	}

	client := abs.NewClient(cfg.URL, cfg.Token)
	if cfg.HeaderFile != "" {
		if err := client.LoadHeadersFromFile(cfg.HeaderFile); err != nil {
			return nil, err
		}
	}
	for _, header := range cfg.Headers {
		if header.Name != "" && header.Value != "" {
			client.SetHeader(header.Name, header.Value)
		}
	}
	return client, nil
}

// ListABSLibraries lists available ABS libraries.
func (s *Service) ListABSLibraries(ctx context.Context, cfg ABSConfigDTO) ([]abs.Library, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	client, err := s.NewABSClient(cfg)
	if err != nil {
		return nil, err
	}
	return client.GetLibraries()
}

// TestABSPathMappings resolves manual or SQLite-derived ABS path mappings.
func (s *Service) TestABSPathMappings(
	ctx context.Context,
	req ABSPathMappingRequest,
) (*ABSPathMappingResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var mappings []abs.PathMapping
	if req.Config.SQLitePath != "" {
		mapper, err := abs.NewPathMapperFromSQLite(req.Config.SQLitePath, req.InputDir)
		if err != nil {
			return nil, err
		}
		mappings = mapper.Mappings
	} else {
		mappings = req.Config.ToPathMappings()
		if len(mappings) == 0 {
			return nil, fmt.Errorf("at least one path mapping or sqlite path is required")
		}
	}

	return &ABSPathMappingResponse{Mappings: pathMappingsToDTO(mappings)}, nil
}

// LoadABSItems loads ABS items as organizer metadata.
func (s *Service) LoadABSItems(
	ctx context.Context,
	req ABSItemsRequest,
) (*ABSItemsResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	provider, err := s.newABSProvider(req.Config)
	if err != nil {
		return nil, err
	}
	if err := provider.LoadAllItems(); err != nil {
		return nil, err
	}
	items, err := provider.GetAllItems()
	if err != nil {
		return nil, err
	}
	return &ABSItemsResponse{Items: items}, nil
}

// LoadABSLibraryState loads raw ABS library item state.
func (s *Service) LoadABSLibraryState(
	ctx context.Context,
	req ABSLibraryStateRequest,
) (*ABSLibraryStateResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	client, err := s.NewABSClient(req.Config)
	if err != nil {
		return nil, err
	}
	libraryID := req.Config.LibraryID
	if libraryID == "" {
		libraryID = "main"
	}
	items, err := client.GetAllLibraryItems(libraryID)
	if err != nil {
		return nil, err
	}
	resp := &ABSLibraryStateResponse{
		LibraryID: libraryID,
		Items:     make([]ABSLibraryItemDTO, 0, len(items)),
	}
	for _, item := range items {
		resp.Items = append(resp.Items, ABSLibraryItemDTO{
			ID:        item.ID,
			Path:      item.Path,
			RelPath:   item.RelPath,
			IsMissing: item.IsMissing,
			IsInvalid: item.IsInvalid,
			MediaType: item.MediaType,
			Title:     item.Media.Metadata.Title,
		})
	}
	return resp, nil
}

// TriggerABSScan triggers an ABS library scan.
func (s *Service) TriggerABSScan(
	ctx context.Context,
	req ABSScanTriggerRequest,
) (*ABSScanTriggerResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	client, err := s.NewABSClient(req.Config)
	if err != nil {
		return nil, err
	}
	libraryID := req.Config.LibraryID
	if libraryID == "" {
		libraryID = "main"
	}
	if err := client.ScanLibrary(libraryID); err != nil {
		return nil, err
	}
	return &ABSScanTriggerResponse{Triggered: true, LibraryID: libraryID}, nil
}

// CleanABSMissing removes library items ABS reports as missing or otherwise having issues.
func (s *Service) CleanABSMissing(
	ctx context.Context,
	req ABSCleanMissingRequest,
) (*ABSCleanMissingResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	client, err := s.NewABSClient(req.Config)
	if err != nil {
		return nil, err
	}
	libraryID := req.Config.LibraryID
	if libraryID == "" {
		libraryID = "main"
	}
	if err := client.RemoveLibraryItemsWithIssues(libraryID); err != nil {
		return nil, err
	}
	return &ABSCleanMissingResponse{Cleaned: true, LibraryID: libraryID}, nil
}

func (s *Service) newABSProvider(cfg ABSConfigDTO) (*abs.MetadataProvider, error) {
	return s.newABSProviderForInput(cfg, "")
}

func (s *Service) newABSProviderForInput(
	cfg ABSConfigDTO,
	inputPath string,
) (*abs.MetadataProvider, error) {
	libraryID := cfg.LibraryID
	if libraryID == "" {
		libraryID = "main"
	}

	var provider *abs.MetadataProvider
	switch {
	case cfg.AllLibraries:
		mappings := cfg.ToPathMappings()
		if len(mappings) == 0 {
			return nil, fmt.Errorf("all-libraries mode requires path mappings")
		}
		provider = abs.NewMetadataProviderAllLibraries(cfg.URL, cfg.Token, mappings)
	case cfg.SQLitePath != "":
		if inputPath == "" && len(cfg.PathMappings) > 0 {
			inputPath = cfg.PathMappings[0].LocalPrefix
		}
		var err error
		provider, err = abs.NewMetadataProviderWithSQLite(
			cfg.URL,
			cfg.Token,
			libraryID,
			cfg.SQLitePath,
			inputPath,
		)
		if err != nil {
			return nil, err
		}
	default:
		mappings := cfg.ToPathMappings()
		if len(mappings) == 0 {
			return nil, fmt.Errorf("path mappings are required")
		}
		provider = abs.NewMetadataProvider(cfg.URL, cfg.Token, libraryID, mappings)
	}

	client, err := s.NewABSClient(cfg)
	if err != nil {
		return nil, err
	}
	provider.SetClient(client)
	return provider, nil
}

func pathMappingsToDTO(mappings []abs.PathMapping) []PathMappingDTO {
	result := make([]PathMappingDTO, 0, len(mappings))
	for _, mapping := range mappings {
		result = append(result, PathMappingDTO{
			ABSPrefix:   mapping.ABSPrefix,
			LocalPrefix: mapping.LocalPrefix,
		})
	}
	return result
}
