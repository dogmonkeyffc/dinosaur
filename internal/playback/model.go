package playback

import (
	"errors"
	"math"
)

var ErrPosition = errors.New("invalid playback position")
var ErrAudio = errors.New("invalid audio track")

type Progress struct {
	MediaID   int64   `json:"media_id"`
	Position  float64 `json:"position"`
	Duration  float64 `json:"duration"`
	Watched   bool    `json:"watched"`
	UpdatedAt int64   `json:"updated_at"`
	Completed bool    `json:"-"`
}

func UpdateProgress(previous Progress, duration float64, position *float64, watched *bool) (Progress, error) {
	if position != nil {
		if math.IsNaN(*position) || math.IsInf(*position, 0) || *position < 0 || *position > 86400 {
			return previous, ErrPosition
		}
		previous.Position = *position
	}
	previous.Duration = duration
	if duration > 0 {
		previous.Position = min(previous.Position, duration)
	}
	// Players send the final position with an explicit end event. The Android
	// player reports milliseconds, so allow for rounding when comparing lengths.
	previous.Completed = watched != nil && *watched && position != nil &&
		(duration <= 0 || *position >= duration-.001)
	if watched != nil {
		previous.Watched = *watched
		if *watched && duration > 0 {
			previous.Position = duration
		}
		if !*watched {
			previous.Position = 0
		}
	}
	return previous, nil
}
func (p Progress) ResumePosition() float64 {
	if p.Watched || p.Position < 1 {
		return 0
	}
	return p.Position
}
