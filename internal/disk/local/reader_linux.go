package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dinosaur/internal/disk"
	"golang.org/x/sys/unix"
)

func (Reader) Read(ctx context.Context) (disk.Report, error) {
	file, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return disk.Report{}, err
	}
	defer file.Close()
	pools := btrfsPools("/sys/fs/btrfs")
	return collect(ctx, file, func(m mount) string { return filesystemKey(m, pools, blockDeviceID) }, readUsage)
}

func deviceID(device uint64) string {
	return fmt.Sprintf("%d:%d", unix.Major(device), unix.Minor(device))
}

func blockDeviceID(path string) string {
	var source unix.Stat_t
	if err := unix.Stat(path, &source); err == nil && source.Mode&unix.S_IFMT == unix.S_IFBLK {
		return deviceID(source.Rdev)
	}
	return ""
}

func btrfsPools(root string) map[string]string {
	pools := map[string]string{}
	paths, _ := filepath.Glob(filepath.Join(root, "*", "devices", "*", "dev"))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		// /sys/fs/btrfs/<filesystem UUID>/devices/<device>/dev
		uuid := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(path))))
		pools[strings.TrimSpace(string(data))] = uuid
	}
	return pools
}

func readUsage(m mount) (disk.Usage, error) {
	fd, err := unix.Open(m.point, unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		return disk.Usage{}, err
	}
	defer unix.Close(fd)
	var info unix.Stat_t
	if err := unix.Fstat(fd, &info); err != nil {
		return disk.Usage{}, err
	}
	if deviceID(info.Dev) != m.device {
		return disk.Usage{}, errHiddenMount
	}
	var capacity unix.Statfs_t
	if err := unix.Fstatfs(fd, &capacity); err != nil {
		return disk.Usage{}, err
	}
	return usageFromStat(m.point, capacity)
}

func usageFromStat(point string, capacity unix.Statfs_t) (disk.Usage, error) {
	blockSize := capacity.Frsize
	if blockSize <= 0 {
		blockSize = capacity.Bsize
	}
	if blockSize <= 0 {
		return disk.Usage{}, fmt.Errorf("invalid filesystem block size for %s", point)
	}
	free := min(capacity.Bfree, capacity.Blocks)
	available := uint64(0)
	if int64(capacity.Bavail) > 0 {
		available = min(capacity.Bavail, free)
	}
	return disk.Usage{
		MountPoint: point, TotalBytes: capacity.Blocks * uint64(blockSize),
		UsedBytes:      (capacity.Blocks - free) * uint64(blockSize),
		AvailableBytes: available * uint64(blockSize),
	}, nil
}
