package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

func TestAutoSubtitleChoosesChineseAndKeepsManualOff(t *testing.T) {
	s, db, admin, root := playbackFixture(t)
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	item.Tracks = append(item.Tracks, media.Track{Index: 2, Type: "subtitle", Codec: "subrip", Language: "eng"}, media.Track{Index: 3, Type: "subtitle", Codec: "hdmv_pgs_subtitle", Language: "chi", Title: "官译简中"})
	if err := db.SaveMedia(context.Background(), item, 1); err != nil {
		t.Fatal(err)
	}
	var info struct {
		Selected *int          `json:"selected_subtitle"`
		Plan     playback.Plan `json:"plan"`
		URL      string        `json:"stream_url"`
	}
	auto := request(t, s, "GET", "/api/playback/1?mode=original&subtitle=auto", nil, &admin, nil)
	status(t, auto, 200)
	if err := json.Unmarshal(auto.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Selected == nil || *info.Selected != -4 || info.Plan.Mode != "transcode" || !strings.Contains(info.URL, "subtitle=-4") {
		t.Fatalf("auto bitmap=%s", auto.Body.String())
	}
	off := request(t, s, "GET", "/api/playback/1?mode=original", nil, &admin, nil)
	status(t, off, 200)
	var disabled struct {
		Selected *int          `json:"selected_subtitle"`
		Plan     playback.Plan `json:"plan"`
	}
	if err := json.Unmarshal(off.Body.Bytes(), &disabled); err != nil {
		t.Fatal(err)
	}
	if disabled.Selected != nil || disabled.Plan.Subtitle != nil || disabled.Plan.Mode != "direct" {
		t.Fatal("manual off was overridden")
	}
	if err := os.WriteFile(filepath.Join(root, "Movie.zh-CN.srt"), []byte("1\n00:00:01,000 --> 00:00:02,000\n字幕\n"), 0600); err != nil {
		t.Fatal(err)
	}
	text := request(t, s, "GET", "/api/playback/1?mode=original&subtitle=auto", nil, &admin, nil)
	status(t, text, 200)
	if err := json.Unmarshal(text.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Selected == nil || *info.Selected != 0 || info.Plan.Mode != "direct" {
		t.Fatalf("auto external=%s", text.Body.String())
	}
	status(t, request(t, s, "GET", "/api/playback/1?subtitle=auto", nil, nil, nil), 401)
}

func TestChineseMovieNameDoesNotImplyChineseSubtitles(t *testing.T) {
	s, db, admin, root := playbackFixture(t)
	item, err := db.MediaByID(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	item.RelativePath = "中文片名.mp4"
	if err := os.WriteFile(filepath.Join(root, item.RelativePath), []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "中文片名.srt"), []byte("subtitle"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveMedia(context.Background(), item, 1); err != nil {
		t.Fatal(err)
	}
	page := request(t, s, "GET", "/api/media", nil, &admin, nil)
	var listing struct {
		Items []media.Item `json:"items"`
	}
	if err := json.Unmarshal(page.Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	var id int64
	for _, record := range listing.Items {
		if record.Filename == "中文片名.mp4" {
			id = record.ID
		}
	}
	if id == 0 {
		t.Fatal("fixture missing")
	}
	response := request(t, s, "GET", fmt.Sprintf("/api/playback/%d?subtitle=auto", id), nil, &admin, nil)
	status(t, response, 200)
	var info struct {
		Selected *int `json:"selected_subtitle"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Selected != nil {
		t.Fatal("Chinese movie name incorrectly classified an unmarked subtitle")
	}
}
