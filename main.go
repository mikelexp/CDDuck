package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type Item struct {
	name  string
	isDir bool
}

type Model struct {
	currentDir    string
	items         []Item
	cursor        int
	offset        int
	filter        string
	filtered      []Item
	width         int
	height        int
	selected      string
	focusName     string
	focusStack    []string
	historyPath   string
	history       []string
	historyOpen   bool
	historyFilter string
	historyCursor int
	historyOffset int
	historyDirty  bool
}

const (
	ansiReset  = "\033[0m"
	ansiBlue   = "\033[34;1m"
	ansiCyan   = "\033[36;1m"
	ansiWhite  = "\033[37;1m"
	ansiBlueBg = "\033[44;37;1m"
)

func main() {
	dir, showVersion, openHistory, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "cdduck:", err)
		os.Exit(2)
	}
	if showVersion {
		fmt.Println(Version)
		return
	}

	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cdduck: no terminal available:", err)
		os.Exit(1)
	}
	defer tty.Close()

	historyPath, err := historyFilePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cdduck: cannot locate config directory:", err)
		os.Exit(1)
	}
	history, err := loadHistory(historyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cdduck: cannot read history:", err)
	}
	m := Model{currentDir: dir, historyPath: historyPath, history: history, historyOpen: openHistory}
	m.loadItems()

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(tty))
	final, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if m, ok := final.(Model); ok {
		if shouldRecordHistory(dir, m.selected) {
			m.history = promoteHistory(m.history, m.selected)
			m.historyDirty = true
		}
		if m.historyDirty {
			if err := saveHistory(m.historyPath, m.history); err != nil {
				fmt.Fprintln(os.Stderr, "cdduck: cannot save history:", err)
			}
		}
		if m.selected != "" {
			fmt.Println(m.selected)
		}
	}
}

func shouldRecordHistory(initialDir, selected string) bool {
	return selected != "" && filepath.Clean(selected) != filepath.Clean(initialDir)
}

func historyFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "cdduck", "history.json"), nil
}

func loadHistory(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var history []string
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}
	return history, nil
}

func saveHistory(path string, history []string) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".history-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func promoteHistory(history []string, selected string) []string {
	selected = filepath.Clean(selected)
	updated := []string{selected}
	seen := map[string]bool{selected: true}
	for _, path := range history {
		path = filepath.Clean(path)
		if !seen[path] {
			updated = append(updated, path)
			seen[path] = true
		}
	}
	return updated
}

func parseArgs(args []string) (string, bool, bool, error) {
	flags := flag.NewFlagSet("cdduck", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	showVersion := flags.Bool("version", false, "print version and exit")
	flags.BoolVar(showVersion, "V", false, "print version and exit")
	openHistory := flags.Bool("H", false, "open directory history")
	if err := flags.Parse(args); err != nil {
		return "", false, false, err
	}

	paths := flags.Args()
	if *showVersion {
		if len(paths) != 0 {
			return "", false, false, errors.New("--version does not accept a path")
		}
		return "", true, false, nil
	}
	if len(paths) > 1 {
		return "", false, false, errors.New("accepts at most one starting path")
	}

	dir := ""
	if len(paths) == 1 {
		var err error
		dir, err = filepath.Abs(paths[0])
		if err != nil {
			return "", false, false, err
		}
	} else {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", false, false, err
		}
	}

	info, err := os.Stat(dir)
	if err != nil {
		return "", false, false, err
	}
	if !info.IsDir() {
		return "", false, false, fmt.Errorf("%q is not a directory", dir)
	}
	return dir, false, *openHistory, nil
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
		m.ensureHistoryVisible()
		return m, nil

	case tea.KeyMsg:
		k := msg.String()
		if m.historyOpen {
			return m.updateHistory(k, msg)
		}

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
			m.historyOpen = true
			m.historyFilter = ""
			m.historyCursor = 0
			m.historyOffset = 0
			m.ensureHistoryVisible()

		case k == "alt+h":
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

func (m *Model) updateHistory(k string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filtered := m.filteredHistory()
	visible := m.historyVisibleRows()
	switch {
	case k == "ctrl+c" || k == "ctrl+q":
		return *m, tea.Quit
	case k == "ctrl+h":
		m.historyOpen = false
	case k == "esc":
		if m.historyFilter != "" {
			m.historyFilter = ""
			m.historyCursor = 0
			m.historyOffset = 0
		} else {
			m.historyOpen = false
		}
	case k == "up":
		if m.historyCursor > 0 {
			m.historyCursor--
		}
	case k == "down":
		if m.historyCursor < len(filtered)-1 {
			m.historyCursor++
		}
	case k == "ctrl+up":
		if m.historyCursor > 0 {
			if m.moveHistoryEntry(filtered, -1) {
				m.historyCursor--
			}
		}
		m.ensureHistoryVisible()
	case k == "ctrl+down":
		if m.historyCursor < len(filtered)-1 {
			if m.moveHistoryEntry(filtered, 1) {
				m.historyCursor++
			}
		}
		m.ensureHistoryVisible()
	case k == "pgup":
		m.historyCursor = max(0, m.historyCursor-visible)
	case k == "pgdown":
		m.historyCursor = min(len(filtered)-1, m.historyCursor+visible)
	case k == "home":
		m.historyCursor, m.historyOffset = 0, 0
	case k == "end":
		m.historyCursor = max(0, len(filtered)-1)
	case k == "ctrl+r":
		if len(filtered) > 0 {
			selected := filtered[m.historyCursor]
			for i, path := range m.history {
				if path == selected {
					m.history = append(m.history[:i], m.history[i+1:]...)
					m.historyDirty = true
					break
				}
			}
			m.ensureHistoryVisible()
		}
	case k == "enter" || k == "alt+enter":
		if len(filtered) > 0 {
			m.currentDir = filtered[m.historyCursor]
			m.cursor = 0
			m.offset = 0
			m.filter = ""
			m.focusName = ""
			m.focusStack = nil
			m.historyOpen = false
			m.loadItems()
			return *m, nil
		}
	case k == "backspace":
		if m.historyFilter != "" {
			runes := []rune(m.historyFilter)
			m.historyFilter = string(runes[:len(runes)-1])
			m.historyCursor, m.historyOffset = 0, 0
		}
	case k == "space":
		m.historyFilter += " "
		m.historyCursor, m.historyOffset = 0, 0
	default:
		if msg.Type == tea.KeyRunes {
			m.historyFilter += string(msg.Runes)
			m.historyCursor, m.historyOffset = 0, 0
		}
	}
	m.ensureHistoryVisible()
	return *m, nil
}

// moveHistoryEntry moves the selected filtered entry across one adjacent entry
// in the full history. With a filter active, hidden entries retain their
// positions relative to the visible entries.
func (m *Model) moveHistoryEntry(filtered []string, delta int) bool {
	if m.historyCursor < 0 || m.historyCursor >= len(filtered) || (delta != -1 && delta != 1) {
		return false
	}
	selected := filtered[m.historyCursor]
	index := -1
	for i, path := range m.history {
		if path == selected {
			index = i
			break
		}
	}
	if index < 0 {
		return false
	}
	filteredIndex := m.historyCursor + delta
	if filteredIndex < 0 || filteredIndex >= len(filtered) {
		return false
	}
	neighbor := filtered[filteredIndex]
	neighborIndex := -1
	for i, path := range m.history {
		if path == neighbor {
			neighborIndex = i
			break
		}
	}
	if neighborIndex < 0 {
		return false
	}
	m.history[index], m.history[neighborIndex] = m.history[neighborIndex], m.history[index]
	m.historyDirty = true
	return true
}

func (m Model) filteredHistory() []string {
	query := strings.ToLower(m.historyFilter)
	filtered := make([]string, 0, len(m.history))
	for _, path := range m.history {
		if query == "" || fuzzyMatch(strings.ToLower(path), query) {
			filtered = append(filtered, path)
		}
	}
	return filtered
}

func (m Model) historyVisibleRows() int {
	modalHeight := m.height * 80 / 100
	if modalHeight > m.height {
		modalHeight = m.height
	}
	if modalHeight < 1 {
		modalHeight = 1
	}
	return max(1, modalHeight-6)
}

func (m *Model) ensureHistoryVisible() {
	filteredLen := len(m.filteredHistory())
	if m.historyCursor >= filteredLen {
		m.historyCursor = max(0, filteredLen-1)
	}
	visible := m.historyVisibleRows()
	if m.historyCursor < m.historyOffset {
		m.historyOffset = m.historyCursor
	}
	if m.historyCursor >= m.historyOffset+visible {
		m.historyOffset = m.historyCursor - visible + 1
	}
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
	base := m.mainView()
	if !m.historyOpen {
		return base
	}
	return m.overlayHistory(base)
}

func (m Model) overlayHistory(base string) string {
	if m.width <= 0 || m.height <= 0 {
		return base
	}
	modalWidth := m.width * 80 / 100
	modalHeight := m.height * 80 / 100
	modalWidth = min(m.width, max(1, modalWidth))
	modalHeight = min(m.height, max(1, modalHeight))
	left := (m.width - modalWidth) / 2
	top := (m.height - modalHeight) / 2
	innerWidth := max(0, modalWidth-2)
	contentWidth := max(0, innerWidth-2)
	visible := max(1, modalHeight-6)
	filtered := m.filteredHistory()
	if m.historyCursor >= len(filtered) {
		m.historyCursor = max(0, len(filtered)-1)
	}

	title := fmt.Sprintf(" History (%d) ", len(filtered))
	if ansi.StringWidth(title) > innerWidth {
		title = truncateCellsWithEllipsis(title, innerWidth)
	}
	titleWidth := ansi.StringWidth(title)
	leftBorder := max(0, (innerWidth-titleWidth)/2)
	topLine := colorize("╔", ansiBlue) +
		colorize(strings.Repeat("═", leftBorder), ansiBlue) +
		colorize(title, ansiWhite) +
		colorize(strings.Repeat("═", max(0, innerWidth-leftBorder-titleWidth)), ansiBlue) +
		colorize("╗", ansiBlue)
	bottomLine := colorize("╚", ansiBlue) + colorize(strings.Repeat("═", innerWidth), ansiBlue) + colorize("╝", ansiBlue)
	frameLine := func(text string) string {
		text = truncateCellsWithEllipsis(text, contentWidth)
		pad := max(0, contentWidth-ansi.StringWidth(text))
		return colorize("║", ansiBlue) + " " + text + strings.Repeat(" ", pad) + " " + colorize("║", ansiBlue)
	}
	emptyLine := colorize("║", ansiBlue) + strings.Repeat(" ", innerWidth) + colorize("║", ansiBlue)

	modal := []string{topLine, emptyLine}
	for row := 0; row < visible; row++ {
		index := m.historyOffset + row
		if index >= len(filtered) {
			modal = append(modal, emptyLine)
			continue
		}
		line := frameLine(filtered[index])
		if index == m.historyCursor {
			selected := truncateCellsWithEllipsis(filtered[index], contentWidth)
			line = colorize("║", ansiBlue) + ansiBlueBg + " " + selected + strings.Repeat(" ", max(0, contentWidth-ansi.StringWidth(selected))) + " " + ansiReset + colorize("║", ansiBlue)
		}
		modal = append(modal, line)
	}
	if len(filtered) == 0 && visible > 0 {
		message := "No matching directories"
		if len(m.history) == 0 && m.historyFilter == "" {
			message = "History is empty"
		}
		modal[2] = frameLine(message)
	}
	modal = append(modal, emptyLine)
	filterText := "Filter: " + m.historyFilter + "█"
	modal = append(modal, frameLine(filterText))
	modal = append(modal, frameLine("↑↓ navigate · Ctrl+↑↓ move · Enter open · Ctrl+R remove · Esc close"))
	modal = append(modal, bottomLine)
	if len(modal) > modalHeight {
		modal = modal[:modalHeight]
	}

	lines := strings.Split(base, "\n")
	for len(lines) < m.height {
		lines = append(lines, "")
	}
	for row, modalLine := range modal {
		lineIndex := top + row
		if lineIndex < 0 || lineIndex >= len(lines) {
			continue
		}
		baseLine := lines[lineIndex]
		if ansi.StringWidth(baseLine) < m.width {
			baseLine += strings.Repeat(" ", m.width-ansi.StringWidth(baseLine))
		}
		lines[lineIndex] = ansi.Cut(baseLine, 0, left) + modalLine + ansi.Cut(baseLine, left+modalWidth, m.width)
	}
	return strings.Join(lines, "\n")
}

func (m Model) mainView() string {
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

	title := " CDDuck v" + Version + " "
	if utf8.RuneCountInString(title) > innerW {
		title = string([]rune(title)[:innerW])
	}
	titleW := utf8.RuneCountInString(title)
	dashTotal := innerW
	lDash := (dashTotal - titleW) / 2
	rDash := dashTotal - lDash - titleW
	topB := colorize("╔", ansiBlue) + colorize(strings.Repeat("═", lDash), ansiBlue) + colorize(title, ansiWhite) + colorize(strings.Repeat("═", rDash), ansiBlue) + colorize("╗", ansiBlue)

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

	emptyLine := colorize("║", ansiBlue) + strings.Repeat(" ", innerW) + colorize("║", ansiBlue) + "\n"

	b.WriteString(emptyLine)

	for i := start; i < end; i++ {
		it := m.filtered[i]
		text := it.name
		if it.isDir && text != ".." {
			text = "[" + text + "]"
		}
		tw := utf8.RuneCountInString(text)
		if tw > contW {
			text = truncateWithEllipsis(text, contW)
			tw = contW
		}

		padding := strings.Repeat(" ", contW-tw)
		if i == m.cursor {
			line := ansiBlueBg + "  " + text + padding + "  " + ansiReset
			b.WriteString(colorize("║", ansiBlue))
			b.WriteString(line)
			b.WriteString(colorize("║", ansiBlue))
			b.WriteByte('\n')
		} else {
			line := "  " + text + padding + "  "
			if it.isDir {
				line = colorize(line, ansiWhite)
			}
			b.WriteString(colorize("║", ansiBlue))
			b.WriteString(line)
			b.WriteString(colorize("║", ansiBlue))
			b.WriteByte('\n')
		}
	}

	for i := end - start; i < vis; i++ {
		b.WriteString(emptyLine)
	}

	b.WriteString(emptyLine)

	b.WriteString(colorize("╚", ansiBlue))
	b.WriteString(colorize(strings.Repeat("═", innerW), ansiBlue))
	b.WriteString(colorize("╝", ansiBlue))
	b.WriteByte('\n')

	path := m.currentDir
	plen := utf8.RuneCountInString(path)
	if plen > m.width {
		path = "…" + string([]rune(path)[plen-m.width+1:])
	}
	b.WriteString(colorize("Current path:", ansiBlue))
	b.WriteString(" ")
	b.WriteString(path)
	b.WriteByte('\n')

	b.WriteString(colorize("Filter:", ansiBlue))
	b.WriteString(" ")
	b.WriteString(m.filter)
	b.WriteString("█")
	b.WriteByte('\n')
	b.WriteString(shortcutsLine(m.width))

	return b.String()
}

func truncateWithEllipsis(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if width == 1 {
		return "…"
	}
	runes := []rune(text)
	return string(runes[:width-1]) + "…"
}

func truncateCellsWithEllipsis(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(text) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}
	return ansi.Cut(text, 0, width-1) + "…"
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
		{"Ctrl+H", true}, {" history  ", false},
		{"Alt+H", true}, {" home  ", false},
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
			b.WriteString(colorize(part.plain, ansiBlue))
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

func colorize(s, code string) string {
	return code + s + ansiReset
}
