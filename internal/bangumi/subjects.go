package bangumi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dinosaur/internal/metadata"
)

type subject struct {
	ID       int64  `json:"id"`
	Type     int    `json:"type"`
	Name     string `json:"name"`
	Chinese  string `json:"name_cn"`
	Summary  string `json:"summary"`
	Date     string `json:"date"`
	Platform string `json:"platform"`
	NSFW     bool   `json:"nsfw"`
	Images   struct {
		Medium string `json:"medium"`
		Large  string `json:"large"`
		Common string `json:"common"`
	} `json:"images"`
	Infobox []struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	} `json:"infobox"`
}

func (c *Client) accepts(s subject) bool {
	if s.NSFW || (s.Type != 2 && s.Type != 6) {
		return false
	}
	movie := s.Platform == "剧场版" || s.Platform == "电影" || strings.EqualFold(s.Platform, "movie")
	if c.kind == "movie" {
		return movie
	}
	switch strings.ToUpper(s.Platform) {
	case "TV", "WEB", "OVA", "OAD", "日剧", "韩剧", "欧美剧", "华语剧", "电视剧", "特摄", "综艺", "电视节目":
		return true
	}
	return false
}

func (c *Client) detail(ctx context.Context, id int64) (subject, error) {
	c.mu.Lock()
	cached, ok := c.cache[id]
	c.mu.Unlock()
	if ok && time.Now().Before(cached.expires) {
		return cached.subject, nil
	}
	var s subject
	if err := c.request(ctx, http.MethodGet, "/v0/subjects/"+strconv.FormatInt(id, 10), nil, &s); err != nil {
		return s, err
	}
	if s.ID != id || strings.TrimSpace(s.Name) == "" {
		return s, metadata.ErrProviderUnavailable
	}
	c.mu.Lock()
	if len(c.cache) >= 64 {
		clear(c.cache)
	}
	c.cache[id] = cachedSubject{subject: s, expires: time.Now().Add(5 * time.Minute)}
	c.mu.Unlock()
	return s, nil
}

func (c *Client) Search(ctx context.Context, _ metadata.Settings, title string, year int) ([]metadata.Candidate, error) {
	body := map[string]any{"keyword": title, "sort": "match", "filter": map[string]any{"type": []int{2, 6}, "nsfw": false}}
	var result struct {
		Data []subject `json:"data"`
	}
	if err := c.request(ctx, http.MethodPost, "/v0/search/subjects?limit=10", body, &result); err != nil {
		return nil, err
	}
	if result.Data == nil {
		return nil, metadata.ErrProviderUnavailable
	}
	candidates := []metadata.Candidate{}
	for _, s := range result.Data {
		if s.ID <= 0 || (s.Type != 2 && s.Type != 6) {
			continue
		}
		// Search's slim subjects omit platform. Confirm the movie/series boundary.
		full, err := c.detail(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		if !c.accepts(full) {
			continue
		}
		info := subjectMovie(full, metadata.Settings{Language: "zh-CN"})
		if year > 0 && info.Year != year && info.Year != year-1 && info.Year != year+1 {
			continue
		}
		candidates = append(candidates, metadata.Candidate{Provider: "bangumi", ID: s.ID, Title: info.Title, OriginalTitle: info.OriginalTitle, Year: info.Year, Overview: info.Overview})
	}
	return candidates, nil
}

func (c *Client) Movie(ctx context.Context, settings metadata.Settings, id int64) (metadata.Movie, error) {
	s, err := c.detail(ctx, id)
	if err != nil {
		return metadata.Movie{}, err
	}
	// Search hands its detail snapshot to the first read only. A later refresh
	// must fetch fresh metadata instead of silently reusing that snapshot.
	c.mu.Lock()
	delete(c.cache, id)
	c.mu.Unlock()
	if !c.accepts(s) {
		return metadata.Movie{}, metadata.ErrProviderNotFound
	}
	return subjectMovie(s, settings), nil
}

func subjectMovie(s subject, settings metadata.Settings) metadata.Movie {
	title, language := s.Name, ""
	if strings.TrimSpace(s.Chinese) != "" {
		title, language = s.Chinese, "zh-CN"
	}
	year := 0
	if len(s.Date) >= 4 {
		year, _ = strconv.Atoi(s.Date[:4])
	}
	info := metadata.Movie{BangumiID: s.ID, Title: title, OriginalTitle: s.Name, Overview: s.Summary, Year: year,
		ReleaseDate: s.Date, Language: language, Status: "matched", PosterPath: s.Images.Large}
	if settings.Language == "en-US" {
		info.Title, info.Language = s.Name, ""
	}
	if s.Images.Medium != "" {
		info.PosterPath = s.Images.Medium
	}
	if info.PosterPath == "" {
		info.PosterPath = s.Images.Common
	}
	if strings.HasPrefix(settings.Language, "zh-") && strings.TrimSpace(s.Chinese) == "" {
		info.Message = "no_chinese_title"
	}
	aliases := []string{s.Name, s.Chinese}
	for _, field := range s.Infobox {
		if field.Key == "别名" {
			var values []struct {
				Value string `json:"v"`
			}
			var value string
			if json.Unmarshal(field.Value, &values) == nil {
				for _, v := range values {
					aliases = append(aliases, v.Value)
				}
			}
			if json.Unmarshal(field.Value, &value) == nil {
				aliases = append(aliases, value)
			}
		}
		if field.Key == "语言" {
			var value string
			if json.Unmarshal(field.Value, &value) == nil {
				info.OriginalLanguage = originalLanguage(value)
			}
		}
	}
	info.Aliases = strings.Join(aliases, "\n")
	return info
}

func originalLanguage(value string) string {
	switch strings.TrimSpace(value) {
	case "日语", "日語", "日本語", "Japanese":
		return "ja"
	case "汉语普通话", "普通话", "汉语", "中文", "国语", "汉语普通话 / 中文", "Chinese":
		return "zh"
	case "粤语", "粵語", "Cantonese":
		return "yue"
	case "英语", "英語", "English":
		return "en"
	case "韩语", "韓語", "Korean":
		return "ko"
	}
	return ""
}
