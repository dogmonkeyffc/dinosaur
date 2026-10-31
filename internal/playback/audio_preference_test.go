package playback

import "testing"

func TestPreferredOriginalAudio(t *testing.T) {
	for _, test := range []struct {
		name, language string
		tracks         []Audio
		want           int
	}{
		{"Original flag before dubbed default", "", []Audio{{Index: 1, Default: true, Dubbed: true}, {Index: 2, Original: true}}, 2},
		{"Explicit original title", "", []Audio{{Index: 1, Title: "国语配音", Default: true}, {Index: 2, Title: "Original DTS"}}, 2},
		{"Japanese original", "ja", []Audio{{Index: 1, Language: "chi", Default: true}, {Index: 2, Language: "jpn"}}, 2},
		{"French original is not English", "fr", []Audio{{Index: 1, Language: "eng", Default: true}, {Index: 2, Language: "fre"}}, 2},
		{"Chinese original is not English", "zh-CN", []Audio{{Index: 1, Language: "eng", Default: true}, {Index: 2, Language: "chi"}}, 2},
		{"Cantonese original", "cn", []Audio{{Index: 1, Language: "chi", Title: "国语", Default: true}, {Index: 2, Language: "chi", Title: "粤语"}}, 2},
		{"Avoid commentary even if marked original", "en", []Audio{{Index: 1, Language: "eng", Original: true, Commentary: true}, {Index: 2, Language: "eng"}}, 2},
		{"Title commentary excluded", "en", []Audio{{Index: 1, Language: "eng", Title: "Director commentary", Default: true}, {Index: 2, Language: "eng"}}, 2},
		{"Unknown original uses file default", "", []Audio{{Index: 1, Language: "eng"}, {Index: 2, Language: "chi", Default: true}}, 2},
		{"No implicit English original", "", []Audio{{Index: 1, Language: "chi"}, {Index: 2, Language: "eng"}}, 1},
		{"Negated original title", "", []Audio{{Index: 1, Title: "no original"}, {Index: 2, Default: true}}, 2},
		{"No audio", "", nil, -1},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := PreferredAudio(test.tracks, test.language); got != test.want {
				t.Fatalf("audio=%d want %d", got, test.want)
			}
		})
	}
}

func TestOriginalAudioPlanAndManualOverride(t *testing.T) {
	source := Source{Filename: "movie.mp4", VideoCodec: "h264", Duration: 600, OriginalLanguage: "ja", Audio: []Audio{{Index: 1, Codec: "aac", Language: "chi", Default: true}, {Index: 2, Codec: "aac", Language: "jpn"}}}
	plan, err := SelectPlan(source, nil, false, 70)
	if err != nil || plan.Audio != 2 || plan.Mode != "remux" || plan.Start != 70 {
		t.Fatalf("original plan=%+v %v", plan, err)
	}
	chosen := 1
	manual, err := SelectPlan(source, &chosen, false, 70)
	if err != nil || manual.Audio != 1 || manual.Mode != "direct" {
		t.Fatalf("manual plan=%+v %v", manual, err)
	}
}
