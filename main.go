package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johngillam/arena/enrichment"
	"github.com/johngillam/arena/model"
	"github.com/johngillam/arena/pihole"
	"github.com/johngillam/arena/renderer"
	"github.com/johngillam/arena/scanner"
	"github.com/johngillam/arena/store"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	Subnets      []string     `yaml:"subnets"`
	DataFile     string       `yaml:"data_file"`
	PiHole       PiHoleConfig `yaml:"pihole"`
	Daemon       bool         `yaml:"daemon"`
	ScanInterval string       `yaml:"scan_interval"`
	HTTPPort     int          `yaml:"http_port"`
	WebDir       string       `yaml:"web_dir"`
}

// PiHoleConfig holds Pi-hole connection settings.
type PiHoleConfig struct {
	Address  string `yaml:"address"`
	APIToken string `yaml:"api_token"`
}

func main() {
	// Load config
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Apply defaults
	if cfg.DataFile == "" {
		cfg.DataFile = "devices.json"
	}
	if cfg.ScanInterval == "" {
		cfg.ScanInterval = "5m"
	}
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = 8080
	}
	if cfg.WebDir == "" {
		cfg.WebDir = "web/dist"
	}

	if len(cfg.Subnets) == 0 {
		log.Fatal("No subnets configured. Edit config.yaml to add your subnet(s).")
	}

	fmt.Printf("🏟  Arena — Home Network Monitor\n")
	fmt.Printf("   Scanning %d subnet(s): %v\n", len(cfg.Subnets), cfg.Subnets)

	// If daemon mode is enabled, start the HTTP server and the scan loop
	if cfg.Daemon {
		// Parse interval
		interval, err := time.ParseDuration(cfg.ScanInterval)
		if err != nil {
			log.Fatalf("Invalid scan_interval '%s': %v", cfg.ScanInterval, err)
		}

		// Perform initial scan immediately
		fmt.Println("\n▶️  Performing initial network scan...")
		runScanWorkflow(cfg)

		// Create cancellable context for elegant shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle OS interrupts
		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			<-sigCh
			fmt.Println("\n🛑 Shutting down Arena daemon...")
			cancel()
			os.Exit(0)
		}()

		// Start HTTP Server
		go startHTTPServer(cfg)

		// Start Ticker Loop
		fmt.Printf("⏱  Starting continuous scan loop every %v\n", interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("\n▶️  [Time: %s] Triggering scheduled network scan...\n", time.Now().Format(time.RFC3339))
				runScanWorkflow(cfg)
			case <-ctx.Done():
				return
			}
		}
	} else {
		// One-shot CLI run
		runScanWorkflow(cfg)
		// Render to CLI only in non-daemon mode
		known, _ := store.Load(cfg.DataFile)
		fmt.Println()
		renderer.Render(os.Stdout, known)
	}
}

// startHTTPServer spins up a basic web server to serve the React frontend and JSON APIs
func startHTTPServer(cfg *Config) {
	mux := http.NewServeMux()

	// 1. Serve devices.json live from disk
	mux.HandleFunc("/api/devices.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Allow CORS for local dev
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, cfg.DataFile)
	})

	// 2. Serve vlans.json (assume it's next to devices.json or config)
	vlansPath := "vlans.json"
	mux.HandleFunc("/api/vlans.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, vlansPath)
	})

	// 3. Serve the built React static files
	fs := http.FileServer(http.Dir(cfg.WebDir))
	mux.Handle("/", fs)

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	fmt.Printf("🌐 HTTP Server running on http://localhost%s (serving %s)\n", addr, cfg.WebDir)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP server crash: %v", err)
	}
}

// runScanWorkflow encapsulates the 6 scanning and enrichment phases
func runScanWorkflow(cfg *Config) {
	var scanned []model.Device

	// Phase 1: Scan all subnets
	for _, subnet := range cfg.Subnets {
		fmt.Printf("   🔍 Scanning %s ...\n", subnet)
		devices, err := scanner.PingSweep(subnet)
		if err != nil {
			log.Printf("   ⚠️  Error scanning %s: %v", subnet, err)
			continue
		}
		fmt.Printf("   ✓  Found %d active devices on %s\n", len(devices), subnet)
		scanned = append(scanned, devices...)
	}

	if len(scanned) == 0 {
		log.Println("   ⚠️  No devices found. Nmap requires root privileges (sudo) for ARP sweeps.")
		return
	}

	// Phase 2: Resolve hostnames via Pi-hole
	if cfg.PiHole.Address != "" {
		fmt.Printf("   🌐 Resolving hostnames via Pi-hole (%s)... ", cfg.PiHole.Address)
		ph := &pihole.Client{
			Address:  cfg.PiHole.Address,
			APIToken: cfg.PiHole.APIToken,
		}

		ips := make([]string, 0, len(scanned))
		for _, d := range scanned {
			ips = append(ips, d.IP)
		}

		hostnames := ph.ResolveHostnames(ips)
		resolved := 0
		for i := range scanned {
			if name, ok := hostnames[scanned[i].IP]; ok {
				scanned[i].Hostname = name
				resolved++
			}
		}

		networkNames := ph.FetchNetworkTable()
		for i := range scanned {
			if scanned[i].Hostname == "" && scanned[i].MAC != "" {
				if name, ok := networkNames[scanned[i].MAC]; ok {
					scanned[i].Hostname = name
					resolved++
				}
			}
		}
		fmt.Printf("Resolved %d hostnames.\n", resolved)
	}

	// Phase 3: Load known devices
	known, err := store.Load(cfg.DataFile)
	if err != nil {
		log.Printf("   ⚠️  Error loading device history: %v", err)
	}

	// Phase 4: Enrich
	devices := enrichment.Enrich(scanned, known)

	// Phase 5: Save
	if err := store.Save(cfg.DataFile, devices); err != nil {
		log.Printf("   ⚠️  Error saving device history: %v", err)
	} else {
		fmt.Printf("   💾 Saved %d devices to %s\n", len(devices), cfg.DataFile)
	}
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}
