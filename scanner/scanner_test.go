package scanner

import (
	"os"
	"testing"
)

func TestParseNmapXML(t *testing.T) {
	data, err := os.ReadFile("../testdata/nmap_ping_sweep.xml")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	devices, err := ParseNmapXML(data)
	if err != nil {
		t.Fatalf("failed to parse XML: %v", err)
	}

	if len(devices) != 4 {
		t.Fatalf("expected 4 devices, got %d", len(devices))
	}

	// Verify first device (router)
	router := devices[0]
	if router.IP != "172.16.44.1" {
		t.Errorf("expected IP 172.16.44.1, got %s", router.IP)
	}
	if router.MAC != "00:1A:2B:3C:4D:5E" {
		t.Errorf("expected MAC 00:1A:2B:3C:4D:5E, got %s", router.MAC)
	}
	if router.Hostname != "router.local" {
		t.Errorf("expected hostname router.local, got %s", router.Hostname)
	}
	if router.Manufacturer != "Netgear" {
		t.Errorf("expected manufacturer Netgear, got %s", router.Manufacturer)
	}

	// Verify pihole
	pihole := devices[1]
	if pihole.IP != "172.16.44.10" {
		t.Errorf("expected IP 172.16.44.10, got %s", pihole.IP)
	}
	if pihole.Hostname != "pihole" {
		t.Errorf("expected hostname pihole, got %s", pihole.Hostname)
	}

	// Verify device with no hostname
	noHost := devices[2]
	if noHost.Hostname != "" {
		t.Errorf("expected empty hostname, got %s", noHost.Hostname)
	}

	// Verify device with no vendor
	noVendor := devices[3]
	if noVendor.Manufacturer != "" {
		t.Errorf("expected empty manufacturer, got %s", noVendor.Manufacturer)
	}

	// All devices should be marked online
	for _, d := range devices {
		if !d.IsOnline {
			t.Errorf("expected device %s to be online", d.IP)
		}
	}
}
