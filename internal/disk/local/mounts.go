// Package local reads disk capacity from the host operating system.
package local

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"dinosaur/internal/disk"
)

type Reader struct{}

type mount struct {
	device string
	root   string
	point  string
	kind   string
	source string
}

var errHiddenMount = errors.New("mount is no longer visible")

var mountEscapes = strings.NewReplacer(`\040`, " ", `\011`, "\t", `\012`, "\n", `\134`, `\`)

func parseMounts(input io.Reader) ([]mount, error) {
	var mounts []mount
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		before, after, ok := strings.Cut(scanner.Text(), " - ")
		fields, filesystem := strings.Fields(before), strings.Fields(after)
		if !ok || len(fields) < 6 || len(filesystem) < 3 {
			return nil, errors.New("invalid mountinfo record")
		}
		m := mount{
			device: fields[2], root: mountEscapes.Replace(fields[3]),
			point: mountEscapes.Replace(fields[4]), kind: filesystem[0],
			source: mountEscapes.Replace(filesystem[1]),
		}
		if localDisk(m) {
			mounts = append(mounts, m)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read mountinfo: %w", err)
	}
	// Prefer a full filesystem mount over binds, then the shortest visible path.
	sort.Slice(mounts, func(i, j int) bool {
		a, b := mounts[i], mounts[j]
		if (a.root == "/") != (b.root == "/") {
			return a.root == "/"
		}
		if len(a.point) != len(b.point) {
			return len(a.point) < len(b.point)
		}
		return a.point < b.point
	})
	return mounts, nil
}

func localDisk(m mount) bool {
	if !strings.HasPrefix(m.source, "/dev/") || strings.HasPrefix(m.device, "7:") {
		return false
	}
	switch m.kind {
	case "tmpfs", "devtmpfs", "squashfs", "iso9660", "udf", "overlay":
		return false
	}
	return true
}

func filesystemKey(m mount, pools map[string]string, sourceDevice func(string) string) string {
	if m.kind != "btrfs" {
		return "device:" + m.device
	}
	// Btrfs subvolumes have distinct st_dev and statfs IDs but share capacity.
	if id := sourceDevice(m.source); id != "" {
		if uuid := pools[id]; uuid != "" {
			return "btrfs:" + uuid
		}
		return "btrfs:device:" + id
	}
	return "btrfs:source:" + m.source
}

func collect(ctx context.Context, input io.Reader, key func(mount) string, probe func(mount) (disk.Usage, error)) (disk.Report, error) {
	report := disk.Report{Disks: []disk.Usage{}}
	mounts, err := parseMounts(input)
	if err != nil {
		return report, err
	}
	seen, failed := map[string]bool{}, map[string]bool{}
	for _, m := range mounts {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		id := key(m)
		if seen[id] {
			continue
		}
		usage, err := probe(m)
		if errors.Is(err, errHiddenMount) {
			continue
		}
		if err != nil {
			failed[id] = true
			continue
		}
		if usage.TotalBytes == 0 {
			continue
		}
		seen[id] = true
		delete(failed, id)
		report.Disks = append(report.Disks, usage)
	}
	report.Incomplete = len(failed) > 0
	sort.Slice(report.Disks, func(i, j int) bool { return report.Disks[i].MountPoint < report.Disks[j].MountPoint })
	return report, ctx.Err()
}
