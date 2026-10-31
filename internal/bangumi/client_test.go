package bangumi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"dinosaur/internal/metadata"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func reply(data string) *http.Response {
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}
}

const animeMovie = `{"id":10,"type":2,"name":"Original","name_cn":"中文片名","date":"2024-07-20","platform":"剧场版","summary":"简介","images":{"large":"https://lain.bgm.tv/pic/cover/l/ab/cd/10_cover.jpg"},"infobox":[{"key":"别名","value":[{"v":"English alias"}]},{"key":"语言","value":"日语"}]}`

func TestSearchAndDetailRespectMediaTypeAndSecrets(t *testing.T) {
	for _, kind := range []string{"movie", "tv"} {
		t.Run(kind, func(t *testing.T) {
			c := New(kind)
			details := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "api.bgm.tv" || !strings.Contains(r.Header.Get("User-Agent"), "dogmonkeyffc/dinosaur/") || r.Header.Get("Authorization") != "" || r.URL.Query().Has("api_key") {
					t.Fatal("invalid upstream identity or leaked credential")
				}
				if r.Method == "POST" {
					var body struct {
						Keyword string
						Filter  struct {
							Type []int
							NSFW bool
						}
					}
					if json.NewDecoder(r.Body).Decode(&body) != nil || body.Keyword != "中文片名" || len(body.Filter.Type) != 2 || body.Filter.NSFW {
						t.Fatal("invalid search body")
					}
					return reply(`{"data":[{"id":10,"type":2},{"id":11,"type":2},{"id":12,"type":1}]}`), nil
				}
				details++
				if strings.HasSuffix(r.URL.Path, "/11") {
					return reply(`{"id":11,"type":2,"name":"Series","name_cn":"中文剧名","date":"2024-01-01","platform":"TV"}`), nil
				}
				return reply(animeMovie), nil
			})
			settings := metadata.Settings{Credential: "tmdb-secret", CredentialType: "api_key", Language: "zh-CN"}
			items, err := c.Search(context.Background(), settings, "中文片名", 2024)
			if err != nil || len(items) != 1 || items[0].Provider != "bangumi" {
				t.Fatalf("search: %+v %v", items, err)
			}
			want := int64(10)
			if kind == "tv" {
				want = 11
			}
			if items[0].ID != want {
				t.Fatalf("wrong media type: %+v", items)
			}
			info, err := c.Movie(context.Background(), settings, want)
			if err != nil || info.BangumiID != want || info.TMDBID != 0 || details != 2 {
				t.Fatalf("detail/cache: %+v %v count=%d", info, err, details)
			}
			if kind == "movie" && (info.OriginalLanguage != "ja" || info.Language != "zh-CN" || !strings.Contains(info.Aliases, "English alias")) {
				t.Fatalf("metadata: %+v", info)
			}
			if _, err := c.Movie(context.Background(), settings, want); err != nil || details != 3 {
				t.Fatalf("refresh reused stale snapshot: %v count=%d", err, details)
			}
			if _, err := c.Movie(context.Background(), settings, 21-want); !errors.Is(err, metadata.ErrProviderNotFound) {
				t.Fatalf("wrong-type manual ID accepted: %v", err)
			}
		})
	}
}

func TestProviderFailuresAndCancellation(t *testing.T) {
	for status, want := range map[int]error{401: metadata.ErrProviderAuth, 404: metadata.ErrProviderNotFound, 429: metadata.ErrProviderRateLimited, 503: metadata.ErrProviderUnavailable, 302: metadata.ErrProviderUnavailable} {
		c := New("movie")
		c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { r := reply(`{}`); r.StatusCode = status; return r, nil })
		_, err := c.Search(context.Background(), metadata.Settings{}, "Name", 2024)
		if !errors.Is(err, want) {
			t.Fatalf("status %d: %v", status, err)
		}
	}
	c := New("movie")
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		t.Fatal("cancelled request reached network")
		return nil, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Search(ctx, metadata.Settings{}, "Name", 2024)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}

func TestNoInventedOriginalLanguageAndInvalidPayloads(t *testing.T) {
	info := subjectMovie(subject{ID: 1, Type: 2, Name: "日本語の名前", Chinese: "中文", Date: "2020-01-01"}, metadata.Settings{Language: "zh-TW"})
	if info.OriginalLanguage != "" || info.Language != "zh-CN" {
		t.Fatalf("inferred language from localized title: %+v", info)
	}
	for _, payload := range []string{`no JSON`, strings.Repeat("x", 2*1024*1024+1)} {
		c := New("movie")
		c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return reply(payload), nil })
		if _, err := c.Search(context.Background(), metadata.Settings{}, "Name", 0); !errors.Is(err, metadata.ErrProviderUnavailable) {
			t.Fatalf("invalid payload accepted: %v", err)
		}
	}
}

func TestUnknownPlatformsAndNonVideoSubjectsAreExcluded(t *testing.T) {
	for _, s := range []subject{{Type: 2, Platform: "MV"}, {Type: 6, Platform: "其他"}, {Type: 1, Platform: "TV"}, {Type: 2, Platform: "TV", NSFW: true}, {Type: 2}} {
		if New("tv").accepts(s) || New("movie").accepts(s) {
			t.Fatalf("non-movie/series accepted: %+v", s)
		}
	}
	info := subjectMovie(subject{Name: "Original"}, metadata.Settings{Language: "zh-CN"})
	if info.Message != "no_chinese_title" {
		t.Fatal("missing Chinese title not reported")
	}
}
