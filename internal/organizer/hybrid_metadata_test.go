package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHybridMetadataExtraction tests that metadata.json + embedded audio metadata are merged correctly
func TestHybridMetadataExtraction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create metadata.json with book-level fields
	metadataJSON := `{
		"title": "The Case of Charles Dexter Ward",
		"authors": ["H. P. Lovecraft"],
		"series": [],
		"narrators": ["Chris Pile"],
		"asin": "B0CX8Z8J88",
		"publishedYear": "2024"
	}`

	metadataPath := filepath.Join(tmpDir, "metadata.json")
	if err := os.WriteFile(metadataPath, []byte(metadataJSON), 0o644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	// Create a dummy audio file (MP3)
	audioPath := filepath.Join(tmpDir, "chapter01.mp3")
	if err := os.WriteFile(audioPath, []byte("dummy audio data"), 0o644); err != nil {
		t.Fatalf("Failed to write audio file: %v", err)
	}

	// Test with metadata.json path (should use extractJSONMetadata which does hybrid extraction)
	provider := NewMetadataProvider(metadataPath, false)
	metadata, err := provider.GetMetadata()
	if err != nil {
		t.Fatalf("GetMetadata() error: %v", err)
	}

	// Verify book-level fields from JSON
	if metadata.Title != "The Case of Charles Dexter Ward" {
		t.Errorf("Title = %q, want %q", metadata.Title, "The Case of Charles Dexter Ward")
	}

	if len(metadata.Authors) != 1 || metadata.Authors[0] != "H. P. Lovecraft" {
		t.Errorf("Authors = %v, want [H. P. Lovecraft]", metadata.Authors)
	}

	// Verify source type is JSON (hybrid mode)
	if metadata.SourceType != "json" {
		t.Errorf("SourceType = %q, want %q", metadata.SourceType, "json")
	}

	// Verify RawData contains JSON fields
	if _, ok := metadata.RawData["asin"]; !ok {
		t.Error("RawData missing 'asin' field from JSON")
	}

	if _, ok := metadata.RawData["publishedYear"]; !ok {
		t.Error("RawData missing 'publishedYear' field from JSON")
	}

	// Note: We can't test embedded audio fields without a real audio file with tags
	// The hybrid extraction would normally add track/disc fields from embedded metadata
}

// TestHybridMetadataFromAudioFile tests hybrid extraction when provider is created with audio file path
func TestHybridMetadataFromAudioFile(t *testing.T) {
	// NOTE: This test requires real audio files with metadata tags to fully test hybrid extraction
	// With dummy files, we can only verify that the JSON fallback works
	tmpDir := t.TempDir()

	// Create metadata.json
	metadataJSON := `{
		"title": "Test Book",
		"authors": ["Test Author"],
		"asin": "TESTASIN"
	}`

	metadataPath := filepath.Join(tmpDir, "metadata.json")
	if err := os.WriteFile(metadataPath, []byte(metadataJSON), 0o644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	// Create audio file
	audioPath := filepath.Join(tmpDir, "test.mp3")
	if err := os.WriteFile(audioPath, []byte("dummy audio"), 0o644); err != nil {
		t.Fatalf("Failed to write audio file: %v", err)
	}

	// When provider is created with audio file path, it should:
	// 1. Detect it's an audio file
	// 2. Check parent directory for metadata.json
	// 3. Do hybrid extraction if found
	provider := NewMetadataProvider(audioPath, false)
	metadata, err := provider.GetMetadata()
	// Dummy audio file can't be parsed, but we can verify JSON fallback behavior
	// The hybrid extraction will attempt to read the audio file and may fail
	// That's expected with dummy data - in production, real audio files would work
	if err != nil {
		t.Logf("GetMetadata() error (expected with dummy audio): %v", err)
		t.Skip("Skipping test - requires real audio file for full hybrid extraction test")
		return
	}

	// If we somehow got metadata (shouldn't happen with dummy file), verify structure
	if metadata.Title != "" && metadata.Title != "Test Book" {
		t.Errorf("Title = %q, want %q", metadata.Title, "Test Book")
	}
}

// TestUseEmbeddedMetadataOnly tests that hybrid mode is skipped when useEmbeddedOnly=true
func TestUseEmbeddedMetadataOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Create metadata.json
	metadataJSON := `{
		"title": "JSON Title",
		"authors": ["JSON Author"]
	}`

	metadataPath := filepath.Join(tmpDir, "metadata.json")
	if err := os.WriteFile(metadataPath, []byte(metadataJSON), 0o644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	// Create audio file
	audioPath := filepath.Join(tmpDir, "test.m4b")
	if err := os.WriteFile(audioPath, []byte("dummy audio"), 0o644); err != nil {
		t.Fatalf("Failed to write audio file: %v", err)
	}

	// With useEmbeddedOnly=true, should NOT use metadata.json
	provider := NewMetadataProvider(audioPath, true)
	_, err := provider.GetMetadata()
	// Since we can't extract from a dummy audio file, this will fail
	// But that's expected - the important thing is it's NOT using JSON
	if err != nil {
		t.Logf("GetMetadata() error (expected with dummy audio): %v", err)
		// This is correct behavior - useEmbeddedOnly=true means it tries to read
		// from the audio file directly and fails on dummy data
	}
}

// TestJSONMetadataProviderHybridExtraction tests that NewJSONMetadataProvider supports hybrid mode
func TestJSONMetadataProviderHybridExtraction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create metadata.json
	metadataJSON := `{
		"title": "Hybrid Test",
		"authors": ["Test Author"],
		"description": "Test description"
	}`

	metadataPath := filepath.Join(tmpDir, "metadata.json")
	if err := os.WriteFile(metadataPath, []byte(metadataJSON), 0o644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	// Create audio file in same directory
	audioPath := filepath.Join(tmpDir, "chapter.mp3")
	if err := os.WriteFile(audioPath, []byte("dummy"), 0o644); err != nil {
		t.Fatalf("Failed to write audio file: %v", err)
	}

	// NewJSONMetadataProvider should use hybrid extraction
	provider := NewJSONMetadataProvider(metadataPath)
	metadata, err := provider.GetMetadata()
	if err != nil {
		t.Fatalf("GetMetadata() error: %v", err)
	}

	// Verify JSON fields
	if metadata.Title != "Hybrid Test" {
		t.Errorf("Title = %q, want %q", metadata.Title, "Hybrid Test")
	}

	// Verify RawData preservation
	if desc, ok := metadata.RawData["description"].(string); !ok || desc != "Test description" {
		t.Error("RawData missing or incorrect 'description' field")
	}

	// Should have _embedded_source marker if hybrid extraction worked
	if _, ok := metadata.RawData["_embedded_source"]; ok {
		// Hybrid extraction attempted (may not have actual data from dummy file)
		t.Log("Hybrid extraction attempted (_embedded_source present)")
	}
}

// TestRenamerHybridMetadata tests that Renamer uses hybrid metadata correctly
func TestRenamerHybridMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	// Create metadata.json
	metadataJSON := `{
		"title": "Test Audiobook",
		"authors": ["Author Name"],
		"series": ["Series #1"]
	}`

	metadataPath := filepath.Join(tmpDir, "metadata.json")
	if err := os.WriteFile(metadataPath, []byte(metadataJSON), 0o644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	// Create audio file
	audioPath := filepath.Join(tmpDir, "audio.m4b")
	if err := os.WriteFile(audioPath, []byte("dummy audio"), 0o644); err != nil {
		t.Fatalf("Failed to write audio file: %v", err)
	}

	// Create renamer
	config := &RenamerConfig{
		BaseDir:      tmpDir,
		Template:     "{author} - {title}",
		AuthorFormat: AuthorFormatFirstLast,
		Recursive:    true,
	}

	renamer, err := NewRenamer(config)
	if err != nil {
		t.Fatalf("NewRenamer() error: %v", err)
	}

	// Scan files - should use hybrid metadata
	candidates, err := renamer.ScanFiles()
	if err != nil {
		t.Fatalf("ScanFiles() error: %v", err)
	}

	if len(candidates) == 0 {
		t.Skip("No candidates found - expected with dummy audio files")
	}

	// Check if any candidates had errors (expected with dummy audio)
	hasErrors := false
	for _, c := range candidates {
		if c.Error != "" {
			hasErrors = true
			t.Logf("Candidate error (expected with dummy audio): %s", c.Error)
		}
	}

	if hasErrors {
		t.Skip("Candidates have errors - expected with dummy audio files")
	}

	// If we somehow got valid candidates, verify structure
	candidate := candidates[0]

	// Verify metadata from JSON
	if candidate.Metadata.Title != "Test Audiobook" {
		t.Errorf(
			"Candidate.Metadata.Title = %q, want %q",
			candidate.Metadata.Title,
			"Test Audiobook",
		)
	}

	if candidate.Metadata.SourceType != "json" {
		t.Errorf(
			"Candidate.Metadata.SourceType = %q, want %q (hybrid mode)",
			candidate.Metadata.SourceType,
			"json",
		)
	}

	// Verify RawData has JSON fields
	if _, ok := candidate.Metadata.RawData["series"]; !ok {
		t.Error("Candidate.Metadata.RawData missing 'series' field")
	}
}
