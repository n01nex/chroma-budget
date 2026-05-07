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
	width := 60 // default width, will be overridden by Place

	return lipgloss.NewStyle().
		Width(width).
		Height(12).
		Background(lipgloss.Color("#1a1a2e")).
		Foreground(lipgloss.Color("#eaeaea")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4"))
}

// titleStyle returns the style for the modal title
func (m ModalConfig) titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Width(60).
		Align(lipgloss.Center)
}

// messageStyle returns the style for the modal message
func (m ModalConfig) messageStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#eaeaea"))
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

	// Render message - don't pre-wrap, let lipgloss handle width
	message := cfg.messageStyle().Render(cfg.Message)

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
func NewTargetInfoModal(name string, target, planned, realized, sidePocket, remaining float64) ModalOverlayState {
	return ModalOverlayState{
		Active: true,
		Config: ModalConfig{
			Title: "Target Entry Details",
			Message: fmt.Sprintf("%s\n\nTarget: %.2f\nPlanned: %.2f\nRealized: %.2f\nSide Pocket: %.2f\nRemaining: %.2f",
				name, target, planned, realized, sidePocket, remaining),
			Type: InfoModal,
		},
		Selected: 0,
	}
}
