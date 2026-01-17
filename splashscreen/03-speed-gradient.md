# ⚡ LazyCurl Splashscreen — Speed Gradient Intense

> Style puissant avec dégradé latéral évoquant la puissance et la rapidité

---

## 📐 Déclinaisons par taille

### XL — Full Banner (Terminal 120+ cols)

```
░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░▶▶▶

░              ██╗      █████╗ ███████╗██╗   ██╗ ██████╗██╗   ██╗██████╗ ██╗
░▒             ██║     ██╔══██╗╚══███╔╝╚██╗ ██╔╝██╔════╝██║   ██║██╔══██╗██║
░▒▓            ██║     ███████║  ███╔╝  ╚████╔╝ ██║     ██║   ██║██████╔╝██║
░▒             ██║     ██╔══██║ ███╔╝    ╚██╔╝  ██║     ██║   ██║██╔══██╗██║
░              ███████╗██║  ██║███████╗   ██║   ╚██████╗╚██████╔╝██║  ██║███████╗
               ╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝
                                        ⚡ HTTP TUI Client ⚡

◀◀◀░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░
```

### L — Standard (Terminal 80 cols)

```
░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░▶▶▶

░       ██╗      █████╗ ███████╗██╗   ██╗ ██████╗██╗   ██╗██████╗ ██╗
░▒      ██║     ██╔══██╗╚══███╔╝╚██╗ ██╔╝██╔════╝██║   ██║██╔══██╗██║
░▒▓     ██║     ███████║  ███╔╝  ╚████╔╝ ██║     ██║   ██║██████╔╝██║
░▒      ██║     ██╔══██║ ███╔╝    ╚██╔╝  ██║     ██║   ██║██╔══██╗██║
░       ███████╗██║  ██║███████╗   ██║   ╚██████╗╚██████╔╝██║  ██║███████╗
        ╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝
                              ⚡ HTTP TUI Client ⚡

◀◀◀░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░
```

### M — Compact (Terminal 60 cols)

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

### S — Mini (Terminal 40 cols)

```
░▒▓█━━━━━━━━━━━━━━━━━█▓▒░▶▶

░    ██╗    ██████╗
░▒   ██║   ██╔════╝
░▒▓  ██║   ██║      LC
░▒   ██║   ██║
░    █████╗╚██████╗
     ╚════╝ ╚═════╝
       ⚡ HTTP TUI ⚡

◀◀░▒▓█━━━━━━━━━━━━━━━━█▓▒░
```

### XS — One-liner

```
░▒▓█━━━▶  LC  ◀━━━█▓▒░  LazyCurl
```

```
░▒▓███━━━━━━▶ LazyCurl ◀━━━━━━███▓▒░
```

```
⚡░▒▓█━▶ LC ◀━█▓▒░⚡
```

---

## 🎬 Perspectives d'Animation

### Animation 1 : Pulse du gradient (Power surge)

```
Frame 1:  ░░░░▒▒▒▓▓▓████━━━━━━━━━━━━━━━████▓▓▓▒▒▒░░░░
Frame 2:  ░░░▒▒▒▓▓▓█████━━━━━━━━━━━━━━█████▓▓▓▒▒▒░░░
Frame 3:  ░░▒▒▓▓▓██████━━━━━━━━━━━━━━██████▓▓▓▒▒░░
Frame 4:  ░▒▒▓▓███████━━━━━━━━━━━━━━███████▓▓▒▒░
Frame 5:  ▒▓▓████████━━━━━━━━━━━━━━████████▓▓▒
Frame 6:  ░▒▒▓▓███████━━━━━━━━━━━━━━███████▓▓▒▒░   <- retour
```

**Implémentation Rust:**

```rust
struct SpeedGradientState {
    intensity: f32,  // 0.0 - 1.0
    direction: bool, // true = increasing
    phase: GradientPhase,
}

enum GradientPhase {
    Idle,
    Charging,
    Peak,
    Discharge,
}

impl SpeedGradientState {
    fn tick(&mut self) {
        match self.phase {
            GradientPhase::Idle => {
                // Subtle breathing
                self.intensity = 0.3 + 0.1 * (self.tick as f32 * 0.1).sin();
            }
            GradientPhase::Charging => {
                self.intensity = (self.intensity + 0.05).min(1.0);
                if self.intensity >= 1.0 {
                    self.phase = GradientPhase::Peak;
                }
            }
            GradientPhase::Peak => {
                // Flash effect
                self.intensity = 1.0;
            }
            GradientPhase::Discharge => {
                self.intensity = (self.intensity - 0.03).max(0.3);
                if self.intensity <= 0.3 {
                    self.phase = GradientPhase::Idle;
                }
            }
        }
    }

    fn render_gradient(&self, width: usize) -> String {
        let chars = ['░', '▒', '▓', '█'];
        let gradient_width = (width as f32 * self.intensity * 0.5) as usize;

        let mut result = String::new();
        for i in 0..gradient_width.min(4) {
            result.push(chars[i]);
        }
        result
    }
}
```

### Animation 2 : Streaming data effect

```
Frame 1:  ░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░▶▶▶
Frame 2:  ░▒▓█━━━━━━━━━━━●━━━━━━━━━━━━━━━━━━━█▓▒░▶▶▶
Frame 3:  ░▒▓█━━━━━━━━━━━━━━●━━━━━━━━━━━━━━━━█▓▒░▶▶▶
Frame 4:  ░▒▓█━━━━━━━━━━━━━━━━━●━━━━━━━━━━━━━█▓▒░▶▶▶
Frame 5:  ░▒▓█━━━━━━━━━━━━━━━━━━━━●━━━━━━━━━━█▓▒░▶▶▶
Frame 6:  ░▒▓█━━━━━━━━━━━━━━━━━━━━━━━●━━━━━━━█▓▒░▶▶▶
Frame 7:  ░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━●━━━━█▓▒░▶▶▶
Frame 8:  ░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━●━█▓▒░▶▶▶
```

### Animation 3 : Energy burst (Request sent)

```
Frame 1 (Charge):
░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░

Frame 2:
░▒▓██━━━━━━━━━━━━━━━━━━━━━━━━━━━━━██▓▒░

Frame 3:
░▒▓███━━━━━━━━━━━━━━━━━━━━━━━━━━━███▓▒░

Frame 4 (Burst):
░▒▓████━━━━━━━━━━━━━━━━━━━━━━━━████▓▒░▶▶▶▶▶
    ⚡⚡⚡ SENDING ⚡⚡⚡

Frame 5 (Release):
░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░▶
```

### Animation 4 : Vertical gradient bars (Side accent)

```
Frame 1:    Frame 2:    Frame 3:    Frame 4:
░           ░▒          ░▒▓         ░▒▓█
░▒          ░▒▓         ░▒▓█        ░▒▓█
░▒▓         ░▒▓█        ░▒▓█        ░▒▓
░▒          ░▒▓         ░▒▓         ░▒
░           ░           ░           ░
```

---

## 🖥️ Intégration Dashboard

### Layout 1 : Power header avec métriques

```
┌─────────────────────────────────────────────────────────────────────────────┐
│░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░▶▶▶│
│░                                                                         ░│
│░▒     ██╗       ██████╗                                                ▒░│
│░▒▓    ██║      ██╔════╝     ┌────────────────────────────────────┐   ▓▒░│
│░▒▓    ██║      ██║          │ ⚡ Performance                     │   ▓▒░│
│░▒     ██║      ██║          │   Requests/min: 847               │    ▒░│
│░      ███████╗ ╚██████╗     │   Avg latency:  42ms              │     ░│
│       ╚══════╝  ╚═════╝     │   Success rate: 99.8%             │      │
│                             └────────────────────────────────────┘      │
│░                        ⚡ HTTP TUI Client ⚡                          ░│
│◀◀◀░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░│
└─────────────────────────────────────────────────────────────────────────────┘
```

### Layout 2 : Startup avec progress bar intégré

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                                                                             │
│         ░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░▶▶▶       │
│                                                                             │
│         ░       ██╗       ██████╗                                           │
│         ░▒      ██║      ██╔════╝                                           │
│         ░▒▓     ██║      ██║         LazyCurl                               │
│         ░▒      ██║      ██║                                                │
│         ░       ███████╗ ╚██████╗                                           │
│                 ╚══════╝  ╚═════╝                                           │
│                     ⚡ HTTP TUI Client ⚡                                    │
│                                                                             │
│         ◀◀◀░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░          │
│                                                                             │
│                      ░▒▓████████████████░░░░░░░░░░░▓▒░  78%                 │
│                            Initializing HTTP engine...                      │
│                                                                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Layout 3 : Working state avec gradient sidebar

```
┌─────────────────────────────────────────────────────────────────────────────┐
│⚡░▒▓█━▶ LC ◀━█▓▒░⚡  LazyCurl v0.1.0      ● Ready      14:32:07            │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Layout 4 : Status indicators avec gradient

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ ░▒▓█━━━ LazyCurl ━━━█▓▒░                                                    │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 🎨 Variantes de couleur (Catppuccin Mocha)

### Gradient Blue → Lavender (Default)

```rust
// Gradient principal
const GRAD_1: Color = Color::Rgb(49, 50, 68);      // ░  Surface0
const GRAD_2: Color = Color::Rgb(69, 71, 90);      // ▒  Surface1
const GRAD_3: Color = Color::Rgb(88, 91, 112);     // ▓  Surface2
const GRAD_4: Color = Color::Rgb(137, 180, 250);   // █  Blue

const LINE: Color = Color::Rgb(180, 190, 254);     // ━  Lavender
const ARROW: Color = Color::Rgb(249, 226, 175);    // ▶  Yellow
const BOLT: Color = Color::Rgb(249, 226, 175);     // ⚡ Yellow
```

### Success Gradient (Green theme)

```rust
const SUCCESS_1: Color = Color::Rgb(49, 50, 68);   // ░  Surface0
const SUCCESS_2: Color = Color::Rgb(69, 71, 90);   // ▒  Surface1
const SUCCESS_3: Color = Color::Rgb(148, 226, 213);// ▓  Teal
const SUCCESS_4: Color = Color::Rgb(166, 227, 161);// █  Green
```

### Error Gradient (Red theme)

```rust
const ERROR_1: Color = Color::Rgb(49, 50, 68);     // ░  Surface0
const ERROR_2: Color = Color::Rgb(88, 91, 112);    // ▒  Surface2
const ERROR_3: Color = Color::Rgb(235, 160, 172);  // ▓  Maroon
const ERROR_4: Color = Color::Rgb(243, 139, 168);  // █  Red
```

### Warning Gradient (Yellow theme)

```rust
const WARN_1: Color = Color::Rgb(49, 50, 68);      // ░  Surface0
const WARN_2: Color = Color::Rgb(88, 91, 112);     // ▒  Surface2
const WARN_3: Color = Color::Rgb(250, 179, 135);   // ▓  Peach
const WARN_4: Color = Color::Rgb(249, 226, 175);   // █  Yellow
```

---

## 📦 Code d'intégration Ratatui

```rust
use ratatui::{
    prelude::*,
    widgets::{Block, Paragraph},
};

pub struct SpeedGradientSplash {
    tick: usize,
    intensity: f32,
    size: SplashSize,
    mode: GradientMode,
}

#[derive(Clone, Copy)]
pub enum GradientMode {
    Default,
    Success,
    Error,
    Warning,
    Loading,
}

impl SpeedGradientSplash {
    pub fn new(size: SplashSize) -> Self {
        Self {
            tick: 0,
            intensity: 0.5,
            size,
            mode: GradientMode::Default,
        }
    }

    pub fn set_mode(&mut self, mode: GradientMode) {
        self.mode = mode;
    }

    pub fn tick(&mut self) {
        self.tick += 1;

        // Pulsing effect
        self.intensity = 0.5 + 0.3 * (self.tick as f32 * 0.05).sin();
    }

    pub fn render(&self, area: Rect, buf: &mut Buffer) {
        let colors = self.get_colors();

        // Render top border
        self.render_border(area, buf, true, &colors);

        // Render side gradient
        self.render_side_gradient(area, buf, &colors);

        // Render logo
        self.render_logo(area, buf);

        // Render bottom border
        self.render_border(area, buf, false, &colors);
    }

    fn get_colors(&self) -> GradientColors {
        match self.mode {
            GradientMode::Default => GradientColors {
                g1: Color::Rgb(49, 50, 68),
                g2: Color::Rgb(69, 71, 90),
                g3: Color::Rgb(88, 91, 112),
                g4: Color::Rgb(137, 180, 250),
                line: Color::Rgb(180, 190, 254),
                accent: Color::Rgb(249, 226, 175),
            },
            GradientMode::Success => GradientColors {
                g1: Color::Rgb(49, 50, 68),
                g2: Color::Rgb(69, 71, 90),
                g3: Color::Rgb(148, 226, 213),
                g4: Color::Rgb(166, 227, 161),
                line: Color::Rgb(166, 227, 161),
                accent: Color::Rgb(166, 227, 161),
            },
            GradientMode::Error => GradientColors {
                g1: Color::Rgb(49, 50, 68),
                g2: Color::Rgb(88, 91, 112),
                g3: Color::Rgb(235, 160, 172),
                g4: Color::Rgb(243, 139, 168),
                line: Color::Rgb(243, 139, 168),
                accent: Color::Rgb(243, 139, 168),
            },
            GradientMode::Warning => GradientColors {
                g1: Color::Rgb(49, 50, 68),
                g2: Color::Rgb(88, 91, 112),
                g3: Color::Rgb(250, 179, 135),
                g4: Color::Rgb(249, 226, 175),
                line: Color::Rgb(249, 226, 175),
                accent: Color::Rgb(249, 226, 175),
            },
            GradientMode::Loading => GradientColors {
                g1: Color::Rgb(49, 50, 68),
                g2: Color::Rgb(137, 180, 250),
                g3: Color::Rgb(180, 190, 254),
                g4: Color::Rgb(203, 166, 247),
                line: Color::Rgb(203, 166, 247),
                accent: Color::Rgb(249, 226, 175),
            },
        }
    }

    fn render_border(&self, area: Rect, buf: &mut Buffer, top: bool, colors: &GradientColors) {
        let y = if top { area.y } else { area.bottom() - 1 };
        let gradient_chars = ['░', '▒', '▓', '█'];
        let gradient_colors = [colors.g1, colors.g2, colors.g3, colors.g4];

        // Left gradient
        for (i, (&ch, &color)) in gradient_chars.iter().zip(&gradient_colors).enumerate() {
            if area.x + i as u16 < area.right() {
                buf.get_mut(area.x + i as u16, y)
                    .set_char(ch)
                    .set_fg(color);
            }
        }

        // Center line
        let line_start = area.x + 4;
        let line_end = area.right() - 4;
        for x in line_start..line_end {
            buf.get_mut(x, y)
                .set_char('━')
                .set_fg(colors.line);
        }

        // Right gradient (reversed)
        for (i, (&ch, &color)) in gradient_chars.iter().zip(&gradient_colors).enumerate().rev() {
            let x = area.right() - 1 - (3 - i as u16);
            if x >= area.x {
                buf.get_mut(x, y)
                    .set_char(ch)
                    .set_fg(color);
            }
        }

        // Arrows
        if top {
            for i in 0..3 {
                let x = area.right() - 1 + i;
                if x < area.right() + 3 {
                    buf.get_mut(area.right(), y)
                        .set_char('▶')
                        .set_fg(colors.accent);
                }
            }
        }
    }

    fn render_side_gradient(&self, area: Rect, buf: &mut Buffer, colors: &GradientColors) {
        let gradient_chars = ['░', '▒', '▓', '█', '▓', '▒', '░'];
        let center_y = area.y + area.height / 2;

        for (i, &ch) in gradient_chars.iter().enumerate() {
            let y = center_y - 3 + i as u16;
            if y >= area.y && y < area.bottom() {
                let color = match i {
                    0 | 6 => colors.g1,
                    1 | 5 => colors.g2,
                    2 | 4 => colors.g3,
                    3 => colors.g4,
                    _ => colors.g1,
                };
                buf.get_mut(area.x, y)
                    .set_char(ch)
                    .set_fg(color);
            }
        }
    }
}

struct GradientColors {
    g1: Color,
    g2: Color,
    g3: Color,
    g4: Color,
    line: Color,
    accent: Color,
}
```

---

## 🔧 Configuration suggérée

```toml
# lazycurl.toml
[splash]
style = "speed-gradient"
size = "auto"
animation = true
animation_speed = 80  # ms per frame
duration = 2000

[splash.gradient]
mode = "default"  # default, success, error, warning, loading
pulse = true
pulse_speed = 0.05
side_bars = true

[splash.colors]
gradient_1 = "#313244"  # Surface0
gradient_2 = "#45475a"  # Surface1
gradient_3 = "#585b70"  # Surface2
gradient_4 = "#89b4fa"  # Blue
line = "#b4befe"        # Lavender
accent = "#f9e2af"      # Yellow (for ⚡ and ▶)
```

---

## 📏 Status-based gradient selection

```rust
fn gradient_for_status(status: u16) -> GradientMode {
    match status {
        200..=299 => GradientMode::Success,
        300..=399 => GradientMode::Warning,
        400..=599 => GradientMode::Error,
        _ => GradientMode::Default,
    }
}

// Usage in request display
fn render_request_status(&self, status: u16, buf: &mut Buffer) {
    let mode = gradient_for_status(status);
    let colors = self.get_colors_for_mode(mode);

    // Render status with appropriate gradient
    self.render_status_bar(status, colors, buf);
}
```

---

## 🔌 Composants réutilisables

### Separateur horizontal avec gradient

```
░▒▓█━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━█▓▒░
```

### Status badge

```
░▒▓█ 200 OK █▓▒░    (Success)
░▒▓▓ 404    ▓▓▒░    (Error)
░▒▓█ POST   █▓▒░    (Method)
```

### Progress bar avec gradient

```
░▒▓████████████████████░░░░░░░░░░░▓▒░  75%
```

### Section header

```
░▒▓█━━━ Request ━━━█▓▒░
```
