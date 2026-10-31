package agent

import (
	"context"
	"strings"
	"time"

	"dinosaur/internal/metadata"
)

type Repository interface {
	AgentTarget(context.Context, string, int64) (Target, error)
	MetadataRecords(context.Context, metadata.ReportFilter) (metadata.RecordPage, error)
	ApplyAgentMetadata(context.Context, Key, string, int64, int64, metadata.Movie) (int64, error)
}

type Service struct{ Repository Repository }

func ValidateKey(input CreateKey) (CreateKey, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Access == "" {
		input.Access = "read"
	}
	if input.ExpiresInDays == 0 {
		input.ExpiresInDays = 90
	}
	if input.Name == "" || !validText(input.Name, 80, false) || input.LibraryID < 0 || (input.Access != "read" && input.Access != "read_write") || input.ExpiresInDays < 1 || input.ExpiresInDays > 365 {
		return input, ErrInput
	}
	return input, nil
}

func (s Service) Target(ctx context.Context, key Key, kind string, id int64, write bool) (Target, error) {
	if (kind != "movie" && kind != "tv") || id < 1 {
		return Target{}, ErrInput
	}
	target, err := s.Repository.AgentTarget(ctx, kind, id)
	if err != nil {
		return target, err
	}
	if !key.Allows(target.LibraryID, write) {
		return Target{}, ErrForbidden
	}
	return target, nil
}

func (s Service) Tasks(ctx context.Context, key Key, filter metadata.ReportFilter) (metadata.RecordPage, error) {
	if key.LibraryID > 0 {
		if filter.LibraryID > 0 && filter.LibraryID != key.LibraryID {
			return metadata.RecordPage{}, ErrForbidden
		}
		filter.LibraryID = key.LibraryID
	}
	return s.Repository.MetadataRecords(ctx, filter)
}

func (s Service) Update(ctx context.Context, key Key, kind string, id, version int64, input Input, source string) (int64, error) {
	if version < 0 {
		return 0, ErrInput
	}
	if _, err := s.Target(ctx, key, kind, id, true); err != nil {
		return 0, err
	}
	info, err := input.Movie(source)
	if err != nil {
		return 0, err
	}
	return s.Repository.ApplyAgentMetadata(ctx, key, kind, id, version, info)
}

func NewKey(userID int64, input CreateKey, prefix string) Key {
	now := time.Now()
	return Key{UserID: userID, Name: input.Name, LibraryID: input.LibraryID, Access: input.Access, Prefix: prefix, CreatedAt: now.Unix(), ExpiresAt: now.Add(time.Duration(input.ExpiresInDays) * 24 * time.Hour).Unix()}
}

func (s Service) Artwork(ctx context.Context, key Key, target Target, expected int64, poster string) (int64, error) {
	if !key.Allows(target.LibraryID, true) {
		return 0, ErrForbidden
	}
	if target.Metadata == nil || !target.Metadata.Matched() {
		return 0, ErrMetadataRequired
	}
	if expected != target.Version {
		return target.Version, metadata.ErrConflict
	}
	if len(poster) != 64 || strings.ContainsFunc(poster, func(r rune) bool { return !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') }) {
		return 0, ErrInput
	}
	info := *target.Metadata
	info.Source, info.Status, info.Message, info.PosterKey = "agent", "matched", "", poster
	return s.Repository.ApplyAgentMetadata(ctx, key, target.Kind, target.ID, expected, info)
}
