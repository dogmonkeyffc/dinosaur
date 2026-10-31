package nfo

import (
	"strings"
	"testing"

	"dinosaur/internal/agent"
)

func TestRoundTripMovieAndSeries(t *testing.T) {
	input := agent.Input{Title: "中文 <片名>", OriginalTitle: "Original & Title", Overview: "剧情\n第二行", Year: 1999, ReleaseDate: "1999-06-05", OriginalLanguage: "ja", Language: "zh-CN", SourceURL: "https://example.com/subject/1", Aliases: []string{"Alias"}, IDs: map[string]string{"tmdb": "4291", "imdb": "tt0199683"}}
	info, err := input.Movie("agent")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"movie", "tv"} {
		data, err := Encode(info, kind)
		if err != nil {
			t.Fatal(err)
		}
		out, err := Decode(data, kind)
		if err != nil || out.Title != input.Title || out.Overview != input.Overview || out.IDs["tmdb"] != "4291" || out.IDs["imdb"] != "tt0199683" || out.OriginalLanguage != "ja" || out.Aliases[0] != "Alias" {
			t.Fatalf("roundtrip: %+v %v", out, err)
		}
	}
	parsed, err := Decode([]byte(`<movie><title>Title</title><unknown>ignored</unknown><uniqueid type="themoviedb">123</uniqueid></movie>`), "movie")
	if err != nil || parsed.IDs["tmdb"] != "123" {
		t.Fatal("Kodi ID alias not accepted")
	}
}

func TestRejectUnsafeAndAmbiguousDocuments(t *testing.T) {
	for _, value := range []string{
		`<!DOCTYPE movie [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><movie><title>&xxe;</title></movie>`,
		`<movie><title>One</title></movie><movie><title>Two</title></movie>`,
		`<movie><title>Title</title></movie>garbage`,
		`<movie xmlns="https://example.com"><title>Title</title></movie>`,
		`<tvshow><title>Series</title></tvshow>`,
		`<movie><title>Title</title><uniqueid type="tmdb">1</uniqueid><uniqueid type="tmdb">2</uniqueid></movie>`,
		`<movie><title>Title</title>` + strings.Repeat(`<nested>`, 40) + strings.Repeat(`</nested>`, 40) + `</movie>`,
		strings.Repeat("x", MaxSize+1),
	} {
		if _, err := Decode([]byte(value), "movie"); err == nil {
			t.Fatal("unsafe/invalid NFO accepted")
		}
	}
}
