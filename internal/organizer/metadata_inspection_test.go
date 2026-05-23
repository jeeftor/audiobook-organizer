package organizer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type testMetadataProvider struct {
	metadata Metadata
	err      error
}

func (p testMetadataProvider) GetMetadata() (Metadata, error) {
	return p.metadata, p.err
}

func TestExtractMappedMetadata_AppliesFieldMapping(t *testing.T) {
	mapping := FieldMapping{
		TitleField:   "album",
		SeriesField:  "title",
		AuthorFields: []string{"artist", "album_artist"},
		TrackField:   "track_number",
	}
	provider := testMetadataProvider{
		metadata: Metadata{
			Title:   "Original Title",
			Album:   "Mapped Album",
			Authors: []string{"Original Author"},
			RawData: map[string]interface{}{
				"artist":       "Mapped Artist",
				"album_artist": "Mapped Album Artist",
				"track_number": float64(9),
			},
		},
	}

	metadata, err := ExtractMappedMetadata(provider, mapping)
	if err != nil {
		t.Fatalf("ExtractMappedMetadata() error = %v", err)
	}

	if metadata.Title != "Mapped Album" {
		t.Errorf("Title = %q, want Mapped Album", metadata.Title)
	}
	if len(metadata.Series) != 1 || metadata.Series[0] != "Original Title" {
		t.Errorf("Series = %v, want [Original Title]", metadata.Series)
	}
	if got := metadata.Authors; len(got) != 2 ||
		got[0] != "Mapped Artist" ||
		got[1] != "Mapped Album Artist" {
		t.Errorf("Authors = %v, want [Mapped Artist Mapped Album Artist]", got)
	}
	if metadata.TrackNumber != 9 {
		t.Errorf("TrackNumber = %d, want 9", metadata.TrackNumber)
	}
}

func TestPrepareMetadata_UsesSharedMappedExtraction(t *testing.T) {
	mapping := FieldMapping{
		TitleField:   "album",
		AuthorFields: []string{"artist"},
	}
	provider := testMetadataProvider{
		metadata: Metadata{
			Title:   "Original Title",
			Album:   "Shared Album",
			Authors: []string{"Original Author"},
			RawData: map[string]interface{}{
				"artist": "Shared Artist",
			},
		},
	}
	organizer := &Organizer{
		config: OrganizerConfig{
			FieldMapping: mapping,
		},
	}

	fromOrganizer, err := organizer.prepareMetadata(provider)
	if err != nil {
		t.Fatalf("prepareMetadata() error = %v", err)
	}
	fromHelper, err := ExtractMappedMetadata(provider, mapping)
	if err != nil {
		t.Fatalf("ExtractMappedMetadata() error = %v", err)
	}

	if fromOrganizer.Title != fromHelper.Title {
		t.Errorf(
			"prepareMetadata title = %q, helper title = %q",
			fromOrganizer.Title,
			fromHelper.Title,
		)
	}
	if len(fromOrganizer.Authors) == 0 ||
		len(fromOrganizer.Authors) != len(fromHelper.Authors) ||
		fromOrganizer.Authors[0] != fromHelper.Authors[0] {
		t.Errorf(
			"prepareMetadata authors = %v, helper authors = %v",
			fromOrganizer.Authors,
			fromHelper.Authors,
		)
	}
}

func TestExtractMappedMetadata_ReturnsProviderError(t *testing.T) {
	wantErr := errors.New("broken metadata")
	_, err := ExtractMappedMetadata(testMetadataProvider{err: wantErr}, DefaultFieldMapping())
	if !errors.Is(err, wantErr) {
		t.Fatalf("ExtractMappedMetadata() error = %v, want %v", err, wantErr)
	}
}

func TestInspectMetadataDirectory_ReportsExtractionErrors(t *testing.T) {
	tmpDir := t.TempDir()
	audioPath := filepath.Join(tmpDir, "broken.mp3")
	if err := os.WriteFile(audioPath, []byte("not real audio"), 0o644); err != nil {
		t.Fatalf("failed to write test audio file: %v", err)
	}

	output, err := InspectMetadataDirectory(tmpDir, MetadataInspectionConfig{
		UseEmbeddedMetadata: true,
	})
	if err != nil {
		t.Fatalf("InspectMetadataDirectory() error = %v", err)
	}

	if output.Summary.FilesScanned != 1 {
		t.Fatalf("FilesScanned = %d, want 1", output.Summary.FilesScanned)
	}
	if output.Summary.Errors != 1 {
		t.Fatalf("Errors = %d, want 1", output.Summary.Errors)
	}
	if len(output.Files) != 1 {
		t.Fatalf("Files length = %d, want 1", len(output.Files))
	}
	if output.Files[0].Path != audioPath {
		t.Errorf("Path = %q, want %q", output.Files[0].Path, audioPath)
	}
	if output.Files[0].SourceType != "audio" {
		t.Errorf("SourceType = %q, want audio", output.Files[0].SourceType)
	}
	if output.Files[0].Error == "" {
		t.Fatal("expected extraction error")
	}
}

func TestInspectMetadataDirectory_PreservesHybridMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	fixturePath := filepath.Join(
		"..",
		"..",
		"testdata",
		"mp3flat",
		"charlesdexterward_01_lovecraft_64kb.mp3",
	)
	fixtureBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read MP3 fixture %s: %v", fixturePath, err)
	}
	audioPath := filepath.Join(tmpDir, "book.mp3")
	if err := os.WriteFile(audioPath, fixtureBytes, 0o644); err != nil {
		t.Fatalf("failed to write test audio file: %v", err)
	}
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	metadataContent := `{
		"title": "Hybrid Book",
		"authors": ["Hybrid Author"],
		"series": ["Hybrid Series"]
	}`
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0o644); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}

	output, err := InspectMetadataDirectory(tmpDir, MetadataInspectionConfig{})
	if err != nil {
		t.Fatalf("InspectMetadataDirectory() error = %v", err)
	}

	if output.Summary.FilesScanned != 1 {
		t.Fatalf("FilesScanned = %d, want 1", output.Summary.FilesScanned)
	}
	file := output.Files[0]
	if file.SourceType != "json" {
		t.Errorf("SourceType = %q, want json", file.SourceType)
	}
	if file.Title != "Hybrid Book" {
		t.Errorf("Title = %q, want Hybrid Book", file.Title)
	}
	if file.TrackNumber != 1 {
		t.Errorf("TrackNumber = %d, want embedded track number 1", file.TrackNumber)
	}
	if got := file.RawData["_embedded_source"]; got != audioPath {
		t.Errorf("RawData[_embedded_source] = %v, want %s", got, audioPath)
	}
}
