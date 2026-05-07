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
	MonthYear  time.Time     // The actual month/year of this entry (for edit form correctness)
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

// =============================================================================
// COMPOSABLE VIEWS
// =============================================================================

// BudgetPanelView is a composable view for the monthly budget panel
type BudgetPanelView struct {
	Title         string
	Rows          []rowData
	Selected      int
	ScrollOffset  int
	VisibleHeight int
	IsFocused     bool
	TotalIncome   float64
	TotalExpenses float64
}

// NewBudgetPanelView creates a new budget panel view
func NewBudgetPanelView(title string, rows []rowData, selected, scrollOffset, visibleHeight int, isFocused bool, totalIncome, totalExpenses float64) *BudgetPanelView {
	return &BudgetPanelView{
		Title:         title,
		Rows:          rows,
		Selected:      selected,
		ScrollOffset:  scrollOffset,
		VisibleHeight: visibleHeight,
		IsFocused:     isFocused,
		TotalIncome:   totalIncome,
		TotalExpenses: totalExpenses,
	}
}

// Render returns the string representation of the budget panel
func (v *BudgetPanelView) Render() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Width(36).
		Align(lipgloss.Center).
		MarginBottom(1)

	panelTitle := titleStyle.Render(v.Title)

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
	endIndex := v.ScrollOffset + v.VisibleHeight
	if endIndex > len(v.Rows) {
		endIndex = len(v.Rows)
	}

	// Render each visible row
	for i := v.ScrollOffset; i < endIndex; i++ {
		row := v.Rows[i]
		displayIndex := i - v.ScrollOffset
		isSelected := displayIndex == v.Selected-v.ScrollOffset

		// Handle separators
		if row.Category == "Separator" {
			sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
			tableLines = append(tableLines, sepStyle.Render("─"+strings.Repeat("─", 16)+" "+strings.Repeat("─", 12)))
			continue
		}

		// Get row style
		rowStyle := v.getRowStyle(row, isSelected)

		nameContent := rowStyle.Width(18).Render(row.Name)
		valueContent := rowStyle.Width(14).Render(fmt.Sprintf("%.2f", row.Value))

		tableLines = append(tableLines, nameContent+" "+valueContent)
	}

	// Add padding rows if needed (to match right panel height)
	paddingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a"))
	visibleCount := endIndex - v.ScrollOffset
	for i := visibleCount; i < v.VisibleHeight; i++ {
		tableLines = append(tableLines, paddingStyle.Render(" "))
	}

	// Join table lines - use consistent border style to prevent dimension changes
	borderColor := lipgloss.Color("240")
	if v.IsFocused {
		borderColor = lipgloss.Color("#FFFF00")
	}
	tableStr := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Render(strings.Join(tableLines, "\n"))

	return panelTitle + "\n" + tableStr
}

// getRowStyle returns the appropriate lipgloss style for a budget row
func (v *BudgetPanelView) getRowStyle(row rowData, isSelected bool) lipgloss.Style {
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
		if v.IsFocused {
			// Focused panel: bright yellow selection
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#FFFF00")).
				Bold(true)
			if row.Realized {
				style = style.Strikethrough(true)
			}
			return style
		} else {
			// Unfocused panel but selected: dim highlight
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

	// No border on individual rows - borders are only on the table container
	return coloredStyle
}

// TargetPanelView is a composable view for the yearly targets panel
type TargetPanelView struct {
	Title         string
	Rows          []targetRowData
	Selected      int
	ScrollOffset  int
	VisibleHeight int
	IsFocused     bool
}

// NewTargetPanelView creates a new targets panel view
func NewTargetPanelView(title string, rows []targetRowData, selected, scrollOffset, visibleHeight int, isFocused bool) *TargetPanelView {
	return &TargetPanelView{
		Title:         title,
		Rows:          rows,
		Selected:      selected,
		ScrollOffset:  scrollOffset,
		VisibleHeight: visibleHeight,
		IsFocused:     isFocused,
	}
}

// Render returns the string representation of the targets panel
func (v *TargetPanelView) Render() string {
	// Find the max name length for dynamic column sizing
	maxNameLen := 10 // minimum width
	for _, row := range v.Rows {
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
		Width(maxNameLen + 50).
		Align(lipgloss.Center).
		MarginBottom(1)

	panelTitle := titleStyle.Render(v.Title)

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
	endIndex := v.ScrollOffset + v.VisibleHeight
	if endIndex > len(v.Rows) {
		endIndex = len(v.Rows)
	}

	// Render each visible row
	for i := v.ScrollOffset; i < endIndex; i++ {
		row := v.Rows[i]
		displayIndex := i - v.ScrollOffset
		isSelected := displayIndex == v.Selected-v.ScrollOffset

		rowStyle := v.getRowStyle(isSelected)

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
	visibleCount := endIndex - v.ScrollOffset
	paddingWidth := maxNameLen + 50 // Total width of all columns plus spaces
	for i := visibleCount; i < v.VisibleHeight; i++ {
		tableLines = append(tableLines, paddingStyle.Render(strings.Repeat(" ", paddingWidth)))
	}

	// Join table lines - use consistent border style to prevent dimension changes
	borderColor := lipgloss.Color("240")
	if v.IsFocused {
		borderColor = lipgloss.Color("#FFFF00")
	}
	tableStr := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Render(strings.Join(tableLines, "\n"))

	return panelTitle + "\n" + tableStr
}

// getRowStyle returns the appropriate lipgloss style for a target row
func (v *TargetPanelView) getRowStyle(isSelected bool) lipgloss.Style {
	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Apply selection highlight when active and selected
	if isSelected {
		if v.IsFocused {
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

	// No border on individual rows - borders are only on the table container
	return baseStyle
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

	// Modal overlay state (info modals - deprecated, using formOverlay instead)
	modalOverlay ModalOverlayState

	// Form overlay state (dual-purpose for edit/create)
	formOverlay FormOverlayState
}

// budgetView returns a composable view for the left panel
func (m *chromaModel) budgetView() *BudgetPanelView {
	return NewBudgetPanelView(
		fmt.Sprintf("%s %d", time.Month(m.month).String(), m.year),
		m.leftRows,
		m.leftSelected,
		m.leftScrollOffset,
		m.leftVisibleHeight,
		m.activePanel == LeftPanel,
		m.totalIncome,
		m.totalExpenses,
	)
}

// targetView returns a composable view for the right panel
func (m *chromaModel) targetView() *TargetPanelView {
	return NewTargetPanelView(
		fmt.Sprintf("Targets %d", m.year),
		m.rightRows,
		m.rightSelected,
		m.rightScrollOffset,
		m.rightVisibleHeight,
		m.activePanel == RightPanel,
	)
}

// =============================================================================
// DATA BUILDING FUNCTIONS
// =============================================================================

// buildRowsFromBudget extracts and organizes budget entries for a given month/year
// Returns the rows, total income, and total expenses
func buildRowsFromBudget(b *Data.Budget, month, year int) ([]rowData, float64, float64) {
	var incomeRows, mandatoryRows, flexibleRows []rowData
	var totalIncome, totalExpenses float64

	for _, entry := range b.Entries {
		if int(entry.MonthYear.Month()) == month && entry.MonthYear.Year() == year {
			row := rowData{
				Name:       entry.Name,
				Value:      entry.Value,
				Category:   entry.Type,
				Percentage: 0,
				Realized:   entry.Realized,
				MonthYear:  entry.MonthYear,
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

	// 7. Buffer line - Income minus (Mandatory + Flexible)
	buffer := totalIncome - totalExpenses
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

	case tea.KeyMsg:
		// Handle form overlay input if active
		if m.formOverlay.Active {
			keyStr := msg.String()

			// Handle backspace
			if keyStr == "backspace" {
				if m.formOverlay.FieldIndex < len(m.formOverlay.Config.Fields) {
					field := &m.formOverlay.Config.Fields[m.formOverlay.FieldIndex]
					if !field.ReadOnly && len(field.Value) > 0 {
						field.Value = field.Value[:len(field.Value)-1]
					}
				}
				return m, nil
			}

			// Handle enter key - save form data
			if keyStr == "enter" {
				var err error
				switch m.formOverlay.Config.Type {
				case CreateEntry:
					name, month, year, value, catStr, realized, e := ParseEntryForm(m.formOverlay.Config.Fields)
					if e != nil {
						err = e
					} else {
						cat := Data.Category(catStr)
						err = m.budget.NewEntry(name, month, year, value, cat, realized)
					}
				case EditEntry:
					// Use OriginalID for lookup - don't allow ID change via form fields
					id := m.formOverlay.OriginalID
					_, name, month, year, value, catStr, realized, e := ParseEntryFormWithID(m.formOverlay.Config.Fields)
					if e != nil {
						err = e
					} else {
						m.budget.UpdateEntry(id, name, month, year, value, Data.Category(catStr), realized)
					}
				case CreateTarget:
					name, year, value, sidePocket, e := ParseTargetForm(m.formOverlay.Config.Fields)
					if e != nil {
						err = e
					} else {
						err = m.budget.NewTarget(name, year, value, sidePocket)
					}
				case EditTarget:
					id, name, _, value, sidePocket, e := ParseTargetFormWithID(m.formOverlay.Config.Fields)
					if e != nil {
						err = e
					} else {
						m.budget.UpdateTarget(id, name, value, sidePocket)
					}
				}
				if err == nil {
					m.formOverlay.Active = false
					m.formOverlay.FieldIndex = 0
					m.reloadRows() // Refresh display after save
					if m.formOverlay.OnSave != nil {
						m.formOverlay.OnSave()
					}
				} else {
					// Show error in form
					m.formOverlay.Error = err.Error()
				}
				return m, nil
			}

			// Handle escape
			if keyStr == "esc" {
				m.formOverlay.Active = false
				m.formOverlay.FieldIndex = 0
				return m, nil
			}

			// Handle tab
			if keyStr == "tab" {
				if m.formOverlay.FieldIndex < len(m.formOverlay.Config.Fields) {
					m.formOverlay.FieldIndex = len(m.formOverlay.Config.Fields)
				} else {
					m.formOverlay.FieldIndex = 0
				}
				return m, nil
			}

			// Handle navigation keys
			switch keyStr {
			case "up":
				if m.formOverlay.FieldIndex > 0 {
					m.formOverlay.FieldIndex--
				} else if m.formOverlay.FieldIndex == len(m.formOverlay.Config.Fields) {
					m.formOverlay.FieldIndex = len(m.formOverlay.Config.Fields) - 1
				}
				return m, nil

			case "down":
				if m.formOverlay.FieldIndex < len(m.formOverlay.Config.Fields)-1 {
					m.formOverlay.FieldIndex++
				}
				return m, nil

			case "left":
				// For Category and Realized fields, toggle/cycle the value
				if m.formOverlay.FieldIndex < len(m.formOverlay.Config.Fields) {
					field := &m.formOverlay.Config.Fields[m.formOverlay.FieldIndex]
					if !field.ReadOnly {
						if field.Name == "Realized" {
							if field.Value == "true" {
								field.Value = "false"
							} else {
								field.Value = "true"
							}
						} else if field.Name == "Category" {
							// Cycle backwards: Flexible -> Mandatory -> Income -> Flexible
							switch field.Value {
							case "Income":
								field.Value = "Flexible"
							case "Mandatory":
								field.Value = "Income"
							case "Flexible":
								field.Value = "Mandatory"
							}
						}
					}
				}
				return m, nil

			case "right":
				// For Category and Realized fields, toggle/cycle the value
				if m.formOverlay.FieldIndex < len(m.formOverlay.Config.Fields) {
					field := &m.formOverlay.Config.Fields[m.formOverlay.FieldIndex]
					if !field.ReadOnly {
						if field.Name == "Realized" {
							if field.Value == "true" {
								field.Value = "false"
							} else {
								field.Value = "true"
							}
						} else if field.Name == "Category" {
							// Cycle forwards: Income -> Mandatory -> Flexible -> Income
							switch field.Value {
							case "Income":
								field.Value = "Mandatory"
							case "Mandatory":
								field.Value = "Flexible"
							case "Flexible":
								field.Value = "Income"
							}
						}
					}
				}
				return m, nil
			}

			// Handle character input for editable fields (regular keys)
			// Only add characters for non-special fields (Category and Realized use left/right)
			if m.formOverlay.FieldIndex < len(m.formOverlay.Config.Fields) {
				field := &m.formOverlay.Config.Fields[m.formOverlay.FieldIndex]
				if !field.ReadOnly && len(keyStr) == 1 {
					// Skip for Category and Realized fields (use left/right arrows instead)
					if field.Name != "Category" && field.Name != "Realized" {
						field.Value += keyStr
					}
				}
			}
			return m, nil
		}

		// Normal navigation when form is not active
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit

		case "e", "E":
			// Open form for creating new Entry
			cfg := NewEntryFormConfig("", 0, m.month, m.year, "Income", false, true)
			m.formOverlay = FormOverlayState{
				Active:     true,
				Config:     cfg,
				FieldIndex: 0,
			}
			return m, nil

		case "t", "T":
			// Open form for creating new Target
			cfg := NewTargetFormConfig("", 0, 0, m.year, true)
			m.formOverlay = FormOverlayState{
				Active:     true,
				Config:     cfg,
				FieldIndex: 0,
			}
			return m, nil

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
			// Open edit form for selected row
			if m.activePanel == LeftPanel {
				if len(m.leftRows) > 0 && m.leftSelected < len(m.leftRows) {
					row := m.leftRows[m.leftSelected]
					// Skip separators
					if row.Category != "Separator" && row.Name != "─" {
						entryMonth := int(row.MonthYear.Month())
						entryYear := row.MonthYear.Year()
						cfg := NewEntryFormConfig(row.Name, row.Value, entryMonth, entryYear, string(row.Category), row.Realized, false)
						// Store original ID for lookup during update (in case user modifies name/month/year)
						originalID := row.Name + fmt.Sprintf("-%04d-%02d", entryYear, entryMonth)
						m.formOverlay = FormOverlayState{
							Active:     true,
							Config:     cfg,
							FieldIndex: 0,
							OriginalID: originalID,
						}
					}
				}
			} else {
				if len(m.rightRows) > 0 && m.rightSelected < len(m.rightRows) {
					row := m.rightRows[m.rightSelected]
					cfg := NewTargetFormConfig(row.Name, row.Value, row.SidePocket, m.year, false)
					m.formOverlay = FormOverlayState{
						Active:     true,
						Config:     cfg,
						FieldIndex: 0,
					}
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

// View renders the dual-panel colored table using composable views
func (m chromaModel) View() tea.View {
	// Build left panel (monthly budget) using composable view
	leftView := m.budgetView().Render()

	// Build right panel (yearly targets) using composable view
	rightView := m.targetView().Render()

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

	baseView := "\n" + combined + "\n" + helpText + "\n"

	// Render form overlay if active
	if m.formOverlay.Active {
		formView := RenderForm(m.formOverlay, &FormError{Message: m.formOverlay.Error})
		// Center the form over the base view
		baseView = lipgloss.Place(
			100, // max width
			50,  // max height
			lipgloss.Center,
			lipgloss.Center,
			formView,
		)
	}

	return tea.NewView(baseView)
}

// renderHelpText returns the help text at the bottom of the view
func (m chromaModel) renderHelpText() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	helpText := helpStyle.Render(
		"↑/↓ or W/S: Navigate" +
			" | ←/→ or A/D: Change month" +
			" | Tab: Switch panel" +
			" | Enter: Edit" +
			" | E/T: New Entry/Target" +
			" | Q: Quit",
	)

	return "  " + helpText
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
