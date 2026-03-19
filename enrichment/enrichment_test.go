package enrichment

import (
	"testing"
	"time"

	"github.com/johngillam/arena/model"
)

func TestLookupOUI(t *testing.T) {
	tests := []struct {
		mac      string
		expected string
	}{
		{"B8:27:EB:AA:BB:CC", "Raspberry Pi Foundation"},
		{"DC:A6:32:11:22:33", "Raspberry Pi Trading"},
		{"00:1A:2B:3C:4D:5E", ""}, // Not in our curated DB
		{"00:17:88:AA:BB:CC", "Philips Hue"},
		{"5C:CF:7F:12:34:56", "Espressif (ESP)"},
		{"b8:27:eb:aa:bb:cc", "Raspberry Pi Foundation"}, // lowercase
		{"B8-27-EB-AA-BB-CC", "Raspberry Pi Foundation"}, // dashes
		{"", ""},
		{"AB", ""},
	}

	for _, tc := range tests {
		got := LookupOUI(tc.mac)
		if got != tc.expected {
			t.Errorf("LookupOUI(%q) = %q, want %q", tc.mac, got, tc.expected)
		}
	}
}

func TestEnrich_NewDevices(t *testing.T) {
	scanned := []model.Device{
		{IP: "172.16.44.1", MAC: "00:09:5B:AA:BB:CC", Hostname: "router"},
		{IP: "172.16.44.10", MAC: "B8:27:EB:11:22:33", Hostname: "pihole"},
	}

	result := Enrich(scanned, nil)

	if len(result) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(result))
	}

	for _, d := range result {
		if !d.IsNew {
			t.Errorf("expected device %s to be IsNew", d.IP)
		}
		if !d.IsOnline {
			t.Errorf("expected device %s to be IsOnline", d.IP)
		}
	}

	// Check manufacturer resolution
	if result[0].Manufacturer != "Netgear" {
		t.Errorf("expected Netgear, got %s", result[0].Manufacturer)
	}
	if result[1].Manufacturer != "Raspberry Pi Foundation" {
		t.Errorf("expected Raspberry Pi Foundation, got %s", result[1].Manufacturer)
	}
}

func TestEnrich_KnownDeviceUpdated(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	known := []model.Device{
		{
			IP:           "172.16.44.10",
			MAC:          "B8:27:EB:11:22:33",
			Hostname:     "pihole",
			Manufacturer: "Raspberry Pi Foundation",
			FirstSeen:    past,
			LastSeen:     past,
			IsOnline:     false,
		},
	}

	scanned := []model.Device{
		{IP: "172.16.44.10", MAC: "B8:27:EB:11:22:33", Hostname: "pihole"},
	}

	result := Enrich(scanned, known)

	if len(result) != 1 {
		t.Fatalf("expected 1 device, got %d", len(result))
	}

	d := result[0]
	if d.IsNew {
		t.Error("expected known device NOT to be IsNew")
	}
	if !d.IsOnline {
		t.Error("expected device to be online")
	}
	if d.FirstSeen != past {
		t.Error("expected FirstSeen to be preserved")
	}
	if !d.LastSeen.After(past) {
		t.Error("expected LastSeen to be updated")
	}
}

func TestEnrich_OfflineDevice(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	known := []model.Device{
		{
			IP:       "172.16.44.50",
			MAC:      "DC:A6:32:11:22:33",
			Hostname: "camera",
			FirstSeen: past,
			LastSeen:  past,
			IsOnline:  true,
		},
	}

	// Scan returns empty — device is gone
	scanned := []model.Device{}

	result := Enrich(scanned, known)

	if len(result) != 1 {
		t.Fatalf("expected 1 device, got %d", len(result))
	}

	if result[0].IsOnline {
		t.Error("expected offline device to be marked offline")
	}
}
