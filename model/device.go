package model

import "time"

// Device represents a single network device discovered on the LAN.
type Device struct {
	IP           string    `json:"ip"`
	MAC          string    `json:"mac"`
	Hostname     string    `json:"hostname"`
	Manufacturer string    `json:"manufacturer"`
	OS           string    `json:"os"`
	OpenPorts    []Port    `json:"open_ports,omitempty"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	IsNew        bool      `json:"is_new"`
	IsOnline     bool      `json:"is_online"`
}

// Port represents an open network port on a device.
type Port struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
}
