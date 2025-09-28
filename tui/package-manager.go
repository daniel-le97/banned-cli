package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type skrrt struct {
	packages      []string
	retryPackages []string
	word          string
	install       func(string) tea.Cmd
	index         int
	width         int
	height        int
	spinner       spinner.Model
	progress      progress.Model
	done          bool
	onError 	 func(error)
	onSuccess	 func(string)
}
type model struct {
	packages      []string
	retryPackages []string
	word          string
	install       func(string) tea.Cmd
	index         int
	width         int
	height        int
	spinner       spinner.Model
	progress      progress.Model
	done          bool
}

var (
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

func NewModel(word string, packages []string, InstallFunc func(string) tea.Cmd) model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return model{
		word:     word,
		install:  InstallFunc,
		packages: packages,
		spinner:  s,
		progress: p,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.install(m.packages[m.index]), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case InstalledPkgMsg:
		pkg := m.packages[m.index]
		if m.index >= len(m.packages)-1 {
			// Everything's been installed. We're done!
			m.done = true
			return m, tea.Sequence(
				tea.Printf("%s %s", checkMark, pkg), // print the last success message
				tea.Quit,                            // exit the program
			)
		}
		// Update progress bar
		m.index++
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.packages)))

		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s %s", checkMark, pkg), // print success message above our program
			m.install(m.packages[m.index]),      // download the next package
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	n := len(m.packages)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("Done! %sed %d videos.\n", m.word, n))
	}

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+pkgCount))

	pkgName := currentPkgNameStyle.Render(m.packages[m.index])
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(fmt.Sprintf("%s ing ", m.word) + pkgName)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+pkgCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + pkgCount
}

type InstalledPkgMsg string

type errMsg struct{ err error }

type error interface {
	Error() string
}

func (e errMsg) Error() string { return e.err.Error() }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
