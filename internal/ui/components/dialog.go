package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// DialogType represents the type of dialog
type DialogType int

const (
	DialogInput DialogType = iota
	DialogConfirm
	DialogNewRequest
	DialogEditRequest
)

// HTTP methods for request creation
var httpMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

// Dialog represents a modal dialog component
type Dialog struct {
	visible     bool
	dialogType  DialogType
	title       string
	message     string
	inputValue  string
	cursorPos   int
	confirmText string
	cancelText  string
	action      string // Action identifier for the callback
	targetNode  *TreeNode

	// For new request dialog
	methodIndex int    // Selected HTTP method index
	urlValue    string // URL endpoint
	focusField  int    // 0=name, 1=method, 2=url
}

// DialogResultMsg is sent when a dialog is completed
type DialogResultMsg struct {
	Action    string
	Confirmed bool
	Value     string
	Method    string // HTTP method for new request
	URL       string // URL endpoint for new request
	Node      *TreeNode
}

// NewDialog creates a new dialog component
func NewDialog() *Dialog {
	return &Dialog{
		visible:     false,
		confirmText: "OK",
		cancelText:  "Cancel",
	}
}

// ShowInput shows an input dialog
func (d *Dialog) ShowInput(title, message, defaultValue, action string, node *TreeNode) {
	d.visible = true
	d.dialogType = DialogInput
	d.title = title
	d.message = message
	d.inputValue = defaultValue
	d.cursorPos = len(defaultValue)
	d.action = action
	d.targetNode = node
	d.focusField = 0
}

// ShowNewRequest shows a new request dialog with method selector and URL
func (d *Dialog) ShowNewRequest(action string, node *TreeNode) {
	d.visible = true
	d.dialogType = DialogNewRequest
	d.title = "New Request"
	d.message = ""
	d.inputValue = "New Request"
	d.cursorPos = len(d.inputValue)
	d.methodIndex = 0 // GET by default
	d.urlValue = "{{base_url}}/endpoint"
	d.action = action
	d.targetNode = node
	d.focusField = 0 // Start on name field
}

// ShowConfirm shows a confirmation dialog
func (d *Dialog) ShowConfirm(title, message, action string, node *TreeNode) {
	d.visible = true
	d.dialogType = DialogConfirm
	d.title = title
	d.message = message
	d.inputValue = ""
	d.action = action
	d.targetNode = node
}

// ShowEditRequest shows an edit request dialog with existing values
func (d *Dialog) ShowEditRequest(node *TreeNode) {
	d.visible = true
	d.dialogType = DialogEditRequest
	d.title = "Edit Request"
	d.message = ""
	d.inputValue = node.Name
	d.cursorPos = len(d.inputValue)
	// Find method index
	d.methodIndex = 0
	for i, m := range httpMethods {
		if m == node.HTTPMethod {
			d.methodIndex = i
			break
		}
	}
	d.urlValue = node.URL
	d.action = "edit_request"
	d.targetNode = node
	d.focusField = 0
}

// Hide hides the dialog
func (d *Dialog) Hide() {
	d.visible = false
}

// IsVisible returns whether the dialog is visible
func (d *Dialog) IsVisible() bool {
	return d.visible
}

// Update handles messages for the dialog
func (d *Dialog) Update(msg tea.Msg) (*Dialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Cancel dialog
			d.Hide()
			return d, func() tea.Msg {
				return DialogResultMsg{
					Action:    d.action,
					Confirmed: false,
					Value:     "",
					Node:      d.targetNode,
				}
			}

		case "enter":
			// Confirm dialog
			d.Hide()
			method := ""
			url := ""
			if d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest {
				method = httpMethods[d.methodIndex]
				url = d.urlValue
			}
			return d, func() tea.Msg {
				return DialogResultMsg{
					Action:    d.action,
					Confirmed: true,
					Value:     d.inputValue,
					Method:    method,
					URL:       url,
					Node:      d.targetNode,
				}
			}

		case "tab", "down":
			// Move to next field in request dialogs
			if d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest {
				d.focusField = (d.focusField + 1) % 3
				// Update cursor position for the new field
				if d.focusField == 2 {
					d.cursorPos = len(d.urlValue)
				} else if d.focusField == 0 {
					d.cursorPos = len(d.inputValue)
				}
			}

		case "shift+tab", "up":
			// Move to previous field in request dialogs
			if d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest {
				d.focusField = (d.focusField + 2) % 3
				if d.focusField == 2 {
					d.cursorPos = len(d.urlValue)
				} else if d.focusField == 0 {
					d.cursorPos = len(d.inputValue)
				}
			}

		case "left", "h":
			if (d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest) && d.focusField == 1 {
				// Change method
				d.methodIndex = (d.methodIndex + len(httpMethods) - 1) % len(httpMethods)
			} else if msg.String() == "left" {
				if d.cursorPos > 0 {
					d.cursorPos--
				}
			}

		case "right", "l":
			if (d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest) && d.focusField == 1 {
				// Change method
				d.methodIndex = (d.methodIndex + 1) % len(httpMethods)
			} else if msg.String() == "right" {
				currentValue := d.getCurrentValue()
				if d.cursorPos < len(currentValue) {
					d.cursorPos++
				}
			}

		case "backspace":
			if d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest {
				if d.focusField == 0 && len(d.inputValue) > 0 && d.cursorPos > 0 {
					d.inputValue = d.inputValue[:d.cursorPos-1] + d.inputValue[d.cursorPos:]
					d.cursorPos--
				} else if d.focusField == 2 && len(d.urlValue) > 0 && d.cursorPos > 0 {
					d.urlValue = d.urlValue[:d.cursorPos-1] + d.urlValue[d.cursorPos:]
					d.cursorPos--
				}
			} else if d.dialogType == DialogInput && len(d.inputValue) > 0 && d.cursorPos > 0 {
				d.inputValue = d.inputValue[:d.cursorPos-1] + d.inputValue[d.cursorPos:]
				d.cursorPos--
			}

		case "home", "ctrl+a":
			d.cursorPos = 0

		case "end", "ctrl+e":
			d.cursorPos = len(d.getCurrentValue())

		default:
			// Insert character
			if len(msg.String()) == 1 {
				char := msg.String()
				if d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest {
					if d.focusField == 0 {
						d.inputValue = d.inputValue[:d.cursorPos] + char + d.inputValue[d.cursorPos:]
						d.cursorPos++
					} else if d.focusField == 2 {
						d.urlValue = d.urlValue[:d.cursorPos] + char + d.urlValue[d.cursorPos:]
						d.cursorPos++
					}
				} else if d.dialogType == DialogInput {
					d.inputValue = d.inputValue[:d.cursorPos] + char + d.inputValue[d.cursorPos:]
					d.cursorPos++
				}
			}
		}
	}

	return d, nil
}

// getCurrentValue returns the current field value based on focus
func (d *Dialog) getCurrentValue() string {
	if d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest {
		if d.focusField == 2 {
			return d.urlValue
		}
	}
	return d.inputValue
}

// View renders the dialog
func (d *Dialog) View(screenWidth, screenHeight int) string {
	if !d.visible {
		return ""
	}

	// Dialog dimensions
	dialogWidth := 56
	if dialogWidth > screenWidth-4 {
		dialogWidth = screenWidth - 4
	}

	// Build dialog content based on type
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		Width(dialogWidth - 4).
		Align(lipgloss.Center)
	content.WriteString(titleStyle.Render(d.title))
	content.WriteString("\n")

	if d.dialogType == DialogNewRequest || d.dialogType == DialogEditRequest {
		content.WriteString(d.renderNewRequestForm(dialogWidth - 4))
	} else if d.dialogType == DialogInput {
		content.WriteString(d.renderInputForm(dialogWidth - 4))
	} else if d.dialogType == DialogConfirm {
		content.WriteString(d.renderConfirmForm(dialogWidth - 4))
	}

	// Buttons
	content.WriteString("\n")
	confirmStyle := lipgloss.NewStyle().
		Background(styles.Lavender).
		Foreground(styles.Base).
		Padding(0, 2).
		Bold(true)

	cancelStyle := lipgloss.NewStyle().
		Background(styles.Surface0).
		Foreground(styles.Text).
		Padding(0, 2)

	buttons := confirmStyle.Render("Enter") + "  " + cancelStyle.Render("Esc")
	content.WriteString(lipgloss.NewStyle().Width(dialogWidth - 4).Align(lipgloss.Center).Render(buttons))

	// Dialog box style - transparent background, only border (matching modal.go)
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Lavender).
		Padding(1, 2).
		Width(dialogWidth)

	// Return just the dialog box, centering is handled by caller overlay
	return dialogStyle.Render(content.String())
}

// renderNewRequestForm renders the new request form fields
func (d *Dialog) renderNewRequestForm(width int) string {
	var content strings.Builder

	labelStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext1)

	inputStyle := lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.Surface0).
		Width(width).
		Padding(0, 1)

	activeInputStyle := lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.Surface1).
		Width(width).
		Padding(0, 1)

	// Path info (read-only, shows location in tree)
	pathInfo := d.getTreePath()
	if pathInfo != "" {
		pathStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Italic(true)
		content.WriteString("\n")
		content.WriteString(pathStyle.Render("Path: " + pathInfo))
	}

	// Name field
	content.WriteString("\n")
	content.WriteString(labelStyle.Render("Name: "))
	nameInput := d.inputValue
	if d.focusField == 0 {
		nameInput = d.renderWithCursor(d.inputValue, d.cursorPos)
		content.WriteString(activeInputStyle.Render(nameInput))
	} else {
		content.WriteString(inputStyle.Render(nameInput))
	}

	// Method field
	content.WriteString("\n")
	content.WriteString(labelStyle.Render("Method: "))
	methodDisplay := d.renderMethodSelector(width, d.focusField == 1)
	content.WriteString(methodDisplay)

	// URL field
	content.WriteString("\n")
	content.WriteString(labelStyle.Render("URL: "))
	urlInput := d.urlValue
	if d.focusField == 2 {
		urlInput = d.renderWithCursor(d.urlValue, d.cursorPos)
		content.WriteString(activeInputStyle.Render(urlInput))
	} else {
		content.WriteString(inputStyle.Render(urlInput))
	}

	content.WriteString("\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Italic(true).
		Width(width).
		Align(lipgloss.Center)
	content.WriteString(helpStyle.Render("Tab: next • h/l: method"))

	return content.String()
}

// getTreePath returns the path in the tree where the request will be created
func (d *Dialog) getTreePath() string {
	if d.targetNode == nil {
		return "/"
	}

	// Build path from node to root
	var parts []string
	current := d.targetNode
	for current != nil {
		parts = append([]string{current.Name}, parts...)
		current = current.Parent
	}

	if len(parts) == 0 {
		return "/"
	}
	return strings.Join(parts, " › ")
}

// renderInputForm renders a simple input form
func (d *Dialog) renderInputForm(width int) string {
	var content strings.Builder

	if d.message != "" {
		messageStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext1).
			Width(width)
		content.WriteString("\n")
		content.WriteString(messageStyle.Render(d.message))
	}

	content.WriteString("\n")
	inputStyle := lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.Surface1).
		Width(width).
		Padding(0, 1)

	inputDisplay := d.renderWithCursor(d.inputValue, d.cursorPos)
	content.WriteString(inputStyle.Render(inputDisplay))
	content.WriteString("\n")

	return content.String()
}

// renderConfirmForm renders a confirmation dialog
func (d *Dialog) renderConfirmForm(width int) string {
	var content strings.Builder

	messageStyle := lipgloss.NewStyle().
		Foreground(styles.Text).
		Width(width).
		Align(lipgloss.Center)
	content.WriteString("\n")
	content.WriteString(messageStyle.Render(d.message))
	content.WriteString("\n")

	return content.String()
}

// renderMethodSelector renders the HTTP method selector
func (d *Dialog) renderMethodSelector(width int, active bool) string {
	// Only show the selected method with arrows for navigation
	method := httpMethods[d.methodIndex]
	bg, fg := d.getMethodColors(method)

	methodStyle := lipgloss.NewStyle().
		Background(bg).
		Foreground(fg).
		Bold(true).
		Padding(0, 1)

	arrowStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0)

	// Show: ◀ METHOD ▶
	content := arrowStyle.Render("◀ ") + methodStyle.Render(method) + arrowStyle.Render(" ▶")

	// No background on container - transparent like other fields
	return content
}

// getMethodColors returns the background and foreground colors for an HTTP method
func (d *Dialog) getMethodColors(method string) (lipgloss.Color, lipgloss.Color) {
	switch method {
	case "GET":
		return styles.MethodGetBg, styles.MethodGetFg
	case "POST":
		return styles.MethodPostBg, styles.MethodPostFg
	case "PUT":
		return styles.MethodPutBg, styles.MethodPutFg
	case "PATCH":
		return styles.MethodPatchBg, styles.MethodPatchFg
	case "DELETE":
		return styles.MethodDeleteBg, styles.MethodDeleteFg
	case "HEAD":
		return styles.MethodHeadBg, styles.MethodHeadFg
	case "OPTIONS":
		return styles.MethodOptionsBg, styles.MethodOptionsFg
	default:
		return styles.Surface1, styles.Text
	}
}

// renderWithCursor renders text with a cursor at the specified position
func (d *Dialog) renderWithCursor(text string, cursorPos int) string {
	if cursorPos >= len(text) {
		return text + "█"
	}
	return text[:cursorPos] + "█" + text[cursorPos+1:]
}
