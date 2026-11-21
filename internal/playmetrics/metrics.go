// Package playmetrics validates anonymous playback performance measurements.
package playmetrics

import (
	"fmt"
	"math"
	"regexp"
)

const (
	maxDimension = 16384
	maxDuration  = 30 * 24 * 60 * 60 * 1000
	maxStartup   = 60 * 60 * 1000
	maxStalls    = 1_000_000
	maxFrames    = 1_000_000_000
	maxSpeed     = 1000
)

var uuidV4 = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// ValidID reports whether id uses the canonical hyphenated form of UUID v4.
func ValidID(id string) bool { return len(id) == 36 && uuidV4.MatchString(id) }

// ValidSpeed reports whether speed is a positive finite multiple of real time.
func ValidSpeed(speed float64) bool {
	return speed > 0 && speed <= maxSpeed && !math.IsNaN(speed)
}

// Snapshot describes a single playback attempt. Stalls, StallMS, PlayedMS and
// the two frame counters are cumulative. BufferMS is a point-in-time gauge.
// An unknown video dimension is zero; an unavailable optional metric is nil.
// StreamID is only used for server-side correlation and is not part of Item.
type Snapshot struct {
	Metrics
	StreamID *string
}

type Metrics struct {
	ID            string
	Mode          string
	Quality       string
	BurnSubtitle  *bool
	Width         int64
	Height        int64
	StartupMS     *int64
	Stalls        int64
	StallMS       int64
	PlayedMS      int64
	BufferMS      int64
	DroppedFrames *int64
	TotalFrames   *int64
}

// Item is the anonymous diagnostics view. UpdatedAt is a Unix second.
type Item struct {
	Metrics
	UpdatedAt      int64
	TranscodeSpeed *float64
}

func (s Snapshot) Validate() error {
	if !ValidID(s.ID) {
		return fmt.Errorf("id must be a UUID v4")
	}
	if s.StreamID != nil && !ValidID(*s.StreamID) {
		return fmt.Errorf("stream_id must be a UUID v4")
	}
	switch s.Mode {
	case "direct", "remux", "transcode":
	default:
		return fmt.Errorf("invalid playback mode")
	}
	switch s.Quality {
	case "auto", "original", "1080", "720", "480":
	default:
		return fmt.Errorf("invalid playback quality")
	}
	for _, field := range []struct {
		name  string
		value *int64
		max   int64
	}{
		{"width", &s.Width, maxDimension}, {"height", &s.Height, maxDimension},
		{"startup_ms", s.StartupMS, maxStartup}, {"stalls", &s.Stalls, maxStalls},
		{"stall_ms", &s.StallMS, maxDuration}, {"played_ms", &s.PlayedMS, maxDuration},
		{"buffer_ms", &s.BufferMS, maxDuration}, {"dropped_frames", s.DroppedFrames, maxFrames},
		{"total_frames", s.TotalFrames, maxFrames},
	} {
		if field.value != nil && (*field.value < 0 || *field.value > field.max) {
			return fmt.Errorf("%s must be between 0 and %d", field.name, field.max)
		}
	}
	return nil
}
