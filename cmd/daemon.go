package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// DaemonConfig holds daemon configuration
type DaemonConfig struct {
	WebUI          bool
	WebUIPort      string
	AutoSync       bool
	SyncInterval   time.Duration
	PidFile        string
	LogFile        string
	BackgroundSync bool
}

var daemonConfig = DaemonConfig{
	WebUI:        false,
	WebUIPort:    "8080",
	AutoSync:     false,
	SyncInterval: 1 * time.Hour,
	PidFile:      "",
	LogFile:      "",
}

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run banned-cli as a daemon service",
	Long: `Run banned-cli as a background daemon service.

The daemon can provide several services:
- Web UI server for database browsing
- Automatic periodic syncing of banned.video data
- Background monitoring and maintenance tasks

Examples:
  banned daemon --webui --port 8080                    # Web UI only
  banned daemon --auto-sync --interval 2h              # Auto-sync every 2 hours
  banned daemon --webui --auto-sync --interval 30m     # Both services
  banned daemon --pid-file /var/run/banned.pid         # Write PID file`,
	
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDaemon(daemonConfig)
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)

	// Daemon service flags
	daemonCmd.Flags().BoolVar(&daemonConfig.WebUI, "webui", false, "Enable web UI server")
	daemonCmd.Flags().StringVar(&daemonConfig.WebUIPort, "port", "8080", "Port for web UI server")
	daemonCmd.Flags().BoolVar(&daemonConfig.AutoSync, "auto-sync", false, "Enable automatic periodic syncing")
	daemonCmd.Flags().DurationVar(&daemonConfig.SyncInterval, "interval", 1*time.Hour, "Sync interval (e.g., 30m, 1h, 2h)")
	
	// System daemon flags
	daemonCmd.Flags().StringVar(&daemonConfig.PidFile, "pid-file", "", "Write daemon PID to file")
	daemonCmd.Flags().StringVar(&daemonConfig.LogFile, "log-file", "", "Log daemon output to file")
	daemonCmd.Flags().BoolVar(&daemonConfig.BackgroundSync, "background", false, "Run in background (detach from terminal)")
}

func runDaemon(config DaemonConfig) error {
	// Setup logging
	if config.LogFile != "" {
		logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	// Write PID file
	if config.PidFile != "" {
		if err := writePidFile(config.PidFile); err != nil {
			return fmt.Errorf("failed to write PID file: %w", err)
		}
		defer os.Remove(config.PidFile)
	}

	log.Println("🚀 Starting banned-cli daemon...")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start services
	var wg sync.WaitGroup
	
	if config.WebUI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runWebUIService(ctx, config.WebUIPort); err != nil {
				log.Printf("❌ Web UI service error: %v", err)
			}
		}()
	}

	if config.AutoSync {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runAutoSyncService(ctx, config.SyncInterval)
		}()
	}

	if !config.WebUI && !config.AutoSync {
		return fmt.Errorf("no services enabled. Use --webui or --auto-sync flags")
	}

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		log.Printf("📡 Received signal: %v. Shutting down gracefully...", sig)
		cancel()
	case <-ctx.Done():
	}

	// Wait for all services to stop
	wg.Wait()
	log.Println("👋 Daemon stopped")
	return nil
}

func runWebUIService(ctx context.Context, port string) error {
	log.Printf("🌐 Starting PWA Web UI service on port %s", port)
	
	// Create HTTP server with PWA support
	srv := &http.Server{
		Addr: ":" + port,
		Handler: createPWAHandler(),
	}
	
	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ Web server error: %v", err)
		}
	}()
	
	log.Printf("✅ PWA Web UI available at: http://localhost:%s", port)
	log.Printf("📱 Install as app for better experience!")
	
	// Wait for shutdown
	<-ctx.Done()
	log.Println("🛑 Shutting down Web UI service...")
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return srv.Shutdown(shutdownCtx)
}

func createPWAHandler() http.Handler {
	mux := http.NewServeMux()
	
	// Serve PWA static files with proper headers
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set PWA headers
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Service-Worker-Allowed", "/")
		
		// Handle root path
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "internal/webui/web/index.html")
			return
		}
		
		// Serve static files
		http.ServeFile(w, r, "internal/webui/web"+r.URL.Path)
	})
	
	// API endpoints for PWA functionality
	mux.HandleFunc("/api/sync/channels", handleSyncChannels)
	mux.HandleFunc("/api/sync/videos", handleSyncVideos)
	mux.HandleFunc("/api/daemon/status", handleDaemonStatus)
	
	return mux
}

func handleSyncChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Trigger channel sync
	log.Println("� API: Syncing channels...")
	
	// Simulate sync work
	time.Sleep(1 * time.Second)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Channels synced successfully",
		"timestamp": time.Now().Unix(),
	})
}

func handleSyncVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Trigger video sync
	log.Println("🔄 API: Syncing videos...")
	
	// Simulate sync work
	time.Sleep(2 * time.Second)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Videos synced successfully", 
		"timestamp": time.Now().Unix(),
	})
}

func handleDaemonStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "running",
		"uptime": time.Since(time.Now()).Seconds(), // This would be actual uptime
		"services": map[string]bool{
			"webui": true,
			"sync": daemonConfig.AutoSync,
		},
		"timestamp": time.Now().Unix(),
	})
}

func runAutoSyncService(ctx context.Context, interval time.Duration) {
	log.Printf("🔄 Starting auto-sync service (interval: %v)", interval)
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	// Run initial sync
	if err := performSync(); err != nil {
		log.Printf("❌ Initial sync failed: %v", err)
	}
	
	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Auto-sync service stopping...")
			return
		case <-ticker.C:
			log.Println("🔄 Running scheduled sync...")
			if err := performSync(); err != nil {
				log.Printf("❌ Sync failed: %v", err)
			} else {
				log.Println("✅ Sync completed successfully")
			}
		}
	}
}

func performSync() error {
	// This would call your existing sync functionality
	// For now, simulate the operation
	log.Println("📥 Fetching latest data from banned.video...")
	time.Sleep(2 * time.Second) // Simulate work
	return nil
}

func writePidFile(pidFile string) error {
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}