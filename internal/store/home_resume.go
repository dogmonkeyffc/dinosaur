package store

import (
	"context"
	"encoding/json"
	"sort"

	"dinosaur/internal/media"
)

const resumeLibraryFilter = "(? IS NULL OR m.library_id IN (SELECT value FROM json_each(?)))"

func (s *Store) ContinueHome(ctx context.Context, userID int64, libraryIDs []int64) ([]media.Item, error) {
	var scope any
	if libraryIDs != nil {
		encoded, err := json.Marshal(libraryIDs)
		if err != nil {
			return nil, err
		}
		scope = string(encoded)
	}
	items, err := s.continueMedia(ctx, userID, "", scope)
	if err != nil {
		return nil, err
	}
	episodes, err := s.continueSeries(ctx, userID, scope, true)
	if err != nil {
		return nil, err
	}
	items = append(items, episodes...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].LastPlayedAt == items[j].LastPlayedAt {
			return items[i].ID > items[j].ID
		}
		return items[i].LastPlayedAt > items[j].LastPlayedAt
	})
	return items[:min(12, len(items))], nil
}
