// Package picker is a full-terminal Azure subscription selector.
package picker

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/hdisk13/zxsubs/internal/azure"
)

const (
	headerLines = 4
	footerLines = 2
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("51"))

	filterLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244"))

	filterTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229"))

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(true)

	nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	disabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	currentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)
)

type model struct {
	all      []azure.Subscription
	filtered []azure.Subscription
	cursor   int
	offset   int
	filter   string
	width    int
	height   int
	chosen   *azure.Subscription
	canceled bool
}

// Run opens the full-screen picker. ok is false when the user cancels.
func Run(subs []azure.Subscription) (azure.Subscription, bool, error) {
	if !stdinIsTerminal() {
		return azure.Subscription{}, false, fmt.Errorf("zxsubs needs an interactive terminal")
	}
	p := tea.NewProgram(newModel(subs), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return azure.Subscription{}, false, err
	}
	m, ok := final.(model)
	if !ok {
		return azure.Subscription{}, false, fmt.Errorf("unexpected picker model")
	}
	if m.canceled || m.chosen == nil {
		return azure.Subscription{}, false, nil
	}
	return *m.chosen, true, nil
}

func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func newModel(subs []azure.Subscription) model {
	m := model{
		all:    subs,
		width:  80,
		height: 24,
	}
	m.applyFilter()
	for i, s := range m.filtered {
		if s.IsDefault {
			m.cursor = i
			break
		}
	}
	m.ensureVisible()
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureVisible()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.canceled = true
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}

		case "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.ensureVisible()
			}

		case "pgup":
			m.cursor -= m.visible()
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureVisible()

		case "pgdown":
			m.cursor += m.visible()
			if m.cursor > len(m.filtered)-1 {
				m.cursor = max(0, len(m.filtered)-1)
			}
			m.ensureVisible()

		case "home":
			m.cursor = 0
			m.ensureVisible()

		case "end":
			m.cursor = max(0, len(m.filtered)-1)
			m.ensureVisible()

		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.filtered) {
				chosen := m.filtered[m.cursor]
				m.chosen = &chosen
				return m, tea.Quit
			}

		case "backspace":
			if m.filter != "" {
				r := []rune(m.filter)
				m.filter = string(r[:len(r)-1])
				m.applyFilter()
			}

		case "ctrl+u":
			if m.filter != "" {
				m.filter = ""
				m.applyFilter()
			}

		default:
			if msg.Type == tea.KeyRunes {
				m.filter += string(msg.Runes)
				m.applyFilter()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("zxsubs"))
	b.WriteString("  ")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%d/%d subscriptions", len(m.filtered), len(m.all))))
	b.WriteByte('\n')

	if m.filter == "" {
		b.WriteString(filterLabelStyle.Render("Filter: ") + mutedStyle.Render("type to search"))
	} else {
		b.WriteString(filterLabelStyle.Render("Filter: ") + filterTextStyle.Render(m.filter) + "█")
	}
	b.WriteByte('\n')
	b.WriteByte('\n')

	vis := m.visible()
	if len(m.filtered) == 0 {
		b.WriteString(emptyStyle.Render("No subscriptions match that filter."))
		b.WriteByte('\n')
	} else {
		end := min(m.offset+vis, len(m.filtered))
		for i := m.offset; i < end; i++ {
			b.WriteString(m.renderRow(i))
			b.WriteByte('\n')
		}
	}

	used := headerLines + min(vis, max(1, len(m.filtered)))
	for i := used; i < m.height-footerLines; i++ {
		b.WriteByte('\n')
	}

	b.WriteString(helpStyle.Render("↑/↓ move   enter select   type to filter   q/esc cancel"))
	b.WriteByte('\n')
	return b.String()
}

func (m model) renderRow(i int) string {
	sub := m.filtered[i]
	nameWidth := max(12, m.width-40)
	name := padName(truncate(sub.Name, nameWidth), nameWidth)
	badge := "       "
	if sub.IsDefault {
		badge = "current"
	}
	line := fmt.Sprintf("%s  %s  …%s  %s  tenant …%s",
		name, badge, azure.ShortID(sub.ID), sub.State, azure.ShortID(sub.TenantID))

	if i == m.cursor {
		return selectedRowStyle.Width(max(0, m.width)).Render("❯ " + line)
	}

	style := nameStyle
	if !strings.EqualFold(sub.State, "Enabled") {
		style = disabledStyle
	}
	if sub.IsDefault {
		return "  " + style.Render(name) + "  " + currentStyle.Render("current") +
			mutedStyle.Render(fmt.Sprintf("  …%s  %s  tenant …%s",
				azure.ShortID(sub.ID), sub.State, azure.ShortID(sub.TenantID)))
	}
	return "  " + style.Render(line)
}

func padName(name string, width int) string {
	r := []rune(name)
	if len(r) >= width {
		return name
	}
	return name + strings.Repeat(" ", width-len(r))
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width == 1 {
		return string(r[:1])
	}
	return string(r[:width-1]) + "…"
}

func (m *model) applyFilter() {
	var keepID string
	if m.cursor >= 0 && m.cursor < len(m.filtered) {
		keepID = m.filtered[m.cursor].ID
	}

	next := make([]azure.Subscription, 0, len(m.all))
	for _, s := range m.all {
		if Match(m.filter, s.Name, s.ID, s.State, s.TenantID) {
			next = append(next, s)
		}
	}
	m.filtered = next

	m.cursor = 0
	if keepID != "" {
		for i, s := range m.filtered {
			if s.ID == keepID {
				m.cursor = i
				break
			}
		}
	}
	m.ensureVisible()
}

func (m *model) ensureVisible() {
	vis := m.visible()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+vis {
		m.offset = m.cursor - vis + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m model) visible() int {
	return max(1, m.height-headerLines-footerLines)
}
