package TUI

/* PACKAGES DOCS:
https://pkg.go.dev/charm.land/lipgloss/v2@v2.0.3
https://pkg.go.dev/charm.land/bubbles/v2
https://pkg.go.dev/github.com/charmbracelet/bubbletea/v2
*/

import (
	"StructData/Internal/Data"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// =============================================================================
// CONSTANTS & CONFIGURATION
// =============================================================================

const (
	// PanelPadding controls the horizontal distance between the two tables
	PanelPadding = 2
)

// =============================================================================
// TYPES
// =============================================================================

// Panel represents which table panel has focus
type Panel int

const (
	LeftPanel  Panel = iota // Monthly Budget entries (left side)
	RightPanel              // Yearly Targets (right side)
)

// rowData holds the data for each rendered row in the budget table
type rowData struct {
	Name       string
	Value      float64
	Category   Data.Category // Income, Mandatory, Flexible, or "Separator"/"Buffer"
	Percentage float64       // Normalized percentage (0-1) for color intensity
	Realized   bool          // Whether this entry has been realized (for strikethrough)
}

// targetRowData holds the data for each rendered row in the Targets table
type targetRowData struct {
	Name            string
	Value           float64
	Year            int
	CurrentPlanned  float64
	CurrentRealized float64
	SidePocket      float64
	Remaining       float64
	Percentage      float64 // Normalized percentage (0-1) for color intensity
}

// chromaModel is the tea.Model for the custom colored dual-panel table view
type chromaModel struct {
	budget *Data.Budget // Reference to budget data for month navigation
	month  int
	year   int
	quit   bool

	// Left panel (monthly budget entries)
	leftRows          []rowData
	leftSelected      int
	leftScrollOffset  int
	leftVisibleHeight int

	// Right panel (yearly targets)
	rightRows          []targetRowData
	rightSelected      int
	rightScrollOffset  int
	rightVisibleHeight int

	// Focus management
	activePanel Panel

	// Shared styling data
	totalIncome   float64
	totalExpenses float64
}

// =============================================================================
// DATA BUILDING FUNCTIONS
// =============================================================================

// buildRowsFromBudget extracts and organizes budget entries for a given month/year
// Returns the rows, total income, and total expenses
func buildRowsFromBudget(b *Data.Budget, month, year int) ([]rowData, float64, float64) {
	var incomeRows, mandatoryRows, flexibleRows []rowData
	var buffer float64
	var totalIncome, totalExpenses float64

	for _, entry := range b.Entries {
		if int(entry.MonthYear.Month()) == month && entry.MonthYear.Year() == year {
			row := rowData{
				Name:       entry.Name,
				Value:      entry.Value,
				Category:   entry.Type,
				Percentage: 0,
				Realized:   entry.Realized,
			}

			switch entry.Type {
			case Data.Income:
				incomeRows = append(incomeRows, row)
				totalIncome += entry.Value
			case Data.Mandatory:
				mandatoryRows = append(mandatoryRows, row)
				totalExpenses += entry.Value
			case Data.Flexible:
				flexibleRows = append(flexibleRows, row)
				totalExpenses += entry.Value
			}
			buffer += entry.Value
		}
	}

	// Build the full rows slice in desired order
	var rows []rowData

	// 1. Income entries
	rows = append(rows, incomeRows...)

	// 2. Separator after Income
	if len(mandatoryRows) > 0 || len(flexibleRows) > 0 || len(incomeRows) > 0 {
		rows = append(rows, rowData{Name: "─", Value: 0, Category: "Separator", Percentage: 0})
	}

	// 3. Mandatory entries
	rows = append(rows, mandatoryRows...)

	// 4. Separator after Mandatory
	if len(flexibleRows) > 0 {
		rows = append(rows, rowData{Name: "─", Value: 0, Category: "Separator", Percentage: 0})
	}

	// 5. Flexible entries
	rows = append(rows, flexibleRows...)

	// 6. Separator before Buffer
	if len(rows) > 0 {
		rows = append(rows, rowData{Name: "─", Value: 0, Category: "Separator", Percentage: 0})
	}

	// 7. Buffer line
	rows = append(rows, rowData{Name: "Buffer", Value: buffer, Category: "Buffer", Percentage: 0, Realized: false})

	// Compute normalized percentages
	for i := range rows {
		if rows[i].Category == Data.Income && totalIncome > 0 {
			rows[i].Percentage = rows[i].Value / totalIncome
		}
	}
	for i := range rows {
		if rows[i].Category == Data.Mandatory && totalExpenses > 0 {
			rows[i].Percentage = rows[i].Value / totalExpenses
		}
	}
	for i := range rows {
		if rows[i].Category == Data.Flexible && totalExpenses > 0 {
			rows[i].Percentage = rows[i].Value / totalExpenses
		}
	}

	return rows, totalIncome, totalExpenses
}

// buildTargetRows extracts and organizes Targets data for a given year
func buildTargetRows(b *Data.Budget, year int) []targetRowData {
	var rows []targetRowData

	for _, t := range b.Targets {
		if t.Year == year {
			row := targetRowData{
				Name:            t.EntryName,
				Value:           t.Value,
				Year:            t.Year,
				CurrentPlanned:  t.CurrentPlanned,
				CurrentRealized: t.CurrentRealized,
				SidePocket:      t.SidePocket,
				Remaining:       t.Remaining,
				Percentage:      0,
			}
			rows = append(rows, row)
		}
	}

	// Compute normalized percentages for Remaining column
	var totalRemaining float64
	for i := range rows {
		if rows[i].Remaining > 0 {
			totalRemaining += rows[i].Remaining
		}
	}
	for i := range rows {
		if rows[i].Remaining > 0 && totalRemaining > 0 {
			rows[i].Percentage = rows[i].Remaining / totalRemaining
		}
	}

	return rows
}

// =============================================================================
// MODEL FACTORY
// =============================================================================

// RenderCustomTable returns a tea.Model for the new custom colored dual-panel table view
func RenderCustomTable(b *Data.Budget, month int, year int) tea.Model {
	leftRows, totalIncome, totalExpenses := buildRowsFromBudget(b, month, year)
	rightRows := buildTargetRows(b, year)

	return chromaModel{
		budget:             b,
		month:              month,
		year:               year,
		leftRows:           leftRows,
		leftSelected:       0,
		leftScrollOffset:   0,
		leftVisibleHeight:  12,
		rightRows:          rightRows,
		rightSelected:      0,
		rightScrollOffset:  0,
		rightVisibleHeight: 12,
		activePanel:        LeftPanel,
		totalIncome:        totalIncome,
		totalExpenses:      totalExpenses,
		quit:               false,
	}
}

// reloadRows rebuilds the rows data from the budget for the current month/year
func (m *chromaModel) reloadRows() {
	m.leftRows, m.totalIncome, m.totalExpenses = buildRowsFromBudget(m.budget, m.month, m.year)
	m.rightRows = buildTargetRows(m.budget, m.year)
}

// =============================================================================
// TEA MODEL INTERFACE
// =============================================================================

// Init initializes the model
func (m chromaModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m chromaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Calculate available height for each panel (accounting for headers and margins)
		// Reduce height by 25% to make TUI more compact
		reducedHeight := int(float64(msg.Height) * 0.75)
		m.leftVisibleHeight = reducedHeight - 7 // Title, header, help, padding
		m.rightVisibleHeight = reducedHeight - 7
		if m.leftVisibleHeight < 3 {
			m.leftVisibleHeight = 3
		}
		if m.rightVisibleHeight < 3 {
			m.rightVisibleHeight = 3
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit

		case "tab":
			// Switch focus between panels
			if m.activePanel == LeftPanel {
				m.activePanel = RightPanel
			} else {
				m.activePanel = LeftPanel
			}
			return m, nil

		case "up", "w", "W":
			m.navigateUp()
			return m, nil

		case "down", "s", "S":
			m.navigateDown()
			return m, nil

		case "left", "a", "A":
			// Navigate to previous month
			if m.month == 1 {
				m.month = 12
				m.year--
			} else {
				m.month--
			}
			m.reloadRows()

		case "right", "d", "D":
			// Navigate to next month
			if m.month == 12 {
				m.month = 1
				m.year++
			} else {
				m.month++
			}
			m.reloadRows()

		case "enter":
			// Placeholder handler for future row editing
			if m.activePanel == LeftPanel {
				if len(m.leftRows) > 0 && m.leftSelected < len(m.leftRows) {
					selectedRow := m.leftRows[m.leftSelected]
					return m, tea.Printf("Selected Budget: %s (%.2f) - Category: %s",
						selectedRow.Name, selectedRow.Value, selectedRow.Category)
				}
			} else {
				if len(m.rightRows) > 0 && m.rightSelected < len(m.rightRows) {
					selectedRow := m.rightRows[m.rightSelected]
					return m, tea.Printf("Selected Target: %s (Remaining: %.2f)",
						selectedRow.Name, selectedRow.Remaining)
				}
			}
		}

		// Ensure scroll offsets are within bounds
		m.clampScrollOffsets()
	}

	return m, nil
}

// navigateUp moves selection up in the active panel
func (m *chromaModel) navigateUp() {
	if m.activePanel == LeftPanel {
		if m.leftSelected > 0 {
			m.leftSelected--
			if m.leftSelected < m.leftScrollOffset {
				m.leftScrollOffset = m.leftSelected
			}
		}
	} else {
		if m.rightSelected > 0 {
			m.rightSelected--
			if m.rightSelected < m.rightScrollOffset {
				m.rightScrollOffset = m.rightSelected
			}
		}
	}
}

// navigateDown moves selection down in the active panel
func (m *chromaModel) navigateDown() {
	if m.activePanel == LeftPanel {
		if m.leftSelected < len(m.leftRows)-1 {
			m.leftSelected++
			if m.leftSelected >= m.leftScrollOffset+m.leftVisibleHeight {
				m.leftScrollOffset = m.leftSelected - m.leftVisibleHeight + 1
			}
		}
	} else {
		if m.rightSelected < len(m.rightRows)-1 {
			m.rightSelected++
			if m.rightSelected >= m.rightScrollOffset+m.rightVisibleHeight {
				m.rightScrollOffset = m.rightSelected - m.rightVisibleHeight + 1
			}
		}
	}
}

// clampScrollOffsets ensures scroll offsets stay within valid bounds
func (m *chromaModel) clampScrollOffsets() {
	// Left panel
	if len(m.leftRows) > m.leftVisibleHeight {
		maxScroll := len(m.leftRows) - m.leftVisibleHeight
		if m.leftScrollOffset > maxScroll {
			m.leftScrollOffset = maxScroll
		}
	} else {
		m.leftScrollOffset = 0
	}
	if m.leftScrollOffset < 0 {
		m.leftScrollOffset = 0
	}
	if m.leftSelected >= len(m.leftRows) {
		m.leftSelected = len(m.leftRows) - 1
	}
	if m.leftSelected < 0 {
		m.leftSelected = 0
	}

	// Right panel
	if len(m.rightRows) > m.rightVisibleHeight {
		maxScroll := len(m.rightRows) - m.rightVisibleHeight
		if m.rightScrollOffset > maxScroll {
			m.rightScrollOffset = maxScroll
		}
	} else {
		m.rightScrollOffset = 0
	}
	if m.rightScrollOffset < 0 {
		m.rightScrollOffset = 0
	}
	if m.rightSelected >= len(m.rightRows) {
		m.rightSelected = len(m.rightRows) - 1
	}
	if m.rightSelected < 0 {
		m.rightSelected = 0
	}
}

// =============================================================================
// VIEW RENDERING
// =============================================================================

// View renders the dual-panel colored table
func (m chromaModel) View() tea.View {
	// Build left panel (monthly budget)
	leftTitle := fmt.Sprintf("%s %d", time.Month(m.month).String(), m.year)
	leftView := m.renderBudgetPanel(leftTitle, true)

	// Build right panel (yearly targets)
	rightTitle := fmt.Sprintf("Targets %d", m.year)
	rightView := m.renderTargetPanel(rightTitle, true)

	// Combine panels side by side with padding
	leftStyled := lipgloss.NewStyle().Render(leftView)
	rightStyled := lipgloss.NewStyle().Render(rightView)

	combined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyled,
		strings.Repeat(" ", PanelPadding),
		rightStyled,
	)

	// Help text
	helpText := m.renderHelpText()

	return tea.NewView("\n" + combined + "\n" + helpText + "\n")
}

// renderBudgetPanel renders the left panel (monthly budget entries)
func (m chromaModel) renderBudgetPanel(title string, isActive bool) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Width(36).
		Align(lipgloss.Center).
		MarginBottom(1)

	panelTitle := titleStyle.Render(title)

	// Build table lines
	var tableLines []string

	// Column headers
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#555555"))

	nameHeader := headerStyle.Width(18).Render("Name")
	valueHeader := headerStyle.Width(14).Render("Value")
	tableLines = append(tableLines, nameHeader+" "+valueHeader)

	// Calculate visible range
	endIndex := m.leftScrollOffset + m.leftVisibleHeight
	if endIndex > len(m.leftRows) {
		endIndex = len(m.leftRows)
	}

	// Render each visible row
	for i := m.leftScrollOffset; i < endIndex; i++ {
		row := m.leftRows[i]
		displayIndex := i - m.leftScrollOffset
		isSelected := displayIndex == m.leftSelected-m.leftScrollOffset

		// Handle separators
		if row.Category == "Separator" {
			sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
			tableLines = append(tableLines, sepStyle.Render("─"+strings.Repeat("─", 16)+" "+strings.Repeat("─", 12)))
			continue
		}

		// Get row style
		rowStyle := m.getBudgetRowStyle(row, isSelected, isActive)

		nameContent := rowStyle.Width(18).Render(row.Name)
		valueContent := rowStyle.Width(14).Render(fmt.Sprintf("%.2f", row.Value))

		tableLines = append(tableLines, nameContent+" "+valueContent)
	}

	// Add padding rows if needed (to match right panel height)
	paddingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a"))
	visibleCount := endIndex - m.leftScrollOffset
	for i := visibleCount; i < m.leftVisibleHeight; i++ {
		tableLines = append(tableLines, paddingStyle.Render(" "))
	}

	// Join table lines
	tableStr := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Render(strings.Join(tableLines, "\n"))

	return panelTitle + "\n" + tableStr
}

// renderTargetPanel renders the right panel (yearly targets)
func (m chromaModel) renderTargetPanel(title string, isActive bool) string {
	// Find the max name length for dynamic column sizing
	maxNameLen := 10 // minimum width
	for _, row := range m.rightRows {
		if len(row.Name) > maxNameLen {
			maxNameLen = len(row.Name)
		}
	}
	if maxNameLen > 20 {
		maxNameLen = 20 // cap at 20
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#F4A556")).
		Width(maxNameLen + 50). // Extra width for all columns
		Align(lipgloss.Center).
		MarginBottom(1)

	panelTitle := titleStyle.Render(title)

	// Build table lines
	var tableLines []string

	// Column headers
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#555555"))

	nameHeader := headerStyle.Width(maxNameLen).Render("Name")
	targetHeader := headerStyle.Width(9).Render("Target")
	plannedHeader := headerStyle.Width(9).Render("Planned")
	realizedHeader := headerStyle.Width(9).Render("Realized")
	sidePocketHeader := headerStyle.Width(9).Render("SidePoc")
	remainingHeader := headerStyle.Width(9).Render("Remaining")
	tableLines = append(tableLines, nameHeader+" "+targetHeader+" "+plannedHeader+" "+realizedHeader+" "+sidePocketHeader+" "+remainingHeader)

	// Calculate visible range
	endIndex := m.rightScrollOffset + m.rightVisibleHeight
	if endIndex > len(m.rightRows) {
		endIndex = len(m.rightRows)
	}

	// Render each visible row
	for i := m.rightScrollOffset; i < endIndex; i++ {
		row := m.rightRows[i]
		displayIndex := i - m.rightScrollOffset
		isSelected := displayIndex == m.rightSelected-m.rightScrollOffset

		rowStyle := m.getTargetRowStyle(row, isSelected, isActive)

		nameContent := rowStyle.Width(maxNameLen).Render(row.Name)
		targetContent := rowStyle.Width(9).Render(fmt.Sprintf("%.2f", row.Value))
		plannedContent := rowStyle.Width(9).Render(fmt.Sprintf("%.2f", row.CurrentPlanned))
		realizedContent := rowStyle.Width(9).Render(fmt.Sprintf("%.2f", row.CurrentRealized))
		sidePocketContent := rowStyle.Width(9).Render(fmt.Sprintf("%.2f", row.SidePocket))
		remainingContent := rowStyle.Width(9).Render(fmt.Sprintf("%.2f", row.Remaining))

		tableLines = append(tableLines, nameContent+" "+targetContent+" "+plannedContent+" "+realizedContent+" "+sidePocketContent+" "+remainingContent)
	}

	// Add padding rows if needed (to match left panel height)
	paddingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a"))
	visibleCount := endIndex - m.rightScrollOffset
	paddingWidth := maxNameLen + 50 // Total width of all columns plus spaces
	for i := visibleCount; i < m.rightVisibleHeight; i++ {
		tableLines = append(tableLines, paddingStyle.Render(strings.Repeat(" ", paddingWidth)))
	}


	// Join table lines
	tableStr := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Render(strings.Join(tableLines, "\n"))

	return panelTitle + "\n" + tableStr
}

// renderHelpText returns the help text at the bottom of the view
func (m chromaModel) renderHelpText() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	focusIndicator := " [LEFT]"
	if m.activePanel == RightPanel {
		focusIndicator = " [RIGHT]"
	}

	helpText := helpStyle.Render(
		"↑/↓ or W/S: Navigate rows" + focusIndicator +
			" | ←/→ or A/D: Change month" +
			" | Tab: Switch panel" +
			" | Enter: Info" +
			" | Q: Quit",
	)

	return "  " + helpText
}

// =============================================================================
// STYLING HELPERS
// =============================================================================

// getBudgetRowStyle returns the appropriate lipgloss style for a budget row
func (m chromaModel) getBudgetRowStyle(row rowData, isSelected, isActive bool) lipgloss.Style {
	// Base style with white text
	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	var coloredStyle lipgloss.Style

	switch row.Category {
	case Data.Income:
		// Green background: intensity scales with percentage (0.2 to 0.9 green)
		greenIntensity := int(row.Percentage * 255)
		if greenIntensity > 255 {
			greenIntensity = 255
		}
		if greenIntensity < 50 {
			greenIntensity = 50
		}
		colorHex := fmt.Sprintf("#00%02X00", greenIntensity)
		coloredStyle = baseStyle.Background(lipgloss.Color(colorHex))

	case Data.Mandatory:
		// Red background: intensity increases as value grows relative to total expenses
		redIntensity := 255 - int(row.Percentage*155)
		if redIntensity < 100 {
			redIntensity = 100
		}
		colorHex := fmt.Sprintf("#%02X0000", redIntensity)
		coloredStyle = baseStyle.Background(lipgloss.Color(colorHex))

	case Data.Flexible:
		// Blue background: darkens as value rises
		blueIntensity := 255 - int(row.Percentage*155)
		if blueIntensity < 100 {
			blueIntensity = 100
		}
		colorHex := fmt.Sprintf("#0000%02X", blueIntensity)
		coloredStyle = baseStyle.Background(lipgloss.Color(colorHex))

	case "Buffer":
		if row.Value != 0 {
			coloredStyle = baseStyle.Background(lipgloss.Color("#BBBB00"))
		} else {
			coloredStyle = baseStyle.Background(lipgloss.Color("#555555"))
		}

	case "Separator":
		coloredStyle = baseStyle
	}

	// Apply selection highlight when active and selected
	if isSelected {
		if isActive {
			// Active panel: bright yellow selection
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#FFFF00")).
				Bold(true)
			if row.Realized {
				style = style.Strikethrough(true)
			}
			return style
		} else {
			// Inactive panel but selected: dim highlight
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC")).
				Background(lipgloss.Color("#333333"))
			if row.Realized {
				style = style.Strikethrough(true)
			}
			return style
		}
	}

	// Apply strikethrough for realized entries (even when not selected)
	if row.Realized {
		coloredStyle = coloredStyle.Strikethrough(true)
	}

	// Inactive panel: add subtle border to indicate it's not focused
	if !isActive {
		return coloredStyle.BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("236"))
	}

	return coloredStyle
}

// getTargetRowStyle returns the appropriate lipgloss style for a target row
// Note: Currently no background coloring as per requirements
func (m chromaModel) getTargetRowStyle(row targetRowData, isSelected, isActive bool) lipgloss.Style {
	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Apply selection highlight when active and selected
	if isSelected {
		if isActive {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#FFFF00")).
				Bold(true)
		} else {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC")).
				Background(lipgloss.Color("#333333"))
		}
	}

	if !isActive {
		return baseStyle.BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("236"))
	}

	return baseStyle
}

// =============================================================================
// RUNNER
// =============================================================================

// RunCustomTable runs the custom colored table view
func RunCustomTable(b *Data.Budget, month int, year int) {
	m := RenderCustomTable(b, month, year)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running table:", err)
		os.Exit(1)
	}
}
