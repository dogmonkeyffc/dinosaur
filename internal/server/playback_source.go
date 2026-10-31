package server

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"dinosaur/internal/media"
	"dinosaur/internal/playback"
)

type playbackSource struct {
	Item media.Item
	Root string
	Path string
}

func (s *Server) playbackSource(w http.ResponseWriter, r *http.Request) (playbackSource, bool) {
	id, ok := libraryID(w, r)
	if !ok {
		return playbackSource{}, false
	}
	item, err := s.store.MediaByID(r.Context(), id)
	if err != nil {
		s.playbackError(w, err)
		return playbackSource{}, false
	}
	library, err := s.store.LibraryByID(r.Context(), item.LibraryID)
	if err != nil {
		s.playbackError(w, err)
		return playbackSource{}, false
	}
	path, err := media.ResolveFile(library.RootPath, item.RelativePath)
	if err != nil {
		fail(w, 404, "media_unavailable", "media file unavailable")
		return playbackSource{}, false
	}
	return playbackSource{Item: item, Root: library.RootPath, Path: path}, true
}
func requestPlan(r *http.Request, item media.Item) (playback.Plan, error) {
	start := 0.0
	var err error
	if value := r.URL.Query().Get("start"); value != "" {
		start, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return playback.Plan{}, playback.ErrPosition
		}
	}
	var audio *int
	if value := r.URL.Query().Get("audio"); value != "" {
		index, err := strconv.Atoi(value)
		if err != nil {
			return playback.Plan{}, playback.ErrAudio
		}
		audio = &index
	}
	source := playback.Source{OriginalLanguage: item.OriginalLanguage, Filename: item.Filename, Duration: item.Duration, VideoCodec: item.VideoCodec}
	for _, track := range item.Tracks {
		if track.Type == "audio" {
			source.Audio = append(source.Audio, playback.Audio{Index: track.Index, Codec: track.Codec, Language: track.Language, Title: track.Title, Default: track.Default != nil && *track.Default, Original: track.Original, Commentary: track.Commentary, Dubbed: track.Dubbed})
		}
	}
	mode := r.URL.Query().Get("mode")
	if mode != "" && mode != "auto" && mode != "compatible" && mode != "original" {
		return playback.Plan{}, errors.New("invalid playback mode")
	}
	plan, err := playback.SelectPlan(source, audio, mode == "compatible", start)
	if err == nil && mode == "original" {
		plan.Mode = "direct"
	}
	if err == nil && r.URL.Query().Has("subtitle") {
		id, parseErr := strconv.Atoi(r.URL.Query().Get("subtitle"))
		if parseErr != nil {
			return plan, playback.ErrSubtitle
		}
		return playback.WithSubtitle(plan, id, item.Tracks)
	}
	return plan, err
}
func streamURL(id int64, plan playback.Plan, compatible bool) string {
	query := url.Values{}
	if plan.Mode == "direct" {
		query.Set("mode", "original")
	} else if compatible {
		query.Set("mode", "compatible")
	}
	if plan.Subtitle != nil {
		query.Set("subtitle", strconv.Itoa(-*plan.Subtitle-1))
	}
	if plan.Audio >= 0 {
		query.Set("audio", strconv.Itoa(plan.Audio))
	}
	return "/api/playback/" + strconv.FormatInt(id, 10) + "/stream?" + query.Encode()
}
