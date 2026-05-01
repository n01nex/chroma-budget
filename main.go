package main

import (
	"StructData/Internal/Data"
	"fmt"
	"time"
)

func main() {
	newBudget := Data.Budget{
		ID:      "01",
		Name:    "Federico's Budget",
		Entries: nil,
		Targets: nil,
	}

	entry1 := Data.Entry{
		ID:        "01",
		Name:      "Income",
		MonthYear: time.Date(2026, 01, 1, 0, 0, 0, 0, time.UTC),
		Value:     8750,
		Type:      Data.Income,
		Realized:  false,
	}

	newBudget.Entries = append(newBudget.Entries, entry1)

	fmt.Println(newBudget)
}
