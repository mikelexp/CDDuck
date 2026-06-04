package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	name  string
	isDir bool
}

type Model struct {
	currentDir string
	items      []Item
	cursor     int
	offset     int
	filter     string
	filtered   []Item
	width      int
	height     int
	selected   string
	focusName  string
	focusStack []string
}

var (
	dirStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	keyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
)

func main() {
	if hasVersionArg(os.Args[1:]) {
		fmt.Println(Version)
		return
	}

	dir, _ := os.Getwd()

	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cdduck: no terminal available:", err)
		os.Exit(1)
	}
	defer tty.Close()

	m := Model{currentDir: dir}
	m.loadItems()

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(tty))
	final, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if m, ok := final.(Model); ok && m.selected != "" {
		fmt.Println(m.selected)
	}
}

func hasVersionArg(args []string) bool {
	for _, arg := range args {
		if arg == "-version" || arg == "--version" {
			return true
		}
	}
	return false
}

func (m *Model) loadItems() {
	entries, err := os.ReadDir(m.currentDir)

	m.items = []Item{}
	if m.currentDir != "/" {
		m.items = append(m.items, Item{name: "..", isDir: true})
	}

	if err != nil {
		m.applyFilter()
		return
	}

	for _, e := range entries {
		name := e.Name()
		m.items = append(m.items, Item{name: name, isDir: e.IsDir()})
	}

	sort.Slice(m.items, func(i, j int) bool {
		if m.items[i].isDir != m.items[j].isDir {
			return m.items[i].isDir
		}
		return strings.ToLower(m.items[i].name) < strings.ToLower(m.items[j].name)
	})

	m.applyFilter()
	if m.focusName != "" {
		for i, it := range m.filtered {
			if it.name == m.focusName {
				m.cursor = i
				m.offset = 0
				m.ensureVisible()
				break
			}
		}
		m.focusName = ""
	}
}

func fuzzyMatch(s, pat string) bool {
	pi := 0
	for si := 0; si < len(s) && pi < len(pat); si++ {
		if s[si] == pat[pi] {
			pi++
		}
	}
	return pi == len(pat)
}

func (m *Model) applyFilter() {
	q := strings.ToLower(m.filter)
	m.filtered = nil
	for _, it := range m.items {
		if q == "" || fuzzyMatch(strings.ToLower(it.name), q) {
			m.filtered = append(m.filtered, it)
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureVisible()
}

func (m *Model) ensureVisible() {
	visible := m.contentHeight()
	if visible < 1 {
		visible = 1
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.offset+visible <= m.cursor {
		m.offset = m.cursor - visible + 1
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureVisible()
		return m, nil

	case tea.KeyMsg:
		k := msg.String()

		switch {
		case k == "ctrl+c" || k == "ctrl+q":
			return m, tea.Quit

		case k == "esc":
			if m.filter != "" {
				m.filter = ""
				m.cursor = 0
				m.offset = 0
				m.applyFilter()
				return m, nil
			}
			m.selected = m.currentDir
			return m, tea.Quit

		case k == "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}

		case k == "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.ensureVisible()
			}

		case k == "pgup":
			visible := max(1, m.contentHeight())
			m.cursor = max(0, m.cursor-visible)
			m.ensureVisible()

		case k == "pgdown":
			visible := max(1, m.contentHeight())
			m.cursor = min(len(m.filtered)-1, m.cursor+visible)
			m.ensureVisible()

		case k == "home":
			m.cursor = 0
			m.offset = 0

		case k == "end":
			m.cursor = len(m.filtered) - 1
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureVisible()

		case k == "enter":
			m.doEnter()

		case k == "alt+enter":
			m.doSelect()
			return m, tea.Quit

		case k == "ctrl+h":
			if home, err := os.UserHomeDir(); err == nil && home != "" {
				m.currentDir = home
				m.cursor = 0
				m.offset = 0
				m.filter = ""
				m.focusName = ""
				m.focusStack = nil
				m.loadItems()
			}

		case k == "backspace":
			if m.filter != "" {
				runes := []rune(m.filter)
				m.filter = string(runes[:len(runes)-1])
				m.applyFilter()
			} else if m.currentDir != "/" {
				if n := len(m.focusStack); n > 0 {
					m.focusName = m.focusStack[n-1]
					m.focusStack = m.focusStack[:n-1]
				} else {
					m.focusName = filepath.Base(m.currentDir)
				}
				m.currentDir = filepath.Dir(m.currentDir)
				m.cursor = 0
				m.offset = 0
				m.filter = ""
				m.loadItems()
			}

		case k == "space":
			m.filter += " "
			m.applyFilter()

		default:
			if msg.Type == tea.KeyRunes {
				m.filter += string(msg.Runes)
				m.applyFilter()
			}
		}
	}

	return m, nil
}

func (m *Model) doEnter() {
	if len(m.filtered) == 0 {
		return
	}
	it := m.filtered[m.cursor]
	if it.name == ".." {
		if n := len(m.focusStack); n > 0 {
			m.focusName = m.focusStack[n-1]
			m.focusStack = m.focusStack[:n-1]
		} else {
			m.focusName = filepath.Base(m.currentDir)
		}
		m.currentDir = filepath.Dir(m.currentDir)
	} else if it.isDir {
		m.focusStack = append(m.focusStack, it.name)
		m.currentDir = filepath.Join(m.currentDir, it.name)
	} else {
		return
	}
	m.cursor = 0
	m.offset = 0
	m.filter = ""
	m.loadItems()
}

func (m *Model) doSelect() {
	m.selected = m.currentDir
	if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		it := m.filtered[m.cursor]
		if it.isDir && it.name != ".." {
			m.selected = filepath.Join(m.currentDir, it.name)
		}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	innerW := m.width - 2
	if innerW < 10 {
		innerW = 10
	}
	contW := innerW - 4
	if contW < 1 {
		contW = 1
	}

	title := "  CDDuck  "
	titleW := utf8.RuneCountInString(title)
	dashTotal := innerW
	lDash := (dashTotal - titleW) / 2
	rDash := dashTotal - lDash - titleW
	frameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	topB := frameStyle.Render("╔") + frameStyle.Render(strings.Repeat("═", lDash)) + titleStyle.Render(title) + frameStyle.Render(strings.Repeat("═", rDash)) + frameStyle.Render("╗")

	vis := m.contentHeight()
	if vis < 1 {
		vis = 1
	}

	start := m.offset
	end := m.offset + vis
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	var b strings.Builder
	b.WriteString(topB)
	b.WriteByte('\n')

	emptyLine := frameStyle.Render("║") + strings.Repeat(" ", innerW) + frameStyle.Render("║") + "\n"

	b.WriteString(emptyLine)

	for i := start; i < end; i++ {
		it := m.filtered[i]
		text := it.name
		if it.isDir && text != ".." {
			text = "[" + text + "]"
		}
		tw := utf8.RuneCountInString(text)
		if tw > contW {
			text = string([]rune(text)[:contW])
			tw = contW
		}

		padding := strings.Repeat(" ", contW-tw)
		if i == m.cursor {
			line := "\033[44;37;1m  " + text + padding + "  \033[0m"
			b.WriteString(frameStyle.Render("║"))
			b.WriteString(line)
			b.WriteString(frameStyle.Render("║"))
			b.WriteByte('\n')
		} else {
			line := "  " + text + padding + "  "
			if it.isDir {
				line = dirStyle.Render(line)
			}
			b.WriteString(frameStyle.Render("║"))
			b.WriteString(line)
			b.WriteString(frameStyle.Render("║"))
			b.WriteByte('\n')
		}
	}

	for i := end - start; i < vis; i++ {
		b.WriteString(emptyLine)
	}

	b.WriteString(emptyLine)

	b.WriteString(frameStyle.Render("╚"))
	b.WriteString(frameStyle.Render(strings.Repeat("═", innerW)))
	b.WriteString(frameStyle.Render("╝"))
	b.WriteByte('\n')

	path := m.currentDir
	plen := utf8.RuneCountInString(path)
	if plen > m.width {
		path = "…" + string([]rune(path)[plen-m.width+1:])
	}
	b.WriteString(keyStyle.Render("Current path:"))
	b.WriteString(" ")
	b.WriteString(path)
	b.WriteByte('\n')

	b.WriteString(keyStyle.Render("Filter:"))
	b.WriteString(" ")
	b.WriteString(m.filter)
	b.WriteString("█")
	b.WriteByte('\n')
	b.WriteString(shortcutsLine(m.width))

	return b.String()
}

func (m Model) contentHeight() int {
	return m.height - 7
}

func shortcutsLine(width int) string {
	type segment struct {
		plain  string
		styled bool
	}

	parts := []segment{
		{"↑↓", true}, {" nav  ", false},
		{"PgUp/PgDn", true}, {" page  ", false},
		{"Home/End", true}, {" first/last  ", false},
		{"Enter", true}, {" open  ", false},
		{"Alt+Enter", true}, {" select  ", false},
		{"Backspace", true}, {" up/filter  ", false},
		{"Esc", true}, {" clear/exit  ", false},
		{"Ctrl+H", true}, {" home  ", false},
		{"Ctrl+C/Q", true}, {" quit", false},
	}

	var b strings.Builder
	visible := 0
	for _, part := range parts {
		count := utf8.RuneCountInString(part.plain)
		if width > 0 && visible+count > width {
			if width > 1 && visible < width {
				b.WriteString("…")
			}
			break
		}
		if part.styled {
			b.WriteString(keyStyle.Render(part.plain))
		} else {
			b.WriteString(part.plain)
		}
		visible += count
	}
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
