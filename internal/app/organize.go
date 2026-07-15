package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// OrganizeRequest requests an organization preview or execution.
type OrganizeRequest struct {
	Config OrganizerConfigDTO `json:"config"`
}

const metadataSourceABS = "abs"

// OrganizePreviewResponse contains a dry-run organization summary.
type OrganizePreviewResponse struct {
	Summary organizer.Summary `json:"summary"`
	LogPath string            `json:"log_path,omitempty"`
}

// OrganizeRunResponse contains an executed organization summary.
type OrganizeRunResponse struct {
	Summary organizer.Summary `json:"summary"`
	LogPath string            `json:"log_path,omitempty"`
}

// PreviewOrganize runs the organizer in dry-run mode.
func (s *Service) PreviewOrganize(
	ctx context.Context,
	req OrganizeRequest,
) (*OrganizePreviewResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	org, err := s.executeOrganize(req, true)
	if err != nil {
		return nil, err
	}
	return &OrganizePreviewResponse{Summary: org.GetSummary()}, nil
}

// RunOrganize runs the organizer with filesystem mutations enabled.
func (s *Service) RunOrganize(
	ctx context.Context,
	req OrganizeRequest,
) (*OrganizeRunResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	org, err := s.executeOrganize(req, false)
	if err != nil {
		return nil, err
	}
	return &OrganizeRunResponse{Summary: org.GetSummary(), LogPath: org.GetLogPath()}, nil
}

func (s *Service) executeOrganize(req OrganizeRequest, dryRun bool) (*organizer.Organizer, error) {
	config := req.Config.ToOrganizerConfig()
	config.DryRun = dryRun
	org, err := organizer.NewOrganizer(&config)
	if err != nil {
		return nil, err
	}
	if req.Config.MetadataSource != metadataSourceABS {
		if err := org.Execute(); err != nil {
			return nil, err
		}
		return org, nil
	}

	if err := org.ResolvePaths(); err != nil {
		return nil, err
	}
	provider, err := s.newABSProviderForInput(req.Config.ABS, org.BaseDir())
	if err != nil {
		return nil, err
	}
	if err := provider.LoadAllItems(); err != nil {
		return nil, fmt.Errorf("loading ABS items: %w", err)
	}
	items, err := provider.GetAllItems()
	if err != nil {
		return nil, fmt.Errorf("getting ABS metadata: %w", err)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].SourcePath < items[j].SourcePath
	})

	processed := 0
	startTime := time.Now()
	for _, item := range items {
		sourcePath := item.SourcePath
		if sourcePath == "" {
			continue
		}
		resolvedSourcePath, err := filepath.EvalSymlinks(sourcePath)
		if err != nil {
			if config.SkipErrors || os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("resolving ABS item path %s: %w", sourcePath, err)
		}
		sourcePath = resolvedSourcePath
		if !isPathWithin(org.BaseDir(), sourcePath) || !org.IsAllowedSourcePath(sourcePath) {
			continue
		}
		if err := org.OrganizePathWithMetadata(sourcePath, item); err != nil {
			if config.SkipErrors {
				continue
			}
			return nil, fmt.Errorf("organizing ABS item %s: %w", sourcePath, err)
		}
		processed++
	}
	if processed == 0 {
		return nil, fmt.Errorf("no mapped ABS items found under %s", org.BaseDir())
	}
	if err := org.Finish(startTime); err != nil {
		return nil, err
	}
	return org, nil
}

func isPathWithin(basePath string, candidatePath string) bool {
	rel, err := filepath.Rel(basePath, candidatePath)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
