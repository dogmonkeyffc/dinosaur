package television

import (
	"dinosaur/internal/media"
	"testing"
)

func TestResumeEpisodesAcrossSeasonsAndSpecials(t *testing.T) {
	items := []media.Item{
		{ID: 1, SeasonNumber: 0, EpisodeNumber: 1},
		{ID: 2, SeasonNumber: 1, EpisodeNumber: 1},
		{ID: 3, SeasonNumber: 1, EpisodeNumber: 2},
		{ID: 4, SeasonNumber: 2, EpisodeNumber: 1},
	}
	if ResumeEpisode(nil) != nil {
		t.Fatal("empty series returned an episode")
	}
	if ResumeEpisode(items).ID != 2 {
		t.Fatal("special displaced main season")
	}
	items[1].LastPlayedAt = 5
	if ResumeEpisode(items).ID != 2 {
		t.Fatal("marking unwatched advanced past the first episode")
	}
	items[1].Watched, items[1].LastPlayedAt = true, 10
	if ResumeEpisode(items).ID != 3 {
		t.Fatal("did not advance to next episode")
	}
	items[2].Watched, items[2].LastPlayedAt = true, 20
	if ResumeEpisode(items).ID != 4 {
		t.Fatal("did not advance across seasons")
	}
	items[1].Watched, items[1].Position, items[1].LastPlayedAt = false, 15, 30
	if ResumeEpisode(items).ID != 2 {
		t.Fatal("partial episode was lost")
	}
	items[0].Position, items[0].LastPlayedAt = 20, 40
	if ResumeEpisode(items).ID != 1 {
		t.Fatal("most recent partial special was lost")
	}
}
