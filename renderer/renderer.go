package renderer

import (
	"fmt"
	"io"
	"net"
	"sort"
	"strings"

	"github.com/johngillam/arena/model"
	"github.com/olekukonko/tablewriter"
)

// Render writes a formatted table of devices to the given writer.
func Render(w io.Writer, devices []model.Device) {
	// Sort by IP address
	sort.Slice(devices, func(i, j int) bool {
		ipA := net.ParseIP(devices[i].IP)
		ipB := net.ParseIP(devices[j].IP)
		if ipA == nil || ipB == nil {
			return devices[i].IP < devices[j].IP
		}
		return bytesLess(ipA, ipB)
	})

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"IP", "MAC", "Hostname", "Manufacturer", "Ports", "Status"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("│")
	table.SetRowSeparator("─")
	table.SetHeaderLine(true)
	table.SetBorder(true)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(false)

	online := 0
	offline := 0
	newCount := 0

	for _, d := range devices {
		status := "✓ Online"
		if !d.IsOnline {
			status = "✗ Offline"
			offline++
		} else {
			online++
		}
		if d.IsNew {
			status = "★ NEW"
			newCount++
		}

		ports := formatPorts(d.OpenPorts)
		hostname := d.Hostname
		if hostname == "" {
			hostname = "—"
		}
		manufacturer := d.Manufacturer
		if manufacturer == "" {
			manufacturer = "Unknown"
		}

		table.Append([]string{
			d.IP,
			d.MAC,
			hostname,
			manufacturer,
			ports,
			status,
		})
	}

	table.Render()

	// Summary line
	fmt.Fprintf(w, "\n  %d devices total  │  %d online  │  %d offline  │  %d new\n\n",
		len(devices), online, offline, newCount)
}

// formatPorts returns a compact string listing open ports.
func formatPorts(ports []model.Port) string {
	if len(ports) == 0 {
		return "—"
	}
	var parts []string
	for _, p := range ports {
		if p.Service != "" {
			parts = append(parts, fmt.Sprintf("%d/%s", p.Number, p.Service))
		} else {
			parts = append(parts, fmt.Sprintf("%d/%s", p.Number, p.Protocol))
		}
	}
	return strings.Join(parts, ", ")
}

func bytesLess(a, b net.IP) bool {
	a = a.To16()
	b = b.To16()
	for i := range a {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return false
}
