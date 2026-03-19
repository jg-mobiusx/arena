package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/johngillam/arena/model"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "devices.json")

	now := time.Now().Truncate(time.Second)
	devices := []model.Device{
		{
			IP:           "172.16.44.1",
			MAC:          "00:1A:2B:3C:4D:5E",
			Hostname:     "router.local",
			Manufacturer: "Netgear",
			FirstSeen:    now,
			LastSeen:     now,
			IsOnline:     true,
		},
		{
			IP:           "172.16.44.10",
			MAC:          "B8:27:EB:AA:BB:CC",
			Hostname:     "pihole",
			Manufacturer: "Raspberry Pi Foundation",
			OpenPorts: []model.Port{
				{Number: 80, Protocol: "tcp", Service: "http"},
				{Number: 53, Protocol: "udp", Service: "dns"},
			},
			FirstSeen: now,
			LastSeen:  now,
			IsOnline:  true,
		},
	}

	// Save
	if err := Save(path, devices); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(loaded))
	}

	if loaded[0].IP != "172.16.44.1" {
		t.Errorf("expected IP 172.16.44.1, got %s", loaded[0].IP)
	}
	if loaded[1].Hostname != "pihole" {
		t.Errorf("expected hostname pihole, got %s", loaded[1].Hostname)
	}
	if len(loaded[1].OpenPorts) != 2 {
		t.Errorf("expected 2 ports, got %d", len(loaded[1].OpenPorts))
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	devices, err := Load("/nonexistent/path/devices.json")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("expected empty slice, got %d devices", len(devices))
	}
}

func TestLoadCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.json")

	if err := os.WriteFile(path, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}
