package server

import (
	"net/http"
	"strconv"
	"strings"
)

func resumeLibraryIDs(w http.ResponseWriter, r *http.Request) ([]int64, bool) {
	query := r.URL.Query()
	if !query.Has("library_ids") {
		return nil, true
	}
	result := []int64{}
	if value := query.Get("library_ids"); value != "" {
		for _, field := range strings.Split(value, ",") {
			id, err := strconv.ParseInt(field, 10, 64)
			if err != nil || id <= 0 {
				fail(w, 400, "invalid_filter", "invalid media library ID")
				return nil, false
			}
			result = append(result, id)
		}
	}
	return result, true
}
