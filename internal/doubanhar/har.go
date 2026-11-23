package doubanhar

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"dinosaur/internal/history"
)

type Record struct {
	Entry     history.Entry
	PosterURL string
}

type Report struct {
	Records []Record
	Posters map[string][]byte
	Total   int
	Starts  []int
	Profile string
}

type content struct {
	Text     string `json:"text"`
	Encoding string `json:"encoding"`
}
type harEntry struct {
	Request struct {
		URL string `json:"url"`
	} `json:"request"`
	Response struct {
		Status  int     `json:"status"`
		Content content `json:"content"`
	} `json:"response"`
}

var collection = regexp.MustCompile(`^/people/([0-9]+)/collect/?$`)

func Load(reader io.Reader) (Report, error) {
	var data struct {
		Log struct {
			Entries []harEntry `json:"entries"`
		} `json:"log"`
	}
	result := Report{Posters: map[string][]byte{}}
	if err := json.NewDecoder(io.LimitReader(reader, 128*1024*1024)).Decode(&data); err != nil {
		return result, err
	}
	seen := map[string]bool{}
	for _, e := range data.Log.Entries {
		if e.Response.Status != 200 {
			continue
		}
		u, err := url.Parse(e.Request.URL)
		if err != nil || u.Scheme != "https" || u.User != nil {
			continue
		}
		if posterURL(e.Request.URL) {
			body, err := decode(e.Response.Content)
			if err == nil && len(body) > 0 && len(body) <= 2*1024*1024 {
				result.Posters[e.Request.URL] = body
			}
			continue
		}
		match := collection.FindStringSubmatch(u.Path)
		if u.Hostname() != "movie.douban.com" || match == nil {
			continue
		}
		if result.Profile != "" && result.Profile != match[1] {
			return result, errors.New("HAR contains multiple Douban profiles")
		}
		result.Profile = match[1]
		body, err := decode(e.Response.Content)
		if err != nil {
			return result, err
		}
		records, total, err := parsePage(string(body))
		if err != nil {
			return result, err
		}
		result.Total = max(result.Total, total)
		start := 0
		if raw := u.Query().Get("start"); raw != "" {
			_, err = fmtScanStart(raw, &start)
			if err != nil {
				return result, err
			}
		}
		result.Starts = append(result.Starts, start)
		for _, record := range records {
			key := record.Entry.SourceID + ":" + record.Entry.WatchedOn
			if !seen[key] {
				seen[key] = true
				result.Records = append(result.Records, record)
			}
		}
	}
	if len(result.Records) == 0 {
		return result, errors.New("HAR contains no valid watched records")
	}
	sort.Ints(result.Starts)
	return result, nil
}

func decode(c content) ([]byte, error) {
	if c.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(c.Text)
	}
	if c.Encoding != "" {
		return nil, errors.New("unsupported HAR encoding")
	}
	return []byte(c.Text), nil
}

func posterURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.User == nil &&
		strings.HasSuffix(u.Hostname(), ".doubanio.com") && strings.HasPrefix(u.Path, "/view/photo/s_ratio_poster/public/")
}
