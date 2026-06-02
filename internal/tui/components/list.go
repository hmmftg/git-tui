package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item in the list
type ListItem struct {
	title       string
	description string
	data        interface{}
}

func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string  { return i.description }
func (i ListItem) FilterValue() string { return i.title }
func (i ListItem) Data() interface{}   { return i.data }

// ListModel wraps bubbles/list with project-specific styling
type ListModel struct {
	list   list.Model
	styles ListStyles
}

// ListStyles holds the styles for the list component
type ListStyles struct {
	Title        lipgloss.Style
	Cursor       lipgloss.Style
	Item         lipgloss.Style
	SelectedItem lipgloss.Style
}

// DefaultListStyles creates default list styles from project Styles
func DefaultListStyles(baseStyles interface{}) ListStyles {
	// Accept interface{} to avoid circular import with tui package
	return ListStyles{
		Title:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7B68EE")),
		Cursor:       lipgloss.NewStyle().Foreground(lipgloss.Color("#00CED1")),
		Item:         lipgloss.NewStyle().PaddingLeft(2),
		SelectedItem: lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#FFD700")),
	}
}

// NewList creates a new list model with the given items
func NewList(items []string, width, height int, styles ListStyles) ListModel {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = ListItem{title: item}
	}

	l := list.New(listItems, list.NewDefaultDelegate(), width, height)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	// Apply custom styles to delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = styles.Item
	delegate.Styles.SelectedTitle = styles.SelectedItem
	delegate.Styles.NormalDesc = lipgloss.NewStyle()
	delegate.Styles.SelectedDesc = lipgloss.NewStyle()

	l.SetDelegate(delegate)

	return ListModel{
		list:   l,
		styles: styles,
	}
}

// NewListWithData creates a list with items that have associated data
func NewListWithData(items []ListItem, width, height int, styles ListStyles) ListModel {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	l := list.New(listItems, list.NewDefaultDelegate(), width, height)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	// Apply custom styles
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = styles.Item
	delegate.Styles.SelectedTitle = styles.SelectedItem
	delegate.Styles.NormalDesc = lipgloss.NewStyle()
	delegate.Styles.SelectedDesc = lipgloss.NewStyle()

	l.SetDelegate(delegate)

	return ListModel{
		list:   l,
		styles: styles,
	}
}

// Update handles messages and updates the list
func (l ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	return l, cmd
}

// View renders the list
func (l ListModel) View() string {
	return l.list.View()
}

// SelectedItem returns the currently selected item title
func (l ListModel) SelectedItem() string {
	item := l.list.SelectedItem()
	if item == nil {
		return ""
	}
	return item.(ListItem).Title()
}

// SelectedItemData returns the data associated with the selected item
func (l ListModel) SelectedItemData() interface{} {
	item := l.list.SelectedItem()
	if item == nil {
		return nil
	}
	return item.(ListItem).Data()
}

// SetItems updates the list items
func (l *ListModel) SetItems(items []string) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = ListItem{title: item}
	}
	l.list.SetItems(listItems)
}

// Cursor returns the current cursor index
func (l ListModel) Cursor() int {
	return l.list.Index()
}

// SetCursor sets the cursor to the specified index
func (l *ListModel) SetCursor(index int) {
	if index >= 0 && index < len(l.list.Items()) {
		l.list.Select(index)
	}
}

// Length returns the number of items in the list
func (l ListModel) Length() int {
	return len(l.list.Items())
}

// Title returns the item at the given index
func (l ListModel) Title(index int) string {
	items := l.list.Items()
	if index < 0 || index >= len(items) {
		return ""
	}
	return items[index].(ListItem).Title()
}

// IsEmpty returns true if the list has no items
func (l ListModel) IsEmpty() bool {
	return len(l.list.Items()) == 0
}

// ViewWithTitle renders the list with a title header
func (l ListModel) ViewWithTitle(title string) string {
	titleStyle := l.styles.Title
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		"",
		l.View(),
	)
}

// ViewWithCursor renders a simple list with custom cursor (for non-bubbles list use cases)
func ViewWithCursor(items []string, cursor int, styles ListStyles, currentBranch string, branchChecker func(string) bool) string {
	var lines []string

	for i, item := range items {
		cursorStr := "  "
		if i == cursor {
			cursorStr = styles.Cursor.Render("▸ ")
		}

		display := item
		if branchChecker != nil && branchChecker(item) {
			display = lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("%s (current)", item))
		}

		lines = append(lines, fmt.Sprintf("%s%s", cursorStr, display))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
