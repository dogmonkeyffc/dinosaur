package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"dinosaur/internal/agent"
	"dinosaur/internal/metadata"
	"dinosaur/internal/nfo"
)

func (s *Server) agentTasks(w http.ResponseWriter, r *http.Request) {
	filter, ok := metadataReportFilter(w, r, 20, "pending matched unmatched needs_review failed warning all")
	if !ok {
		return
	}
	if filter.Status == "" {
		filter.Status = "pending"
	} else if filter.Status == "all" {
		filter.Status = ""
	}
	page, err := s.agentService().Tasks(r.Context(), agentKey(r), filter)
	if s.agentError(w, r, err) {
		return
	}
	writeJSON(w, 200, page)
}

func (s *Server) agentTarget(w http.ResponseWriter, r *http.Request) {
	target, ok := s.scopedAgentTarget(w, r, false)
	if !ok {
		return
	}
	writeJSON(w, 200, target)
}

func targetInfo(target agent.Target) metadata.Movie {
	if target.Metadata != nil {
		return *target.Metadata
	}
	return metadata.Movie{Title: target.Title, Year: target.Year}
}

func (s *Server) agentMetadata(w http.ResponseWriter, r *http.Request) {
	target, ok := s.scopedAgentTarget(w, r, false)
	if !ok {
		return
	}
	writeJSON(w, 200, agent.FromMovie(targetInfo(target)))
}

func (s *Server) agentNFO(w http.ResponseWriter, r *http.Request) {
	target, ok := s.scopedAgentTarget(w, r, false)
	if !ok {
		return
	}
	data, err := nfo.Encode(targetInfo(target), target.Kind)
	if s.agentError(w, r, err) {
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+target.Kind+"-"+strconv.FormatInt(target.ID, 10)+`.nfo"`)
	w.WriteHeader(200)
	_, _ = w.Write(data)
}

func decodeAgentMetadata(w http.ResponseWriter, r *http.Request, kind string) (agent.Input, string, bool) {
	var input agent.Input
	content, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || (content != "application/json" && content != "application/xml" && content != "text/xml") {
		fail(w, 415, "unsupported_media_type", "use application/json or application/xml")
		return input, "", false
	}
	r.Body = http.MaxBytesReader(w, r.Body, nfo.MaxSize)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var large *http.MaxBytesError
		if errors.As(err, &large) {
			fail(w, 413, "body_too_large", "metadata exceeds 128 KiB")
		} else {
			fail(w, 400, "invalid_agent_input", "could not read metadata")
		}
		return input, "", false
	}
	source := "agent"
	if content != "application/json" {
		source = "nfo"
		input, err = nfo.Decode(data, kind)
	} else {
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&input)
		if err == nil {
			var extra any
			if decoder.Decode(&extra) != io.EOF {
				err = agent.ErrInput
			}
		}
		if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
			err = agent.ErrInput
		}
	}
	if err != nil {
		fail(w, 400, "invalid_agent_input", "invalid metadata JSON or NFO")
		return input, "", false
	}
	return input, source, true
}

func (s *Server) putAgentMetadata(w http.ResponseWriter, r *http.Request) {
	target, ok := s.scopedAgentTarget(w, r, true)
	if !ok {
		return
	}
	expected, ok := agentVersion(w, r)
	if !ok {
		return
	}
	input, source, ok := decodeAgentMetadata(w, r, target.Kind)
	if !ok {
		return
	}
	version, err := s.agentService().Update(r.Context(), agentKey(r), target.Kind, target.ID, expected, input, source)
	if errors.Is(err, metadata.ErrConflict) {
		w.Header().Set("ETag", `"`+strconv.FormatInt(version, 10)+`"`)
	}
	if s.agentError(w, r, err) {
		return
	}
	w.Header().Set("ETag", `"`+strconv.FormatInt(version, 10)+`"`)
	writeJSON(w, 200, map[string]any{"version": version, "source": source, "confirmed": true})
}
