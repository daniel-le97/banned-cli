package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daniel-le97/banned-cli/db"
	bolt "go.etcd.io/bbolt"
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

// Database bucket names and their display names
var dbBuckets = map[string]string{
	"settings":  "Settings",
	"channels":  "Channels",
	"videos":    "Videos",
	"downloads": "Downloads",
}

type databaseViewModel struct {
	db            *bolt.DB
	table         table.Model
	currentBucket string
	buckets       []string
	bucketIndex   int
	width         int
	height        int
	error         error
	rawData       []interface{} // Store raw data for selected rows
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
		key.WithHelp("←/h", "previous bucket"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("→/l", "next bucket"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch bucket"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter/space", "print selected"),
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

func newDatabaseViewModel(database *bolt.DB) *databaseViewModel {
	// Ensure settings is first since we know it has data
	buckets := []string{"settings", "channels", "videos", "downloads"}

	m := &databaseViewModel{
		db:            database,
		buckets:       buckets,
		currentBucket: "settings",
		bucketIndex:   0,
	}

	// Initialize with the settings bucket
	m.loadBucket(m.currentBucket)

	return m
}

func (m *databaseViewModel) Init() tea.Cmd {
	return nil
}

func (m *databaseViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Use more of the terminal space - account for title (1) + status bar (1) + margins (1)
		tableHeight := m.height - 3
		if tableHeight < 5 {
			tableHeight = 5
		}

		// Use nearly full width, just leave small margins
		m.table.SetWidth(m.width - 2)
		m.table.SetHeight(tableHeight)
		return m, nil

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			// Handle mouse clicks on table rows
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Enter):
			// Print selected row data to console
			m.printSelectedItem()
			return m, nil

		case key.Matches(msg, keys.Left), key.Matches(msg, keys.Tab):
			m.bucketIndex--
			if m.bucketIndex < 0 {
				m.bucketIndex = len(m.buckets) - 1
			}
			m.currentBucket = m.buckets[m.bucketIndex]
			m.loadBucket(m.currentBucket)
			return m, nil

		case key.Matches(msg, keys.Right):
			m.bucketIndex++
			if m.bucketIndex >= len(m.buckets) {
				m.bucketIndex = 0
			}
			m.currentBucket = m.buckets[m.bucketIndex]
			m.loadBucket(m.currentBucket)
			return m, nil
		}

		// Pass other keys to the table
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *databaseViewModel) View() string {
	if m.error != nil {
		return fmt.Sprintf("❌ Error: %v\nPress 'q' to quit.", m.error)
	}

	recordCount := len(m.table.Rows())
	if recordCount == 0 {
		return fmt.Sprintf("📵 Bucket '%s' is empty\n\nPress ←/→ to switch buckets or 'q' to quit.", dbBuckets[m.currentBucket])
	}

	// Title bar - use full width
	titleText := fmt.Sprintf(" 📊 Database Viewer - %s (%d records) ", dbBuckets[m.currentBucket], recordCount)
	titleBar := titleBarStyle.Width(m.width).Render(titleText)

	// Table view - use full available space
	tableView := m.table.View()

	// Status bar with navigation help - use full width
	statusText := fmt.Sprintf(" ←/→: Switch buckets | ↑/↓: Navigate rows | Enter/Space: Print selected | q: Quit | Current: %s ", m.currentBucket)
	statusBar := statusBarStyle.Width(m.width).Render(statusText)

	return fmt.Sprintf("%s\n%s\n%s", titleBar, tableView, statusBar)
}

// printSelectedItem prints the selected row's raw data to stdout
func (m *databaseViewModel) printSelectedItem() {
	selectedRow := m.table.Cursor()
	if selectedRow < 0 || selectedRow >= len(m.rawData) {
		fmt.Printf("No item selected or invalid selection\n")
		return
	}

	selectedData := m.rawData[selectedRow]

	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("Selected %s Record #%d:\n", dbBuckets[m.currentBucket], selectedRow+1)
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	switch data := selectedData.(type) {
	case map[string]string:
		// Settings data
		for key, value := range data {
			fmt.Printf("Key: %s\n", key)
			fmt.Printf("Value: %s\n", value)
			fmt.Printf(strings.Repeat("-", 40) + "\n")
		}
	case db.Channel:
		// Channel data
		prettyJSON, _ := json.MarshalIndent(data, "", "  ")
		fmt.Printf("%s\n", string(prettyJSON))
	case db.Video:
		// Video data
		prettyJSON, _ := json.MarshalIndent(data, "", "  ")
		fmt.Printf("%s\n", string(prettyJSON))
	case db.Download:
		// Download data
		prettyJSON, _ := json.MarshalIndent(data, "", "  ")
		fmt.Printf("%s\n", string(prettyJSON))
	default:
		fmt.Printf("Raw data: %+v\n", selectedData)
	}

	fmt.Printf(strings.Repeat("=", 80) + "\n")
}

func (m *databaseViewModel) loadBucket(bucketName string) {
	var columns []table.Column
	var rows []table.Row

	// Clear and reinitialize raw data storage
	m.rawData = make([]interface{}, 0)

	err := m.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket '%s' not found", bucketName)
		}

		switch bucketName {
		case "settings":
			// Calculate responsive column widths based on terminal width
			keyWidth := m.width / 3
			valueWidth := m.width - keyWidth - 4 // Leave some padding
			if keyWidth < 15 {
				keyWidth = 15
			}
			if valueWidth < 25 {
				valueWidth = 25
			}

			columns = []table.Column{
				{Title: "Key", Width: keyWidth},
				{Title: "Value", Width: valueWidth},
			}

			bucket.ForEach(func(k, v []byte) error {
				// Store raw data for this row
				rawItem := map[string]string{
					string(k): string(v),
				}
				m.rawData = append(m.rawData, rawItem)

				rows = append(rows, table.Row{
					string(k),
					string(v),
				})
				return nil
			})

		case "channels":
			// Calculate responsive column widths
			idWidth := m.width / 5
			titleWidth := m.width / 2
			videosWidth := 10
			viewsWidth := 12
			liveWidth := 8

			if idWidth < 20 {
				idWidth = 20
			}
			if titleWidth < 30 {
				titleWidth = 30
			}

			columns = []table.Column{
				{Title: "ID", Width: idWidth},
				{Title: "Title", Width: titleWidth},
				{Title: "Videos", Width: videosWidth},
				{Title: "Views", Width: viewsWidth},
				{Title: "Live", Width: liveWidth},
			}

			bucket.ForEach(func(k, v []byte) error {
				var channel db.Channel
				if err := json.Unmarshal(v, &channel); err != nil {
					return nil // Skip invalid records
				}

				// Store raw channel data
				m.rawData = append(m.rawData, channel)

				liveStatus := "No"
				if channel.IsLive {
					liveStatus = "Yes"
				}

				rows = append(rows, table.Row{
					truncateString(channel.ID, idWidth-2),
					truncateString(channel.Title, titleWidth-2),
					fmt.Sprintf("%.0f", channel.TotalVideos),
					fmt.Sprintf("%.0f", channel.TotalVideoViews),
					liveStatus,
				})
				return nil
			})

		case "videos":
			// Calculate responsive column widths
			idWidth := m.width / 6
			titleWidth := m.width / 2
			durationWidth := 12
			viewsWidth := 10
			likesWidth := 8
			sizeWidth := 12

			if idWidth < 20 {
				idWidth = 20
			}
			if titleWidth < 25 {
				titleWidth = 25
			}

			columns = []table.Column{
				{Title: "ID", Width: idWidth},
				{Title: "Title", Width: titleWidth},
				{Title: "Duration", Width: durationWidth},
				{Title: "Views", Width: viewsWidth},
				{Title: "Likes", Width: likesWidth},
				{Title: "Size", Width: sizeWidth},
			}

			bucket.ForEach(func(k, v []byte) error {
				var record struct {
					Video     db.Video `json:"video"`
					ChannelID string   `json:"channel_id"`
				}
				if err := json.Unmarshal(v, &record); err != nil {
					return nil // Skip invalid records
				}

				// Store raw video data
				m.rawData = append(m.rawData, record.Video)

				duration := formatDuration(record.Video.VideoDuration)
				fileSize := formatFileSizeForViewer(record.Video.FileSize)

				rows = append(rows, table.Row{
					truncateString(record.Video.ID, idWidth-2),
					truncateString(record.Video.Title, titleWidth-2),
					duration,
					fmt.Sprintf("%d", record.Video.PlayCount),
					fmt.Sprintf("%d", record.Video.LikeCount),
					fileSize,
				})
				return nil
			})

		case "downloads":
			// Calculate responsive column widths
			idWidth := m.width / 6
			titleWidth := m.width / 3
			statusWidth := 12
			sizeWidth := 12
			torrentWidth := 10
			createdWidth := 15

			if idWidth < 20 {
				idWidth = 20
			}
			if titleWidth < 25 {
				titleWidth = 25
			}

			columns = []table.Column{
				{Title: "ID", Width: idWidth},
				{Title: "Title", Width: titleWidth},
				{Title: "Status", Width: statusWidth},
				{Title: "Size", Width: sizeWidth},
				{Title: "Torrent", Width: torrentWidth},
				{Title: "Created", Width: createdWidth},
			}

			bucket.ForEach(func(k, v []byte) error {
				var download db.Download
				if err := json.Unmarshal(v, &download); err != nil {
					return nil // Skip invalid records
				}

				// Store raw download data
				m.rawData = append(m.rawData, download)

				torrentStatus := "No"
				if download.TorrentCreated {
					torrentStatus = "Yes"
				}

				created := download.CreatedAt.Format("2006-01-02")
				fileSize := formatFileSizeForViewer(download.FileSize)

				rows = append(rows, table.Row{
					truncateString(download.ID, idWidth-2),
					truncateString(download.Title, titleWidth-2),
					download.Status,
					fileSize,
					torrentStatus,
					created,
				})
				return nil
			})
		}

		return nil
	})

	if err != nil {
		m.error = err
		return
	}

	// Create and configure the table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height-6),
	)

	// Apply custom styles
	s := table.DefaultStyles()
	s.Header = headerStyle
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m.table = t
	m.error = nil
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "0:00"
	}

	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func formatFileSizeForViewer(bytes int64) string {
	if bytes == 0 {
		return "Unknown"
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
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// RunDatabaseViewer starts the interactive database viewer
func RunDatabaseViewer() error {
	database, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Printf("Starting interactive database viewer...\n")
	fmt.Printf("💡 Mouse selection and text copying are enabled in your terminal.\n")
	fmt.Printf("📋 Press Enter/Space on any row to print its full data to console.\n")
	fmt.Printf("⌨️  Use ←/→ to switch buckets, ↑/↓ to navigate rows, 'q' to quit.\n\n")

	model := newDatabaseViewModel(database)

	// Use alt screen and enable mouse support for better full-screen experience
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run database viewer: %w", err)
	}

	fmt.Printf("\n✅ Database viewer closed.\n")

	return nil
}
