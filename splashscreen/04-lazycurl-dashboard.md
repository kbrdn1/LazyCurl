# 🚀 LazyCurl — Dashboard & Splashscreen

> Documentation pour l'implémentation du dashboard de démarrage style LazyVim

---

## 📋 Table des matières

1. [Dashboard Workspace Selector](#-dashboard-workspace-selector)
2. [Déclinaisons de splashscreen](#-déclinaisons-de-splashscreen)
3. [Librairies d'animation](#-librairies-danimation)
4. [Implémentation Go/Bubble Tea](#-implémentation-gobubble-tea)
5. [Issue Template](#-issue-template)

---

## 🏠 Dashboard Workspace Selector

### Design LazyVim-like

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                                                                             │
│         ░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░          │
│                                                                             │
│                  ██╗       ██████╗                                          │
│                  ██║      ██╔════╝                                          │
│                  ██║      ██║         LazyCurl                              │
│                  ██║      ██║                                               │
│                  ███████╗ ╚██████╗    ⚡ HTTP TUI Client                    │
│                  ╚══════╝  ╚═════╝                                          │
│                                                                             │
│         ◀◀◀░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░           │
│                                                                             │
│                                                                             │
│           ┌─────────────────────────────────────────────────┐               │
│           │                                                 │               │
│           │    Recent Workspaces                            │               │
│           │                                                 │               │
│           │    ▸ 󰉋  my-api-project       ~/dev/my-api       │               │
│           │      󰉋  e-commerce-backend   ~/work/shop        │               │
│           │      󰉋  school-management    ~/projects/ecole   │               │
│           │      󰉋  card-game-api        ~/dev/nuxt-game    │               │
│           │                                                 │               │
│           └─────────────────────────────────────────────────┘               │
│                                                                             │
│                                                                             │
│           ┌─────────────────────────────────────────────────┐               │
│           │                                                 │               │
│           │     Find Workspace          f                   │               │
│           │     New Workspace           n                   │               │
│           │     Open Current Dir         .                  │               │
│           │     Recent Files             r                  │               │
│           │     Config                   c                  │               │
│           │    󰒲  Lazy                   l                  │               │
│           │     Quit                     q                  │               │
│           │                                                 │               │
│           └─────────────────────────────────────────────────┘               │
│                                                                             │
│                                                                             │
│                                     v0.1.0                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Layout responsive (80 cols)

```
┌──────────────────────────────────────────────────────────────────────┐
│                                                                      │
│      ░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░              │
│                                                                      │
│               ██╗       ██████╗                                      │
│               ██║      ██╔════╝    LazyCurl                          │
│               ██║      ██║         ⚡ HTTP TUI Client                │
│               ██║      ██║                                           │
│               ███████╗ ╚██████╗                                      │
│               ╚══════╝  ╚═════╝                                      │
│                                                                      │
│      ◀◀◀░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░               │
│                                                                      │
│        ┌────────────────────────────────────────────────┐            │
│        │  Recent Workspaces                             │            │
│        │                                                │            │
│        │  ▸ 󰉋  my-api-project         ~/dev/my-api     │            │
│        │    󰉋  e-commerce-backend     ~/work/shop      │            │
│        │    󰉋  school-management      ~/projects/ecole │            │
│        └────────────────────────────────────────────────┘            │
│                                                                      │
│        ┌────────────────────────────────────────────────┐            │
│        │   Find Workspace   f     Recent Files  r     │            │
│        │   New Workspace    n     Config        c     │            │
│        │   Open Current     .    󰒲  Lazy         l     │            │
│        │   Quit             q                          │            │
│        └────────────────────────────────────────────────┘            │
│                                                                      │
│                      v0.1.0  ·  MIT License                          │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### Layout compact (60 cols)

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│   ░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░          │
│                                                      │
│        ██╗       ██████╗    LazyCurl                │
│        ██║      ██╔════╝    ⚡ HTTP TUI             │
│        ██║      ██║                                  │
│        ██║      ██║                                  │
│        ███████╗ ╚██████╗                            │
│        ╚══════╝  ╚═════╝                            │
│                                                      │
│   ◀◀◀░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░          │
│                                                      │
│   ┌────────────────────────────────────────┐        │
│   │  Workspaces                            │        │
│   │  ▸ my-api-project                      │        │
│   │    e-commerce-backend                  │        │
│   │    school-management                   │        │
│   └────────────────────────────────────────┘        │
│                                                      │
│    f Find   n New   . Open   q Quit               │
│                                                      │
│                    v0.1.0                            │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

## 🎨 Déclinaisons de Splashscreen

### Style 1: Speed Gradient (Recommandé)

```
░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░▶▶▶

░       ██╗       ██████╗
░▒      ██║      ██╔════╝
░▒▓     ██║      ██║         LazyCurl
░▒      ██║      ██║
░       ███████╗ ╚██████╗
        ╚══════╝  ╚═════╝
            ⚡ HTTP TUI Client ⚡

◀◀◀░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░
```

### Style 2: Motion Blur

```
░░░▒▒▒▓▓▓███━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━▶▶▶

        ██╗       ██████╗
        ██║      ██╔════╝
        ██║      ██║         LazyCurl
        ██║      ██║
        ███████╗ ╚██████╗
        ╚══════╝  ╚═════╝
                HTTP TUI Client

◀◀◀━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━███▓▓▓▒▒▒░░░
```

### Style 3: Particles

```
 ·  ∙  • ━━━━━━━━━━━━━━━━━━━━━━━━━━━ •  ∙  · ▶▶▶

        ██╗       ██████╗
  ·  ∙  ██║      ██╔════╝
 ∙  •   ██║      ██║         LazyCurl
  ·  ∙  ██║      ██║
        ███████╗ ╚██████╗
        ╚══════╝  ╚═════╝
                HTTP TUI Client

 ◀◀◀ ·  ∙  • ━━━━━━━━━━━━━━━━━━━━━━━ •  ∙  ·
```

### One-liners (header compact)

```
░▒▓█━━━▶  LC  ◀━━━█▓▒░  LazyCurl v0.1.0
```

```
⚡░▒▓█━▶ LazyCurl ◀━█▓▒░⚡  HTTP TUI Client
```

```
·∙•━━━▶  LC  ◀━━━•∙·  LazyCurl
```

---

## 📚 Librairies d'Animation

### 🎼 Harmonica (Recommandé)

**Repository:** `github.com/charmbracelet/harmonica`

**Description:** Librairie d'animation spring légère et efficace, créée par Charmbracelet (même équipe que Bubble Tea). Parfaite pour des animations fluides et naturelles.

**Avantages:**

- ✅ Même écosystème que Bubble Tea/Lipgloss
- ✅ Très légère (pas de dépendances externes)
- ✅ Animations physics-based (spring damping)
- ✅ Framework-agnostic
- ✅ Fonctionne parfaitement en TUI

**Installation:**

```bash
go get github.com/charmbracelet/harmonica
```

**Exemple d'utilisation avec Bubble Tea:**

```go
package main

import (
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/harmonica"
    "github.com/charmbracelet/lipgloss"
)

const (
    fps       = 60
    frequency = 7.0
    damping   = 0.15
)

type frameMsg time.Time

func animate() tea.Cmd {
    return tea.Tick(time.Second/fps, func(t time.Time) tea.Msg {
        return frameMsg(t)
    })
}

type model struct {
    x      float64
    xVel   float64
    spring harmonica.Spring
    width  int
}

func initialModel() model {
    return model{
        spring: harmonica.NewSpring(harmonica.FPS(fps), frequency, damping),
    }
}

func (m model) Init() tea.Cmd {
    return tea.Sequentially(
        tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg { return nil }),
        animate(),
    )
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        return m, nil
    case frameMsg:
        targetX := float64(m.width - 12)
        m.x, m.xVel = m.spring.Update(m.x, m.xVel, targetX)
        return m, animate()
    case tea.KeyMsg:
        return m, tea.Quit
    }
    return m, nil
}

func (m model) View() string {
    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#89b4fa")).
        Background(lipgloss.Color("#1e1e2e"))

    return style.Render("█") // Animated block
}
```

**Paramètres de spring:**

| Paramètre | Description | Valeur typique |
|-----------|-------------|----------------|
| `frequency` | Vitesse de l'animation (angular frequency) | 5.0 - 10.0 |
| `damping` | Rebond de l'animation (damping ratio) | 0.1 - 1.0 |

- **damping < 1.0** → Under-damped (rebond, oscillation)
- **damping = 1.0** → Critically-damped (pas de rebond, plus rapide)
- **damping > 1.0** → Over-damped (pas de rebond, plus lent)

---

### 🫧 Bubbles Progress (Built-in)

**Repository:** `github.com/charmbracelet/bubbles/progress`

Le composant progress de Bubbles utilise Harmonica en interne pour les animations fluides.

```go
import "github.com/charmbracelet/bubbles/progress"

// Progress bar avec animation
p := progress.New(
    progress.WithDefaultGradient(),
    progress.WithWidth(40),
)
```

---

### 🎬 Animation Patterns pour LazyCurl

#### Pattern 1: Fade-in du logo

```go
type splashModel struct {
    opacity  float64
    opacVel  float64
    spring   harmonica.Spring
    revealed int // Nombre de lignes révélées
}

func (m splashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case frameMsg:
        m.opacity, m.opacVel = m.spring.Update(m.opacity, m.opacVel, 1.0)
        m.revealed = int(m.opacity * float64(logoHeight))
        return m, animate()
    }
    return m, nil
}
```

#### Pattern 2: Sliding gradient

```go
type gradientState struct {
    offset    float64
    offsetVel float64
    spring    harmonica.Spring
}

func (g *gradientState) tick() {
    targetOffset := 1.0 // Position finale
    g.offset, g.offsetVel = g.spring.Update(g.offset, g.offsetVel, targetOffset)
}

func (g gradientState) render(width int) string {
    chars := []rune{'░', '▒', '▓', '█'}
    visibleWidth := int(g.offset * float64(width))

    var result strings.Builder
    for i := 0; i < visibleWidth && i < 4; i++ {
        result.WriteRune(chars[i])
    }
    return result.String()
}
```

#### Pattern 3: Pulsing (idle animation)

```go
type pulseModel struct {
    phase float64
    tick  int
}

func (m pulseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case frameMsg:
        m.tick++
        m.phase = 0.5 + 0.3*math.Sin(float64(m.tick)*0.05)
        return m, animate()
    }
    return m, nil
}
```

---

## 🔧 Implémentation Go/Bubble Tea

### Structure de fichiers suggérée

```
internal/
└── ui/
    ├── dashboard/
    │   ├── dashboard.go      # Main dashboard model
    │   ├── splash.go         # Splashscreen component
    │   ├── workspaces.go     # Workspace list component
    │   └── actions.go        # Action buttons component
    ├── components/
    │   └── animated/
    │       ├── gradient.go   # Animated gradient borders
    │       ├── logo.go       # Animated logo reveal
    │       └── pulse.go      # Pulsing animations
    └── styles/
        └── catppuccin.go     # Theme colors
```

### Model principal du Dashboard

```go
package dashboard

import (
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/harmonica"
    "github.com/charmbracelet/lipgloss"
)

type State int

const (
    StateSplash State = iota
    StateReady
)

type Model struct {
    state       State
    width       int
    height      int

    // Animation
    spring      harmonica.Spring
    logoOpacity float64
    logoVel     float64

    // Components
    workspaces  []Workspace
    selected    int

    // Timing
    splashStart time.Time
}

type Workspace struct {
    Name string
    Path string
}

func New() Model {
    return Model{
        state:       StateSplash,
        spring:      harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.5),
        splashStart: time.Now(),
        workspaces: []Workspace{
            {Name: "my-api-project", Path: "~/dev/my-api"},
            {Name: "e-commerce-backend", Path: "~/work/shop"},
            {Name: "school-management", Path: "~/projects/ecole"},
        },
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        tea.EnterAltScreen,
        animate(),
    )
}

func animate() tea.Cmd {
    return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

type tickMsg time.Time

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil

    case tickMsg:
        if m.state == StateSplash {
            m.logoOpacity, m.logoVel = m.spring.Update(m.logoOpacity, m.logoVel, 1.0)

            // Transition après 2s ou animation complète
            if time.Since(m.splashStart) > 2*time.Second && m.logoOpacity > 0.95 {
                m.state = StateReady
            }
            return m, animate()
        }
        return m, nil

    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "j", "down":
            if m.selected < len(m.workspaces)-1 {
                m.selected++
            }
        case "k", "up":
            if m.selected > 0 {
                m.selected--
            }
        case "enter", " ":
            // Open workspace
            return m, openWorkspace(m.workspaces[m.selected])
        case "n":
            // New workspace
        case "f":
            // Find workspace
        }
    }
    return m, nil
}

func (m Model) View() string {
    if m.state == StateSplash {
        return m.renderSplash()
    }
    return m.renderDashboard()
}

func (m Model) renderSplash() string {
    // Gradient animé basé sur logoOpacity
    gradientWidth := int(m.logoOpacity * 40)
    gradient := renderGradient(gradientWidth)

    logo := renderLogo(m.logoOpacity)

    return lipgloss.Place(
        m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        lipgloss.JoinVertical(lipgloss.Center,
            gradient,
            "",
            logo,
            "",
            reverseGradient(gradientWidth),
        ),
    )
}

func (m Model) renderDashboard() string {
    // Full dashboard with workspaces
    header := renderCompactHeader()
    workspaceList := m.renderWorkspaces()
    actions := renderActions()
    footer := renderFooter()

    content := lipgloss.JoinVertical(lipgloss.Center,
        header,
        "",
        workspaceList,
        "",
        actions,
        "",
        footer,
    )

    return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderWorkspaces() string {
    var items []string

    for i, ws := range m.workspaces {
        cursor := "  "
        if i == m.selected {
            cursor = "▸ "
        }

        style := lipgloss.NewStyle()
        if i == m.selected {
            style = style.Foreground(lipgloss.Color("#89b4fa")).Bold(true)
        } else {
            style = style.Foreground(lipgloss.Color("#cdd6f4"))
        }

        item := fmt.Sprintf("%s󰉋  %-20s  %s", cursor, ws.Name, ws.Path)
        items = append(items, style.Render(item))
    }

    box := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#45475a")).
        Padding(1, 2).
        Width(50)

    title := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#cdd6f4")).
        Bold(true).
        Render("Recent Workspaces")

    return box.Render(lipgloss.JoinVertical(lipgloss.Left,
        title,
        "",
        strings.Join(items, "\n"),
    ))
}
```

### Couleurs Catppuccin Mocha

```go
package styles

import "github.com/charmbracelet/lipgloss"

var (
    // Base colors
    Base     = lipgloss.Color("#1e1e2e")
    Mantle   = lipgloss.Color("#181825")
    Crust    = lipgloss.Color("#11111b")

    // Surface colors
    Surface0 = lipgloss.Color("#313244")
    Surface1 = lipgloss.Color("#45475a")
    Surface2 = lipgloss.Color("#585b70")

    // Overlay colors
    Overlay0 = lipgloss.Color("#6c7086")
    Overlay1 = lipgloss.Color("#7f849c")
    Overlay2 = lipgloss.Color("#9399b2")

    // Text colors
    Text     = lipgloss.Color("#cdd6f4")
    Subtext0 = lipgloss.Color("#a6adc8")
    Subtext1 = lipgloss.Color("#bac2de")

    // Accent colors
    Blue     = lipgloss.Color("#89b4fa")
    Lavender = lipgloss.Color("#b4befe")
    Mauve    = lipgloss.Color("#cba6f7")
    Teal     = lipgloss.Color("#94e2d5")
    Green    = lipgloss.Color("#a6e3a1")
    Yellow   = lipgloss.Color("#f9e2af")
    Peach    = lipgloss.Color("#fab387")
    Red      = lipgloss.Color("#f38ba8")
)

// Gradient chars for animations
var GradientChars = []rune{'░', '▒', '▓', '█'}

var GradientColors = []lipgloss.Color{
    Surface0, Surface1, Surface2, Blue,
}
```

---

## 📝 Issue Template

### Title: `✨ feat(ui): Add animated dashboard with workspace selector`

### Description

```markdown
## 📋 Summary

Implement an animated dashboard/splashscreen at startup, inspired by LazyVim's dashboard design. This will provide a welcoming entry point to LazyCurl with quick access to recent workspaces.

## 🎯 Goals

- [ ] Animated splashscreen with LazyCurl branding
- [ ] Recent workspaces list with quick selection
- [ ] Keyboard shortcuts for common actions
- [ ] Smooth spring-based animations using Harmonica
- [ ] Responsive layout (adapt to terminal size)

## 🎨 Design

### Splashscreen (2s duration)
- Animated gradient borders (speed gradient style)
- Logo fade-in with spring animation
- Catppuccin Mocha color scheme

### Dashboard
- Header: Compact logo with version
- Main: Recent workspaces list (vim navigation j/k)
- Actions: Quick access buttons (f, n, ., r, c, l, q)
- Footer: Version, license info

## 📦 Dependencies

### New dependency
```go
go get github.com/charmbracelet/harmonica
```

Harmonica is a lightweight spring animation library from Charmbracelet (same ecosystem as Bubble Tea). It's efficient and perfect for TUI animations.

**Why Harmonica?**

- Same team as Bubble Tea/Lipgloss
- Physics-based animations (natural feel)
- Minimal overhead (~60fps with low CPU)
- Already used by bubbles/progress component

## 📁 Files to create/modify

### New files

- `internal/ui/dashboard/dashboard.go` - Main dashboard model
- `internal/ui/dashboard/splash.go` - Splashscreen component
- `internal/ui/dashboard/workspaces.go` - Workspace list
- `internal/ui/dashboard/actions.go` - Action buttons
- `internal/ui/components/animated/gradient.go` - Animated gradients
- `internal/ui/components/animated/logo.go` - Logo animations

### Modified files

- `cmd/lazycurl/main.go` - Start with dashboard
- `go.mod` - Add harmonica dependency

## 🔧 Implementation Notes

### Animation settings

```go
const (
    fps       = 60
    frequency = 6.0  // Animation speed
    damping   = 0.5  // Bounce (< 1 = bouncy, 1 = no bounce)
)
```

### State machine

```
StateSplash (2s) → StateReady (dashboard visible)
                 ↓
              User selects workspace
                 ↓
              StateWorking (main app)
```

### Responsive breakpoints

- XL: > 120 cols (full animations, detailed layout)
- L: 80-120 cols (standard layout)
- M: 60-80 cols (compact layout)
- S: < 60 cols (minimal, text-only)

## 🎬 Animation sequence

1. **Frame 0-30 (0.5s)**: Gradient borders slide in
2. **Frame 30-90 (1s)**: Logo fades in with spring
3. **Frame 90-120 (0.5s)**: Hold + subtle pulse
4. **Frame 120+**: Transition to dashboard

## 📸 Mockups

See attached documentation for full mockups:

- `docs/dashboard/mockups.md`

## 🏷️ Labels

`enhancement` `ui` `animation` `good-first-issue`

## 📚 References

- [Harmonica docs](https://pkg.go.dev/github.com/charmbracelet/harmonica)
- [LazyVim dashboard](https://www.lazyvim.org/plugins/ui)
- [Bubbles progress](https://github.com/charmbracelet/bubbles/tree/master/progress)

```

---

## 📊 Performance Considerations

### Animation Budget

| Component | CPU Impact | Memory |
|-----------|------------|--------|
| Harmonica spring | ~0.1% | 64 bytes/spring |
| 60fps tick | ~1-2% | Minimal |
| Gradient render | ~0.5% | String allocs |
| **Total** | **< 3%** | **< 1KB** |

### Optimizations

1. **Skip frames on slow terminals:**
```go
if time.Since(lastRender) < time.Second/30 {
    return m, nil // Skip this frame
}
```

2. **Cache static content:**

```go
var cachedLogo string
func init() {
    cachedLogo = generateLogo()
}
```

3. **Disable animations in CI/non-interactive:**

```go
if !term.IsTerminal(int(os.Stdout.Fd())) {
    m.animationsEnabled = false
}
```

---

## ✅ Checklist pour l'implémentation

- [ ] Ajouter `harmonica` au `go.mod`
- [ ] Créer la structure de fichiers dashboard/
- [ ] Implémenter le splashscreen animé
- [ ] Implémenter la liste des workspaces
- [ ] Ajouter les raccourcis clavier
- [ ] Tester sur différentes tailles de terminal
- [ ] Documenter dans le README
- [ ] Ajouter tests unitaires
- [ ] Performance testing (< 3% CPU)
