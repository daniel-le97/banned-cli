/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUI Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(lipgloss.Color("#FAFAFA"))

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#EE6FF8")).
				Bold(true)

	paginationStyle = list.DefaultStyles().PaginationStyle.
			PaddingLeft(4)

	helpStyle = list.DefaultStyles().HelpStyle.
			PaddingLeft(4).
			PaddingBottom(1)

	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

// Menu items
type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("▶ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// Main menu model
type menuModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menuModel) View() string {
	if m.choice != "" {
		return fmt.Sprintf("\nSelected: %s\n", m.choice)
	}
	if m.quitting {
		return fmt.Sprintf("\n%s\n", "Goodbye! 👋")
	}

	header := titleStyle.Render(fmt.Sprintf("%s - Interactive Menu", AppName))

	return docStyle.Render(header + "\n\n" + m.list.View())
}

// Create the main menu
func NewMenu() menuModel {
	items := []list.Item{
		item("Browse Channels"),
		item("Browse Videos"),
		item("Download from URL"),
		item(fmt.Sprintf("Install %s to $PATH", AppName)),
		item("Update application"),
		item("Settings"),
		item("Exit"),
	}

	const defaultWidth = 20
	const listHeight = 14

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = fmt.Sprintf("%s - banned.video Content Manager", AppName)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return menuModel{list: l}
}

// Run the interactive TUI
func RunTUI() error {
	m := NewMenu()
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Handle the user's choice
	if finalModel, ok := finalModel.(menuModel); ok && finalModel.choice != "" {
		return handleMenuChoice(finalModel.choice)
	}

	return nil
}

// Handle the selected menu choice
func handleMenuChoice(choice string) error {
	switch choice {
	case "Browse Channels":
		return runChannelsFlow()
	case "Browse Videos":
		return runVideosFlow()
	case "Download from URL":
		return runDownloadFlow()
	case fmt.Sprintf("Install %s to $PATH", AppName):
		return runInstallFlow()
	case "Update application":
		return runUpdateFlow()
	case "Settings":
		return runSettingsFlow()
	case "Exit":
		fmt.Println("Goodbye! 👋")
		return nil
	}
	return nil
}

// Download flow with TUI
func runDownloadFlow() error {
	fmt.Printf("\n🚀 %s Download Manager\n", AppName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\nThis feature will be implemented with the download command.")
	fmt.Println("Usage: banned download <url>")
	return nil
}

// Install flow with TUI
func runInstallFlow() error {
	fmt.Printf("\n📦 %s Installation\n", AppName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create an interactive install process
	return createInstallTUI()
}

// Update flow with TUI
func runUpdateFlow() error {
	fmt.Printf("\n⬆️  %s Update Manager\n", AppName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create an interactive update process
	return createUpdateTUI()
}

// Settings flow
func runSettingsFlow() error {
	fmt.Printf("\n⚙️  %s Settings\n", AppName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Display current settings from database
	settings := []string{"download_dir", "max_concurrent_downloads", "retry_attempts", "user_agent"}

	fmt.Println("\nCurrent Settings:")
	for _, key := range settings {
		value, err := GetSetting(key)
		if err != nil {
			fmt.Printf("  %s: <error: %v>\n", key, err)
		} else {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Println("\nSettings management is available through the database!")
	fmt.Printf("Database location: %s\n", getDatabaseLocation())
	return nil
}

// Helper function to get database location for display
func getDatabaseLocation() string {
	// Get database path using the same logic as the database module
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "unknown"
		}
		dataDir = filepath.Join(homeDir, ".local", "share")
	}

	appDataDir := filepath.Join(dataDir, AppName)
	dbPath := filepath.Join(appDataDir, AppName+".db")

	return dbPath
}

// Install TUI
func createInstallTUI() error {
	items := []list.Item{
		item("Create symlink (recommended)"),
		item("Add to shell profile"),
		item("Cancel"),
	}

	l := list.New(items, itemDelegate{}, 50, 10)
	l.Title = "How would you like to install?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	model := menuModel{list: l}
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	if finalModel, ok := finalModel.(menuModel); ok && finalModel.choice != "" {
		switch finalModel.choice {
		case "Create symlink (recommended)":
			fmt.Printf("\n🔗 Installing %s via symlink...\n", AppName)
			return installBinary(false)
		case "Add to shell profile":
			fmt.Printf("\n📝 Installing %s via shell profile...\n", AppName)
			return installBinary(true)
		case "Cancel":
			fmt.Println("Installation cancelled.")
			return nil
		}
	}
	return nil
}

// Update TUI
func createUpdateTUI() error {
	items := []list.Item{
		item("Update from default source"),
		item("Update from custom source"),
		item("Cancel"),
	}

	l := list.New(items, itemDelegate{}, 50, 10)
	l.Title = "How would you like to update?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	model := menuModel{list: l}
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	if finalModel, ok := finalModel.(menuModel); ok && finalModel.choice != "" {
		switch finalModel.choice {
		case "Update from default source":
			fmt.Printf("\n⬆️  Updating %s from default source...\n", AppName)
			return updateBinary("")
		case "Update from custom source":
			fmt.Print("\n🔗 Enter custom source URL: ")
			var sourceURL string
			fmt.Scanln(&sourceURL)
			if sourceURL != "" {
				fmt.Printf("⬆️  Updating %s from %s...\n", AppName, sourceURL)
				return updateBinary(sourceURL)
			}
			return nil
		case "Cancel":
			fmt.Println("Update cancelled.")
			return nil
		}
	}
	return nil
}

// Channels flow with TUI
func runChannelsFlow() error {
	fmt.Printf("\n🎬 %s Channel Browser\n", AppName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get channels from database
	channels, err := GetAllChannelsFromDB()
	if err != nil {
		fmt.Printf("❌ Failed to get channels from database: %v\n", err)
		return err
	}

	if len(channels) == 0 {
		fmt.Printf("📭 No channels found in database.\n")
		fmt.Printf("💡 Use 'banned fetch channels' to fetch from API first.\n")
		return nil
	}

	// Create channel selection menu
	return createChannelSelectionTUI(channels)
}

// Videos flow with TUI
func runVideosFlow() error {
	fmt.Printf("\n🎥 %s Video Browser\n", AppName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// First get channels to allow user to select one
	channels, err := GetAllChannelsFromDB()
	if err != nil {
		fmt.Printf("❌ Failed to get channels from database: %v\n", err)
		return err
	}

	if len(channels) == 0 {
		fmt.Printf("📭 No channels found in database.\n")
		fmt.Printf("💡 Use 'banned fetch channels' to fetch from API first.\n")
		return nil
	}

	// Let user select a channel first
	selectedChannel, err := selectChannelForVideos(channels)
	if err != nil {
		return err
	}

	if selectedChannel == nil {
		fmt.Println("No channel selected.")
		return nil
	}

	// Get videos for selected channel
	videos, err := GetChannelVideos(selectedChannel.ID, 50, 0) // Get first 50 videos
	if err != nil {
		fmt.Printf("❌ Failed to get videos for channel '%s': %v\n", selectedChannel.Title, err)
		return err
	}

	if len(videos) == 0 {
		fmt.Printf("📭 No videos found for channel '%s'.\n", selectedChannel.Title)
		fmt.Printf("💡 Use 'banned fetch channel %s' to fetch videos from API.\n", selectedChannel.ID)
		return nil
	}

	// Create video selection menu
	return createVideoSelectionTUI(selectedChannel, videos)
}

// Create channel selection TUI
func createChannelSelectionTUI(channels []Channel) error {
	fmt.Printf("\n📋 Found %d channels:\n\n", len(channels))

	// Display channels with numbers
	for i, channel := range channels {
		fmt.Printf("%d. 🎬 %s\n", i+1, channel.Title)
		if channel.Summary != "" {
			fmt.Printf("   📝 %s\n", channel.Summary)
		}
		fmt.Printf("   🆔 %s\n\n", channel.ID)
	}

	fmt.Print("Enter channel number to view details (or 0 to go back): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice == 0 {
		return nil
	}

	if choice < 1 || choice > len(channels) {
		fmt.Printf("❌ Invalid choice. Please select 1-%d\n", len(channels))
		return nil
	}

	selectedChannel := channels[choice-1]
	return displayChannelDetails(selectedChannel)
}

// Select channel for videos view
func selectChannelForVideos(channels []Channel) (*Channel, error) {
	fmt.Printf("\n📋 Select a channel to browse videos:\n\n")

	// Display channels with numbers
	for i, channel := range channels {
		fmt.Printf("%d. 🎬 %s\n", i+1, channel.Title)
	}

	fmt.Print("\nEnter channel number (or 0 to go back): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice == 0 {
		return nil, nil
	}

	if choice < 1 || choice > len(channels) {
		fmt.Printf("❌ Invalid choice. Please select 1-%d\n", len(channels))
		return nil, nil
	}

	return &channels[choice-1], nil
}

// Create video selection TUI
func createVideoSelectionTUI(channel *Channel, videos []Video) error {
	fmt.Printf("\n📹 Videos in '%s' (%d total):\n\n", channel.Title, len(videos))

	// Display videos with numbers
	for i, video := range videos {
		fmt.Printf("%d. 🎥 %s\n", i+1, video.Title)
		if video.Summary != "" {
			summary := video.Summary
			if len(summary) > 100 {
				summary = summary[:100] + "..."
			}
			fmt.Printf("   📝 %s\n", summary)
		}
		if video.VideoDuration > 0 {
			minutes := int(video.VideoDuration / 60)
			seconds := int(video.VideoDuration) % 60
			fmt.Printf("   ⏱️  Duration: %d:%02d\n", minutes, seconds)
		}
		if video.DirectURL != "" {
			fmt.Printf("   🔗 %s\n", video.DirectURL)
		}
		fmt.Printf("   🆔 %s\n\n", video.ID)
	}

	fmt.Print("Enter video number to download (or 0 to go back): ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice == 0 {
		return nil
	}

	if choice < 1 || choice > len(videos) {
		fmt.Printf("❌ Invalid choice. Please select 1-%d\n", len(videos))
		return nil
	}

	selectedVideo := videos[choice-1]
	return handleVideoSelection(selectedVideo)
}

// Display channel details
func displayChannelDetails(channel Channel) error {
	fmt.Printf("\n🎬 Channel Details: %s\n", channel.Title)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if channel.Summary != "" {
		fmt.Printf("📝 Summary: %s\n", channel.Summary)
	}
	if channel.TextInfo != "" {
		fmt.Printf("ℹ️  Info: %s\n", channel.TextInfo)
	}
	if channel.Avatar != "" {
		fmt.Printf("🖼️  Avatar: %s\n", channel.Avatar)
	}
	if channel.IsLive {
		fmt.Printf("🔴 Status: LIVE\n")
	}

	// Show social links
	if channel.Links != nil {
		fmt.Printf("\n🔗 Links:\n")
		if channel.Links.Website != "" {
			fmt.Printf("   🌐 Website: %s\n", channel.Links.Website)
		}
		if channel.Links.Twitter != "" {
			fmt.Printf("   🐦 Twitter: %s\n", channel.Links.Twitter)
		}
		if channel.Links.Facebook != "" {
			fmt.Printf("   📘 Facebook: %s\n", channel.Links.Facebook)
		}
		if channel.Links.Telegram != "" {
			fmt.Printf("   📱 Telegram: %s\n", channel.Links.Telegram)
		}
	}

	fmt.Printf("\n🆔 Channel ID: %s\n", channel.ID)

	// Get video count for this channel
	_, err := GetDB()
	if err == nil {
		videoCount, err := CountVideosForChannel(channel.ID)
		if err == nil {
			fmt.Printf("📹 Videos in database: %d\n", videoCount)
		}
	}

	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()
	return nil
}

// Handle video selection (download option)
func handleVideoSelection(video Video) error {
	fmt.Printf("\n🎥 Selected Video: %s\n", video.Title)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if video.Summary != "" {
		fmt.Printf("📝 Summary: %s\n", video.Summary)
	}
	if video.VideoDuration > 0 {
		minutes := int(video.VideoDuration / 60)
		seconds := int(video.VideoDuration) % 60
		fmt.Printf("⏱️  Duration: %d:%02d\n", minutes, seconds)
	}
	if video.DirectURL != "" {
		fmt.Printf("🔗 Direct URL: %s\n", video.DirectURL)
	}

	fmt.Printf("\nWould you like to download this video? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		fmt.Printf("🚀 Starting download of '%s'...\n", video.Title)

		if video.DirectURL == "" {
			fmt.Printf("❌ No direct URL available for this video\n")
			return nil
		}

		// Create filename from title
		filename := fmt.Sprintf("%s.mp4", video.Title)
		// Clean filename of invalid characters
		filename = strings.ReplaceAll(filename, "/", "_")
		filename = strings.ReplaceAll(filename, "\\", "_")
		filename = strings.ReplaceAll(filename, ":", "_")

		err := DownloadVideo(video.DirectURL, filename)
		if err != nil {
			fmt.Printf("❌ Download failed: %v\n", err)
			return err
		}

		fmt.Printf("✅ Successfully downloaded '%s'\n", filename)

		// Ask if user wants to create torrent
		fmt.Printf("\nCreate torrent file? (y/N): ")
		var torrentResponse string
		fmt.Scanln(&torrentResponse)

		if strings.ToLower(torrentResponse) == "y" || strings.ToLower(torrentResponse) == "yes" {
			err := createTorrentWithMktorrent(filename, "", nil)
			if err != nil {
				fmt.Printf("⚠️  Failed to create torrent: %v\n", err)
			} else {
				fmt.Printf("✅ Torrent file created successfully\n")
			}
		}
	}

	return nil
}

// Channel list item for TUI
type channelItem struct {
	channel Channel
}

func (i channelItem) FilterValue() string {
	return i.channel.Title + " " + i.channel.Summary
}

func (i channelItem) Title() string {
	return "🎬 " + i.channel.Title
}

func (i channelItem) Description() string {
	desc := i.channel.Summary
	if desc == "" {
		desc = "No description available"
	}
	// Add channel ID and live status
	if i.channel.IsLive {
		desc += " 🔴 LIVE"
	}
	desc += fmt.Sprintf(" • ID: %s", i.channel.ID)
	return desc
}

// Channel browser model
type channelBrowserModel struct {
	list          list.Model
	textInput     textinput.Model
	allChannels   []Channel
	choice        *Channel
	quitting      bool
	searchFocused bool
}

func (m channelBrowserModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m channelBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Use full terminal width and height minus space for title, help, status bar, and search input
		h := msg.Height - 6 // Account for title, help, status bar, search input, and padding
		if h < 1 {
			h = 1
		}
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(h)
		m.textInput.Width = msg.Width - 10 // Leave some margin for search input
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.searchFocused {
				m.searchFocused = false
				m.textInput.Blur()
				return m, nil
			} else {
				m.quitting = true
				return m, tea.Quit
			}

		case "/", "ctrl+f":
			// Focus search input
			m.searchFocused = true
			m.textInput.Focus()
			return m, textinput.Blink

		case "enter":
			if m.searchFocused {
				// Apply search filter
				m.searchFocused = false
				m.textInput.Blur()
				m.filterChannels()
				return m, nil
			} else if item, ok := m.list.SelectedItem().(channelItem); ok {
				m.choice = &item.channel
				return m, tea.Quit
			}
		}

		// Handle search input when focused
		if m.searchFocused {
			m.textInput, cmd = m.textInput.Update(msg)
			// Filter as user types
			m.filterChannels()
			return m, cmd
		}
	}

	// Update list when not in search mode
	if !m.searchFocused {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

// filterChannels filters the channel list based on search input
func (m *channelBrowserModel) filterChannels() {
	searchTerm := strings.ToLower(strings.TrimSpace(m.textInput.Value()))

	if searchTerm == "" {
		// Show all channels if search is empty
		items := make([]list.Item, len(m.allChannels))
		for i, channel := range m.allChannels {
			items[i] = channelItem{channel: channel}
		}
		m.list.SetItems(items)
		return
	}

	// Filter channels based on search term
	var filteredItems []list.Item
	for _, channel := range m.allChannels {
		if strings.Contains(strings.ToLower(channel.Title), searchTerm) ||
			strings.Contains(strings.ToLower(channel.Summary), searchTerm) ||
			strings.Contains(strings.ToLower(channel.ID), searchTerm) {
			filteredItems = append(filteredItems, channelItem{channel: channel})
		}
	}

	m.list.SetItems(filteredItems)
}

func (m channelBrowserModel) View() string {
	if m.choice != nil {
		return ""
	}

	// Create search bar styling
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("236")).
		Padding(0, 1)

	if m.searchFocused {
		searchStyle = searchStyle.BorderForeground(lipgloss.Color("69"))
	}

	// Build the search section
	searchPrompt := "Search: "
	if !m.searchFocused {
		searchPrompt += "Press '/' to search"
	}

	searchSection := searchStyle.Render(searchPrompt + m.textInput.View())

	// Show filtered count if searching
	statusLine := ""
	if m.textInput.Value() != "" {
		totalItems := len(m.allChannels)
		filteredItems := len(m.list.Items())
		statusLine = fmt.Sprintf("Showing %d of %d channels", filteredItems, totalItems)
	}

	// Combine list view with search section
	content := m.list.View()
	if statusLine != "" {
		content += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(statusLine)
	}
	content += "\n" + searchSection

	return "\n" + content
}

// RunChannelBrowserTUI creates and runs the channel browser TUI
func RunChannelBrowserTUI(channels []Channel) error {
	// Convert channels to list items
	items := make([]list.Item, len(channels))
	for i, channel := range channels {
		items[i] = channelItem{channel: channel}
	}

	// Create list model with minimal initial size - will be updated when we get window size
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("📺 %s Channels (%d total)", AppName, len(channels))
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false) // We'll handle filtering ourselves
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// Create text input for search
	ti := textinput.New()
	ti.Placeholder = "Type to search channels..."
	ti.CharLimit = 50

	// Create model
	m := channelBrowserModel{
		list:        l,
		textInput:   ti,
		allChannels: channels,
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Handle the user's choice
	if finalModel, ok := finalModel.(channelBrowserModel); ok && finalModel.choice != nil {
		return handleChannelSelection(*finalModel.choice)
	}

	return nil
}

// Handle channel selection from the TUI
func handleChannelSelection(channel Channel) error {
	fmt.Printf("\n🎬 Selected Channel: %s\n", channel.Title)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Display channel details
	if channel.Summary != "" {
		fmt.Printf("📝 Summary: %s\n", channel.Summary)
	}
	if channel.TextInfo != "" {
		fmt.Printf("ℹ️  Info: %s\n", channel.TextInfo)
	}
	if channel.Avatar != "" {
		fmt.Printf("🖼️  Avatar: %s\n", channel.Avatar)
	}
	if channel.IsLive {
		fmt.Printf("🔴 Status: LIVE\n")
	}

	// Show social links
	if channel.Links != nil {
		fmt.Printf("\n🔗 Links:\n")
		if channel.Links.Website != "" {
			fmt.Printf("   🌐 Website: %s\n", channel.Links.Website)
		}
		if channel.Links.Twitter != "" {
			fmt.Printf("   🐦 Twitter: %s\n", channel.Links.Twitter)
		}
		if channel.Links.Facebook != "" {
			fmt.Printf("   📘 Facebook: %s\n", channel.Links.Facebook)
		}
		if channel.Links.Telegram != "" {
			fmt.Printf("   📱 Telegram: %s\n", channel.Links.Telegram)
		}
		if channel.Links.Gab != "" {
			fmt.Printf("   🗣️  Gab: %s\n", channel.Links.Gab)
		}
		if channel.Links.Minds != "" {
			fmt.Printf("   🧠 Minds: %s\n", channel.Links.Minds)
		}
		if channel.Links.SubscribeStar != "" {
			fmt.Printf("   ⭐ SubscribeStar: %s\n", channel.Links.SubscribeStar)
		}
	}

	fmt.Printf("\n🆔 Channel ID: %s\n", channel.ID)

	// Get video count for this channel
	_, err := GetDB()
	if err == nil {
		videoCount, err := CountVideosForChannel(channel.ID)
		if err == nil {
			fmt.Printf("📹 Videos in database: %d\n", videoCount)
		}
	}

	// Ask what to do next
	fmt.Printf("\nWhat would you like to do?\n")
	fmt.Printf("1. View videos for this channel\n")
	fmt.Printf("2. Fetch latest videos from API\n")
	fmt.Printf("3. Back to channel list\n")
	fmt.Printf("0. Exit\n")
	fmt.Print("\nChoice (0-3): ")

	var choice int
	fmt.Scanf("%d", &choice)

	switch choice {
	case 1:
		// View videos for this channel
		videos, err := GetChannelVideos(channel.ID, 50, 0)
		if err != nil {
			fmt.Printf("❌ Failed to get videos: %v\n", err)
			return err
		}

		if len(videos) == 0 {
			fmt.Printf("📭 No videos found for this channel.\n")
			fmt.Printf("💡 Try option 2 to fetch from API.\n")
		} else {
			return createVideoSelectionTUI(&channel, videos)
		}

	case 2:
		// Fetch latest videos from API
		fmt.Printf("🔄 Fetching videos for '%s'...\n", channel.Title)
		client := NewClient("https://api.banned.video/graphql")
		videos, err := client.FetchVideos(channel.ID, 0, 50)
		if err != nil {
			fmt.Printf("❌ Failed to fetch videos: %v\n", err)
			return err
		}
		fmt.Printf("✅ Fetched %d videos\n", len(videos))

		if len(videos) > 0 {
			return createVideoSelectionTUI(&channel, videos)
		}

	case 3:
		// Return to channel list - this could be implemented by returning a special error
		// or by restructuring the flow
		fmt.Println("Returning to channel list...")
		return nil

	case 0:
		fmt.Println("Goodbye! 👋")
		return nil

	default:
		fmt.Println("Invalid choice.")
	}

	return nil
}
