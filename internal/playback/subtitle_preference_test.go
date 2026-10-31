package playback

import "testing"

func TestSubtitleLanguageRecognition(t *testing.T) {
	for _, test := range []struct{ language, title, want string }{
		{"chi", "", "zh"}, {"zho", "", "zh"}, {"ZH_CN", "", "zh"}, {"zh-Hans", "", "zh"}, {"zh-Hant", "", "zh"},
		{"cmn", "", "zh"}, {"yue", "", "zh"}, {"und", "官译简中", "zh"}, {"und", "繁體中文字幕", "zh"},
		{"", "CHS", "zh"}, {"", "cht", "zh"}, {"und", "English.CHS", "zh"}, {"eng", "中英双语", "zh"},
		{"", "Chinese", "zh"}, {"en-US", "", "en"}, {"und", "English", "en"},
		{"und", "", ""}, {"", "外挂字幕", ""}, {"", "French", ""}, {"", "Nachspiel", ""},
		{"und", "双语", ""}, {"und", "无中文字幕", ""}, {"und", "No Chinese subtitles", ""},
	} {
		if got := subtitleLanguage(test.language, test.title); got != test.want {
			t.Errorf("%q/%q = %q, want %q", test.language, test.title, got, test.want)
		}
	}
}

func TestPreferredSubtitlePriorities(t *testing.T) {
	for _, test := range []struct {
		name    string
		options []SubtitleOption
		want    *int
	}{
		{"Chinese beats English", []SubtitleOption{{ID: 1, Language: "eng", Delivery: "text"}, {ID: 2, Language: "chi", Delivery: "text"}}, subtitleID(2)},
		{"Chinese bitmap beats English text", []SubtitleOption{{ID: 1, Language: "eng", Delivery: "text"}, {ID: -4, Language: "chi", Delivery: "burn"}}, subtitleID(-4)},
		{"Chinese text avoids unnecessary bitmap transcoding", []SubtitleOption{{ID: -4, Language: "chi", Title: "官译简中", Delivery: "burn"}, {ID: 0, Title: "CHT", Delivery: "text"}}, subtitleID(0)},
		{"Simplified within equivalent text", []SubtitleOption{{ID: 1, Language: "chi", Title: "繁體", Delivery: "text"}, {ID: 2, Language: "chi", Title: "简体", Delivery: "text"}}, subtitleID(2)},
		{"Full captions before forced", []SubtitleOption{{ID: 1, Language: "chi", Title: "forced", Delivery: "text"}, {ID: 2, Language: "chi", Delivery: "burn"}}, subtitleID(2)},
		{"Explicit English fallback", []SubtitleOption{{ID: 0, Title: "en", Delivery: "text"}}, subtitleID(0)},
		{"Unknown language stays disabled", []SubtitleOption{{ID: 0, Delivery: "text"}}, nil},
		{"Unsupported track stays disabled", []SubtitleOption{{ID: 0, Language: "chi", Delivery: "unknown"}}, nil},
		{"Empty list stays disabled", nil, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := PreferredSubtitle(test.options)
			if (got == nil) != (test.want == nil) || got != nil && *got != *test.want {
				t.Fatalf("preferred=%v, want %v", got, test.want)
			}
		})
	}
}
func subtitleID(id int) *int { return &id }
