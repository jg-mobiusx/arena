package pihole

import (
	"testing"
)

func TestResolveHostnames_LivePiHole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Pi-hole test in short mode")
	}

	c := &Client{Address: "172.16.44.11"}

	// Test with a few known IPs
	ips := []string{"172.16.44.11", "172.16.44.50", "172.16.44.200"}
	results := c.ResolveHostnames(ips)

	// We should get at least one result back from Pi-hole's DNS
	if len(results) == 0 {
		t.Log("Warning: no hostnames resolved - Pi-hole may not be reachable")
	}

	for ip, hostname := range results {
		t.Logf("  %s → %s", ip, hostname)
	}
}

func TestFetchNetworkTable_NoToken(t *testing.T) {
	c := &Client{Address: "172.16.44.11"}
	results := c.FetchNetworkTable()
	// Without a token, should return empty map gracefully
	if results == nil {
		t.Error("expected empty map, got nil")
	}
}
