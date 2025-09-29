package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bolt "go.etcd.io/bbolt"

	"github.com/daniel-le97/banned-cli/db"
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

	dbTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Bold(true).
			Padding(0, 1)
)

// Database buckets
var dbBuckets = []string{"settings", "channels", "videos", "downloads"}

var dbBucketTitles = map[string]string{
	"settings":  "Settings",
	"channels":  "Channels",
	"videos":    "Videos",
	"downloads": "Downloads",
}

type databaseViewModel struct {
	db             *bolt.DB
	table          table.Model
	currentBucket  int
	width          int
	height         int
	rawData        []interface{}
	error          error
	keys           keyMap
	showDetails    bool
	currentDetails string
	detailsText    string
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Tab    key.Binding
	Enter  key.Binding
	Toggle key.Binding
	Quit   key.Binding
	Help   key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
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
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "view details"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "toggle details panel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// copyToClipboard copies text to the system clipboard using a pure Go library
func copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

func newDatabaseViewModel(database *bolt.DB) *databaseViewModel {
	m := &databaseViewModel{
		db:            database,
		currentBucket: 0,
		keys:          newKeyMap(),
		showDetails:   false,
		detailsText:   "Select an item from the table above to view its details here.",
	}
	m.loadBucket(dbBuckets[0])
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

		// Calculate layout dimensions
		var tableHeight, textAreaHeight int

		if m.showDetails {
			// Split screen: 60% table, 40% text area
			tableHeight = int(float64(m.height-4) * 0.6)
			textAreaHeight = m.height - tableHeight - 6
		} else {
			// Full screen for table
			tableHeight = m.height - 4
			textAreaHeight = 0
		}

		if tableHeight < 5 {
			tableHeight = 5
		}
		if textAreaHeight < 3 && m.showDetails {
			textAreaHeight = 3
		} // Update table dimensions
		if m.table.Width() == 0 {
			m.table = m.table
		} else {
			m.table.SetWidth(m.width - 4)
			m.table.SetHeight(tableHeight)
		}

		// Text area dimensions will be handled by rendering

		// Reload bucket to recalculate column widths
		m.loadBucket(dbBuckets[m.currentBucket])

	case tea.KeyMsg:
		// Handle clipboard copy when in details view
		if m.showDetails {
			switch msg.Type {
			case tea.KeyEnter, tea.KeyCtrlC:
				if m.currentDetails != "" {
					if err := copyToClipboard(m.currentDetails); err != nil {
						fmt.Printf("Failed to copy to clipboard: %v\n", err)
					} else {
						fmt.Printf("✅ Copied to clipboard!\n")
					}
				}
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Left):
			m.currentBucket--
			if m.currentBucket < 0 {
				m.currentBucket = len(dbBuckets) - 1
			}
			m.loadBucket(dbBuckets[m.currentBucket])
			return m, nil

		case key.Matches(msg, m.keys.Right):
			m.currentBucket++
			if m.currentBucket >= len(dbBuckets) {
				m.currentBucket = 0
			}
			m.loadBucket(dbBuckets[m.currentBucket])
			return m, nil

		case key.Matches(msg, m.keys.Toggle):
			m.showDetails = !m.showDetails
			// Trigger a resize to recalculate layout
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}

		case key.Matches(msg, m.keys.Enter):
			m.updateDetailsArea()
			if !m.showDetails {
				m.showDetails = true
				// Trigger a resize to show the details area
				return m, func() tea.Msg {
					return tea.WindowSizeMsg{Width: m.width, Height: m.height}
				}
			}
			return m, nil
		}
	}

	// Update components
	var cmds []tea.Cmd

	// Update table
	m.table, cmd = m.table.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// No textarea updates needed for simple text rendering

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m *databaseViewModel) View() string {
	if m.error != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit.", m.error)
	}

	if len(m.rawData) == 0 {
		return fmt.Sprintf("Bucket '%s' is empty\n\nPress ←/→ to switch buckets or 'q' to quit.", dbBuckets[m.currentBucket])
	}

	var view strings.Builder

	// Title
	title := fmt.Sprintf("Database Viewer - %s (%d items)",
		dbBucketTitles[dbBuckets[m.currentBucket]], len(m.rawData))
	view.WriteString(dbTitleStyle.Render(title))
	view.WriteString("\n\n")

	// Table
	view.WriteString(baseStyle.Render(m.table.View()))

	// Details area if enabled
	if m.showDetails {
		view.WriteString("\n\n")
		view.WriteString(headerStyle.Render("Details (Press Enter or Ctrl+C to copy, mouse selection works too!)"))
		view.WriteString("\n")
		// Render details as simple selectable text (no borders like your working example)
		view.WriteString(m.detailsText)
	}

	// Help
	view.WriteString("\n")
	helpText := "←/→ switch buckets • ↑/↓ navigate • enter view details • d toggle details • q quit"
	if m.showDetails {
		helpText = "←/→ switch buckets • ↑/↓ navigate • enter/ctrl+c copy all • d toggle • q quit"
	}
	view.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(helpText))

	return view.String()
}

// updateDetailsArea updates the details text with the selected item's details
func (m *databaseViewModel) updateDetailsArea() {
	selectedRow := m.table.Cursor()
	if selectedRow < 0 || selectedRow >= len(m.rawData) {
		m.currentDetails = "No item selected"
		m.detailsText = m.currentDetails
		return
	}

	selectedData := m.rawData[selectedRow]
	details := m.formatItemDetails(selectedData, selectedRow+1)
	m.currentDetails = details
	m.detailsText = details
}

// formatItemDetails formats the selected item's data into a readable string
func (m *databaseViewModel) formatItemDetails(selectedData interface{}, rowNum int) string {
	var details strings.Builder

	details.WriteString(fmt.Sprintf("--- %s #%d ---\n\n", dbBuckets[m.currentBucket], rowNum))

	switch data := selectedData.(type) {
	case map[string]string:
		for key, value := range data {
			details.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}

	case db.Channel:
		details.WriteString(fmt.Sprintf("ID: %s\n", data.ID))
		details.WriteString(fmt.Sprintf("Title: %s\n", data.Title))
		if data.Summary != "" {
			details.WriteString(fmt.Sprintf("Summary: %s\n", data.Summary))
		}
		details.WriteString(fmt.Sprintf("Videos: %.0f | Views: %.0f | Likes: %.0f | Live: %t\n",
			data.TotalVideos, data.TotalVideoViews, data.TotalLikes, data.IsLive))

	case db.Video:
		details.WriteString(fmt.Sprintf("ID: %s\n", data.ID))
		details.WriteString(fmt.Sprintf("Title: %s\n", data.Title))
		if data.Summary != "" {
			details.WriteString(fmt.Sprintf("Summary: %s\n", data.Summary))
		}
		details.WriteString(fmt.Sprintf("Duration: %s | Views: %d | Likes: %d | Published: %t\n",
			formatDuration(data.VideoDuration), data.PlayCount, data.LikeCount, data.Published))
		if data.FileSize > 0 {
			details.WriteString(fmt.Sprintf("Size: %s\n", formatFileSizeForViewer(data.FileSize)))
		}

	case db.Download:
		details.WriteString(fmt.Sprintf("ID: %s\n", data.ID))
		details.WriteString(fmt.Sprintf("Title: %s\n", data.Title))
		details.WriteString(fmt.Sprintf("File: %s\n", data.Filename))
		details.WriteString(fmt.Sprintf("Status: %s | Size: %s | Torrent: %t\n",
			data.Status, formatFileSizeForViewer(data.FileSize), data.TorrentCreated))
		details.WriteString(fmt.Sprintf("Created: %s\n", data.CreatedAt.Format("2006-01-02 15:04:05")))

	default:
		prettyJSON, err := json.MarshalIndent(selectedData, "", "  ")
		if err == nil {
			details.WriteString(string(prettyJSON))
		} else {
			details.WriteString(fmt.Sprintf("%+v", selectedData))
		}
	}

	return details.String()
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
	fmt.Printf("Press Enter/Space on any row to view details, 'd' to toggle details panel.\n")
	fmt.Printf("In details view: Press Enter/Ctrl+C to copy all, or use mouse selection.\n")
	fmt.Printf("Use ←/→ to switch buckets, ↑/↓ to navigate rows, 'q' to quit.\n\n")

	model := newDatabaseViewModel(database)

	// Simple program setup like your working example
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run database viewer: %w", err)
	}

	fmt.Printf("\nDatabase viewer closed.\n")

	return nil
}
