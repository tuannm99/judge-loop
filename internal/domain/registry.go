package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// ProblemManifest is the normalized record for a problem in the registry.
// Provider statements are not fetched automatically, but an optional
// author-written Markdown description can be stored locally.
type ProblemManifest struct {
	Provider            Provider          `json:"provider"`
	ExternalID          string            `json:"external_id"`
	Slug                string            `json:"slug"`
	Title               string            `json:"title"`
	Difficulty          Difficulty        `json:"difficulty"`
	Tags                []string          `json:"tags"`
	SourceURL           string            `json:"source_url"`
	EstimatedTime       int               `json:"estimated_time"`
	DescriptionMarkdown string            `json:"description_markdown"`
	StarterCode         map[string]string `json:"starter_code"`
	Version             int               `json:"version"`
}

func (m *ProblemManifest) UnmarshalJSON(data []byte) error {
	type rawProblemManifest struct {
		Provider            Provider          `json:"provider"`
		ExternalID          string            `json:"external_id"`
		Slug                string            `json:"slug"`
		Title               string            `json:"title"`
		Difficulty          Difficulty        `json:"difficulty"`
		Tags                []string          `json:"tags"`
		LegacyPatternTags   []string          `json:"pattern_tags"`
		SourceURL           string            `json:"source_url"`
		EstimatedTime       int               `json:"estimated_time"`
		DescriptionMarkdown string            `json:"description_markdown"`
		StarterCode         map[string]string `json:"starter_code"`
		Version             int               `json:"version"`
	}

	var raw rawProblemManifest
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*m = ProblemManifest{
		Provider:            raw.Provider,
		ExternalID:          raw.ExternalID,
		Slug:                raw.Slug,
		Title:               raw.Title,
		Difficulty:          raw.Difficulty,
		Tags:                normalizeManifestTags(append(raw.Tags, raw.LegacyPatternTags...)),
		SourceURL:           raw.SourceURL,
		EstimatedTime:       raw.EstimatedTime,
		DescriptionMarkdown: raw.DescriptionMarkdown,
		StarterCode:         raw.StarterCode,
		Version:             raw.Version,
	}
	return nil
}

func normalizeManifestTags(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
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
