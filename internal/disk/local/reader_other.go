//go:build !linux

package local

import (
	"context"

	"dinosaur/internal/disk"
)

func (Reader) Read(context.Context) (disk.Report, error) {
	return disk.Report{}, disk.ErrUnsupported
}
