package domain

import "time"

// ProblemManifest is the normalized metadata record for a problem in the registry.
// Full problem statements are not stored; this is metadata only.
type ProblemManifest struct {
	Provider      Provider   `json:"provider"`
	ExternalID    string     `json:"external_id"`
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Difficulty    Difficulty `json:"difficulty"`
	Tags          []string   `json:"tags"`
	PatternTags   []string   `json:"pattern_tags"`
	SourceURL     string     `json:"source_url"`
	EstimatedTime int        `json:"estimated_time"`
	Version       int        `json:"version"`
}

// ManifestRef is a pointer to a provider or track manifest in the index.
type ManifestRef struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
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
