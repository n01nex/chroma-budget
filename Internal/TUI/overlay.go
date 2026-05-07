package TUI

/* MODAL OVERLAY SYSTEM
This module provides a modal dialog overlay for displaying
information and getting user acknowledgment.
*/

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// =============================================================================
// MODAL TYPES
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
	Active   bool
	Config   ModalConfig
	Selected int // Selected button index (0 = OK/confirm)
}

// =============================================================================
// MODAL STYLING
// =============================================================================

// modalStyle returns the base style for a modal dialog
func (m ModalConfig) modalStyle() lipgloss.Style {
	width := 50
	if len(m.Message) > 40 {
		width = len(m.Message) + 10
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(10).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#1a1a2e")).
		Foreground(lipgloss.Color("#eaeaea")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		MarginTop(2)
}

// titleStyle returns the style for the modal title
func (m ModalConfig) titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Width(m.modalStyle().GetWidth()).
		Align(lipgloss.Center)
}

// messageStyle returns the style for the modal message
func (m ModalConfig) messageStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(m.modalStyle().GetWidth()).
		Align(lipgloss.Center).
		MarginTop(2)
}

// buttonStyle returns the style for a button based on selection state
func buttonStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFF00")).
			Bold(true).
			Padding(0, 2)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 2)
}

// =============================================================================
// MODAL RENDERING
// =============================================================================

// RenderModal renders a modal overlay centered on the screen
func RenderModal(state ModalOverlayState) string {
	if !state.Active {
		return ""
	}

	cfg := state.Config

	// Build the modal components
	title := cfg.titleStyle().Render(cfg.Title)

	// Wrap message if too long
	message := cfg.messageStyle().Render(wrapMessage(cfg.Message, 45))

	// OK button
	okButton := buttonStyle(state.Selected == 0).Render("[ OK ]")

	// Assemble the modal
	modalBody := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		message,
		"",
		okButton,
		"",
	)

	return cfg.modalStyle().Render(modalBody)
}

// wrapMessage wraps a message string to fit within a width
func wrapMessage(msg string, width int) string {
	if len(msg) <= width {
		return msg
	}

	// Simple word wrap
	words := ""
	result := ""
	for i, char := range msg {
		words += string(char)
		if (i+1)%width == 0 {
			result += words + "\n"
			words = ""
		}
	}
	if len(words) > 0 {
		result += words
	}
	return result
}

// OverlayDimensions calculates the overlay dimensions based on terminal size
func OverlayDimensions(terminalWidth, terminalHeight int) (overlayWidth, overlayHeight int) {
	overlayWidth = 60
	overlayHeight = 12

	if overlayWidth > terminalWidth-4 {
		overlayWidth = terminalWidth - 4
	}
	if overlayHeight > terminalHeight-4 {
		overlayHeight = terminalHeight - 4
	}

	return overlayWidth, overlayHeight
}

// =============================================================================
// MODAL FACTORY FUNCTIONS
// =============================================================================

// NewBudgetInfoModal creates a modal for displaying budget entry info
func NewBudgetInfoModal(name string, value float64, category string) ModalOverlayState {
	return ModalOverlayState{
		Active: true,
		Config: ModalConfig{
			Title:   "Budget Entry Details",
			Message: fmt.Sprintf("%s\n\nValue: %.2f\nCategory: %s", name, value, category),
			Type:    InfoModal,
		},
		Selected: 0,
	}
}

// NewTargetInfoModal creates a modal for displaying target entry info
func NewTargetInfoModal(name string, remaining float64) ModalOverlayState {
	return ModalOverlayState{
		Active: true,
		Config: ModalConfig{
			Title:   "Target Entry Details",
			Message: fmt.Sprintf("%s\n\nRemaining: %.2f", name, remaining),
			Type:    InfoModal,
		},
		Selected: 0,
	}
}