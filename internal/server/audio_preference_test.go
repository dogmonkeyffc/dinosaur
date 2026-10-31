package server

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/metadata"
	"dinosaur/internal/playback"
)

func TestOriginalLanguageAudioAndManualSelection(t *testing.T) {
	s, db, admin, _ := playbackFixture(t)
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	item.Tracks = []media.Track{{Index: 1, Type: "audio", Codec: "aac", Language: "chi"}, {Index: 2, Type: "audio", Codec: "aac", Language: "jpn"}}
	if err := db.SaveMedia(context.Background(), item, 1); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMovieMetadata(context.Background(), 1, metadata.Movie{TMDBID: 1, Title: "中文译名", Language: "zh-CN", OriginalLanguage: "ja", Status: "matched"}); err != nil {
		t.Fatal(err)
	}
	response := request(t, s, "GET", "/api/playback/1?subtitle=auto", nil, &admin, nil)
	status(t, response, 200)
	var info struct {
		Plan playback.Plan `json:"plan"`
		URL  string        `json:"stream_url"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Plan.Audio != 2 || info.Plan.Mode != "remux" || !strings.Contains(info.URL, "audio=2") {
		t.Fatalf("original language ignored: %s", response.Body.String())
	}
	manual := request(t, s, "GET", "/api/playback/1?audio=1", nil, &admin, nil)
	status(t, manual, 200)
	if err := json.Unmarshal(manual.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Plan.Audio != 1 {
		t.Fatal("manual selection overridden")
	}
}
