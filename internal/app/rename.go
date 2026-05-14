package app

import (
	"context"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// RenameRequest requests a rename preview.
type RenameRequest struct {
	Config RenameConfigDTO `json:"config"`
}

// RenamePreviewResponse contains rename candidates and a summary.
type RenamePreviewResponse struct {
	Candidates []organizer.RenameCandidate `json:"candidates"`
	Summary    organizer.RenameSummary     `json:"summary"`
}

// PreviewRename scans files and returns rename candidates without mutating files.
func (s *Service) PreviewRename(
	ctx context.Context,
	req RenameRequest,
) (*RenamePreviewResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	config := req.Config.ToRenamerConfig()
	config.DryRun = true
	renamer, err := organizer.NewRenamer(&config)
	if err != nil {
		return nil, err
	}
	candidates, err := renamer.ScanFiles()
	if err != nil {
		return nil, err
	}
	return &RenamePreviewResponse{
		Candidates: candidates,
		Summary:    renamer.GetSummary(),
	}, nil
}
