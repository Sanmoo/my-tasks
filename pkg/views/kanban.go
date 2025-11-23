package views

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

type Kanban struct {
	ProjectName string
	Columns     []Column
}

type Column struct {
	Name  string
	Tasks []string
}

func (k *Kanban) Render() {
	// Calculate terminal width for layout
	width, _, _ := pterm.GetTerminalSize()
	if width == 0 {
		width = 80 // Default width if terminal size can't be determined
	}

	// Create header
	header := pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue))
	header.Println(k.ProjectName)

	// If we have too many columns for the terminal width, show them vertically
	if len(k.Columns) > 3 || width < 100 {
		k.renderVertical()
	} else {
		k.renderHorizontal(width)
	}
}

func (k *Kanban) renderVertical() {
	for _, column := range k.Columns {
		k.renderColumn(column)
		fmt.Println()
	}
}

func (k *Kanban) renderHorizontal(terminalWidth int) {
	// Calculate column width
	colWidth := (terminalWidth - (len(k.Columns) * 3)) / len(k.Columns)
	if colWidth < 20 {
		colWidth = 20
	}

	// Create table data
	tableData := pterm.TableData{}

	// Add header row
	headerRow := []string{}
	for _, column := range k.Columns {
		headerRow = append(headerRow, k.formatColumnHeader(column.Name))
	}
	tableData = append(tableData, headerRow)

	// Find maximum number of tasks in any column
	maxTasks := 0
	for _, column := range k.Columns {
		if len(column.Tasks) > maxTasks {
			maxTasks = len(column.Tasks)
		}
	}

	// Add task rows
	for i := 0; i < maxTasks; i++ {
		row := []string{}
		for _, column := range k.Columns {
			if i < len(column.Tasks) {
				row = append(row, k.formatTask(column.Tasks[i]))
			} else {
				row = append(row, "")
			}
		}
		tableData = append(tableData, row)
	}

	// Render table
	pterm.DefaultTable.WithBoxed(true).WithData(tableData).Render()
}

func (k *Kanban) renderColumn(column Column) {
	// Determine column color based on content
	var boxStyle *pterm.Style

	switch {
	case strings.Contains(column.Name, "🏃"):
		boxStyle = pterm.NewStyle(pterm.FgLightYellow)
	case strings.Contains(column.Name, "✅"):
		boxStyle = pterm.NewStyle(pterm.FgLightGreen)
	case strings.Contains(column.Name, "🗓️"):
		boxStyle = pterm.NewStyle(pterm.FgLightMagenta)
	default:
		boxStyle = pterm.NewStyle(pterm.FgLightBlue)
	}

	// Create box for column
	box := pterm.DefaultBox.
		WithBoxStyle(boxStyle).
		WithTitle(k.formatColumnHeader(column.Name)).
		WithTitleTopCenter()

	// Format tasks
	content := ""
	for _, task := range column.Tasks {
		content += k.formatTask(task) + "\n"
	}

	if content == "" {
		content = "No tasks"
	}

	box.Println(content)
}

func (k *Kanban) formatColumnHeader(name string) string {
	// Add color to column headers based on content
	switch {
	case strings.Contains(name, "🏃"):
		return pterm.LightYellow(name)
	case strings.Contains(name, "✅"):
		return pterm.LightGreen(name)
	case strings.Contains(name, "🗓️"):
		return pterm.LightMagenta(name)
	default:
		return pterm.LightBlue(name)
	}
}

func (k *Kanban) formatTask(task string) string {
	// Simple task formatting - could be enhanced with tag highlighting, etc.
	return "• " + task
}
