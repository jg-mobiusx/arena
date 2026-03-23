package enrichment

import (
	"strings"
	"time"

	"github.com/johngillam/arena/model"
)

// Enrich merges newly scanned devices with previously known devices.
// It sets IsNew for devices not seen before, updates timestamps, and resolves manufacturers via OUI lookup.
func Enrich(scanned []model.Device, known []model.Device) []model.Device {
	now := time.Now()

	// Build lookup of known devices by MAC
	knownByMAC := make(map[string]*model.Device, len(known))
	for i := range known {
		mac := strings.ToUpper(known[i].MAC)
		knownByMAC[mac] = &known[i]
	}

	// Track which known devices are still seen in this scan
	seen := make(map[string]bool)

	var result []model.Device

	for _, s := range scanned {
		mac := strings.ToUpper(s.MAC)
		seen[mac] = true

		if existing, ok := knownByMAC[mac]; ok {
			// Known device — update it
			existing.IP = s.IP
			
			// Flag for re-probing if it just returned online, OR if it has never been probed natively.
			if !existing.IsOnline || existing.LastProbed.IsZero() {
				existing.NeedsProbe = true
			}
			
			existing.IsOnline = true
			existing.LastSeen = now
			existing.IsNew = false

			// Update hostname if we got a better one
			if s.Hostname != "" {
				existing.Hostname = s.Hostname
			}

			// Enrich manufacturer if missing
			if existing.Manufacturer == "" {
				existing.Manufacturer = resolveManufacturer(s)
			}

			// Update ports if scan returned them
			if len(s.OpenPorts) > 0 {
				existing.OpenPorts = s.OpenPorts
			}

			// Update OS if scan returned it
			if s.OS != "" {
				existing.OS = s.OS
			}

			result = append(result, *existing)
		} else {
			// New device
			s.FirstSeen = now
			s.LastSeen = now
			s.IsNew = true
			s.IsOnline = true
			s.NeedsProbe = true
			s.Manufacturer = resolveManufacturer(s)
			result = append(result, s)
		}
	}

	// Mark known devices not in this scan as offline
	for mac, d := range knownByMAC {
		if !seen[mac] {
			d.IsOnline = false
			d.IsNew = false
			result = append(result, *d)
		}
	}

	return result
}

// resolveManufacturer determines the manufacturer from the device's data.
// Prefers the vendor hint from Nmap, falls back to OUI lookup.
func resolveManufacturer(d model.Device) string {
	if d.Manufacturer != "" {
		return d.Manufacturer
	}
	if d.MAC != "" {
		if mfr := LookupOUI(d.MAC); mfr != "" {
			return mfr
		}
	}
	return ""
}
