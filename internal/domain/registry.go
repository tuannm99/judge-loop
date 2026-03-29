package domain

import "time"

// ProblemManifest is the normalized metadata record for a problem in the registry.
// Full problem statements are not stored; this is metadata only.
type ProblemManifest struct {
	Provider      Provider
	ExternalID    string
	Slug          string
	Title         string
	Difficulty    Difficulty
	Tags          []string
	PatternTags   []string
	SourceURL     string
	EstimatedTime int // minutes
	Version       int // incremented when metadata changes
}

// ManifestRef is a pointer to a provider or track manifest in the index.
type ManifestRef struct {
	Name     string
	Path     string // relative path within registry
	Checksum string // sha256:... for change detection
}

// RegistryVersion tracks the currently synced registry state.
type RegistryVersion struct {
	Version   string
	UpdatedAt time.Time
	Manifests []ManifestRef
}

// Track is a curated ordered list of problems.
type Track struct {
	Name        string
	Title       string
	Description string
	Problems    []TrackProblem
}

// TrackProblem is an ordered entry in a Track.
type TrackProblem struct {
	Provider Provider
	Slug     string
	Order    int
}

// RoadmapStage is one stage of a learning roadmap.
type RoadmapStage struct {
	Name     string
	Title    string
	Track    string   // track name
	Problems []string // slugs
}

// RoadmapPlan is a multi-stage learning roadmap.
type RoadmapPlan struct {
	Name   string
	Title  string
	Stages []RoadmapStage
}
