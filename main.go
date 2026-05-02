package main

import (
	"StructData/Internal/Data"
	"fmt"

	"github.com/google/uuid"
)

func main() {
	newBudget := Data.Budget{
		ID:      uuid.New(),
		Name:    "Federico's Budget",
		Entries: nil,
		Targets: nil,
	}

	err := newBudget.NewEntry(
		"Income",
		1,
		2026,
		8750,
		Data.Income,
		false)
	if err != nil {
		fmt.Println(err)
	}

	err = newBudget.NewTarget(
		"Income",
		2026,
		50000.0,
		2000.0)

	fmt.Println(newBudget)

	err = newBudget.NewEntry(
		"Income",
		2,
		2026,
		9000.0,
		Data.Income,
		true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("_____")
	fmt.Println(newBudget)
}
