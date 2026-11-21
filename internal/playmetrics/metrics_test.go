package playmetrics

import (
	"math"
	"testing"
)

func TestValidID(t *testing.T) {
	for id, want := range map[string]bool{
		"00000000-0000-4000-8000-000000000000":   true,
		"ABCDEFAB-0123-4ABC-BABC-0123456789AB":   true,
		"ABCDEFAB-0123-5ABC-BABC-0123456789AB":   false,
		"ABCDEFAB-0123-4ABC-7ABC-0123456789AB":   false,
		"{ABCDEFAB-0123-4ABC-BABC-0123456789AB}": false,
		"ABCDEFAB01234ABCBABC0123456789AB":       false,
		"":                                       false,
	} {
		if ValidID(id) != want {
			t.Errorf("ValidID(%q) != %t", id, want)
		}
	}
}

func TestSnapshotValidation(t *testing.T) {
	good := Snapshot{Metrics: Metrics{ID: "00000000-0000-4000-8000-000000000000", Mode: "remux", Quality: "720"}}
	if err := good.Validate(); err != nil {
		t.Fatal(err)
	}
	negative, longStartup, frames := int64(-1), int64(maxStartup+1), int64(maxFrames+1)
	badID := "bad"
	for name, mutate := range map[string]func(*Snapshot){
		"id":             func(s *Snapshot) { s.ID = badID },
		"stream id":      func(s *Snapshot) { s.StreamID = &badID },
		"mode":           func(s *Snapshot) { s.Mode = "web" },
		"quality":        func(s *Snapshot) { s.Quality = "4k" },
		"width":          func(s *Snapshot) { s.Width = maxDimension + 1 },
		"height":         func(s *Snapshot) { s.Height = -1 },
		"startup":        func(s *Snapshot) { s.StartupMS = &longStartup },
		"stalls":         func(s *Snapshot) { s.Stalls = maxStalls + 1 },
		"stall duration": func(s *Snapshot) { s.StallMS = -1 },
		"played":         func(s *Snapshot) { s.PlayedMS = maxDuration + 1 },
		"buffer":         func(s *Snapshot) { s.BufferMS = -1 },
		"dropped frames": func(s *Snapshot) { s.DroppedFrames = &negative },
		"total frames":   func(s *Snapshot) { s.TotalFrames = &frames },
	} {
		t.Run(name, func(t *testing.T) {
			bad := good
			mutate(&bad)
			if err := bad.Validate(); err == nil {
				t.Fatal("invalid snapshot accepted")
			}
		})
	}
}

func TestValidSpeed(t *testing.T) {
	if !ValidSpeed(0.1) || !ValidSpeed(maxSpeed) {
		t.Fatal("positive bounded speeds rejected")
	}
	for _, speed := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1), maxSpeed + 1} {
		if ValidSpeed(speed) {
			t.Errorf("invalid speed %v accepted", speed)
		}
	}
}
