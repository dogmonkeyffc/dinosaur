package server

import (
	"net/http"
	"strconv"
	"strings"

	"dinosaur/internal/media"
)

func mediaFilter(w http.ResponseWriter, r *http.Request) (media.Filter, bool) {
	filter := media.Filter{Page: 1, PageSize: 24, Query: strings.TrimSpace(r.URL.Query().Get("q"))}
	for name, target := range map[string]*int{"page": &filter.Page, "page_size": &filter.PageSize} {
		value := r.URL.Query().Get(name)
		if value == "" {
			continue
		}
		number, err := strconv.Atoi(value)
		if err != nil || number < 1 || number > 1_000_000 || (name == "page_size" && number > 100) {
			fail(w, 400, "invalid_filter", "invalid pagination parameters")
			return filter, false
		}
		*target = number
	}
	if value := r.URL.Query().Get("library_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			fail(w, 400, "invalid_filter", "invalid library ID")
			return filter, false
		}
		filter.LibraryID = id
	}
	if len(filter.Query) > 256 {
		fail(w, 400, "invalid_filter", "search query is too long")
		return filter, false
	}
	return filter, true
}
