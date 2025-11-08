package views

import (
	"fmt"
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
	fmt.Printf("========================== %s ==========================\n\n", k.ProjectName)
	for _, column := range k.Columns {
		for _, task := range column.Tasks {
			fmt.Printf("[%s] %s\n", column.Name, task)
		}
	}
	fmt.Printf("\n========================== %s ==========================\n", k.ProjectName)
}
