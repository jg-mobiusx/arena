package scanner

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"time"

	"github.com/johngillam/arena/model"
)

// nmapRun is the top-level XML element from nmap -oX output.
type nmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []nmapHost `xml:"host"`
}

type nmapHost struct {
	Status    nmapStatus    `xml:"status"`
	Addresses []nmapAddress `xml:"address"`
	Hostnames nmapHostnames `xml:"hostnames"`
	Ports     nmapPorts     `xml:"ports"`
}

type nmapStatus struct {
	State string `xml:"state,attr"`
}

type nmapAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
	Vendor   string `xml:"vendor,attr"`
}

type nmapHostnames struct {
	Names []nmapHostname `xml:"hostname"`
}

type nmapHostname struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type nmapPorts struct {
	Ports []nmapPort `xml:"port"`
}

type nmapPort struct {
	Protocol string      `xml:"protocol,attr"`
	PortID   int         `xml:"portid,attr"`
	Service  nmapService `xml:"service"`
	State    nmapState   `xml:"state"`
}

type nmapService struct {
	Name string `xml:"name,attr"`
}

type nmapState struct {
	State string `xml:"state,attr"`
}

// ParseNmapXML parses raw Nmap XML output into a slice of Devices.
func ParseNmapXML(data []byte) ([]model.Device, error) {
	var run nmapRun
	if err := xml.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("parsing nmap XML: %w", err)
	}

	now := time.Now()
	var devices []model.Device

	for _, host := range run.Hosts {
		if host.Status.State != "up" {
			continue
		}

		d := model.Device{
			IsOnline: true,
			LastSeen: now,
		}

		for _, addr := range host.Addresses {
			switch addr.AddrType {
			case "ipv4":
				d.IP = addr.Addr
			case "mac":
				d.MAC = addr.Addr
				if addr.Vendor != "" {
					d.Manufacturer = addr.Vendor
				}
			}
		}

		if len(host.Hostnames.Names) > 0 {
			d.Hostname = host.Hostnames.Names[0].Name
		}

		for _, p := range host.Ports.Ports {
			if p.State.State == "open" {
				d.OpenPorts = append(d.OpenPorts, model.Port{
					Number:   p.PortID,
					Protocol: p.Protocol,
					Service:  p.Service.Name,
				})
			}
		}

		devices = append(devices, d)
	}

	return devices, nil
}

// PingSweep runs nmap -sn against the given subnet and returns discovered devices.
func PingSweep(subnet string) ([]model.Device, error) {
	out, err := exec.Command("nmap", "-sn", "-oX", "-", subnet).Output()
	if err != nil {
		return nil, fmt.Errorf("nmap ping sweep: %w", err)
	}
	return ParseNmapXML(out)
}

// ServiceScan runs a targeted nmap service/OS detection scan against a single IP.
func ServiceScan(ip string) ([]model.Port, string, error) {
	out, err := exec.Command("nmap", "-sV", "-O", "--osscan-guess", "-oX", "-", ip).Output()
	if err != nil {
		return nil, "", fmt.Errorf("nmap service scan: %w", err)
	}

	var run nmapRun
	if err := xml.Unmarshal(out, &run); err != nil {
		return nil, "", fmt.Errorf("parsing service scan XML: %w", err)
	}

	var ports []model.Port
	var osGuess string

	for _, host := range run.Hosts {
		for _, p := range host.Ports.Ports {
			if p.State.State == "open" {
				ports = append(ports, model.Port{
					Number:   p.PortID,
					Protocol: p.Protocol,
					Service:  p.Service.Name,
				})
			}
		}
	}

	return ports, osGuess, nil
}
