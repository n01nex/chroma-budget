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

// rowData holds the data for each rendered row along with category info for color mapping
type rowData struct {
	Name       string
	Value      float64
	Category   Data.Category // Income, Mandatory, Flexible, or "Buffer"
	Percentage float64       // Normalized percentage (0-1) for color intensity
}

// chromaModel is the tea.Model for the custom colored table view
type chromaModel struct {
	month         int
	year          int
	rows          []rowData // All data rows including separators and buffer
	selected      int       // Currently selected row index
	scrollOffset  int       // For scrolling when rows exceed visible area
	visibleHeight int       // Number of visible rows
	totalIncome   float64   // Total income for normalization
	totalExpenses float64   // Total expenses for normalization
	quit          bool      // Flag to signal quit
}

// RenderCustomTable returns a tea.Model for the new custom colored table view
func RenderCustomTable(b *Data.Budget, month int, year int) tea.Model {
	// Collect and categorize entries for the given month/year
	var incomeRows, mandatoryRows, flexibleRows []rowData
	var buffer float64
	var totalIncome, totalExpenses float64

	for _, entry := range b.Entries {
		if int(entry.MonthYear.Month()) == month && entry.MonthYear.Year() == year {
			row := rowData{
				Name:       entry.Name,
				Value:      entry.Value,
				Category:   entry.Type,
				Percentage: 0, // Will be computed after totals are known
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

	// 2. Separator after Income (empty row with special handling)
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
	rows = append(rows, rowData{Name: "Buffer", Value: buffer, Category: "Buffer", Percentage: 0})

	// Compute normalized percentages for color intensity
	// Income: percentage of total income
	for i := range rows {
		if rows[i].Category == Data.Income && totalIncome > 0 {
			rows[i].Percentage = rows[i].Value / totalIncome
		}
	}

	// Mandatory: percentage of total expenses (inverse - higher value = more intense red)
	for i := range rows {
		if rows[i].Category == Data.Mandatory && totalExpenses > 0 {
			rows[i].Percentage = rows[i].Value / totalExpenses
		}
	}

	// Flexible: percentage of total expenses (inverse - higher value = darker blue)
	for i := range rows {
		if rows[i].Category == Data.Flexible && totalExpenses > 0 {
			rows[i].Percentage = rows[i].Value / totalExpenses
		}
	}

	// Buffer: 0 if zero, non-zero otherwise (handled separately in rendering)
	_ = buffer // Already set in the buffer row

	return chromaModel{
		month:         month,
		year:          year,
		rows:          rows,
		selected:      0,
		scrollOffset:  0,
		visibleHeight: 12, // Default, will be updated in Init
		totalIncome:   totalIncome,
		totalExpenses: totalExpenses,
		quit:          false,
	}
}

// Init initializes the model
func (m chromaModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m chromaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.visibleHeight = msg.Height - 6 // Account for title and help area
		if m.visibleHeight < 5 {
			m.visibleHeight = 5
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit

		case "up", "w", "W":
			if m.selected > 0 {
				m.selected--
				// Adjust scroll offset if needed
				if m.selected < m.scrollOffset {
					m.scrollOffset = m.selected
				}
			}

		case "down", "s", "S":
			if m.selected < len(m.rows)-1 {
				m.selected++
				// Adjust scroll offset if needed
				if m.selected >= m.scrollOffset+m.visibleHeight {
					m.scrollOffset = m.selected - m.visibleHeight + 1
				}
			}

		case "enter":
			// Placeholder handler for future row editing
			selectedRow := m.rows[m.selected]
			return m, tea.Printf("Selected: %s (%.2f) - Category: %s", selectedRow.Name, selectedRow.Value, selectedRow.Category)
		}

		// Ensure scroll offset is within bounds
		if m.scrollOffset > len(m.rows)-m.visibleHeight && len(m.rows) > m.visibleHeight {
			m.scrollOffset = len(m.rows) - m.visibleHeight
		}
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
	}

	return m, nil
}

// View renders the custom colored table
func (m chromaModel) View() tea.View {
	// Title
	monthYear := fmt.Sprintf("%s %d", time.Month(m.month).String(), m.year)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Width(36).
		Align(lipgloss.Center).
		MarginBottom(1)

	title := titleStyle.Render(monthYear)

	// Build the table view
	var tableLines []string

	// Column headers
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#555555"))

	nameHeader := headerStyle.Width(20).Render("Name")
	valueHeader := headerStyle.Width(12).Render("Value")
	tableLines = append(tableLines, nameHeader+" "+valueHeader)

	// Calculate which rows are visible
	endIndex := m.scrollOffset + m.visibleHeight
	if endIndex > len(m.rows) {
		endIndex = len(m.rows)
	}

	// Render each visible row
	for i := m.scrollOffset; i < endIndex; i++ {
		row := m.rows[i]
		displayIndex := i - m.scrollOffset // Index within visible area

		// Handle separators
		if row.Category == "Separator" {
			sepStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#444444"))
			tableLines = append(tableLines, sepStyle.Render("─"+strings.Repeat("─", 18)+" "+strings.Repeat("─", 10)))
			continue
		}

		// Get the style for this row based on category and intensity
		rowStyle := m.getRowStyle(row, displayIndex == m.selected-m.scrollOffset)

		nameContent := rowStyle.Width(20).Render(row.Name)
		valueContent := rowStyle.Width(12).Render(fmt.Sprintf("%.2f", row.Value))

		tableLines = append(tableLines, nameContent+" "+valueContent)
	}

	// Join table lines
	tableStr := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Render(strings.Join(tableLines, "\n"))

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	helpText := helpStyle.Render("↑/↓ or W/S: Navigate | Enter: Edit | Q: Quit")

	return tea.NewView("\n" + title + "\n" + tableStr + "\n  " + helpText + "\n")
}

// getRowStyle returns the appropriate lipgloss style for a row based on its category and percentage
func (m chromaModel) getRowStyle(row rowData, isSelected bool) lipgloss.Style {
	// Base style with white text
	baseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	var coloredStyle lipgloss.Style

	switch row.Category {
	case Data.Income:
		// Green background: intensity scales with percentage (0.2 to 0.9 green)
		// Higher income proportion = brighter green
		// Range: #005500 (dark) to #00FF00 (bright)
		greenIntensity := int(row.Percentage * 255)
		if greenIntensity > 255 {
			greenIntensity = 255
		}
		if greenIntensity < 0 {
			greenIntensity = 0
		}
		colorHex := fmt.Sprintf("#00%02X00", greenIntensity)
		coloredStyle = baseStyle.Background(lipgloss.Color(colorHex))

	case Data.Mandatory:
		// Red background: intensity increases as value grows relative to total expenses
		// Higher percentage = more intense red (darker red, like #FF0000 to #990000)
		redIntensity := 255 - int(row.Percentage*155)
		if redIntensity < 100 {
			redIntensity = 100
		}
		colorHex := fmt.Sprintf("#%02X0000", redIntensity)
		coloredStyle = baseStyle.Background(lipgloss.Color(colorHex))

	case Data.Flexible:
		// Blue background: darkens as value rises
		// Higher percentage = darker blue
		// Range: #0000FF (bright) to #000055 (dark)
		blueIntensity := 255 - int(row.Percentage*155)
		if blueIntensity < 100 {
			blueIntensity = 100
		}
		colorHex := fmt.Sprintf("#0000%02X", blueIntensity)
		coloredStyle = baseStyle.Background(lipgloss.Color(colorHex))

	case "Buffer":
		if row.Value != 0 {
			// Yellow background if non-zero
			coloredStyle = baseStyle.Background(lipgloss.Color("#BBBB00"))
		} else {
			// Gray background if zero
			coloredStyle = baseStyle.Background(lipgloss.Color("#555555"))
		}
	}

	// Apply selection highlight
	if isSelected {
		selectionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFF00")).
			Bold(true)
		return selectionStyle
	}

	return coloredStyle
}

// RunCustomTable runs the custom colored table view
func RunCustomTable(b *Data.Budget, month int, year int) {
	m := RenderCustomTable(b, month, year)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running table:", err)
		os.Exit(1)
	}
}
