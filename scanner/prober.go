package scanner

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/johngillam/arena/model"
)

var titleRegex = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
var targetPorts = []int{22, 80, 443, 53, 5000, 8080, 8443, 8006, 3000, 8123} // 8006 Proxmox, 5000 Synology, 8123 HA

// ProbeDevices runs a concurrent port and HTTP banner grabber on flagged devices
func ProbeDevices(devices []model.Device) []model.Device {
	var wg sync.WaitGroup
	result := make([]model.Device, len(devices))
	copy(result, devices)

	// Custom fast HTTP client
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Allow untrusted certs on LAN
			DisableKeepAlives: true,
		},
	}

	// Limit concurrent device probes to prevent macOS "too many open files"
	// 10 devices * 10 ports = max 100 concurrent sockets
	devSem := make(chan struct{}, 10)

	for i := range result {
		if !result[i].NeedsProbe || !result[i].IsOnline {
			continue
		}

		wg.Add(1)
		go func(idx int) {
			devSem <- struct{}{}        // acquire slot
			defer func() { <-devSem }() // release slot
			defer wg.Done()
			
			dev := &result[idx]
			dev.LastProbed = time.Now()
			
			// Sweep target ports concurrently per device
			var portWg sync.WaitGroup
			var mu sync.Mutex
			openPorts := make([]model.Port, 0)

			for _, port := range targetPorts {
				portWg.Add(1)
				go func(p int) {
					defer portWg.Done()
					address := net.JoinHostPort(dev.IP, formatInt(p))
					conn, err := net.DialTimeout("tcp", address, 2500*time.Millisecond)
					if err == nil {
						conn.Close()
						
						mu.Lock()
						openPorts = append(openPorts, model.Port{Number: p, Protocol: "tcp", Service: "open"})
						mu.Unlock()

						// If HTTP/HTTPS port is open, scrape title and server header
						if p == 80 || p == 443 || p == 8080 || p == 8443 || p == 8006 || p == 5000 {
							scrapeHTTP(client, dev, p)
						}
					}
				}(port)
			}
			portWg.Wait()

			// Save open ports
			if len(openPorts) > 0 {
				dev.OpenPorts = openPorts
			}
			dev.NeedsProbe = false // clear flag
		}(i)
	}

	wg.Wait()
	return result
}

func formatInt(i int) string {
	// Custom fast int to string for net.JoinHostPort
	switch i {
	case 22: return "22"
	case 53: return "53"
	case 80: return "80"
	case 443: return "443"
	case 3000: return "3000"
	case 5000: return "5000"
	case 8006: return "8006"
	case 8080: return "8080"
	case 8123: return "8123"
	case 8443: return "8443"
	}
	// fallback (though slice is fixed)
	return "80"
}

func scrapeHTTP(client *http.Client, dev *model.Device, port int) {
	scheme := "http"
	if port == 443 || port == 8443 || port == 8006 {
		scheme = "https"
	}
	url := scheme + "://" + dev.IP
	if port != 80 && port != 443 {
		url += ":" + formatInt(port)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Arena-Scout/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if dev.ServerHeader == "" {
		if srv := resp.Header.Get("Server"); srv != "" {
			dev.ServerHeader = srv
		}
	}

	// Read first 4KB to find title
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err == nil && dev.HttpTitle == "" {
		matches := titleRegex.FindStringSubmatch(string(body))
		if len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			// clean up title
			title = strings.ReplaceAll(title, "\n", "")
			title = strings.ReplaceAll(title, "\r", "")
			dev.HttpTitle = title
		}
	}
}
