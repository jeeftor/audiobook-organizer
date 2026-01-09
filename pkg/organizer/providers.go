// Package organizer provides public APIs for metadata providers
package organizer

import (
	"github.com/jeeftor/audiobook-organizer/internal/organizer"
)

// Re-export metadata provider types
type (
	JSONMetadataProvider  = organizer.JSONMetadataProvider
	EPUBMetadataProvider  = organizer.EPUBMetadataProvider
	AudioMetadataProvider = organizer.AudioMetadataProvider
	MetadataFormatter     = organizer.MetadataFormatter
)

// Re-export metadata provider constructors
var (
	NewJSONMetadataProvider  = organizer.NewJSONMetadataProvider
	NewEPUBMetadataProvider  = organizer.NewEPUBMetadataProvider
	NewAudioMetadataProvider = organizer.NewAudioMetadataProvider
	NewMetadataFormatter     = organizer.NewMetadataFormatter
)
