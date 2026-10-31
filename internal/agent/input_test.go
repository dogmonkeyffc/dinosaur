package agent

import (
	"strings"
	"testing"
)

func TestMetadataDoesNotRequireProviderID(t *testing.T) {
	input := Input{Title: " 中文片名 ", OriginalTitle: "Original", Year: 1999, Overview: "第一行\n第二行", OriginalLanguage: "ja", Language: "zh-CN", Aliases: []string{"Alias"}, IDs: map[string]string{"imdb": "tt123"}, SourceURL: "https://example.com/source"}
	for _, source := range []string{"agent", "nfo"} {
		info, err := input.Movie(source)
		if err != nil || !info.Managed() || !info.Matched() || info.TMDBID != 0 || info.Title != "中文片名" {
			t.Fatalf("%+v %v", info, err)
		}
		out := FromMovie(info)
		if out.IDs["imdb"] != "tt123" || out.Aliases[0] != "Alias" || out.OriginalLanguage != "ja" {
			t.Fatal("lost imported metadata")
		}
	}
}

func TestInputValidation(t *testing.T) {
	mutations := []func(*Input){
		func(v *Input) { v.Title = " " }, func(v *Input) { v.Title = "bad\x00title" }, func(v *Input) { v.Overview = strings.Repeat("x", 32001) },
		func(v *Input) { v.ReleaseDate = "2024-99-99" }, func(v *Input) { v.SourceURL = "javascript:alert(1)" }, func(v *Input) { v.SourceURL = "https://user:secret@example.com" },
		func(v *Input) { v.IDs = map[string]string{"tmdb": "-1"} }, func(v *Input) { v.IDs = map[string]string{"bad key": "1"} }, func(v *Input) { v.OriginalLanguage = "中文" },
	}
	for i, mutate := range mutations {
		input := Input{Title: "Title"}
		mutate(&input)
		if _, err := input.Movie("agent"); err == nil {
			t.Fatalf("invalid input %d accepted", i)
		}
	}
	input := Input{Title: "Title", IDs: map[string]string{"tmdb": "123", "bangumi": "456"}}
	info, err := input.Movie("agent")
	if err != nil || info.TMDBID != 123 || info.BangumiID != 456 {
		t.Fatal("provider IDs conflated")
	}
	if _, err := input.Movie("tmdb"); err == nil {
		t.Fatal("caller can spoof source")
	}
}

func TestKeyValidationAndScope(t *testing.T) {
	if _, err := ValidateKey(CreateKey{Name: "Agent", Access: "read_write", ExpiresInDays: 90}); err != nil {
		t.Fatal(err)
	}
	for _, input := range []CreateKey{{Name: ""}, {Name: "Agent", LibraryID: -1}, {Name: "Agent", Access: "admin"}, {Name: "Agent", ExpiresInDays: 366}} {
		if _, err := ValidateKey(input); err == nil {
			t.Fatalf("bad key accepted: %+v", input)
		}
	}
	key := Key{LibraryID: 2, Access: "read"}
	if !key.Allows(2, false) || key.Allows(1, false) || key.Allows(2, true) {
		t.Fatal("read scope broken")
	}
	key.Access = "read_write"
	if !key.Allows(2, true) || key.Allows(1, true) {
		t.Fatal("library scope broken")
	}
}
