package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Database viewer styles
var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1)

	titleBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3C3C3C")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	dbHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// Database table names and their display names
var dbTables = map[string]string{
	"downloads": "Downloads",
	"settings":  "Settings",
	"channels":  "Channels",
	"videos":    "Videos",
}

type databaseViewModel struct {
	db           *sql.DB
	table        table.Model
	currentTable string
	tables       []string
	tableIndex   int
	width        int
	height       int
	error        error
}

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Tab   key.Binding
	Enter key.Binding
	Quit  key.Binding
	Help  key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("←/h", "previous table"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("→/l", "next table"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch table"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

func newDatabaseViewModel(db *sql.DB) *databaseViewModel {
	// Ensure settings is first since we know it has data
	tables := []string{"settings", "downloads", "channels", "videos"}

	m := &databaseViewModel{
		db:           db,
		tables:       tables,
		currentTable: "settings",
		tableIndex:   0,
	}

	// Initialize with the settings table
	m.loadTable(m.currentTable)

	return m
}

func (m *databaseViewModel) Init() tea.Cmd {
	return nil
}

func (m *databaseViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Left):
			m.previousTable()
			m.loadTable(m.currentTable)

		case key.Matches(msg, keys.Right), key.Matches(msg, keys.Tab):
			m.nextTable()
			m.loadTable(m.currentTable)

		default:
			m.table, cmd = m.table.Update(msg)
		}
	}

	return m, cmd
}

func (m *databaseViewModel) View() string {
	if m.error != nil {
		return fmt.Sprintf("❌ Error loading table '%s': %v\n\nPress 'q' to quit or ←/→ to try other tables.", m.currentTable, m.error)
	}

	recordCount := len(m.table.Rows())
	if recordCount == 0 {
		return fmt.Sprintf("📵 Table '%s' is empty\n\nPress ←/→ to switch tables or 'q' to quit.", dbTables[m.currentTable])
	}

	// Title with current table info - use full width
	titleText := fmt.Sprintf(" 📊 Database Viewer - %s (%d records) ", dbTables[m.currentTable], recordCount)
	title := titleBarStyle.Width(m.width).Render(titleText)

	// Table navigation info - use full width
	navInfo := fmt.Sprintf("Table %d/%d - Use ←/→ or Tab to switch", m.tableIndex+1, len(m.tables))
	navBar := statusBarStyle.Width(m.width).Render(navInfo) // Table view
	tableView := baseStyle.Render(m.table.View())

	// Help text
	helpText := dbHelpStyle.Render("Navigation: ←/→ (switch table) | ↑/↓ (navigate rows) | q (quit)")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		navBar,
		"",
		tableView,
		"",
		helpText,
	)
}

func (m *databaseViewModel) nextTable() {
	m.tableIndex = (m.tableIndex + 1) % len(m.tables)
	m.currentTable = m.tables[m.tableIndex]
}

func (m *databaseViewModel) previousTable() {
	m.tableIndex = (m.tableIndex - 1 + len(m.tables)) % len(m.tables)
	m.currentTable = m.tables[m.tableIndex]
}

func (m *databaseViewModel) updateTableSize() {
	if m.width > 0 && m.height > 0 {
		// Reserve space for title, status bar, help text, and padding
		availableHeight := m.height - 8
		availableWidth := m.width - 4

		if availableHeight < 5 {
			availableHeight = 5
		}
		if availableWidth < 40 {
			availableWidth = 40
		}

		m.table.SetWidth(availableWidth)
		m.table.SetHeight(availableHeight)

		// Recalculate column widths based on available space
		m.updateColumnWidths(availableWidth)
	}
}

func (m *databaseViewModel) updateColumnWidths(availableWidth int) {
	columns := m.table.Columns()
	if len(columns) == 0 {
		return
	}

	// Calculate new column widths based on table type and available space
	var newColumns []table.Column
	switch m.currentTable {
	case "settings":
		newColumns = m.calculateSettingsColumnWidths(availableWidth)
	case "downloads":
		newColumns = m.calculateDownloadsColumnWidths(availableWidth)
	case "channels":
		newColumns = m.calculateChannelsColumnWidths(availableWidth)
	case "videos":
		newColumns = m.calculateVideosColumnWidths(availableWidth)
	default:
		return
	}

	m.table.SetColumns(newColumns)
}

// Helper function to truncate text based on column width
func (m *databaseViewModel) truncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth-3 { // Reserve 3 chars for "..."
		return text
	}
	if maxWidth <= 3 {
		return "..."
	}
	return text[:maxWidth-3] + "..."
}

func (m *databaseViewModel) calculateSettingsColumnWidths(availableWidth int) []table.Column {
	// Settings: Key, Value, Updated
	keyWidth := availableWidth * 30 / 100     // 30%
	valueWidth := availableWidth * 50 / 100   // 50%
	updatedWidth := availableWidth * 20 / 100 // 20%

	if keyWidth < 15 {
		keyWidth = 15
	}
	if valueWidth < 20 {
		valueWidth = 20
	}
	if updatedWidth < 12 {
		updatedWidth = 12
	}

	return []table.Column{
		{Title: "Key", Width: keyWidth},
		{Title: "Value", Width: valueWidth},
		{Title: "Updated", Width: updatedWidth},
	}
}

func (m *databaseViewModel) calculateDownloadsColumnWidths(availableWidth int) []table.Column {
	// Downloads: ID, URL, Title, Status, Created
	idWidth := availableWidth * 8 / 100       // 8%
	urlWidth := availableWidth * 40 / 100     // 40%
	titleWidth := availableWidth * 25 / 100   // 25%
	statusWidth := availableWidth * 12 / 100  // 12%
	createdWidth := availableWidth * 15 / 100 // 15%

	if idWidth < 6 {
		idWidth = 6
	}
	if urlWidth < 20 {
		urlWidth = 20
	}
	if titleWidth < 15 {
		titleWidth = 15
	}
	if statusWidth < 8 {
		statusWidth = 8
	}
	if createdWidth < 12 {
		createdWidth = 12
	}

	return []table.Column{
		{Title: "ID", Width: idWidth},
		{Title: "URL", Width: urlWidth},
		{Title: "Title", Width: titleWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Created", Width: createdWidth},
	}
}

func (m *databaseViewModel) calculateChannelsColumnWidths(availableWidth int) []table.Column {
	// Channels: ID, Title, Website, Live, Updated
	idWidth := availableWidth * 15 / 100      // 15%
	titleWidth := availableWidth * 35 / 100   // 35%
	websiteWidth := availableWidth * 25 / 100 // 25%
	liveWidth := availableWidth * 10 / 100    // 10%
	updatedWidth := availableWidth * 15 / 100 // 15%

	if idWidth < 10 {
		idWidth = 10
	}
	if titleWidth < 20 {
		titleWidth = 20
	}
	if websiteWidth < 15 {
		websiteWidth = 15
	}
	if liveWidth < 6 {
		liveWidth = 6
	}
	if updatedWidth < 12 {
		updatedWidth = 12
	}

	return []table.Column{
		{Title: "ID", Width: idWidth},
		{Title: "Title", Width: titleWidth},
		{Title: "Website", Width: websiteWidth},
		{Title: "Live", Width: liveWidth},
		{Title: "Updated", Width: updatedWidth},
	}
}

func (m *databaseViewModel) calculateVideosColumnWidths(availableWidth int) []table.Column {
	// Videos: ID, Title, Channel, Duration, Plays
	idWidth := availableWidth * 12 / 100       // 12%
	titleWidth := availableWidth * 40 / 100    // 40%
	channelWidth := availableWidth * 25 / 100  // 25%
	durationWidth := availableWidth * 12 / 100 // 12%
	playsWidth := availableWidth * 11 / 100    // 11%

	if idWidth < 10 {
		idWidth = 10
	}
	if titleWidth < 25 {
		titleWidth = 25
	}
	if channelWidth < 15 {
		channelWidth = 15
	}
	if durationWidth < 8 {
		durationWidth = 8
	}
	if playsWidth < 6 {
		playsWidth = 6
	}

	return []table.Column{
		{Title: "ID", Width: idWidth},
		{Title: "Title", Width: titleWidth},
		{Title: "Channel", Width: channelWidth},
		{Title: "Duration", Width: durationWidth},
		{Title: "Plays", Width: playsWidth},
	}
}

func (m *databaseViewModel) loadTable(tableName string) {
	var columns []table.Column
	var rows []table.Row

	switch tableName {
	case "downloads":
		columns, rows = m.loadDownloadsTable()
	case "settings":
		columns, rows = m.loadSettingsTable()
	case "channels":
		columns, rows = m.loadChannelsTable()
	case "videos":
		columns, rows = m.loadVideosTable()
	default:
		m.error = fmt.Errorf("unknown table: %s", tableName)
		return
	}

	// Check if we got data
	if len(columns) == 0 {
		m.error = fmt.Errorf("no columns found for table: %s", tableName)
		return
	}

	// Create a default row if no data
	if len(rows) == 0 {
		rows = []table.Row{{"No data", "Table is empty", "", "", ""}}
	}

	// Create table without fixed dimensions
	m.table = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	// Apply styling
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.table.SetStyles(s)

	// Update table size to fit terminal
	m.updateTableSize()
	m.error = nil
}

func (m *databaseViewModel) loadSettingsTable() ([]table.Column, []table.Row) {
	// Use default widths that will be updated by updateColumnWidths
	columns := m.calculateSettingsColumnWidths(80) // Default fallback width

	query := "SELECT key, value, updated_at FROM settings ORDER BY key"

	dbRows, err := m.db.Query(query)
	if err != nil {
		return columns, []table.Row{{"Error", fmt.Sprintf("Query failed: %v", err), "N/A"}}
	}
	defer dbRows.Close()

	var rows []table.Row
	for dbRows.Next() {
		var key, value, updatedAt string

		if err := dbRows.Scan(&key, &value, &updatedAt); err != nil {
			continue
		}

		// Get current column widths for dynamic truncation
		currentColumns := m.calculateSettingsColumnWidths(m.width)
		valueMaxWidth := currentColumns[1].Width

		// Truncate value dynamically based on column width
		value = m.truncateText(value, valueMaxWidth)

		// Parse and format date
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			updatedAt = t.Format("2006-01-02 15:04")
		}

		rows = append(rows, table.Row{
			key,
			value,
			updatedAt,
		})
	}

	return columns, rows
}

func (m *databaseViewModel) loadDownloadsTable() ([]table.Column, []table.Row) {
	// Use default widths that will be updated by updateColumnWidths
	columns := m.calculateDownloadsColumnWidths(80) // Default fallback width

	query := `
		SELECT id, url, COALESCE(title, 'Unknown'), 
		       COALESCE(status, 'pending'), created_at 
		FROM downloads 
		ORDER BY created_at DESC 
		LIMIT 100`

	dbRows, err := m.db.Query(query)
	if err != nil {
		return columns, []table.Row{{"Error", fmt.Sprintf("Query failed: %v", err), "", "", ""}}
	}
	defer dbRows.Close()

	var rows []table.Row
	for dbRows.Next() {
		var id int
		var url, title, status, createdAt string

		if err := dbRows.Scan(&id, &url, &title, &status, &createdAt); err != nil {
			continue
		}

		// Get current column widths for dynamic truncation
		currentColumns := m.calculateDownloadsColumnWidths(m.width)
		urlMaxWidth := currentColumns[1].Width
		titleMaxWidth := currentColumns[2].Width

		// Truncate dynamically based on column width
		url = m.truncateText(url, urlMaxWidth)
		title = m.truncateText(title, titleMaxWidth)

		// Parse and format date
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			createdAt = t.Format("2006-01-02 15:04")
		}

		rows = append(rows, table.Row{
			strconv.Itoa(id),
			url,
			title,
			status,
			createdAt,
		})
	}

	return columns, rows
}

func (m *databaseViewModel) loadChannelsTable() ([]table.Column, []table.Row) {
	// Use default widths that will be updated by updateColumnWidths
	columns := m.calculateChannelsColumnWidths(80) // Default fallback width

	query := `
		SELECT id, COALESCE(title, 'Unknown'), COALESCE(website, ''), 
		       COALESCE(is_live, 0), COALESCE(updated_at, created_at)
		FROM channels 
		ORDER BY title 
		LIMIT 100`

	dbRows, err := m.db.Query(query)
	if err != nil {
		return columns, []table.Row{{"Error", fmt.Sprintf("Query failed: %v", err), "", "", ""}}
	}
	defer dbRows.Close()

	var rows []table.Row
	for dbRows.Next() {
		var isLive int
		var id, title, website, updatedAt string

		if err := dbRows.Scan(&id, &title, &website, &isLive, &updatedAt); err != nil {
			continue
		}

		// Get current column widths for dynamic truncation
		currentColumns := m.calculateChannelsColumnWidths(m.width)
		idMaxWidth := currentColumns[0].Width
		titleMaxWidth := currentColumns[1].Width
		websiteMaxWidth := currentColumns[2].Width

		// Truncate dynamically based on column width
		id = m.truncateText(id, idMaxWidth)
		title = m.truncateText(title, titleMaxWidth)
		website = m.truncateText(website, websiteMaxWidth)

		// Convert isLive to readable format
		liveStatus := "No"
		if isLive == 1 {
			liveStatus = "Yes"
		}

		// Parse and format date
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			updatedAt = t.Format("2006-01-02 15:04")
		}

		rows = append(rows, table.Row{
			id,
			title,
			website,
			liveStatus,
			updatedAt,
		})
	}

	return columns, rows
}

func (m *databaseViewModel) loadVideosTable() ([]table.Column, []table.Row) {
	// Use default widths that will be updated by updateColumnWidths
	columns := m.calculateVideosColumnWidths(80) // Default fallback width

	query := `
		SELECT v.id, COALESCE(v.title, 'Untitled'), 
		       COALESCE(c.title, 'Unknown'), 
		       COALESCE(v.video_duration, 0), 
		       COALESCE(v.play_count, 0)
		FROM videos v
		LEFT JOIN channels c ON v.channel_id = c.id
		ORDER BY v.play_count DESC 
		LIMIT 100`

	dbRows, err := m.db.Query(query)
	if err != nil {
		return columns, []table.Row{{"Error", fmt.Sprintf("Query failed: %v", err), "", "", ""}}
	}
	defer dbRows.Close()

	var rows []table.Row
	for dbRows.Next() {
		var duration float64
		var playCount int
		var id, title, channel string

		if err := dbRows.Scan(&id, &title, &channel, &duration, &playCount); err != nil {
			continue
		}

		// Get current column widths for dynamic truncation
		currentColumns := m.calculateVideosColumnWidths(m.width)
		idMaxWidth := currentColumns[0].Width
		titleMaxWidth := currentColumns[1].Width
		channelMaxWidth := currentColumns[2].Width

		// Truncate dynamically based on column width
		id = m.truncateText(id, idMaxWidth)
		title = m.truncateText(title, titleMaxWidth)
		channel = m.truncateText(channel, channelMaxWidth)

		// Format duration from float (seconds) to mm:ss
		var durationStr string
		if duration > 0 {
			totalSeconds := int(duration)
			hours := totalSeconds / 3600
			minutes := (totalSeconds % 3600) / 60
			seconds := totalSeconds % 60
			if hours > 0 {
				durationStr = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
			} else {
				durationStr = fmt.Sprintf("%d:%02d", minutes, seconds)
			}
		} else {
			durationStr = "Unknown"
		}

		rows = append(rows, table.Row{
			id,
			title,
			channel,
			durationStr,
			strconv.Itoa(playCount),
		})
	}

	return columns, rows
}
