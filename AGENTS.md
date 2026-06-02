# Bubble Tea and Lipgloss - Agent Documentation

This document provides essential guidance for AI agents working with the Charm Bracelet TUI libraries (Bubble Tea and Lipgloss) in the gitflow-tui project.

## Project Context

**Versions Used:**
- `github.com/charmbracelet/bubbletea v1.3.10`
- `github.com/charmbracelet/lipgloss v1.1.1`
- `github.com/charmbracelet/bubbles v1.0.0` (component library)

**Import Aliases:**
```go
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/lipgloss"
```

---

## Bubble Tea Framework

### Core Architecture (The Elm Architecture)

Bubble Tea programs are built around three core components:

1. **Model** - The application state (any type, typically a struct)
2. **Update** - Handles messages and updates the model
3. **View** - Renders the UI based on current model state

### Model Interface

Every model must implement three methods:

```go
type Model interface {
    Init() Cmd       // Initial command to run
    Update(Msg) (Model, Cmd)  // Handle messages
    View() string    // Render the UI
}
```

### Example Model Structure

```go
type model struct {
    choices  []string    // items
    cursor   int         // cursor position
    selected map[int]struct{}  // selected items
}

func (m model) Init() tea.Cmd {
    return nil  // No initial I/O needed
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.cursor > 0 { m.cursor-- }
        case "down", "j":
            if m.cursor < len(m.choices)-1 { m.cursor++ }
        case "enter", "space":
            // Toggle selection
        }
    }
    return m, nil
}

func (m model) View() string {
    // Build UI string
    return "UI content here"
}
```

### Key Message Types

| Message | Description |
|---------|-------------|
| `tea.KeyMsg` | Keyboard input |
| `tea.KeyPressMsg` | Key press events (v2 style) |
| `tea.WindowSizeMsg` | Terminal resize (`msg.Width`, `msg.Height`) |
| `tea.Quit` | Command to exit program |

### Common Commands

```go
// Quit the program
return m, tea.Quit

// Batch multiple commands
return m, tea.Batch(cmd1, cmd2, cmd3)

// Send a message after delay
return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
    return tickMsg(t)
})
```

### Program Entry Point

```go
func main() {
    p := tea.NewProgram(initialModel(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
```

### Program Options

```go
tea.NewProgram(model,
    tea.WithAltScreen(),      // Use alternate screen buffer
    tea.WithMouse(),          // Enable mouse support
)
```

---

## Lipgloss Styling Library

### Style Definition

```go
import "github.com/charmbracelet/lipgloss"

var style = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    Padding(2, 4).
    Width(22)

lipgloss.Println(style.Render("Hello, World"))
```

### Color Support

```go
// ANSI 16 colors (4-bit)
lipgloss.Color("5")   // magenta
lipgloss.Color("9")   // red
lipgloss.Color("12")  // light blue

// ANSI 256 colors (8-bit)
lipgloss.Color("86")  // aqua
lipgloss.Color("201") // hot pink

// True Color (24-bit hex)
lipgloss.Color("#0000FF")
lipgloss.Color("#04B575")

// Named constants
lipgloss.Black, lipgloss.Red, lipgloss.Green, lipgloss.Yellow
lipgloss.Blue, lipgloss.Magenta, lipgloss.Cyan, lipgloss.White
lipgloss.BrightBlack, lipgloss.BrightRed, ...
```

### Text Formatting

```go
var style = lipgloss.NewStyle().
    Bold(true).
    Italic(true).
    Faint(true).
    Blink(true).
    Strikethrough(true).
    Underline(true).
    Reverse(true)
```

### Underline Styles

```go
UnderlineNone
UnderlineSingle
UnderlineDouble
UnderlineCurly
UnderlineDotted
UnderlineDashed

// Usage:
s := lipgloss.NewStyle().
    UnderlineStyle(lipgloss.UnderlineCurly).
    UnderlineColor(lipgloss.Color("#FF0000"))
```

### Block-Level Formatting

```go
// Padding
lipgloss.NewStyle().
    PaddingTop(2).
    PaddingRight(4).
    PaddingBottom(2).
    PaddingLeft(4)

// CSS-like shorthand
lipgloss.NewStyle().Padding(2)           // All sides
lipgloss.NewStyle().Padding(2, 4)        // Vertical, Horizontal
lipgloss.NewStyle().Padding(1, 4, 2)     // Top, Horizontal, Bottom
lipgloss.NewStyle().Padding(2, 4, 3, 1)  // Top, Right, Bottom, Left

// Margins (same patterns as padding)
lipgloss.NewStyle().Margin(2, 4)

// Custom padding/margin characters
lipgloss.NewStyle().
    Padding(1, 2).
    PaddingChar('·').
    Margin(1, 2).
    MarginChar('░')
```

### Alignment

```go
var style = lipgloss.NewStyle().
    Width(24).
    Align(lipgloss.Left).
    Align(lipgloss.Right).
    Align(lipgloss.Center)
```

### Borders

```go
// Predefined border styles
lipgloss.NormalBorder()
lipgloss.RoundedBorder()
lipgloss.ThickBorder()
lipgloss.DoubleBorder()

// Custom border
var myBorder = lipgloss.Border{
    Top:         "._.:*:",
    Bottom:      "._.:*:",
    Left:        "|*",
    Right:       "|*",
    TopLeft:     "*",
    TopRight:    "*",
    BottomLeft:  "*",
    BottomRight: "*",
}

// Usage
var style = lipgloss.NewStyle().
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("63")).
    BorderTop(true).
    BorderLeft(true)

// Border with gradient
lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForegroundBlend(lipgloss.Color("#FF0000"), lipgloss.Color("#0000FF"))
```

### Layout Utilities

```go
// Join horizontally
lipgloss.JoinHorizontal(lipgloss.Bottom, paragraphA, paragraphB)
lipgloss.JoinHorizontal(lipgloss.Center, paraA, paraB)
lipgloss.JoinHorizontal(0.2, paraA, paraB)  // 20% from top alignment

// Join vertically
lipgloss.JoinVertical(lipgloss.Center, paraA, paraB)

// Place in whitespace
lipgloss.PlaceHorizontal(80, lipgloss.Center, content)
lipgloss.PlaceVertical(30, lipgloss.Bottom, content)
lipgloss.Place(30, 80, lipgloss.Right, lipgloss.Bottom, content)
```

### Measurement

```go
width := lipgloss.Width(block)
height := lipgloss.Height(block)
w, h := lipgloss.Size(block)
```

### Style Inheritance & Copying

```go
// Copy by assignment (pure value type)
styleA := lipgloss.NewStyle().Foreground(lipgloss.Color("219"))
styleB := styleA                    // True copy
styleC := styleA.Blink(true)        // Copy with modifications

// Inheritance (only unset rules are inherited)
baseStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("229")).
    Background(lipgloss.Color("63"))

childStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("201")).
    Inherit(baseStyle)  // Only background is inherited

// Unsetting rules
var style = lipgloss.NewStyle().
    Bold(true).
    UnsetBold().
    Background(lipgloss.Color("227")).
    UnsetBackground()
```

### Rendering Output

```go
// Print functions (auto-downsample colors)
lipgloss.Print(content)
lipgloss.Println(content)
lipgloss.Printf("Hello %s", name)

// To specific output
lipgloss.Fprint(os.Stderr, content)
lipgloss.Fprintf(os.Stderr, "Error: %s", err)

// To string
result := lipgloss.Sprint(content)
result := lipgloss.Sprintf("Value: %d", value)

// Direct style rendering
style := lipgloss.NewStyle().Bold(true)
output := style.Render("text")
```

---

## Tables (lipgloss/table subpackage)

```go
import "github.com/charmbracelet/lipgloss/table"

rows := [][]string{
    {"Chinese", "您好", "你好"},
    {"Japanese", "こんにちは", "やあ"},
}

t := table.New().
    Border(lipgloss.NormalBorder()).
    BorderStyle(lipgloss.NewStyle().Foreground(purple)).
    StyleFunc(func(row, col int) lipgloss.Style {
        switch {
        case row == table.HeaderRow:
            return headerStyle
        case row%2 == 0:
            return evenRowStyle
        default:
            return oddRowStyle
        }
    }).
    Headers("LANGUAGE", "FORMAL", "INFORMAL").
    Rows(rows...)

lipgloss.Println(t)
```

### Table Border Types

```go
table.New().Border(lipgloss.NormalBorder())
table.New().Border(lipgloss.MarkdownBorder())  // Markdown style
table.New().Border(lipgloss.ASCIIBorder())     // ASCII art style
```

---

## Lists (lipgloss/list subpackage)

```go
import "github.com/charmbracelet/lipgloss/list"

// Basic list
l := list.New("A", "B", "C")

// Nested list
l := list.New(
    "A",
    list.New("Artichoke"),
    "B",
    list.New("Bananas", "Barley"),
)

// Styled list
enumeratorStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("99")).
    MarginRight(1)

itemStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("212"))

l := list.New("Glossier", "Nyx", "Mac").
    Enumerator(list.Roman).
    EnumeratorStyle(enumeratorStyle).
    ItemStyle(itemStyle)
```

### List Enumerators

- `list.Arabic` - 1, 2, 3...
- `list.Alphabet` - a, b, c...
- `list.Roman` - I, II, III...
- `list.Bullet` - • • •
- `list.Tree` - tree style

---

## Trees (lipgloss/tree subpackage)

```go
import "github.com/charmbracelet/lipgloss/tree"

// Basic tree
t := tree.Root(".").Child("A", "B", "C")

// Nested tree
t := tree.Root(".").
    Child("macOS").
    Child(
        tree.New().
            Root("Linux").
            Child("NixOS").
            Child("Arch Linux"),
    )

// Styled tree
enumeratorStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("63")).
    MarginRight(1)

rootStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

t := tree.Root("⁜ Makeup").
    Child("Glossier", "Nyx", "Mac").
    Enumerator(tree.RoundedEnumerator).
    EnumeratorStyle(enumeratorStyle).
    RootStyle(rootStyle).
    ItemStyle(itemStyle)
```

### Tree Enumerators

- `tree.DefaultEnumerator` - ├── └──
- `tree.RoundedEnumerator` - rounded corners

---

## Adaptive Colors (Dark/Light Mode)

### With Bubble Tea

```go
func (m model) Init() tea.Cmd {
    return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.BackgroundColorMsg:
        m.styles = newStyles(msg.IsDark())
    }
    return m, nil
}

func newStyles(bgIsDark bool) styles {
    lightDark := lipgloss.LightDark(bgIsDark)
    return styles{
        myStyle: lipgloss.NewStyle().Foreground(lightDark(
            lipgloss.Color("#f1f1f1"),  // Light mode color
            lipgloss.Color("#333333"),  // Dark mode color
        )),
    }
}
```

### Standalone

```go
hasDarkBG := lipgloss.HasDarkBackground(os.Stdin, os.Stderr)
lightDark := lipgloss.LightDark(hasDarkBG)
color := lightDark(
    lipgloss.Color("#C5ADF9"),  // For light backgrounds
    lipgloss.Color("#864EFF"),  // For dark backgrounds
)
```

---

## Color Utilities

```go
c := lipgloss.Color("#EB4268")

// Color manipulation
dark := lipgloss.Darken(c, 0.5)
light := lipgloss.Lighten(c, 0.35)
complement := lipgloss.Complementary(c)
withAlpha := lipgloss.Alpha(c, 0.2)

// Color blending (1D and 2D gradients)
colors := lipgloss.Blend1D(10, lipgloss.Color("#FF0000"), lipgloss.Color("#0000FF"))
colors := lipgloss.Blend2D(80, 24, 45.0, color1, color2, color3)
```

---

## Project-Specific Patterns

### Styles Pattern (from internal/tui/styles.go)

```go
type Styles struct {
    Title      lipgloss.Style
    Header     lipgloss.Style
    Footer     lipgloss.Style
    Menu       lipgloss.Style
    MenuItem   lipgloss.Style
    MenuActive lipgloss.Style
    // ... etc
}

func DefaultStyles() Styles {
    s := Styles{}
    
    s.Title = lipgloss.NewStyle().
        Bold(true).
        Foreground(PrimaryColor).
        Background(BackgroundColor).
        Padding(0, 1).
        Margin(0, 0, 1, 0).
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(PrimaryColor)
    
    // ... define other styles
    
    return s
}
```

### Theme Colors

```go
var (
    PrimaryColor    = lipgloss.Color("#7B68EE") // MediumSlateBlue
    SecondaryColor  = lipgloss.Color("#00CED1") // DarkTurquoise
    SuccessColor    = lipgloss.Color("#32CD32") // LimeGreen
    ErrorColor      = lipgloss.Color("#FF6347") // Tomato
    WarningColor    = lipgloss.Color("#FFD700") // Gold
    InfoColor       = lipgloss.Color("#87CEEB") // SkyBlue
    TextColor       = lipgloss.Color("#FFFFFF") // White
    DimmedColor     = lipgloss.Color("#808080") // Gray
    BackgroundColor = lipgloss.Color("#1a1a2e") // Dark blue
)
```

### Navigation Pattern (from internal/tui/app.go)

```go
// Message types for view navigation
type viewChangeMsg models.ViewType

// Navigation command
func (m *AppModel) navigateTo(view models.ViewType) tea.Cmd {
    return func() tea.Msg {
        return viewChangeMsg(view)
    }
}

// In Update method
case viewChangeMsg:
    m.CurrentView = models.ViewType(msg)
    // Initialize view-specific models
    switch m.CurrentView {
    case models.ViewStatus:
        m.statusModel = NewStatusModel(m.GitService, m.Styles)
        return m, m.statusModel.Init()
    }
```

### Multi-View Architecture

```go
type AppModel struct {
    CurrentView models.ViewType
    Styles      Styles
    Width       int
    Height      int
    
    // View models
    statusModel   *StatusModel
    commitModel   *CommitModel
    // ... etc
}

func (m *AppModel) View() string {
    var content string
    switch m.CurrentView {
    case models.ViewStatus:
        content = m.statusModel.View()
    case models.ViewCommit:
        content = m.commitModel.View()
    }
    
    return lipgloss.JoinVertical(
        lipgloss.Left,
        m.renderHeader(),
        content,
        m.renderFooter(),
    )
}
```

---

## Debugging Tips

### Logging to File

```go
if len(os.Getenv("DEBUG")) > 0 {
    f, err := tea.LogToFile("debug.log", "debug")
    if err != nil {
        fmt.Println("fatal:", err)
        os.Exit(1)
    }
    defer f.Close()
}
```

Then run: `tail -f debug.log`

### Delve Debugger

```bash
# Start headless debugger
dlv debug --headless --api-version=2 --listen=127.0.0.1:43000 .

# Connect from another terminal
dlv connect 127.0.0.1:43000
```

---

## Common Mistakes to Avoid

1. **Don't use `fmt.Print` in View()** - Always return strings from View()
2. **Remember that styles are value types** - Assignment creates copies
3. **Handle `tea.WindowSizeMsg` early** - Store dimensions for responsive layouts
4. **Use `tea.Batch()` for multiple commands** - Don't return multiple commands separately
5. **Check for nil models** - When routing to sub-models, verify they exist

---

## External Resources

- **Bubble Tea GitHub**: https://github.com/charmbracelet/bubbletea
- **Lipgloss GitHub**: https://github.com/charmbracelet/lipgloss
- **Bubbles (Components)**: https://github.com/charmbracelet/bubbles
- **Charm Libraries**: https://charm.land/libs/
- **Go Docs (Bubble Tea)**: https://pkg.go.dev/github.com/charmbracelet/bubbletea
- **Go Docs (Lipgloss)**: https://pkg.go.dev/github.com/charmbracelet/lipgloss

---

## Quick Reference Card

```go
// Imports
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/lipgloss"

// Colors
lipgloss.Color("#FF0000")           // Hex
lipgloss.Color("201")               // ANSI 256
lipgloss.Red                          // Named

// Style building
lipgloss.NewStyle().Bold(true).Foreground(color)

// Layout
lipgloss.JoinHorizontal(pos, a, b)
lipgloss.JoinVertical(pos, a, b)
lipgloss.Place(w, h, x, y, content)

// Program
tea.NewProgram(model)
p.Run()

// Messages
tea.KeyMsg
tea.WindowSizeMsg
tea.Quit

// Commands
return m, tea.Quit
return m, tea.Batch(cmd1, cmd2)
```
