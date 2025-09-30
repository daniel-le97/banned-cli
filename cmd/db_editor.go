package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// dbEditorCmd represents the db editor command
var dbEditorCmd = &cobra.Command{
	Use:   "editor",
	Short: "Open web-based SQLite editor",
	Long: `Open a web-based SQLite database editor in your default browser.

This command starts a local web server with an embedded SQLite editor interface
that allows you to:
- Browse database tables and buckets
- Execute SQL queries
- View and edit data
- Export query results

The web interface will be available at http://localhost:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")

		fmt.Printf("🌐 Starting SQLite web editor...\n")
		fmt.Printf("📊 Database: %s\n", getDBPath())
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
	// Parse templates
	tmpl := template.Must(template.New("editor").Parse(editorHTML))

	// Serve main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Title    string
			Database string
			Port     int
		}{
			Title:    "Banned CLI - Database Editor",
			Database: getDBPath(),
			Port:     port,
		}

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// API endpoints for database operations
	http.HandleFunc("/api/query", handleSQLQuery)
	http.HandleFunc("/api/tables", handleListTables)
	http.HandleFunc("/api/schema", handleGetSchema)
	http.HandleFunc("/api/data", handleGetTableData)

	// Health check endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"database":  getDBPath(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	addr := ":" + strconv.Itoa(port)
	fmt.Printf("🚀 Web server starting on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// API Handlers
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

	// TODO: Execute query against database
	// This will need to be adapted based on whether you're using bbolt or SQLite

	response := map[string]interface{}{
		"success": true,
		"message": "Query executed successfully",
		"query":   req.Query,
		"results": []map[string]interface{}{
			{"note": "Query execution not yet implemented"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListTables(w http.ResponseWriter, r *http.Request) {
	// TODO: List all tables/buckets in the database

	tables := []string{"channels", "videos", "downloads", "settings"}

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

	// TODO: Get schema for specific table/bucket

	schema := map[string]interface{}{
		"table": tableName,
		"columns": []map[string]string{
			{"name": "id", "type": "TEXT", "primary": "true"},
			{"name": "created_at", "type": "DATETIME", "primary": "false"},
		},
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

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "100"
	}

	// TODO: Get data from specific table/bucket

	data := []map[string]interface{}{
		{"id": "example1", "name": "Sample Data", "created_at": time.Now()},
		{"id": "example2", "name": "More Data", "created_at": time.Now()},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"table":   tableName,
		"data":    data,
		"count":   len(data),
	})
}

func getDBPath() string {
	// TODO: Return actual database path
	return "~/.config/banned/banned.db"
}

// HTML template for the database editor
const editorHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: #1a1a1a;
            color: #e0e0e0;
            height: 100vh;
            overflow: hidden;
        }
        
        header {
            background: #2d2d2d;
            padding: 1rem 2rem;
            border-bottom: 1px solid #404040;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        h1 {
            color: #4fc3f7;
            font-size: 1.5rem;
        }
        
        .status {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .indicator {
            color: #4caf50;
            font-size: 1.2rem;
        }
        
        .main-container {
            display: flex;
            height: calc(100vh - 80px);
        }
        
        .sidebar {
            width: 300px;
            background: #2a2a2a;
            border-right: 1px solid #404040;
            padding: 1rem;
            overflow-y: auto;
        }
        
        .sidebar h3 {
            color: #81c784;
            margin-bottom: 1rem;
            font-size: 1.1rem;
        }
        
        .tables-list {
            list-style: none;
            margin-bottom: 2rem;
        }
        
        .table-item {
            padding: 0.75rem;
            margin: 0.25rem 0;
            background: #333;
            border-radius: 4px;
            cursor: pointer;
            transition: background 0.2s;
            border-left: 3px solid #4fc3f7;
        }
        
        .table-item:hover {
            background: #404040;
        }
        
        .table-item.active {
            background: #1976d2;
        }
        
        .quick-actions h4 {
            color: #ffb74d;
            margin-bottom: 0.5rem;
        }
        
        button {
            background: #1976d2;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
            margin: 0.25rem 0.25rem 0.25rem 0;
            font-size: 0.9rem;
            transition: background 0.2s;
        }
        
        button:hover {
            background: #1565c0;
        }
        
        .execute-btn {
            background: #4caf50;
            padding: 0.75rem 1.5rem;
            font-weight: bold;
        }
        
        .execute-btn:hover {
            background: #45a049;
        }
        
        .content {
            flex: 1;
            display: flex;
            flex-direction: column;
        }
        
        .tabs {
            display: flex;
            background: #2a2a2a;
            border-bottom: 1px solid #404040;
        }
        
        .tab {
            background: transparent;
            border: none;
            padding: 1rem 1.5rem;
            color: #b0b0b0;
            cursor: pointer;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        }
        
        .tab:hover {
            color: #e0e0e0;
            background: #333;
        }
        
        .tab.active {
            color: #4fc3f7;
            border-bottom-color: #4fc3f7;
            background: #1a1a1a;
        }
        
        .tab-content {
            display: none;
            flex: 1;
            padding: 1rem;
            overflow: auto;
        }
        
        .tab-content.active {
            display: block;
        }
        
        .table-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1rem;
        }
        
        .table-controls {
            display: flex;
            gap: 1rem;
            align-items: center;
        }
        
        select {
            background: #333;
            color: #e0e0e0;
            border: 1px solid #555;
            padding: 0.5rem;
            border-radius: 4px;
        }
        
        #query-input {
            width: 100%;
            height: 200px;
            background: #2a2a2a;
            color: #e0e0e0;
            border: 1px solid #555;
            border-radius: 4px;
            padding: 1rem;
            font-family: 'Courier New', monospace;
            font-size: 14px;
            resize: vertical;
            margin-bottom: 1rem;
        }
        
        .data-table {
            width: 100%;
            border-collapse: collapse;
            background: #2a2a2a;
            border-radius: 4px;
            overflow: hidden;
        }
        
        .data-table th {
            background: #1976d2;
            color: white;
            padding: 0.75rem;
            text-align: left;
            font-weight: bold;
        }
        
        .data-table td {
            padding: 0.75rem;
            border-bottom: 1px solid #404040;
        }
        
        .data-table tr:nth-child(even) {
            background: #333;
        }
        
        .data-table tr:hover {
            background: #404040;
        }
        
        .placeholder {
            text-align: center;
            color: #777;
            padding: 3rem;
            font-style: italic;
        }
        
        .query-results {
            background: #2a2a2a;
            border-radius: 4px;
            padding: 1rem;
            min-height: 300px;
        }
        
        .error {
            color: #f44336;
            background: #ffebee;
            padding: 1rem;
            border-radius: 4px;
            border-left: 4px solid #f44336;
        }
        
        .success {
            color: #4caf50;
            background: #e8f5e8;
            padding: 1rem;
            border-radius: 4px;
            border-left: 4px solid #4caf50;
        }
        
        .loading {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid #333;
            border-radius: 50%;
            border-top-color: #4fc3f7;
            animation: spin 1s ease-in-out infinite;
        }
        
        @keyframes spin {
            to { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div id="app">
        <header>
            <h1>{{.Title}}</h1>
            <div class="status" id="status">
                <span class="indicator" id="indicator">●</span>
                <span id="status-text">Connected to {{.Database}}</span>
            </div>
        </header>

        <div class="main-container">
            <!-- Sidebar with tables -->
            <div class="sidebar">
                <h3>📊 Tables</h3>
                <ul id="tables-list" class="tables-list">
                    <!-- Tables will be loaded dynamically -->
                </ul>
                
                <div class="quick-actions">
                    <h4>⚡ Quick Actions</h4>
                    <button onclick="refreshTables()">🔄 Refresh</button>
                    <button onclick="showTab('query')">📝 Query</button>
                </div>
            </div>

            <!-- Main content area -->
            <div class="content">
                <div class="tabs">
                    <button class="tab active" onclick="showTab('browser')">📋 Browse Data</button>
                    <button class="tab" onclick="showTab('query')">💻 SQL Query</button>
                    <button class="tab" onclick="showTab('schema')">🏗️ Schema</button>
                </div>

                <!-- Data Browser Tab -->
                <div id="browser-tab" class="tab-content active">
                    <div class="table-header">
                        <h3 id="current-table">Select a table to view data</h3>
                        <div class="table-controls">
                            <button onclick="exportData()">📤 Export</button>
                            <select id="limit-select" onchange="loadTableData()">
                                <option value="50">50 rows</option>
                                <option value="100" selected>100 rows</option>
                                <option value="500">500 rows</option>
                                <option value="1000">1000 rows</option>
                            </select>
                        </div>
                    </div>
                    <div id="data-container">
                        <div class="placeholder">
                            👈 Select a table from the sidebar to view its data
                        </div>
                    </div>
                </div>

                <!-- SQL Query Tab -->
                <div id="query-tab" class="tab-content">
                    <div class="query-header">
                        <h3>💻 SQL Query Editor</h3>
                        <button onclick="executeQuery()" class="execute-btn">▶️ Execute Query</button>
                    </div>
                    <textarea id="query-input" placeholder="Enter your SQL query here...

Example queries:
SELECT * FROM channels LIMIT 10;
SELECT COUNT(*) FROM videos;
SELECT title, file_size FROM videos WHERE file_size > 0;"></textarea>
                    <div id="query-results" class="query-results">
                        <div class="placeholder">
                            ✨ Query results will appear here
                        </div>
                    </div>
                </div>

                <!-- Schema Tab -->
                <div id="schema-tab" class="tab-content">
                    <div class="table-header">
                        <h3>🏗️ Database Schema</h3>
                        <button onclick="loadAllSchemas()">🔄 Refresh</button>
                    </div>
                    <div id="schema-container">
                        <div class="placeholder">
                            📋 Database schema information will appear here
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        let currentTable = '';
        
        // Initialize the app
        document.addEventListener('DOMContentLoaded', function() {
            checkHealth();
            loadTables();
        });
        
        // Health check
        async function checkHealth() {
            try {
                const response = await fetch('/api/health');
                const data = await response.json();
                document.getElementById('status-text').textContent = 'Connected to ' + data.database;
                document.getElementById('indicator').style.color = '#4caf50';
            } catch (error) {
                document.getElementById('status-text').textContent = 'Connection error';
                document.getElementById('indicator').style.color = '#f44336';
            }
        }
        
        // Load tables list
        async function loadTables() {
            try {
                const response = await fetch('/api/tables');
                const data = await response.json();
                
                const tablesList = document.getElementById('tables-list');
                tablesList.innerHTML = '';
                
                data.tables.forEach(table => {
                    const li = document.createElement('li');
                    li.className = 'table-item';
                    li.textContent = table;
                    li.onclick = () => selectTable(table);
                    tablesList.appendChild(li);
                });
            } catch (error) {
                console.error('Failed to load tables:', error);
            }
        }
        
        // Select a table
        function selectTable(tableName) {
            currentTable = tableName;
            document.getElementById('current-table').textContent = tableName;
            
            // Update active state
            document.querySelectorAll('.table-item').forEach(item => {
                item.classList.remove('active');
            });
            event.target.classList.add('active');
            
            // Switch to browser tab and load data
            showTab('browser');
            loadTableData();
        }
        
        // Load table data
        async function loadTableData() {
            if (!currentTable) return;
            
            const limit = document.getElementById('limit-select').value;
            const container = document.getElementById('data-container');
            
            try {
                container.innerHTML = '<div class="loading"></div> Loading data...';
                
                const response = await fetch('/api/data?table=' + currentTable + '&limit=' + limit);
                const data = await response.json();
                
                if (data.success && data.data.length > 0) {
                    container.innerHTML = createDataTable(data.data);
                } else {
                    container.innerHTML = '<div class="placeholder">No data found in this table</div>';
                }
            } catch (error) {
                container.innerHTML = '<div class="error">Error loading data: ' + error.message + '</div>';
            }
        }
        
        // Create data table HTML
        function createDataTable(data) {
            if (!data || data.length === 0) return '<div class="placeholder">No data</div>';
            
            const headers = Object.keys(data[0]);
            let html = '<table class="data-table"><thead><tr>';
            
            headers.forEach(header => {
                html += '<th>' + escapeHtml(header) + '</th>';
            });
            
            html += '</tr></thead><tbody>';
            
            data.forEach(row => {
                html += '<tr>';
                headers.forEach(header => {
                    const value = row[header];
                    html += '<td>' + escapeHtml(String(value)) + '</td>';
                });
                html += '</tr>';
            });
            
            html += '</tbody></table>';
            return html;
        }
        
        // Execute SQL query
        async function executeQuery() {
            const query = document.getElementById('query-input').value.trim();
            if (!query) return;
            
            const resultsContainer = document.getElementById('query-results');
            
            try {
                resultsContainer.innerHTML = '<div class="loading"></div> Executing query...';
                
                const response = await fetch('/api/query', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ query: query })
                });
                
                const data = await response.json();
                
                if (data.success) {
                    if (data.results && data.results.length > 0) {
                        resultsContainer.innerHTML = '<div class="success">Query executed successfully</div>' + 
                                                   createDataTable(data.results);
                    } else {
                        resultsContainer.innerHTML = '<div class="success">Query executed successfully (no results)</div>';
                    }
                } else {
                    resultsContainer.innerHTML = '<div class="error">Error: ' + (data.error || 'Unknown error') + '</div>';
                }
            } catch (error) {
                resultsContainer.innerHTML = '<div class="error">Error executing query: ' + error.message + '</div>';
            }
        }
        
        // Show tab
        function showTab(tabName) {
            // Update tab buttons
            document.querySelectorAll('.tab').forEach(tab => {
                tab.classList.remove('active');
            });
            event.target.classList.add('active');
            
            // Update tab content
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.remove('active');
            });
            document.getElementById(tabName + '-tab').classList.add('active');
        }
        
        // Refresh tables
        function refreshTables() {
            loadTables();
        }
        
        // Export data (placeholder)
        function exportData() {
            alert('Export functionality will be implemented');
        }
        
        // Load all schemas
        async function loadAllSchemas() {
            const container = document.getElementById('schema-container');
            container.innerHTML = '<div class="loading"></div> Loading schemas...';
            
            try {
                const tablesResponse = await fetch('/api/tables');
                const tablesData = await tablesResponse.json();
                
                let html = '';
                
                for (const table of tablesData.tables) {
                    const schemaResponse = await fetch('/api/schema?table=' + table);
                    const schemaData = await schemaResponse.json();
                    
                    html += '<div style="margin-bottom: 2rem;">';
                    html += '<h4 style="color: #4fc3f7; margin-bottom: 0.5rem;">📋 ' + table + '</h4>';
                    
                    if (schemaData.success && schemaData.schema.columns) {
                        html += '<table class="data-table" style="max-width: 600px;"><thead><tr>';
                        html += '<th>Column</th><th>Type</th><th>Primary</th></tr></thead><tbody>';
                        
                        schemaData.schema.columns.forEach(col => {
                            html += '<tr><td>' + col.name + '</td><td>' + col.type + '</td><td>' + 
                                   (col.primary === 'true' ? '✓' : '') + '</td></tr>';
                        });
                        
                        html += '</tbody></table>';
                    }
                    
                    html += '</div>';
                }
                
                container.innerHTML = html;
            } catch (error) {
                container.innerHTML = '<div class="error">Error loading schemas: ' + error.message + '</div>';
            }
        }
        
        // Utility function to escape HTML
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        // Handle Enter key in query textarea
        document.addEventListener('keydown', function(e) {
            if (e.ctrlKey && e.key === 'Enter') {
                executeQuery();
            }
        });
    </script>
</body>
</html>`
