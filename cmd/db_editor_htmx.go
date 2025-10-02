package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/daniel-le97/banned-cli/internal/db"
	"github.com/daniel-le97/banned-cli/internal/webhtmx"
	"github.com/spf13/cobra"
)

// htmxWebAssets gets the embedded HTMX web UI files from internal package
var htmxWebAssets = webhtmx.GetHTMXWebAssets()

// Global templates variable
var globalTemplates *template.Template

// parseTemplates loads all templates from the embedded filesystem
func parseTemplates() (*template.Template, error) {
	templates := template.New("")

	// List of template files to load
	templateFiles := []string{
		"web-htmx/templates/stats.html",
		"web-htmx/templates/tables.html",
		"web-htmx/templates/table-data.html",
		"web-htmx/templates/schema.html",
		"web-htmx/templates/tab-data-empty.html",
		"web-htmx/templates/tab-schema-empty.html",
	}

	for _, file := range templateFiles {
		content, err := htmxWebAssets.FS().ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", file, err)
		}

		// Extract template name from file path - keep the .html extension
		templateName := strings.TrimPrefix(file, "web-htmx/templates/")

		_, err = templates.New(templateName).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", file, err)
		}
	}

	return templates, nil
}

// dbEditorHtmxCmd represents the HTMX db editor command
var dbEditorHtmxCmd = &cobra.Command{
	Use:   "editor-htmx",
	Short: "Open HTMX-based web SQLite editor",
	Long: `Open an HTMX-based web SQLite database editor in your default browser.

This command starts a local web server with an HTMX-powered SQLite editor interface
that allows you to:
- Browse database tables and data with real-time search
- View table schemas and field information
- Explore database statistics and relationships
- Interactive UI with minimal JavaScript using HTMX

The web interface will be available at http://localhost:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		fmt.Printf("🌐 Starting HTMX SQLite web editor...\n")
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

		if err := startHtmxWebEditor(port); err != nil {
			fmt.Printf("❌ Failed to start web server: %v\n", err)
		}
	},
}

func init() {

}

func startHtmxWebEditor(port int) error {
	// Simple port check without complex cleanup
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err == nil {
		conn.Close()
		return fmt.Errorf("port %d is already in use. Please use a different port with --port flag", port)
	}

	// Create a sub-filesystem for the web directory from embedded files
	webSubFS, err := htmxWebAssets.SubFS("web-htmx")
	if err != nil {
		return fmt.Errorf("failed to create web sub-filesystem: %w", err)
	}

	// Parse templates from embedded filesystem
	templates, err := parseTemplates()
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Serve static files from embedded filesystem
	http.Handle("/web-htmx/", http.StripPrefix("/web-htmx/", http.FileServer(http.FS(webSubFS))))

	// Serve main page at /htmx and /htmx/
	http.HandleFunc("/htmx", func(w http.ResponseWriter, r *http.Request) {
		indexHTML, err := htmxWebAssets.FS().ReadFile("web-htmx/index.html")
		if err != nil {
			http.Error(w, "Failed to read index.html: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	})

	http.HandleFunc("/htmx/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/htmx/" {
			indexHTML, err := htmxWebAssets.FS().ReadFile("web-htmx/index.html")
			if err != nil {
				http.Error(w, "Failed to read index.html: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(indexHTML)
		}
	})

	// Set global templates variable
	globalTemplates = templates

	// HTMX-specific endpoints
	http.HandleFunc("/htmx/stats", handleHtmxStats)
	http.HandleFunc("/htmx/tables", handleHtmxTables)
	http.HandleFunc("/htmx/tab/data", handleHtmxTabData)
	http.HandleFunc("/htmx/tab/schema", handleHtmxTabSchema)
	http.HandleFunc("/htmx/table/data/", handleHtmxTableData)
	http.HandleFunc("/htmx/table/schema/", handleHtmxTableSchema)

	// Health check endpoints
	http.HandleFunc("/htmx/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<span class="indicator">●</span><span>Connected</span>`)
	})

	// JSON health check endpoint for compatibility
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		dbPath, _ := db.GetDatabasePath()
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"database":  dbPath,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	serverAddr := ":" + strconv.Itoa(port)
	fmt.Printf("🚀 HTMX Web server starting on %s\n", serverAddr)
	fmt.Printf("🔗 Open in browser: http://localhost:%d/htmx/\n", port)
	return http.ListenAndServe(serverAddr, nil)
}

// Helper function to check if request is from HTMX
func isHtmxRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// HTMX Stats Handler
func handleHtmxStats(w http.ResponseWriter, r *http.Request) {
	database, err := db.GetDB()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	// Get table counts
	type tableCount struct {
		Name  string
		Count int64
	}

	tables := []string{"channels", "videos", "downloads", "settings"}
	var stats []tableCount
	var totalRecords int64

	for _, table := range tables {
		var count int64
		err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			count = 0
		}
		stats = append(stats, tableCount{Name: table, Count: count})
		totalRecords += count
	}

	// Get database file size
	dbPath, _ := db.GetDatabasePath()
	var fileSize int64
	if stat, err := os.Stat(dbPath); err == nil {
		fileSize = stat.Size()
	}

	// Use parsed template

	data := struct {
		Stats        []tableCount
		TotalRecords int64
		FileSize     string
	}{
		Stats:        stats,
		TotalRecords: totalRecords,
		FileSize:     formatFileSizeHtmx(fileSize),
	}

	w.Header().Set("Content-Type", "text/html")
	globalTemplates.ExecuteTemplate(w, "stats.html", data)
}

// HTMX Tables Handler
func handleHtmxTables(w http.ResponseWriter, r *http.Request) {
	database, err := db.GetDB()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	// Get tables and their counts
	type table struct {
		Name  string
		Count int64
	}

	tables := []string{"channels", "videos", "downloads", "settings"}
	var tableList []table

	for _, tableName := range tables {
		var count int64
		err := database.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
		if err != nil {
			count = 0
		}
		tableList = append(tableList, table{Name: tableName, Count: count})
	}

	// Use parsed template
	w.Header().Set("Content-Type", "text/html")
	globalTemplates.ExecuteTemplate(w, "tables.html", tableList)
}

// HTMX Tab Data Handler
func handleHtmxTabData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	globalTemplates.ExecuteTemplate(w, "tab-data-empty.html", nil)
}

// HTMX Tab Schema Handler
func handleHtmxTabSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	globalTemplates.ExecuteTemplate(w, "tab-schema-empty.html", nil)
}

// HTMX Table Data Handler
func handleHtmxTableData(w http.ResponseWriter, r *http.Request) {
	tableName := strings.TrimPrefix(r.URL.Path, "/htmx/table/data/")
	if tableName == "" {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	// Get table data (limit for performance)
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 100", tableName)
	rows, err := database.Query(query)
	if err != nil {
		http.Error(w, "Failed to query table: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns", http.StatusInternalServerError)
		return
	}

	// Collect data
	var data []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, column := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[column] = val
		}
		data = append(data, row)
	}

	// Use parsed template

	templateData := struct {
		TableName string
		Columns   []string
		Data      []map[string]interface{}
	}{
		TableName: tableName,
		Columns:   columns,
		Data:      data,
	}

	w.Header().Set("Content-Type", "text/html")
	err = globalTemplates.ExecuteTemplate(w, "table-data.html", templateData)
	if err != nil {
		http.Error(w, "Template execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// HTMX Table Schema Handler
func handleHtmxTableSchema(w http.ResponseWriter, r *http.Request) {
	tableName := strings.TrimPrefix(r.URL.Path, "/htmx/table/schema/")
	if tableName == "" {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	// Get table schema
	rows, err := database.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		http.Error(w, "Failed to get table schema: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type column struct {
		Name    string
		Type    string
		NotNull bool
		Primary bool
		Default string
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

	// Use parsed template

	templateData := struct {
		TableName string
		Columns   []column
	}{
		TableName: tableName,
		Columns:   columns,
	}

	w.Header().Set("Content-Type", "text/html")
	globalTemplates.ExecuteTemplate(w, "schema.html", templateData)
}

// Helper function to format file size
func formatFileSizeHtmx(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
