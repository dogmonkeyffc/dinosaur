package tmdb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"dinosaur/internal/metadata"
)

func TestMovieSearchAndChineseFallback(t *testing.T) {
	client := New()
	settings := metadata.Settings{CredentialType: "bearer", Credential: "private-secret", Language: "zh-CN"}
	client.HTTP.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer private-secret" || r.URL.Query().Get("language") != "zh-CN" {
			t.Fatal("missing authentication or language")
		}
		body := `{"id":1,"title":"Original","original_title":"Original","original_language":"fr","release_date":"2004-03-17","overview":"English","poster_path":"/abc.jpg","translations":{"translations":[{"iso_639_1":"zh","iso_3166_1":"TW","data":{"title":"中文譯名","overview":"中文簡介"}}]},"alternative_titles":{"titles":[{"title":"Other alias"}]}}`
		if r.URL.Path == "/3/search/movie" {
			if r.URL.Query().Get("query") != "The Chorus" || r.URL.Query().Get("primary_release_year") != "2004" {
				t.Fatal("incorrect search query")
			}
			body = `{"results":[{"id":1,"title":"放牛班的春天","original_title":"Les Choristes","release_date":"2004-03-17"}]}`
		} else if r.URL.Query().Get("append_to_response") != "translations,alternative_titles" {
			t.Fatal("missing translations")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	candidates, err := client.Search(context.Background(), settings, "The Chorus", 2004)
	if err != nil || len(candidates) != 1 || candidates[0].Year != 2004 {
		t.Fatalf("invalid search: %+v %v", candidates, err)
	}
	movie, err := client.Movie(context.Background(), settings, 1)
	if err != nil || movie.Title != "中文譯名" || movie.Language != "zh-TW" || movie.OriginalLanguage != "fr" || movie.Overview != "中文簡介" || !strings.Contains(movie.Aliases, "Other alias") {
		t.Fatalf("invalid localized details: %+v %v", movie, err)
	}
}

func TestChineseTranslationPreferenceAndUntranslated(t *testing.T) {
	movie := metadata.Movie{Title: "Original"}
	translations := []translation{{Language: "zh", Country: "TW"}, {Language: "zh", Country: "CN"}}
	translations[0].Data.Title, translations[1].Data.Title = "繁體譯名", "简体片名"
	localizeChinese(&movie, "zh-CN", translations)
	if movie.Title != "简体片名" {
		t.Fatal("did not prefer selected language")
	}
	client := New()
	client.HTTP.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":1,"title":"Original","original_title":"Original"}`))}, nil
	})
	result, err := client.Movie(context.Background(), metadata.Settings{Credential: "secret", Language: "zh-CN"}, 1)
	if err != nil || result.Title != "Original" || result.Message != "no_chinese_title" {
		t.Fatal("untranslated movie was lost or invented")
	}
	_, err = client.Movie(context.Background(), metadata.Settings{Credential: "secret"}, 2)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal("accepted incorrect provider movie id")
	}
}
