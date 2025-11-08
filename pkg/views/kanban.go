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
	panels := []pterm.Panel{}
	for _, column := range k.Columns {
		panels = append(panels, pterm.Panel{
			Data: pterm.Sprintf("%s\n\n%s", pterm.Bold.Sprint(column.Name), renderCards(column.Tasks)),
		})
	}

	if err := pterm.DefaultPanel.
		WithPanels(pterm.Panels{panels}).
		WithPadding(1).
		Render(); err != nil {
		fmt.Println("render error: ", err)
	}
}

func renderCards(items []string) string {
	if len(items) == 0 {
		return "(empty)"
	}

	// Determine maximum width among items
	maximum := 0
	for _, it := range items {
		// account for possible multi-byte runes by using len([]rune(...))
		l := len([]rune(it))
		if l > maximum {
			maximum = l
		}
	}
	padding := 2 // one space each side
	width := maximum + padding*2

	var b strings.Builder
	for i, it := range items {
		line := padCenter(it, width-padding*2)
		top := "┌" + strings.Repeat("─", width) + "┐\n"
		mid := "│" + strings.Repeat(" ", padding) + line + strings.Repeat(" ", padding) + "│\n"
		b.WriteString(top)
		b.WriteString(mid)
		b.WriteString("└" + strings.Repeat("─", width) + "┘")
		if i < len(items)-1 {
			b.WriteString("\n\n") // gap between cards
		}
	}
	return b.String()
}

func padCenter(s string, w int) string {
	rs := []rune(s)
	l := len(rs)
	if l >= w {
		return string(rs[:w])
	}
	space := w - l
	left := space / 2
	right := space - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
