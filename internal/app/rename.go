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

// RenameRunResponse contains applied rename candidates, summary, and undo log path.
type RenameRunResponse struct {
	Candidates []organizer.RenameCandidate `json:"candidates"`
	Summary    organizer.RenameSummary     `json:"summary"`
	LogPath    string                      `json:"log_path,omitempty"`
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

// RunRename scans files and applies rename candidates.
func (s *Service) RunRename(
	ctx context.Context,
	req RenameRequest,
) (*RenameRunResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	config := req.Config.ToRenamerConfig()
	config.DryRun = false
	config.PromptEnabled = false
	renamer, err := organizer.NewRenamer(&config)
	if err != nil {
		return nil, err
	}
	candidates, err := renamer.ScanFiles()
	if err != nil {
		return nil, err
	}
	if err := renamer.Execute(); err != nil {
		return nil, err
	}
	summary := renamer.GetSummary()
	logPath := ""
	if summary.FilesRenamed > 0 {
		logPath = renamer.GetLogPath()
	}
	return &RenameRunResponse{
		Candidates: candidates,
		Summary:    summary,
		LogPath:    logPath,
	}, nil
}
