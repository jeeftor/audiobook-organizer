package app

import (
	"context"

	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// OrganizeRequest requests an organization preview or execution.
type OrganizeRequest struct {
	Config OrganizerConfigDTO `json:"config"`
}

// OrganizePreviewResponse contains a dry-run organization summary.
type OrganizePreviewResponse struct {
	Summary organizer.Summary `json:"summary"`
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

	config := req.Config.ToOrganizerConfig()
	config.DryRun = true
	org, err := organizer.NewOrganizer(&config)
	if err != nil {
		return nil, err
	}
	if err := org.Execute(); err != nil {
		return nil, err
	}
	return &OrganizePreviewResponse{Summary: org.GetSummary()}, nil
}
