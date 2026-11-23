package doubanhar

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseCapturedHTMLAndPosters(t *testing.T) {
	html := `<h1>看过的影视(110)</h1><div class="item comment-item"><li class="title"><a href="https://movie.douban.com/subject/123/"><em>片名 / Original</em></a></li><li class="intro">2020-01-01 / 演员</li><span class="date">2024-04-21</span><span class="rating5-t"></span><span class="comment">短评 &amp; 记录</span><img src="https://img3.doubanio.com/view/photo/s_ratio_poster/public/a.webp"></div>`
	entry := harEntry{}
	entry.Request.URL = "https://movie.douban.com/people/99/collect?start=15"
	entry.Response.Status = 200
	entry.Response.Content = content{Text: html}
	poster := harEntry{}
	poster.Request.URL = "https://img3.doubanio.com/view/photo/s_ratio_poster/public/a.webp"
	poster.Response.Status = 200
	poster.Response.Content = content{Encoding: "base64", Text: base64.StdEncoding.EncodeToString([]byte("image"))}
	data, _ := json.Marshal(map[string]any{"log": map[string]any{"entries": []harEntry{entry, entry, poster}}})
	report, err := Load(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Records) != 1 || report.Total != 110 || report.Profile != "99" {
		t.Fatalf("wrong report: %+v", report)
	}
	e := report.Records[0].Entry
	if e.Title != "片名" || e.OriginalTitle != "Original" || e.Rating != 5 || e.Review != "短评 & 记录" || e.Year != 2020 {
		t.Fatalf("wrong record: %+v", e)
	}
	if len(report.Posters) != 1 {
		t.Fatal("poster not extracted")
	}
	for _, bad := range []string{strings.Replace(html, "2024-04-21", "2024-02-30", 1), `<div class="item comment-item"></div>`} {
		if _, _, err := parsePage(bad); err == nil {
			t.Fatal("malformed entry accepted")
		}
	}
}
