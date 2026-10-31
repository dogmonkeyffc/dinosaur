package server

import (
	"dinosaur/internal/media"
	"dinosaur/internal/playback"
	"dinosaur/internal/streaming"
	"net/http"
	"strconv"
)

func playbackPreference(r *http.Request, item media.Item, subtitles []streaming.Subtitle) (playback.Plan, *int, error) {
	if r.URL.Query().Get("subtitle") != "auto" {
		plan, err := requestPlan(r, item)
		return plan, nil, err
	}
	options := make([]playback.SubtitleOption, 0, len(subtitles))
	for _, subtitle := range subtitles {
		options = append(options, playback.SubtitleOption{ID: subtitle.ID, Language: subtitle.Language, Title: subtitle.Title, Delivery: subtitle.Delivery})
	}
	selected := playback.PreferredSubtitle(options)
	request := r.Clone(r.Context())
	query := request.URL.Query()
	query.Del("subtitle")
	for _, subtitle := range subtitles {
		if selected != nil && subtitle.ID == *selected && subtitle.Delivery == "burn" {
			query.Set("subtitle", strconv.Itoa(*selected))
		}
	}
	request.URL.RawQuery = query.Encode()
	plan, err := requestPlan(request, item)
	return plan, selected, err
}
