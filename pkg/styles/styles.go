package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Catppuccin Mocha Colors
	// Base colors
	Base   = lipgloss.Color("#1e1e2e") // background
	Mantle = lipgloss.Color("#181825") // darker background
	Crust  = lipgloss.Color("#11111b") // darkest background

	// Text colors
	Text     = lipgloss.Color("#cdd6f4") // main text
	Subtext1 = lipgloss.Color("#bac2de") // dimmed text
	Subtext0 = lipgloss.Color("#a6adc8") // more dimmed

	// Overlay colors
	Surface0 = lipgloss.Color("#313244") // borders
	Surface1 = lipgloss.Color("#45475a") // lighter borders

	// Accent colors
	Lavender = lipgloss.Color("#b4befe") // primary accent
	Mauve    = lipgloss.Color("#cba6f7") // secondary accent
	Pink     = lipgloss.Color("#f5c2e7") // tertiary accent
	Red      = lipgloss.Color("#f38ba8") // errors
	Peach    = lipgloss.Color("#fab387") // warnings
	Yellow   = lipgloss.Color("#f9e2af") // highlights
	Green    = lipgloss.Color("#a6e3a1") // success/active
	Teal     = lipgloss.Color("#94e2d5") // info
	Sky      = lipgloss.Color("#89dceb") // links
	Sapphire = lipgloss.Color("#74c7ec") // special
	Blue     = lipgloss.Color("#89b4fa") // primary actions

	// Legacy color names for compatibility
	PrimaryColor   = Lavender
	SecondaryColor = Blue
	AccentColor    = Mauve
	TextColor      = Text
	MutedColor     = Subtext0
	BorderColor    = Surface0
	ActiveColor    = Lavender

	// Mode colors (vim-style modes) - text always white except INSERT (black)
	ModeNormalBg  = lipgloss.Color("#6798da") // Blue
	ModeNormalFg  = lipgloss.Color("#FFFFFF") // White
	ModeViewBg    = lipgloss.Color("#4c8c49") // Green
	ModeViewFg    = lipgloss.Color("#FFFFFF") // White
	ModeCommandBg = lipgloss.Color("#a45e0e") // Orange
	ModeCommandFg = lipgloss.Color("#FFFFFF") // White
	ModeInsertBg  = lipgloss.Color("#b8bcc2") // Light gray
	ModeInsertFg  = lipgloss.Color("#000000") // Black

	// HTTP method colors - text always white
	MethodHeadBg    = lipgloss.Color("#4c8c49") // Green
	MethodHeadFg    = lipgloss.Color("#FFFFFF") // White
	MethodGetBg     = lipgloss.Color("#4c8c49") // Green
	MethodGetFg     = lipgloss.Color("#FFFFFF") // White
	MethodPostBg    = lipgloss.Color("#a45e0e") // Orange
	MethodPostFg    = lipgloss.Color("#FFFFFF") // White
	MethodPutBg     = lipgloss.Color("#6798da") // Blue
	MethodPutFg     = lipgloss.Color("#FFFFFF") // White
	MethodPatchBg   = lipgloss.Color("#d48cee") // Purple
	MethodPatchFg   = lipgloss.Color("#FFFFFF") // White
	MethodDeleteBg  = lipgloss.Color("#fa827c") // Red/Coral
	MethodDeleteFg  = lipgloss.Color("#FFFFFF") // White
	MethodOptionsBg = lipgloss.Color("#a48e85") // Brown/Taupe
	MethodOptionsFg = lipgloss.Color("#FFFFFF") // White

	// HTTP status colors - response types
	Status2xxBg = lipgloss.Color("#4c8c49") // Green (success)
	Status2xxFg = lipgloss.Color("#FFFFFF") // White
	Status3xxBg = lipgloss.Color("#6798da") // Blue (redirect)
	Status3xxFg = lipgloss.Color("#FFFFFF") // White
	Status4xxBg = lipgloss.Color("#a45e0e") // Orange (client error)
	Status4xxFg = lipgloss.Color("#FFFFFF") // White
	Status5xxBg = lipgloss.Color("#fa827c") // Red/Coral (server error)
	Status5xxFg = lipgloss.Color("#FFFFFF") // White

	// Tree selection colors - using primary color (Lavender) with secondary text (Blue)
	SelectedPanelBg        = lipgloss.Color("#b4befe") // Primary (Lavender) - Selected panel active item
	SelectedPanelFg        = lipgloss.Color("#89b4fa") // Secondary (Blue) text for contrast
	SelectedRequestBg      = lipgloss.Color("#45475a") // Surface1 - Selected request (inactive)
	SelectedRequestFg      = lipgloss.Color("#b4befe") // Primary (Lavender) text
	CurrentCollectionBg    = lipgloss.Color("#b4befe") // Primary (Lavender) - Current collection selected
	CurrentCollectionFg    = lipgloss.Color("#89b4fa") // Secondary (Blue) text for contrast
	CollectionNotCurrentBg = lipgloss.Color("#313244") // Surface0 - Collection not current
	CollectionNotCurrentFg = lipgloss.Color("#cdd6f4") // Text color

	// Environment colors
	SecretColor   = lipgloss.Color("#d48cee") // Purple for secret values
	InactiveColor = lipgloss.Color("#6f747a") // Gray for inactive values
	CheckboxOn    = lipgloss.Color("#a6e3a1") // Green checkbox filled
	CheckboxOff   = lipgloss.Color("#6f747a") // Gray checkbox empty

	// Search colors
	SearchMatch  = lipgloss.Color("#c884e0") // Purple/pink for search matches
	SearchDimmed = lipgloss.Color("#45475a") // Dark gray for non-matching items

	// URL syntax highlighting colors
	URLVariable = lipgloss.Color("#fa827c") // Coral/red for {{variables}}
	URLParam    = lipgloss.Color("#6798da") // Blue for :params
	URLBase     = lipgloss.Color("#cdd6f4") // Text color for base URL

	// Base styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Lavender).
			Background(Mantle).
			Padding(0, 1)

	ActiveTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Lavender).
				Background(Mantle).
				Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Surface0).
			Padding(0)

	ActiveBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(Lavender).
				Padding(0)

	InactiveBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(Surface0).
				Padding(0)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(Mantle).
			Padding(0, 0)

	ItemStyle = lipgloss.NewStyle().
			Foreground(Text).
			PaddingLeft(1)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(Lavender).
				Bold(true).
				PaddingLeft(1)

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(Surface0).
			Padding(0, 0)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Subtext0).
			Italic(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Blue)
)
