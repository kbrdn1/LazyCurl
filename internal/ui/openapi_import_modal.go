package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// Spinner frames for progress indicator
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// OpenAPIImportState represents the current state of the import modal
type OpenAPIImportState int

// Import modal states
const (
	StateInputPath        OpenAPIImportState = iota // Waiting for file path input
	StatePreview                                    // Showing import preview
	StateConfirmOverwrite                           // Asking to confirm overwrite
	StateImporting                                  // Importing in progress
	StateError                                      // Error occurred
)

// OpenAPIImportModal handles the OpenAPI import modal
type OpenAPIImportModal struct {
	pathInput      textinput.Model
	state          OpenAPIImportState
	preview        *api.ImportPreview
	importer       *api.OpenAPIImporter
	error          string
	visible        bool
	width          int
	height         int
	filePath       string
	collectionsDir string

	// Conflict handling (T069)
	conflictPath    string // Path of existing collection that would be overwritten
	suggestedPath   string // Alternative path with numeric suffix
	overwriteChoice int    // 0=overwrite, 1=rename

	// Progress indicator (T070)
	spinnerFrame int
	importing    bool
}

// NewOpenAPIImportModal creates a new OpenAPI import modal
func NewOpenAPIImportModal(collectionsDir string) *OpenAPIImportModal {
	ti := textinput.New()
	ti.Placeholder = "Enter path to OpenAPI spec (JSON/YAML)..."
	ti.CharLimit = 500
	ti.Width = 60

	return &OpenAPIImportModal{
		pathInput:      ti,
		state:          StateInputPath,
		visible:        false,
		width:          80,
		height:         20,
		collectionsDir: collectionsDir,
	}
}

// Show makes the modal visible and focuses the input
func (m *OpenAPIImportModal) Show() {
	m.visible = true
	m.state = StateInputPath
	m.error = ""
	m.preview = nil
	m.importer = nil
	m.filePath = ""
	m.pathInput.Reset()
	m.pathInput.Focus()
}

// Hide hides the modal
func (m *OpenAPIImportModal) Hide() {
	m.visible = false
	m.error = ""
	m.pathInput.Blur()
}

// IsVisible returns whether the modal is visible
func (m *OpenAPIImportModal) IsVisible() bool {
	return m.visible
}

// SetSize updates the modal dimensions
func (m *OpenAPIImportModal) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate input size
	modalWidth := min(90, width-10)
	inputWidth := modalWidth - 10

	m.pathInput.Width = inputWidth
}

// Update handles messages for the OpenAPI import modal
func (m *OpenAPIImportModal) Update(msg tea.Msg) (*OpenAPIImportModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case OpenAPISpinnerTickMsg:
		if m.importing {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			return m, m.spinnerTick()
		}
		return m, nil

	case OpenAPIImportCompleteMsg:
		m.importing = false
		if msg.Error != nil {
			var importErr *api.ImportError
			if errors.As(msg.Error, &importErr) {
				m.error = importErr.Message
			} else {
				m.error = msg.Error.Error()
			}
			m.state = StateError
			return m, nil
		}

		// Count requests
		requestCount := countCollectionRequests(msg.Collection)

		// Success - hide modal and send import message
		m.Hide()
		return m, func() tea.Msg {
			return OpenAPIImportedMsg{
				Collection: msg.Collection,
				Stats: OpenAPIImportStats{
					FolderCount:  len(msg.Collection.Folders),
					RequestCount: requestCount,
					WarningCount: len(m.preview.Warnings),
				},
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.state == StatePreview || m.state == StateConfirmOverwrite {
				// Go back to input
				m.state = StateInputPath
				m.preview = nil
				m.importer = nil
				m.conflictPath = ""
				m.suggestedPath = ""
				m.pathInput.Focus()
				return m, nil
			}
			// Close modal
			m.Hide()
			return m, func() tea.Msg {
				return HideOpenAPIImportModalMsg{}
			}

		case "enter":
			switch m.state {
			case StateInputPath:
				// Load and preview the spec
				return m.loadSpec()

			case StatePreview:
				// Check for conflicts before importing
				return m.checkConflictAndImport()

			case StateConfirmOverwrite:
				// Confirm with selected choice
				return m.executeImport()
			}

		case "tab", "left", "right":
			// Toggle between overwrite/rename choices in confirm state
			if m.state == StateConfirmOverwrite {
				m.overwriteChoice = 1 - m.overwriteChoice
				return m, nil
			}

		case "o", "O":
			// Quick key for overwrite
			if m.state == StateConfirmOverwrite {
				m.overwriteChoice = 0
				return m.executeImport()
			}

		case "r", "R":
			// Quick key for rename
			if m.state == StateConfirmOverwrite {
				m.overwriteChoice = 1
				return m.executeImport()
			}
		}
	}

	// Pass other messages to text input when in input state
	if m.state == StateInputPath {
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// spinnerTick returns a command that ticks the spinner
func (m *OpenAPIImportModal) spinnerTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(_ time.Time) tea.Msg {
		return OpenAPISpinnerTickMsg{}
	})
}

// loadSpec loads and previews the OpenAPI spec
func (m *OpenAPIImportModal) loadSpec() (*OpenAPIImportModal, tea.Cmd) {
	path := strings.TrimSpace(m.pathInput.Value())
	if path == "" {
		m.error = "Please enter a file path"
		return m, nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	// Make path absolute if relative
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err == nil {
			path = filepath.Join(cwd, path)
		}
	}

	m.filePath = path

	// Create importer
	importer, err := api.NewOpenAPIImporterFromFile(path)
	if err != nil {
		var importErr *api.ImportError
		if errors.As(err, &importErr) {
			m.error = importErr.Message
			if importErr.Details != "" {
				m.error += "\n" + importErr.Details
			}
		} else {
			m.error = err.Error()
		}
		m.state = StateError
		return m, nil
	}

	// Get preview
	preview, err := importer.Preview()
	if err != nil {
		var importErr *api.ImportError
		if errors.As(err, &importErr) {
			m.error = importErr.Message
			if importErr.Details != "" {
				m.error += "\n" + importErr.Details
			}
		} else {
			m.error = err.Error()
		}
		m.state = StateError
		return m, nil
	}

	m.importer = importer
	m.preview = preview
	m.state = StatePreview
	m.error = ""
	m.pathInput.Blur()

	return m, nil
}

// checkConflictAndImport checks for existing collections and handles conflicts
func (m *OpenAPIImportModal) checkConflictAndImport() (*OpenAPIImportModal, tea.Cmd) {
	if m.importer == nil {
		m.error = "No spec loaded"
		return m, nil
	}

	// Get the target filename
	filename := sanitizeOpenAPIFilename(m.preview.Title) + ".json"
	savePath := filepath.Join(m.collectionsDir, filename)

	// Check if file already exists
	if _, err := os.Stat(savePath); err == nil {
		// File exists - show conflict dialog
		m.conflictPath = savePath
		m.suggestedPath = m.findAvailablePath(filename)
		m.overwriteChoice = 1 // Default to rename
		m.state = StateConfirmOverwrite
		return m, nil
	}

	// No conflict - proceed with import
	m.conflictPath = ""
	m.suggestedPath = ""
	return m.executeImport()
}

// findAvailablePath generates a unique filename by adding a numeric suffix
func (m *OpenAPIImportModal) findAvailablePath(filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	for i := 1; i <= 99; i++ {
		newName := fmt.Sprintf("%s-%d%s", base, i, ext)
		newPath := filepath.Join(m.collectionsDir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// Fallback with timestamp
	return filepath.Join(m.collectionsDir, fmt.Sprintf("%s-%d%s", base, time.Now().Unix(), ext))
}

// executeImport performs the actual import with progress indicator
func (m *OpenAPIImportModal) executeImport() (*OpenAPIImportModal, tea.Cmd) {
	if m.importer == nil {
		m.error = "No spec loaded"
		return m, nil
	}

	m.state = StateImporting
	m.importing = true
	m.spinnerFrame = 0

	// Determine save path based on conflict resolution
	var savePath string
	if m.conflictPath != "" {
		if m.overwriteChoice == 0 {
			// Overwrite existing
			savePath = m.conflictPath
		} else {
			// Use suggested path
			savePath = m.suggestedPath
		}
	} else {
		// No conflict
		filename := sanitizeOpenAPIFilename(m.preview.Title) + ".json"
		savePath = filepath.Join(m.collectionsDir, filename)
	}

	// Start async import
	importer := m.importer
	collectionsDir := m.collectionsDir

	importCmd := func() tea.Msg {
		// Perform import
		collection, err := importer.ToCollection(api.ImportOptions{
			IncludeExamples: true,
		})
		if err != nil {
			return OpenAPIImportCompleteMsg{Error: err}
		}

		// Ensure collections directory exists
		if err := os.MkdirAll(collectionsDir, 0755); err != nil {
			return OpenAPIImportCompleteMsg{
				Error: fmt.Errorf("failed to create collections directory: %w", err),
			}
		}

		collection.FilePath = savePath
		if err := api.SaveCollection(collection, savePath); err != nil {
			return OpenAPIImportCompleteMsg{
				Error: fmt.Errorf("failed to save collection: %w", err),
			}
		}

		return OpenAPIImportCompleteMsg{
			Collection: collection,
			SavePath:   savePath,
		}
	}

	// Return both spinner tick and import command
	return m, tea.Batch(m.spinnerTick(), importCmd)
}

// View renders the OpenAPI import modal
func (m *OpenAPIImportModal) View() string {
	if !m.visible {
		return ""
	}

	modalWidth := min(90, m.width-10)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		MarginTop(1)

	errorStyle := lipgloss.NewStyle().
		Foreground(styles.Red).
		Bold(true).
		MarginTop(1)

	successStyle := lipgloss.NewStyle().
		Foreground(styles.Green)

	warningStyle := lipgloss.NewStyle().
		Foreground(styles.Yellow)

	labelStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext1).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.Lavender).
		Background(styles.Base)

	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("📄 Import OpenAPI Specification"))
	content.WriteString("\n")

	switch m.state {
	case StateInputPath, StateError:
		content.WriteString(subtitleStyle.Render("Enter the path to your OpenAPI 3.x specification file"))
		content.WriteString("\n\n")

		// Path input
		content.WriteString(labelStyle.Render("File Path:"))
		content.WriteString("\n")
		content.WriteString(m.pathInput.View())

		// Error message
		if m.error != "" {
			content.WriteString("\n")
			content.WriteString(errorStyle.Render("⚠ " + m.error))
		}

		// Help text
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Enter: Load & Preview • Esc: Cancel"))

	case StatePreview:
		if m.preview == nil {
			content.WriteString("\n")
			content.WriteString("Loading preview...")
		} else {
			content.WriteString(subtitleStyle.Render("Review import details before confirming"))
			content.WriteString("\n\n")

			// API Info
			content.WriteString(labelStyle.Render("API: "))
			content.WriteString(valueStyle.Render(m.preview.Title))
			content.WriteString(" ")
			content.WriteString(subtitleStyle.Render("(OpenAPI " + m.preview.SpecVersion + ")"))
			content.WriteString("\n")

			if m.preview.Description != "" {
				desc := m.preview.Description
				if len(desc) > 80 {
					desc = desc[:77] + "..."
				}
				content.WriteString(subtitleStyle.Render(desc))
				content.WriteString("\n")
			}

			content.WriteString("\n")

			// Statistics
			content.WriteString(labelStyle.Render("Endpoints: "))
			content.WriteString(successStyle.Render(fmt.Sprintf("%d", m.preview.EndpointCount)))
			content.WriteString("\n")

			content.WriteString(labelStyle.Render("Folders: "))
			content.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.preview.FolderCount)))
			content.WriteString("\n")

			// Folder details
			if len(m.preview.Folders) > 0 {
				content.WriteString(subtitleStyle.Render("  "))
				var folderNames []string
				for _, f := range m.preview.Folders {
					folderNames = append(folderNames, fmt.Sprintf("%s (%d)", f.Name, f.RequestCount))
				}
				// Limit displayed folders
				if len(folderNames) > 5 {
					content.WriteString(subtitleStyle.Render(strings.Join(folderNames[:5], ", ") + "..."))
				} else {
					content.WriteString(subtitleStyle.Render(strings.Join(folderNames, ", ")))
				}
				content.WriteString("\n")
			}

			// Servers
			if len(m.preview.Servers) > 0 {
				content.WriteString(labelStyle.Render("Base URL: "))
				content.WriteString(valueStyle.Render(m.preview.Servers[0]))
				content.WriteString("\n")
			}

			// Warnings
			if len(m.preview.Warnings) > 0 {
				content.WriteString("\n")
				content.WriteString(warningStyle.Render(fmt.Sprintf("⚠ %d warning(s)", len(m.preview.Warnings))))
				content.WriteString("\n")
				for i, w := range m.preview.Warnings {
					if i >= 3 {
						content.WriteString(subtitleStyle.Render(fmt.Sprintf("  ... and %d more", len(m.preview.Warnings)-3)))
						break
					}
					content.WriteString(subtitleStyle.Render("  • " + w))
					content.WriteString("\n")
				}
			}

			// Help text
			content.WriteString("\n")
			content.WriteString(helpStyle.Render("Enter: Confirm Import • Esc: Back"))
		}

	case StateConfirmOverwrite:
		// Conflict resolution dialog
		content.WriteString(warningStyle.Render("⚠ Collection Already Exists"))
		content.WriteString("\n\n")

		content.WriteString(labelStyle.Render("Existing file: "))
		content.WriteString(valueStyle.Render(filepath.Base(m.conflictPath)))
		content.WriteString("\n\n")

		content.WriteString(subtitleStyle.Render("Choose an action:"))
		content.WriteString("\n\n")

		// Overwrite option
		overwriteStyle := valueStyle
		renameStyle := valueStyle
		if m.overwriteChoice == 0 {
			overwriteStyle = successStyle.Bold(true)
		} else {
			renameStyle = successStyle.Bold(true)
		}

		content.WriteString("  ")
		if m.overwriteChoice == 0 {
			content.WriteString(overwriteStyle.Render("● "))
		} else {
			content.WriteString(valueStyle.Render("○ "))
		}
		content.WriteString(overwriteStyle.Render("[O] Overwrite"))
		content.WriteString(subtitleStyle.Render(" - Replace the existing collection"))
		content.WriteString("\n")

		content.WriteString("  ")
		if m.overwriteChoice == 1 {
			content.WriteString(renameStyle.Render("● "))
		} else {
			content.WriteString(valueStyle.Render("○ "))
		}
		content.WriteString(renameStyle.Render("[R] Rename"))
		content.WriteString(subtitleStyle.Render(" - Save as "))
		content.WriteString(successStyle.Render(filepath.Base(m.suggestedPath)))
		content.WriteString("\n")

		// Help text
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("O/R: Quick select • Tab/←→: Toggle • Enter: Confirm • Esc: Back"))

	case StateImporting:
		content.WriteString("\n\n")

		// Animated spinner
		spinnerStyle := lipgloss.NewStyle().
			Foreground(styles.Lavender).
			Bold(true)

		spinner := spinnerFrames[m.spinnerFrame]
		content.WriteString(spinnerStyle.Render(spinner))
		content.WriteString(" ")
		content.WriteString(subtitleStyle.Render("Importing specification..."))
		content.WriteString("\n\n")

		// Progress details
		if m.preview != nil {
			content.WriteString(labelStyle.Render("Converting "))
			content.WriteString(valueStyle.Render(fmt.Sprintf("%d endpoints", m.preview.EndpointCount)))
			content.WriteString(labelStyle.Render(" to collection format"))
			content.WriteString("\n")
		}

		content.WriteString("\n")
		content.WriteString(subtitleStyle.Render("Please wait..."))
	}

	return modalStyle.Render(content.String())
}

// GetFilePath returns the current file path
func (m *OpenAPIImportModal) GetFilePath() string {
	return m.filePath
}

// SetError sets an error message
func (m *OpenAPIImportModal) SetError(err string) {
	m.error = err
	m.state = StateError
}

// Helper functions

// sanitizeOpenAPIFilename converts a name to a valid filename
func sanitizeOpenAPIFilename(name string) string {
	// Replace spaces and special characters
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	// Remove characters that are not alphanumeric, dash, or underscore
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	filename := result.String()
	if filename == "" {
		filename = "imported-api"
	}

	return filename
}

// countCollectionRequests counts all requests in a collection
func countCollectionRequests(c *api.CollectionFile) int {
	count := len(c.Requests)
	for _, folder := range c.Folders {
		count += countFolderRequestsRecursive(&folder)
	}
	return count
}

// countFolderRequestsRecursive counts requests in a folder and its subfolders
func countFolderRequestsRecursive(f *api.Folder) int {
	count := len(f.Requests)
	for _, subfolder := range f.Folders {
		count += countFolderRequestsRecursive(&subfolder)
	}
	return count
}
