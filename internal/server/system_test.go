package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"dinosaur/internal/disk"
)

type diskReaderFunc func(context.Context) (disk.Report, error)

func (read diskReaderFunc) Read(ctx context.Context) (disk.Report, error) { return read(ctx) }

func TestDiskUsagePermissionsAndFailures(t *testing.T) {
	s, _ := fixture(t)
	setupAdmin(t, s)
	admin := loginAs(t, s, "admin", testPassword)
	status(t, request(t, s, "POST", "/api/admin/users", map[string]string{"username": "viewer", "password": testPassword}, &admin, nil), 201)
	viewer := loginAs(t, s, "viewer", testPassword)
	path := "/api/admin/system/disks"
	reads := 0
	s.disks = diskReaderFunc(func(context.Context) (disk.Report, error) {
		reads++
		return disk.Report{Disks: []disk.Usage{{MountPoint: "/data", TotalBytes: 100, UsedBytes: 60, AvailableBytes: 30}}, Incomplete: true}, nil
	})
	status(t, request(t, s, "GET", path, nil, nil, nil), 401)
	status(t, request(t, s, "GET", path, nil, &viewer, nil), 403)
	if reads != 0 {
		t.Fatal("unauthorized requests read host disk usage")
	}
	response := request(t, s, "GET", path, nil, &admin, nil)
	status(t, response, 200)
	var report disk.Report
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Disks) != 1 || report.Disks[0].AvailableBytes != 30 || !report.Incomplete || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("invalid report: %+v", report)
	}
	s.disks = diskReaderFunc(func(context.Context) (disk.Report, error) { return disk.Report{}, disk.ErrUnsupported })
	status(t, request(t, s, "GET", path, nil, &admin, nil), 501)
	s.disks = diskReaderFunc(func(context.Context) (disk.Report, error) { return disk.Report{}, errors.New("mountinfo unavailable") })
	status(t, request(t, s, "GET", path, nil, &admin, nil), 500)
}
