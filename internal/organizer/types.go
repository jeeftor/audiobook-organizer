package organizer

import "time"

type Metadata struct {
	Authors []string `json:"authors"`
	Title   string   `json:"title"`
	Series  []string `json:"series"`
}

type LogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	SourcePath string    `json:"source_path"`
	TargetPath string    `json:"target_path"`
	Files      []string  `json:"files"`
}

type Summary struct {
	MetadataFound   []string
	MetadataMissing []string
	Moves           []MoveSummary
}

type MoveSummary struct {
	From string
	To   string
}
