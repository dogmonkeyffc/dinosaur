package local

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"dinosaur/internal/disk"
)

func mountRecord(device, root, point, kind, source string) string {
	return fmt.Sprintf("100 1 %s %s %s rw shared:1 - %s %s rw\n", device, root, point, kind, source)
}

func fakeUsage(m mount) (disk.Usage, error) {
	return disk.Usage{MountPoint: m.point, TotalBytes: 100, UsedBytes: 60, AvailableBytes: 30}, nil
}

func TestMountedPartitionsAreCountedOnce(t *testing.T) {
	input := mountRecord("8:1", "/movies", "/movies", "ext4", "/dev/sda1") +
		mountRecord("8:1", "/", "/mnt/media", "ext4", "/dev/disk/by-uuid/media") +
		mountRecord("8:1", "/", "/second-copy", "ext4", "/dev/sda1") +
		mountRecord("8:2", "/", "/mnt/other", "ext4", "/dev/sda2") +
		mountRecord("8:17", "/", "/", "ext4", "/dev/sdb1")
	probes := 0
	report, err := collect(context.Background(), strings.NewReader(input), func(m mount) string {
		return filesystemKey(m, nil, nil)
	}, func(m mount) (disk.Usage, error) {
		probes++
		return fakeUsage(m)
	})
	if err != nil || report.Incomplete || len(report.Disks) != 3 || probes != 3 {
		t.Fatalf("report=%+v probes=%d err=%v", report, probes, err)
	}
	want := []string{"/", "/mnt/media", "/mnt/other"}
	for i, point := range want {
		if report.Disks[i].MountPoint != point {
			t.Fatalf("disk %d: got %q want %q", i, report.Disks[i].MountPoint, point)
		}
	}
}

func TestSharedSubvolumeCapacityIsCountedOnce(t *testing.T) {
	input := mountRecord("0:35", "/@home", "/home", "btrfs", "/dev/sda1") +
		mountRecord("0:34", "/@", "/", "btrfs", "/dev/sdb1") +
		mountRecord("0:36", "/", "/other", "btrfs", "/dev/sdc1")
	// Both backing devices belong to the same multi-device Btrfs filesystem.
	devices := map[string]string{"/dev/sda1": "8:1", "/dev/sdb1": "8:17", "/dev/sdc1": "8:33"}
	pools := map[string]string{"8:1": "shared-uuid", "8:17": "shared-uuid", "8:33": "other-uuid"}
	report, err := collect(context.Background(), strings.NewReader(input), func(m mount) string {
		return filesystemKey(m, pools, func(source string) string { return devices[source] })
	}, fakeUsage)
	if err != nil || len(report.Disks) != 2 || report.Disks[0].MountPoint != "/" || report.Disks[1].MountPoint != "/other" {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}

func TestVirtualMountsAreExcludedAndPathsDecoded(t *testing.T) {
	input := mountRecord("0:1", "/", "/proc", "proc", "proc") +
		mountRecord("0:2", "/", "/run", "tmpfs", "tmpfs") +
		mountRecord("7:1", "/", "/snap/core", "squashfs", "/dev/loop1") +
		mountRecord("7:2", "/", "/mnt/image", "ext4", "/dev/loop2") +
		mountRecord("0:3", "/", "/merged", "overlay", "overlay") +
		mountRecord("11:0", "/", "/media/cdrom", "iso9660", "/dev/sr0") +
		mountRecord("0:4", "/", "/mnt/network", "nfs4", "nas:/share") +
		mountRecord("8:1", "/", `/mnt/media\040disk\134backup`, "ext4", "/dev/sda1")
	mounts, err := parseMounts(strings.NewReader(input))
	if err != nil || len(mounts) != 1 || mounts[0].point != `/mnt/media disk\backup` {
		t.Fatalf("mounts=%+v err=%v", mounts, err)
	}
}

func TestUnreadableMountFallbackAndPartialReport(t *testing.T) {
	input := mountRecord("8:1", "/", "/a", "ext4", "/dev/sda1") +
		mountRecord("8:1", "/movies", "/b", "ext4", "/dev/sda1") +
		mountRecord("8:2", "/", "/hidden", "ext4", "/dev/sda2")
	key := func(m mount) string { return m.device }
	probe := func(m mount) (disk.Usage, error) {
		switch m.point {
		case "/a", "/missing":
			return disk.Usage{}, errors.New("unreadable mount")
		case "/hidden":
			return disk.Usage{}, errHiddenMount
		default:
			return fakeUsage(m)
		}
	}
	report, err := collect(context.Background(), strings.NewReader(input), key, probe)
	if err != nil || report.Incomplete || len(report.Disks) != 1 || report.Disks[0].MountPoint != "/b" {
		t.Fatalf("fallback report=%+v err=%v", report, err)
	}
	input += mountRecord("8:3", "/", "/missing", "ext4", "/dev/sda3")
	report, err = collect(context.Background(), strings.NewReader(input), key, probe)
	if err != nil || !report.Incomplete || len(report.Disks) != 1 {
		t.Fatalf("partial report=%+v err=%v", report, err)
	}
}

func TestEmptyInvalidAndCancelledCollection(t *testing.T) {
	key := func(m mount) string { return m.device }
	report, err := collect(context.Background(), strings.NewReader(""), key, fakeUsage)
	if err != nil || report.Disks == nil || len(report.Disks) != 0 || report.Incomplete {
		t.Fatalf("empty report=%+v err=%v", report, err)
	}
	if _, err := parseMounts(strings.NewReader("invalid record")); err == nil {
		t.Fatal("invalid mountinfo accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := collect(ctx, strings.NewReader(""), key, fakeUsage); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error=%v", err)
	}
}
