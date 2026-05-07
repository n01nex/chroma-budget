package TUI

/* FORM OVERLAY SYSTEM
This module provides a dual-purpose modal overlay for:
1. Viewing and editing existing Entry/Target records
2. Creating new Entry/Target records

The overlay displays all relevant fields with input capability,
and provides Cancel/Confirm actions.
*/

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// =============================================================================
// FORM TYPES
// =============================================================================

// FormType defines the purpose of the form
type FormType int

const (
	EditEntry FormType = iota
	EditTarget
	CreateEntry
	CreateTarget
)

// FormField represents a single field in the form
type FormField struct {
	Name      string
	Value     string
	ReadOnly  bool
	IsNumeric bool
}

// FormConfig holds configuration for a form overlay
type FormConfig struct {
	Title  string
	Type   FormType
	Fields []FormField
}

// FormOverlayState tracks the current form state
type FormOverlayState struct {
	Active     bool
	Config     FormConfig
	FieldIndex int          // Current selected field (navigable inputs)
	OriginalID string       // ID of entry being edited (for edit operations)
	Error      string       // Error message to display in form
	Budget     interface{}  // Reference to budget for update/create operations
	OnSave     func() error // Callback after successful save
}

// =============================================================================
// MODAL TYPES (kept for backward compatibility)
// =============================================================================

// ModalType defines the kind of modal dialog
type ModalType int

const (
	InfoModal ModalType = iota
	ConfirmModal
)

// ModalConfig holds configuration for a modal dialog
type ModalConfig struct {
	Title   string
	Message string
	Type    ModalType
}

// ModalOverlayState tracks the current modal state
type ModalOverlayState struct {
	Active bool
	Config ModalConfig
}

// =============================================================================
// FORM STYLING
// =============================================================================

// formStyle returns the base style for a form overlay
func (f FormConfig) formStyle() lipgloss.Style {
	width := 60
	return lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#1a1a2e"))
}

// titleStyle returns the style for the form title
func (f FormConfig) titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Width(60).
		Align(lipgloss.Center)
}

// fieldLabelStyle returns the style for field labels
func fieldLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Width(18)
}

// fieldValueStyle returns the style for editable field values
func fieldValueStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFF00"))
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
}

// fieldReadOnlyStyle returns the style for read-only field values
func fieldReadOnlyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true)
}

// =============================================================================
// FORM HELPERS
// =============================================================================

// NewEntryFormConfig creates form configuration for editing/creating an Entry
func NewEntryFormConfig(entryName string, value float64, month, year int, cat string, realized bool, isNew bool) FormConfig {
	title := "Edit Entry"
	if isNew {
		title = "New Entry"
	}

	formType := EditEntry
	if isNew {
		formType = CreateEntry
	}

	// Build fields - ID field only shown when editing existing entry
	var fields []FormField
	if !isNew {
		// For editing, include ID as a read-only field (needed for update)
		// ID is reconstructed from name-month-year
		entryID := entryName + fmt.Sprintf("-%04d-%02d", year, month)
		fields = []FormField{
			{Name: "ID", Value: entryID, ReadOnly: true, IsNumeric: false},
			{Name: "Name", Value: entryName, ReadOnly: false, IsNumeric: false},
			{Name: "Month", Value: fmt.Sprintf("%d", month), ReadOnly: false, IsNumeric: true},
			{Name: "Year", Value: fmt.Sprintf("%d", year), ReadOnly: false, IsNumeric: true},
			{Name: "Value", Value: fmt.Sprintf("%.2f", value), ReadOnly: false, IsNumeric: true},
			{Name: "Category", Value: cat, ReadOnly: false, IsNumeric: false},
			{Name: "Realized", Value: boolToString(realized), ReadOnly: false, IsNumeric: false},
		}
	} else {
		// For creating new entry, don't include ID
		fields = []FormField{
			{Name: "Name", Value: entryName, ReadOnly: false, IsNumeric: false},
			{Name: "Month", Value: fmt.Sprintf("%d", month), ReadOnly: false, IsNumeric: true},
			{Name: "Year", Value: fmt.Sprintf("%d", year), ReadOnly: false, IsNumeric: true},
			{Name: "Value", Value: fmt.Sprintf("%.2f", value), ReadOnly: false, IsNumeric: true},
			{Name: "Category", Value: cat, ReadOnly: false, IsNumeric: false},
			{Name: "Realized", Value: boolToString(realized), ReadOnly: false, IsNumeric: false},
		}
	}

	return FormConfig{
		Title:  title,
		Type:   formType,
		Fields: fields,
	}
}

// boolToString converts bool to string representation
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// NewTargetFormConfig creates form configuration for editing/creating a Target
func NewTargetFormConfig(entryName string, targetValue, sidePocket float64, year int, isNew bool) FormConfig {
	title := "Edit Target"
	if isNew {
		title = "New Target"
	}

	formType := EditTarget
	if isNew {
		formType = CreateTarget
	}

	// Build fields - ID field only shown when editing existing entry
	var fields []FormField
	if !isNew {
		// For editing, include ID as a read-only field (needed for update)
		// ID is reconstructed from name-year
		// Note: Year is NOT editable for existing targets - it's set at creation time
		entryID := fmt.Sprintf("%s-%d", entryName, year)
		fields = []FormField{
			{Name: "ID", Value: entryID, ReadOnly: true, IsNumeric: false},
			{Name: "Name", Value: entryName, ReadOnly: false, IsNumeric: false},
			{Name: "Value", Value: fmt.Sprintf("%.2f", targetValue), ReadOnly: false, IsNumeric: true},
			{Name: "SidePocket", Value: fmt.Sprintf("%.2f", sidePocket), ReadOnly: false, IsNumeric: true},
		}
	} else {
		// For creating new entry, don't include ID
		fields = []FormField{
			{Name: "Name", Value: entryName, ReadOnly: false, IsNumeric: false},
			{Name: "Year", Value: fmt.Sprintf("%d", year), ReadOnly: false, IsNumeric: true},
			{Name: "Value", Value: fmt.Sprintf("%.2f", targetValue), ReadOnly: false, IsNumeric: true},
			{Name: "SidePocket", Value: fmt.Sprintf("%.2f", sidePocket), ReadOnly: false, IsNumeric: true},
		}
	}

	return FormConfig{
		Title:  title,
		Type:   formType,
		Fields: fields,
	}
}

// =============================================================================
// FORM RENDERING
// =============================================================================

// FormError holds error message to display in form
type FormError struct {
	Message string
}

// RenderForm renders a form overlay centered on the screen
func RenderForm(state FormOverlayState, err *FormError) string {
	if !state.Active {
		return ""
	}

	cfg := state.Config

	// Build the form components
	title := cfg.titleStyle().Render(cfg.Title)

	// Build field lines
	var fieldLines []string
	for i, field := range cfg.Fields {
		// Skip ID field (hidden)
		if field.Name == "ID" {
			continue
		}

		labelStyle := fieldLabelStyle()
		valueStyle := fieldValueStyle(i == state.FieldIndex)

		label := labelStyle.Render(field.Name + ":")

		var value string
		if field.ReadOnly {
			value = fieldReadOnlyStyle().Width(40).Render(field.Value)
		} else {
			// Handle Realized field as checkbox with left/right hint
			if field.Name == "Realized" {
				if field.Value == "true" {
					value = valueStyle.Width(40).Render("<☑>")
				} else {
					value = valueStyle.Width(40).Render("<☐>")
				}
				// Handle Category field with < > wrapping
			} else if field.Name == "Category" {
				value = valueStyle.Width(40).Render("<" + field.Value + ">")
			} else {
				value = valueStyle.Width(40).Render(field.Value)
			}
		}

		fieldLines = append(fieldLines, label+" "+value)
	}

	// Assemble the form
	formBody := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
	)

	for _, line := range fieldLines {
		formBody += "\n" + lipgloss.JoinHorizontal(lipgloss.Left, "  ", line)
	}

	// Add error message if present
	if err != nil && err.Message != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
		formBody += "\n\n" + errorStyle.Render("Error: "+err.Message)
	}

	// Add help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
	formBody += "\n" + helpStyle.Render("  ↑/↓: Navigate fields | ←/→: Change <Value> | Enter: Save")

	return cfg.formStyle().Render(formBody)
}

// =============================================================================
// ENTRY OPERATIONS
// =============================================================================

// ParseEntryForm parses form fields for creating a new Entry
// Returns: name, month, year, value, category, realized, error
func ParseEntryForm(fields []FormField) (string, int, int, float64, string, bool, error) {
	var name, category string
	var month, year, value float64
	var realized bool

	for _, field := range fields {
		switch field.Name {
		case "Name":
			name = strings.TrimSpace(field.Value)
		case "Month":
			fmt.Sscanf(field.Value, "%f", &month)
		case "Year":
			fmt.Sscanf(field.Value, "%f", &year)
		case "Value":
			fmt.Sscanf(field.Value, "%f", &value)
		case "Category":
			category = strings.TrimSpace(field.Value)
		case "Realized":
			realized = field.Value == "true" || field.Value == "1"
		}
	}

	if name == "" {
		return "", 0, 0, 0, "", false, fmt.Errorf("name is required")
	}

	return name, int(month), int(year), value, category, realized, nil
}

// ParseEntryFormWithID parses form fields for editing an existing Entry
// Returns: id, name, month, year, value, category, realized, error
func ParseEntryFormWithID(fields []FormField) (string, string, int, int, float64, string, bool, error) {
	var id, name, category string
	var month, year, value float64
	var realized bool

	for _, field := range fields {
		switch field.Name {
		case "ID":
			id = strings.TrimSpace(field.Value)
		case "Name":
			name = strings.TrimSpace(field.Value)
		case "Month":
			fmt.Sscanf(field.Value, "%f", &month)
		case "Year":
			fmt.Sscanf(field.Value, "%f", &year)
		case "Value":
			fmt.Sscanf(field.Value, "%f", &value)
		case "Category":
			category = strings.TrimSpace(field.Value)
		case "Realized":
			realized = field.Value == "true" || field.Value == "1"
		}
	}

	if name == "" {
		return "", "", 0, 0, 0, "", false, fmt.Errorf("name is required")
	}
	if id == "" {
		return "", "", 0, 0, 0, "", false, fmt.Errorf("entry ID is required")
	}

	return id, name, int(month), int(year), value, category, realized, nil
}

// =============================================================================
// TARGET OPERATIONS
// =============================================================================

// ParseTargetForm parses form fields for creating a new Target
// Returns: name, year, value, sidePocket, error
func ParseTargetForm(fields []FormField) (string, int, float64, float64, error) {
	var name string
	var year, value, sidePocket float64

	for _, field := range fields {
		switch field.Name {
		case "Name":
			name = strings.TrimSpace(field.Value)
		case "Year":
			fmt.Sscanf(field.Value, "%f", &year)
		case "Value":
			fmt.Sscanf(field.Value, "%f", &value)
		case "SidePocket":
			fmt.Sscanf(field.Value, "%f", &sidePocket)
		}
	}

	if name == "" {
		return "", 0, 0, 0, fmt.Errorf("name is required")
	}

	return name, int(year), value, sidePocket, nil
}

// ParseTargetFormWithID parses form fields for editing an existing Target
// Returns: id, name, year, value, sidePocket, error
func ParseTargetFormWithID(fields []FormField) (string, string, int, float64, float64, error) {
	var id, name string
	var year, value, sidePocket float64

	for _, field := range fields {
		switch field.Name {
		case "ID":
			id = strings.TrimSpace(field.Value)
		case "Name":
			name = strings.TrimSpace(field.Value)
		case "Year":
			fmt.Sscanf(field.Value, "%f", &year)
		case "Value":
			fmt.Sscanf(field.Value, "%f", &value)
		case "SidePocket":
			fmt.Sscanf(field.Value, "%f", &sidePocket)
		}
	}

	if name == "" {
		return "", "", 0, 0, 0, fmt.Errorf("name is required")
	}
	if id == "" {
		return "", "", 0, 0, 0, fmt.Errorf("target ID is required")
	}

	return id, name, int(year), value, sidePocket, nil
}
