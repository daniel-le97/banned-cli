package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/daniel-le97/banned-cli/db"
	"github.com/spf13/cobra"
)

// dbEditorCmd represents the db editor command
var dbEditorCmd = &cobra.Command{
	Use:   "editor",
	Short: "Open web-based SQLite editor",
	Long: `Open a web-based SQLite database editor in your default browser.

This command starts a local web server with an embedded SQLite editor interface
that allows you to:
- Browse database tables and data
- Execute SQL queries with real-time results
- View table schemas and relationships
- Export query results

The web interface will be available at http://localhost:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		fmt.Printf("🌐 Starting SQLite web editor...\n")
		dbPath, err := db.GetDatabasePath()
		if err != nil {
			fmt.Printf("❌ Could not get database path: %v\n", err)
			return
		}
		fmt.Printf("📊 Database: %s\n", dbPath)

		// Check if port is available and cleanup if necessary
		if err := ensurePortAvailable(port); err != nil {
			fmt.Printf("❌ Port preparation failed: %v\n", err)
			return
		}

		fmt.Printf("🔗 Open in browser: http://localhost:%d\n", port)
		fmt.Printf("Press Ctrl+C to stop the server\n\n")

		if err := startWebEditor(port); err != nil {
			fmt.Printf("❌ Failed to start web server: %v\n", err)
		}
	},
}

func init() {
	dbCmd.AddCommand(dbEditorCmd)
	dbEditorCmd.Flags().IntP("port", "p", 8080, "Port for the web server")
}

func startWebEditor(port int) error {
	// Get current working directory to find web files
	webDir := "./web"

	// Serve static files from web directory
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir(webDir))))

	// Serve main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
	})

	// API endpoints for database operations
	http.HandleFunc("/api/query", handleSQLQuery)
	http.HandleFunc("/api/tables", handleListTables)
	http.HandleFunc("/api/schema", handleGetSchema)
	http.HandleFunc("/api/data", handleGetTableData)
	http.HandleFunc("/api/stats", handleGetStats)

	// Health check endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		dbPath, _ := db.GetDatabasePath()
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"database":  dbPath,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	addr := ":" + strconv.Itoa(port)
	fmt.Printf("🚀 Web server starting on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// ensurePortAvailable checks if the port is available and cleans up any existing processes
func ensurePortAvailable(port int) error {
	addr := fmt.Sprintf("localhost:%d", port)

	// First, try to connect to see if something is already running
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		// Port is available
		fmt.Printf("✅ Port %d is available\n", port)
		return nil
	}
	conn.Close()

	// Port is in use, try to identify and clean up
	fmt.Printf("⚠️  Port %d is already in use, attempting cleanup...\n", port)

	// Try to find and kill existing banned db editor processes
	if err := killExistingEditorProcesses(port); err != nil {
		fmt.Printf("⚠️  Could not automatically clean up processes: %v\n", err)
		fmt.Printf("💡 You may need to manually stop the process using port %d\n", port)
	}

	// Wait a moment for cleanup
	time.Sleep(2 * time.Second)

	// Check again if port is now available
	conn, err = net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		// Port is now available
		fmt.Printf("✅ Port %d is now available after cleanup\n", port)
		return nil
	}
	conn.Close()

	// Port is still in use
	return fmt.Errorf("port %d is still in use after cleanup attempt. Please manually stop the process or use a different port with --port flag", port)
}

// killExistingEditorProcesses attempts to find and kill existing banned db editor processes
func killExistingEditorProcesses(port int) error {
	switch runtime.GOOS {
	case "linux", "darwin": // Linux and macOS
		return killUnixProcesses(port)
	case "windows":
		return killWindowsProcesses(port)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// killUnixProcesses kills processes on Unix-like systems (Linux, macOS)
func killUnixProcesses(port int) error {
	// First try to find processes by command pattern
	cmd := exec.Command("pgrep", "-f", "banned db editor")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		pids := strings.TrimSpace(string(output))
		if pids != "" {
			fmt.Printf("🔍 Found existing banned db editor processes: %s\n", pids)

			// Kill the processes
			pidList := strings.Split(pids, "\n")
			args := append([]string{"-TERM"}, pidList...)
			killCmd := exec.Command("kill", args...)
			if err := killCmd.Run(); err == nil {
				fmt.Printf("✅ Successfully terminated existing processes\n")
				return nil
			}
		}
	}

	// Fallback: try to find process by port using lsof or netstat
	cmd = exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port))
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		pid := strings.TrimSpace(string(output))
		if pid != "" {
			fmt.Printf("🔍 Found process %s using port %d\n", pid, port)

			killCmd := exec.Command("kill", "-TERM", pid)
			if err := killCmd.Run(); err == nil {
				fmt.Printf("✅ Successfully terminated process using port\n")
				return nil
			}
		}
	}

	// Try netstat as another fallback
	cmd = exec.Command("sh", "-c", fmt.Sprintf("netstat -tulpn 2>/dev/null | grep :%d | awk '{print $7}' | cut -d'/' -f1", port))
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		pid := strings.TrimSpace(string(output))
		if pid != "" && pid != "-" {
			fmt.Printf("🔍 Found process %s using port %d via netstat\n", pid, port)

			killCmd := exec.Command("kill", "-TERM", pid)
			if err := killCmd.Run(); err == nil {
				fmt.Printf("✅ Successfully terminated process using port\n")
				return nil
			}
		}
	}

	return fmt.Errorf("could not find or kill processes using port %d", port)
}

// killWindowsProcesses kills processes on Windows
func killWindowsProcesses(port int) error {
	// Try to find processes by command pattern first
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq banned.exe", "/FO", "CSV")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "banned.exe") {
				parts := strings.Split(line, ",")
				if len(parts) >= 2 {
					pid := strings.Trim(parts[1], `"`)
					fmt.Printf("🔍 Found banned.exe process with PID: %s\n", pid)

					killCmd := exec.Command("taskkill", "/PID", pid, "/F")
					if err := killCmd.Run(); err == nil {
						fmt.Printf("✅ Successfully terminated process %s\n", pid)
						return nil
					}
				}
			}
		}
	}

	// Fallback: try to find process by port using netstat
	cmd = exec.Command("netstat", "-ano")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, fmt.Sprintf(":%d ", port)) && strings.Contains(line, "LISTENING") {
				parts := strings.Fields(line)
				if len(parts) >= 5 {
					pid := parts[len(parts)-1]
					fmt.Printf("🔍 Found process %s using port %d\n", pid, port)

					killCmd := exec.Command("taskkill", "/PID", pid, "/F")
					if err := killCmd.Run(); err == nil {
						fmt.Printf("✅ Successfully terminated process using port\n")
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("could not find or kill processes using port %d", port)
}

// API Handlers using actual database operations

func handleListTables(w http.ResponseWriter, r *http.Request) {
	database, err := db.GetDB()
	if err != nil {
		http.Error(w, "Database connection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Query SQLite for all table names
	rows, err := database.Query(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		http.Error(w, "Failed to query tables: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tables":  tables,
	})
}

func handleGetSchema(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		http.Error(w, "Database connection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get table schema from SQLite
	rows, err := database.Query(`PRAGMA table_info(` + tableName + `)`)
	if err != nil {
		http.Error(w, "Failed to get table schema: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type column struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		NotNull bool   `json:"notNull"`
		Primary bool   `json:"primary"`
		Default string `json:"default"`
	}

	var columns []column
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk bool
		var dfltValue interface{}

		err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
		if err != nil {
			continue
		}

		defaultStr := ""
		if dfltValue != nil {
			defaultStr = fmt.Sprintf("%v", dfltValue)
		}

		columns = append(columns, column{
			Name:    name,
			Type:    dataType,
			NotNull: notNull,
			Primary: pk,
			Default: defaultStr,
		})
	}

	schema := map[string]interface{}{
		"table":   tableName,
		"columns": columns,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"schema":  schema,
	})
}

func handleGetTableData(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "100"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	database, err := db.GetDB()
	if err != nil {
		http.Error(w, "Database connection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Sanitize table name to prevent SQL injection
	if !isValidTableName(tableName) {
		http.Error(w, "Invalid table name", http.StatusBadRequest)
		return
	}

	// First check if table exists and get row count
	var tableExists bool
	var totalRows int
	err = database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&tableExists)
	if err != nil {
		http.Error(w, "Failed to check table existence: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get total row count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err = database.QueryRow(countQuery).Scan(&totalRows)
	if err != nil {
		http.Error(w, "Failed to count rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize data slice
	var data []map[string]interface{}
	var columns []string

	// If table has no data, we still want to return the structure
	if totalRows == 0 {
		// Get column information even for empty tables
		schemaRows, err := database.Query("PRAGMA table_info(" + tableName + ")")
		if err == nil {
			defer schemaRows.Close()
			for schemaRows.Next() {
				var cid int
				var name, dataType string
				var notNull, pk bool
				var dfltValue interface{}

				if err := schemaRows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err == nil {
					columns = append(columns, name)
				}
			}
		}

		// Return empty data with table info
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"table":     tableName,
			"data":      data, // empty array
			"count":     0,
			"totalRows": totalRows,
			"columns":   columns,
			"isEmpty":   true,
		})
		return
	}

	// Query data from the specified table
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableName, limit)
	rows, err := database.Query(query)
	if err != nil {
		http.Error(w, "Failed to query table data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Get column names
	columns, err = rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare data structure
	for rows.Next() {
		// Create a slice to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan values
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// Create row map
		row := make(map[string]interface{})
		for i, column := range columns {
			val := values[i]
			// Convert byte arrays to strings for JSON serialization
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[column] = val
		}
		data = append(data, row)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"table":     tableName,
		"data":      data,
		"count":     len(data),
		"totalRows": totalRows,
		"columns":   columns,
		"isEmpty":   false,
	})
}

func handleSQLQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		http.Error(w, "Query cannot be empty", http.StatusBadRequest)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   "Database connection failed: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Determine if this is a SELECT query or a modification query
	queryType := strings.ToUpper(strings.Fields(query)[0])

	var results []map[string]interface{}
	var message string
	var rowsAffected int64

	if queryType == "SELECT" || queryType == "PRAGMA" || queryType == "EXPLAIN" {
		// Execute SELECT query
		rows, err := database.Query(query)
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"query":   query,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"error":   "Failed to get columns: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		for rows.Next() {
			// Create a slice to hold column values
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			// Scan values
			if err := rows.Scan(valuePtrs...); err != nil {
				continue
			}

			// Create row map
			row := make(map[string]interface{})
			for i, column := range columns {
				val := values[i]
				// Convert byte arrays to strings for JSON serialization
				if b, ok := val.([]byte); ok {
					val = string(b)
				}
				row[column] = val
			}
			results = append(results, row)
		}

		message = fmt.Sprintf("Query executed successfully. Returned %d rows.", len(results))
	} else {
		// Execute modification query (INSERT, UPDATE, DELETE, etc.)
		result, err := database.Exec(query)
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"query":   query,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		rowsAffected, _ = result.RowsAffected()
		message = fmt.Sprintf("Query executed successfully. %d rows affected.", rowsAffected)
	}

	response := map[string]interface{}{
		"success":      true,
		"message":      message,
		"query":        query,
		"results":      results,
		"rowsAffected": rowsAffected,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetStats(w http.ResponseWriter, r *http.Request) {
	database, err := db.GetDB()
	if err != nil {
		http.Error(w, "Database connection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get database file size
	dbPath, err := db.GetDatabasePath()
	var fileSizeMB float64 = 0
	var fileSizeBytes int64 = 0
	if err == nil {
		if info, err := os.Stat(dbPath); err == nil {
			fileSizeBytes = info.Size()
			fileSizeMB = float64(fileSizeBytes) / (1024 * 1024)
		}
	}

	// Get table counts
	tableCounts := make(map[string]interface{})
	tables := []string{"channels", "videos", "downloads", "settings"}
	totalRows := 0

	for _, table := range tables {
		var count int
		err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			tableCounts[table] = "Error"
		} else {
			tableCounts[table] = count
			totalRows += count
		}
	}

	// Get SQLite version
	var sqliteVersion string
	err = database.QueryRow("SELECT sqlite_version()").Scan(&sqliteVersion)
	if err != nil {
		sqliteVersion = "Unknown"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"tables":       len(tables),
		"totalRows":    totalRows,
		"sizeMB":       fmt.Sprintf("%.2f", fileSizeMB),
		"sizeBytes":    fileSizeBytes,
		"version":      sqliteVersion,
		"tableCounts":  tableCounts,
		"databasePath": dbPath,
	})
}

// Helper functions

func isValidTableName(name string) bool {
	// Simple validation to prevent SQL injection
	validTables := []string{"channels", "videos", "downloads", "settings"}
	for _, validTable := range validTables {
		if name == validTable {
			return true
		}
	}
	return false
}
