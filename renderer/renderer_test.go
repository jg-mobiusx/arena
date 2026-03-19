package renderer

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/johngillam/arena/model"
)

func TestRender_ContainsDeviceData(t *testing.T) {
	now := time.Now()
	devices := []model.Device{
		{
			IP:           "172.16.44.1",
			MAC:          "00:09:5B:AA:BB:CC",
			Hostname:     "router.local",
			Manufacturer: "Netgear",
			IsOnline:     true,
			FirstSeen:    now,
			LastSeen:     now,
		},
		{
			IP:           "172.16.44.10",
			MAC:          "B8:27:EB:11:22:33",
			Hostname:     "pihole",
			Manufacturer: "Raspberry Pi Foundation",
			OpenPorts: []model.Port{
				{Number: 80, Protocol: "tcp", Service: "http"},
				{Number: 53, Protocol: "tcp", Service: "dns"},
			},
			IsOnline:  true,
			FirstSeen: now,
			LastSeen:  now,
		},
		{
			IP:           "172.16.44.50",
			MAC:          "DC:A6:32:11:22:33",
			Hostname:     "",
			Manufacturer: "",
			IsNew:        true,
			IsOnline:     true,
			FirstSeen:    now,
			LastSeen:     now,
		},
		{
			IP:       "172.16.44.200",
			MAC:      "AA:BB:CC:DD:EE:FF",
			Hostname: "old-device",
			IsOnline: false,
		},
	}

	var buf bytes.Buffer
	Render(&buf, devices)
	output := buf.String()

	// Check that key data appears in the output
	expectations := []string{
		"172.16.44.1",
		"172.16.44.10",
		"router.local",
		"pihole",
		"Netgear",
		"Raspberry Pi Foundation",
		"80/http",
		"53/dns",
		"Online",
		"Offline",
		"NEW",
		"Unknown",                // empty manufacturer
		"4 devices total",
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q, it did not.\nOutput:\n%s", exp, output)
		}
	}
}

func TestRender_SortsByIP(t *testing.T) {
	devices := []model.Device{
		{IP: "172.16.44.100", MAC: "AA:AA:AA:AA:AA:AA", IsOnline: true},
		{IP: "172.16.44.1", MAC: "BB:BB:BB:BB:BB:BB", IsOnline: true},
		{IP: "172.16.44.50", MAC: "CC:CC:CC:CC:CC:CC", IsOnline: true},
	}

	var buf bytes.Buffer
	Render(&buf, devices)
	output := buf.String()

	idx1 := strings.Index(output, "172.16.44.1 ")
	idx50 := strings.Index(output, "172.16.44.50")
	idx100 := strings.Index(output, "172.16.44.100")

	if idx1 == -1 || idx50 == -1 || idx100 == -1 {
		t.Fatalf("not all IPs found in output:\n%s", output)
	}

	if !(idx1 < idx50 && idx50 < idx100) {
		t.Errorf("expected IPs sorted: .1 < .50 < .100, but positions were %d, %d, %d", idx1, idx50, idx100)
	}
}
