//go:build abs_e2e

package e2e

import "testing"

// TestRESTHarness_ABSRenameMetadataPreview proves the browser-facing rename
// endpoint can resolve real Docker-backed Audiobookshelf metadata for each
// library file without mutating the mounted library during preview.
func TestRESTHarness_ABSRenameMetadataPreview(t *testing.T) {
	harness := newRESTHarness(t)
	defer harness.server.Close()

	var ctx restScenarioContext
	step(t, "01 reset and scan ABS audiobook library", func(t *testing.T) {
		resetAndInitialScan(t)
		ctx = newRESTScenarioContext(t, harness, plainInstance, audiobooksLibrary)
	})

	step(t, "02 preview rename using mapped ABS API metadata", func(t *testing.T) {
		var response struct {
			Candidates []struct {
				Error    string `json:"Error"`
				Metadata struct {
					Title      string `json:"title"`
					SourceType string `json:"source_type"`
				} `json:"Metadata"`
			} `json:"candidates"`
		}
		harness.postJSON(t, "/api/rename/preview", map[string]any{
			"config": map[string]any{
				"base_dir":        localLibraryPath(plainInstance, audiobooksLibrary),
				"template":        "{author} - {title}",
				"author_format":   "first-last",
				"recursive":       true,
				"preserve_path":   true,
				"metadata_source": "abs",
				"abs":             ctx.config,
			},
		}, &response)
		if len(response.Candidates) == 0 {
			t.Fatal("expected real ABS rename preview candidates")
		}
		for _, candidate := range response.Candidates {
			if candidate.Error != "" {
				t.Fatalf("ABS rename preview candidate error = %q", candidate.Error)
			}
			if candidate.Metadata.SourceType != "abs" || candidate.Metadata.Title == "" {
				t.Fatalf("candidate metadata = %+v, want ABS title metadata", candidate.Metadata)
			}
		}
	})
}
