package tmdb

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"dinosaur/internal/metadata"
)

func TestTVNamespaceSearchAndChineseNames(t *testing.T) {
	client := New()
	provider := Television{Client: client}
	settings := metadata.Settings{CredentialType: "bearer", Credential: "test-placeholder-key", Language: "zh-CN"}
	client.HTTP.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer test-placeholder-key" || r.URL.Query().Get("language") != "zh-CN" {
			t.Fatal("missing credential or language")
		}
		var body string
		switch r.URL.Path {
		case "/3/search/tv":
			if r.URL.Query().Get("first_air_date_year") != "2020" || r.URL.Query().Get("query") != "Example" {
				t.Fatal("incorrect TV year filter")
			}
			body = `{"results":[{"id":1,"name":"剧名","original_name":"Example","first_air_date":"2020-01-02"}]}`
		case "/3/tv/1":
			if r.URL.Query().Get("append_to_response") != "translations,alternative_titles" {
				t.Fatal("missing translations")
			}
			body = `{"id":1,"name":"Example","original_name":"Example","first_air_date":"2020-01-02","translations":{"translations":[{"iso_639_1":"zh","iso_3166_1":"CN","data":{"name":"中文剧名","overview":"中文简介"}}]},"alternative_titles":{"results":[{"title":"Another alias"}]}}`
		default:
			t.Fatalf("unexpected endpoint %s", r.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	results, err := provider.Search(context.Background(), settings, "Example", 2020)
	if err != nil || len(results) != 1 || results[0].Year != 2020 || results[0].OriginalTitle != "Example" {
		t.Fatalf("search=%+v %v", results, err)
	}
	info, err := provider.Movie(context.Background(), settings, 1)
	if err != nil || info.Title != "中文剧名" || info.Overview != "中文简介" || !strings.Contains(info.Aliases, "Another alias") {
		t.Fatalf("details=%+v %v", info, err)
	}
}
