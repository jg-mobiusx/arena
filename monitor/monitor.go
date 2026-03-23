package monitor

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// Status represents the health metrics of the daemon and internet connection
type Status struct {
	LastScan     time.Time `json:"last_scan"`
	PingMs       int       `json:"ping_ms"`
	DownloadMbps float64   `json:"download_mbps"`
}

var (
	currentStatus Status
	mu            sync.RWMutex
)

// UpdateScanTime sets the timestamp of the last completed network scan
func UpdateScanTime(t time.Time) {
	mu.Lock()
	defer mu.Unlock()
	currentStatus.LastScan = t
}

// StartBackgroundJobs begins the continuous ping and speed test monitors
func StartBackgroundJobs() {
	go pingLoop()
	// Optionally hook in a fully-fledged speedtest binary here in the future
}

// pingLoop continuously measures basic internet latency
func pingLoop() {
	// A simple HTTP latency check to a massive CDN (Cloudflare)
	client := http.Client{Timeout: 5 * time.Second}
	ticker := time.NewTicker(60 * time.Second)

	// Do an immediate initial ping
	doPing(client)

	for {
		<-ticker.C
		doPing(client)
	}
}

func doPing(client http.Client) {
	start := time.Now()
	resp, err := client.Get("http://1.1.1.1")
	
	mu.Lock()
	defer mu.Unlock()

	if err == nil {
		resp.Body.Close()
		latency := time.Since(start).Milliseconds()
		currentStatus.PingMs = int(latency)
	} else {
		log.Printf("⚠️  Internet Ping Failed: %v\n", err)
		currentStatus.PingMs = -1 // Indicates Offline / Error
	}
}

// ServeStatus provides the JSON HTTP endpoint for the React dashboard
func ServeStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	mu.RLock()
	data, err := json.Marshal(currentStatus)
	mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
