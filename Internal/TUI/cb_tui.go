package TUI

import (
	"StructData/Internal/Data"
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	Width(38).
	Align(lipgloss.Center).
	MarginBottom(1)

type model struct {
	table table.Model
	month string
	year  int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("You have selected: %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	monthYear := fmt.Sprintf("%s %d", m.month, m.year)
	renderedTitle := titleStyle.Render(monthYear)
	return tea.NewView("\n" + renderedTitle + "\n" + baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n")
}

func RenderTableMonth(b *Data.Budget, month int, year int) {

	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Value", Width: 10},
	}

	// Separate slices for each category
	var incomeRows, mandatoryRows, flexibleRows []table.Row
	var buffer float64

	// First pass: categorize entries and calculate buffer
	for _, entry := range b.Entries {
		if int(entry.MonthYear.Month()) == month && entry.MonthYear.Year() == year {
			line := table.Row{entry.Name, fmt.Sprintf("%.2f", entry.Value)}

			switch entry.Type {
			case Data.Income:
				incomeRows = append(incomeRows, line)
				buffer += entry.Value
			case Data.Mandatory:
				mandatoryRows = append(mandatoryRows, line)
				buffer -= entry.Value
			case Data.Flexible:
				flexibleRows = append(flexibleRows, line)
				buffer -= entry.Value
			}
		}
	}

	// Build rows in desired order with separators
	separator := table.Row{"─", "─"}

	var rows []table.Row

	// 1. Income entries
	rows = append(rows, incomeRows...)

	// 2. Separator after Income
	if len(mandatoryRows) > 0 || len(flexibleRows) > 0 || len(incomeRows) > 0 {
		rows = append(rows, separator)
	}

	// 3. Mandatory entries
	rows = append(rows, mandatoryRows...)

	// 4. Separator after Mandatory
	if len(flexibleRows) > 0 {
		rows = append(rows, separator)
	}

	// 5. Flexible entries
	rows = append(rows, flexibleRows...)

	// 6. Separator before Buffer
	if len(rows) > 0 {
		rows = append(rows, separator)
	}

	// 7. Buffer line
	rows = append(rows, table.Row{"Buffer", fmt.Sprintf("%.2f", buffer)})

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)+2),
		table.WithWidth(36),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		table: t,
		month: time.Month(month).String(),
		year:  year,
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error: ", err, "")
		os.Exit(1)
	}

}
