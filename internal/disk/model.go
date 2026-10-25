// Package disk describes capacity for mounted filesystems.
package disk

import (
	"context"
	"errors"
)

var ErrUnsupported = errors.New("disk usage is not supported on this platform")

type Usage struct {
	MountPoint     string `json:"mount_point"`
	TotalBytes     uint64 `json:"total_bytes"`
	UsedBytes      uint64 `json:"used_bytes"`
	AvailableBytes uint64 `json:"available_bytes"`
}

type Report struct {
	Disks      []Usage `json:"disks"`
	Incomplete bool    `json:"incomplete"`
}

type Reader interface {
	Read(context.Context) (Report, error)
}
