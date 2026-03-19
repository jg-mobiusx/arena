package pihole

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Client provides access to Pi-hole's DNS and API for hostname resolution.
type Client struct {
	// Address is the Pi-hole IP (e.g. "172.16.44.11")
	Address  string
	APIToken string
}

// networkResponse is the Pi-hole v5 API /admin/api.php?network response.
type networkResponse struct {
	Network []networkDevice `json:"network"`
}

type networkDevice struct {
	HWaddr    string              `json:"hwaddr"`
	Name      string              `json:"name"`
	FirstSeen int64               `json:"firstSeen"`
	LastQuery int64               `json:"lastQuery"`
	NumQueries int                `json:"numQueries"`
	IP        networkDeviceIPList `json:"ip"`
}

type networkDeviceIPList struct {
	Entries []string
}

func (n *networkDeviceIPList) UnmarshalJSON(data []byte) error {
	// Pi-hole returns IP as either a single string or array
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		n.Entries = []string{single}
		return nil
	}
	return json.Unmarshal(data, &n.Entries)
}

// ResolveHostnames performs reverse DNS lookups against the Pi-hole DNS server
// for each given IP address. Returns a map of IP → hostname.
func (c *Client) ResolveHostnames(ips []string) map[string]string {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 2 * time.Second}
			return d.DialContext(ctx, "udp", c.Address+":53")
		},
	}

	results := make(map[string]string, len(ips))

	for _, ip := range ips {
		names, err := resolver.LookupAddr(context.Background(), ip)
		if err != nil || len(names) == 0 {
			continue
		}
		// Clean up the hostname: remove trailing dot, strip .local. suffix
		hostname := strings.TrimSuffix(names[0], ".")
		hostname = strings.TrimSuffix(hostname, ".local")
		if hostname != "" && hostname != ip {
			results[ip] = hostname
		}
	}

	return results
}

// FetchNetworkTable queries the Pi-hole API for the network table.
// Returns a map of MAC address (uppercase) → hostname.
// Returns an empty map if the API is unavailable or privacy settings block it.
func (c *Client) FetchNetworkTable() map[string]string {
	if c.APIToken == "" {
		return map[string]string{}
	}

	url := fmt.Sprintf("http://%s/admin/api.php?network&auth=%s", c.Address, c.APIToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return map[string]string{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]string{}
	}

	// Try parsing as array first (Pi-hole returns [] when no data)
	var devices []networkDevice
	if err := json.Unmarshal(body, &devices); err != nil {
		// Try as object with "network" key
		var nr networkResponse
		if err := json.Unmarshal(body, &nr); err != nil {
			return map[string]string{}
		}
		devices = nr.Network
	}

	results := make(map[string]string, len(devices))
	for _, d := range devices {
		if d.Name != "" && d.HWaddr != "" {
			mac := strings.ToUpper(d.HWaddr)
			results[mac] = d.Name
		}
	}

	return results
}
