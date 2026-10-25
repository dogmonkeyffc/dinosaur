package local

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestUsagePreservesReservedSpaceAndClampsUnavailableBlocks(t *testing.T) {
	tests := []struct {
		name      string
		stats     unix.Statfs_t
		total     uint64
		used      uint64
		available uint64
		wantError bool
	}{
		{"reserved blocks", unix.Statfs_t{Blocks: 100, Bfree: 30, Bavail: 20, Frsize: 4096, Bsize: 8192}, 409600, 286720, 81920, false},
		{"block size fallback", unix.Statfs_t{Blocks: 100, Bfree: 30, Bavail: 20, Bsize: 4096}, 409600, 286720, 81920, false},
		{"full disk", unix.Statfs_t{Blocks: 100, Bsize: 4096}, 409600, 409600, 0, false},
		{"negative availability", unix.Statfs_t{Blocks: 100, Bfree: 30, Bavail: ^uint64(0), Bsize: 4096}, 409600, 286720, 0, false},
		{"inconsistent counters", unix.Statfs_t{Blocks: 100, Bfree: 110, Bavail: 120, Bsize: 4096}, 409600, 0, 409600, false},
		{"invalid block size", unix.Statfs_t{Blocks: 100}, 0, 0, 0, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			usage, err := usageFromStat("/data", test.stats)
			if (err != nil) != test.wantError {
				t.Fatalf("err=%v", err)
			}
			if usage.TotalBytes != test.total || usage.UsedBytes != test.used || usage.AvailableBytes != test.available {
				t.Fatalf("unexpected usage: %+v", usage)
			}
		})
	}
}

func TestReadUsageUsesVisibleMount(t *testing.T) {
	point := t.TempDir()
	var info unix.Stat_t
	if err := unix.Stat(point, &info); err != nil {
		t.Fatal(err)
	}
	usage, err := readUsage(mount{point: point, device: deviceID(info.Dev)})
	if err != nil || usage.TotalBytes == 0 || usage.UsedBytes+usage.AvailableBytes > usage.TotalBytes {
		t.Fatalf("usage=%+v err=%v", usage, err)
	}
	if _, err := readUsage(mount{point: point, device: "999:999"}); err != errHiddenMount {
		t.Fatalf("hidden mount error=%v", err)
	}
}

func TestBtrfsPoolDeviceDiscovery(t *testing.T) {
	root := t.TempDir()
	for name, device := range map[string]string{"shared-uuid/devices/sda1": "8:1\n", "shared-uuid/devices/sdb1": "8:17\n", "other-uuid/devices/sdc1": "8:33\n"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "dev"), []byte(device), 0600); err != nil {
			t.Fatal(err)
		}
	}
	pools := btrfsPools(root)
	if len(pools) != 3 || pools["8:1"] != "shared-uuid" || pools["8:17"] != "shared-uuid" || pools["8:33"] != "other-uuid" {
		t.Fatalf("incorrect pool identities: %+v", pools)
	}
}
