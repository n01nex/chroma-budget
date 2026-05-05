package main

import (
	"StructData/Internal/Data"
	"StructData/Internal/Service"
	"StructData/Internal/TUI"
	"path/filepath"
)

const configFile = "config.json"
const confDir = ".chroma-budget/"

func main() {

	// Initialize app with directory, files or load existing config
	cfg, appDir, initErr := Service.Init(configFile, confDir)
	if initErr != nil {
		panic(initErr)
	}

	//Load Budget if it exists
	var budget Data.Budget
	if cfg.LastBudgetPath != "" {
		budgetTmp, err := Data.LoadFromFile(cfg.LastBudgetPath)
		if err != nil {
			panic(err) // TODO: If error, ask if to create a new budget or load another one
		}
		budget = *budgetTmp
	} else {
		budget = *testData() //TODO: Create a new Budget Process - Include selection of where to save it
	}

	//USE BUDGET SECTION - LIVE APP

	TUI.RenderTableMonth(&budget, 2, 2026)

	// ON CLOSURE: SAVE BUDGET + SAVE CONFIG
	if cfg.LastBudgetPath == "" {
		budgetName := budget.Name
		if budgetName == "" {
			budgetName = "default" // or generate UUID-based name
		}
		cfg.LastBudgetPath = filepath.Join(appDir, budgetName+".json")
	}
	err := budget.SaveToFile(cfg.LastBudgetPath)
	if err != nil {
		panic(err)
	}
	err = cfg.Save(filepath.Join(appDir, configFile))
	if err != nil {
		panic(err)
	}

}

// TEST DATA
func testData() *Data.Budget {
	budget := Data.NewBudget("Test Budget 2026")

	// ============================================
	// INCOME ENTRIES (all realized - actual salary)
	// ============================================
	// January through June 2026 - realistic monthly salary
	for month := 1; month <= 6; month++ {
		budget.NewEntry("Salary", month, 2026, 5000.00, Data.Income, true)
	}

	// ============================================
	// MANDATORY EXPENSES
	// ============================================
	// Rent - $1500/month, paid on 1st of each month (realized)
	for month := 1; month <= 6; month++ {
		budget.NewEntry("Rent", month, 2026, 1500.00, Data.Mandatory, true)
	}

	// Utilities - $120/month average, realized after payment
	budget.NewEntry("Utilities", 1, 2026, 115.50, Data.Mandatory, true)  // Jan - paid
	budget.NewEntry("Utilities", 2, 2026, 98.75, Data.Mandatory, true)   // Feb - paid
	budget.NewEntry("Utilities", 3, 2026, 135.20, Data.Mandatory, true)  // Mar - paid
	budget.NewEntry("Utilities", 4, 2026, 110.00, Data.Mandatory, true)  // Apr - paid
	budget.NewEntry("Utilities", 5, 2026, 105.00, Data.Mandatory, true)  // May - paid
	budget.NewEntry("Utilities", 6, 2026, 125.00, Data.Mandatory, false) // Jun - planned

	// Health Insurance - $300/month, auto-deducted (realized)
	for month := 1; month <= 6; month++ {
		budget.NewEntry("Health Insurance", month, 2026, 300.00, Data.Mandatory, true)
	}

	// ============================================
	// FLEXIBLE EXPENSES
	// ============================================
	// Groceries - ~$400/month budget, variable spending
	budget.NewEntry("Groceries", 1, 2026, 385.50, Data.Flexible, true)  // Jan - actual
	budget.NewEntry("Groceries", 2, 2026, 420.00, Data.Flexible, true)  // Feb - actual
	budget.NewEntry("Groceries", 3, 2026, 395.75, Data.Flexible, true)  // Mar - actual
	budget.NewEntry("Groceries", 4, 2026, 410.00, Data.Flexible, true)  // Apr - actual
	budget.NewEntry("Groceries", 5, 2026, 380.00, Data.Flexible, true)  // May - actual
	budget.NewEntry("Groceries", 6, 2026, 400.00, Data.Flexible, false) // Jun - planned

	// Dining Out - variable, some months over/under budget
	budget.NewEntry("Dining Out", 1, 2026, 150.00, Data.Flexible, true)  // Jan
	budget.NewEntry("Dining Out", 2, 2026, 200.00, Data.Flexible, true)  // Feb (Valentine's)
	budget.NewEntry("Dining Out", 3, 2026, 125.00, Data.Flexible, true)  // Mar
	budget.NewEntry("Dining Out", 4, 2026, 175.00, Data.Flexible, true)  // Apr
	budget.NewEntry("Dining Out", 5, 2026, 190.00, Data.Flexible, true)  // May
	budget.NewEntry("Dining Out", 6, 2026, 150.00, Data.Flexible, false) // Jun - planned

	// Entertainment - streaming, hobbies
	budget.NewEntry("Entertainment", 1, 2026, 85.00, Data.Flexible, true)  // Jan
	budget.NewEntry("Entertainment", 2, 2026, 120.00, Data.Flexible, true) // Feb (games)
	budget.NewEntry("Entertainment", 3, 2026, 65.00, Data.Flexible, true)  // Mar
	budget.NewEntry("Entertainment", 4, 2026, 90.00, Data.Flexible, true)  // Apr
	budget.NewEntry("Entertainment", 5, 2026, 75.00, Data.Flexible, true)  // May
	budget.NewEntry("Entertainment", 6, 2026, 80.00, Data.Flexible, false) // Jun - planned

	// ============================================
	// ONE-TIME / IRREGULAR EXPENSES
	// ============================================
	budget.NewEntry("Car Maintenance", 2, 2026, 350.00, Data.Mandatory, true) // Feb - oil change
	budget.NewEntry("New Laptop", 4, 2026, 1200.00, Data.Flexible, true)      // Apr - major purchase

	// ============================================
	// SAVINGS ENTRIES (monthly contributions)
	// ============================================
	// All savings entries are planned (not yet realized) except Jan which was auto-transferred
	budget.NewEntry("Savings", 1, 2026, 500.00, Data.Flexible, true)  // Jan - deposited
	budget.NewEntry("Savings", 2, 2026, 500.00, Data.Flexible, false) // Feb - planned
	budget.NewEntry("Savings", 3, 2026, 500.00, Data.Flexible, false) // Mar - planned
	budget.NewEntry("Savings", 4, 2026, 500.00, Data.Flexible, false) // Apr - planned
	budget.NewEntry("Savings", 5, 2026, 500.00, Data.Flexible, false) // May - planned
	budget.NewEntry("Savings", 6, 2026, 500.00, Data.Flexible, false) // Jun - planned

	// ============================================
	// TARGETS (Personal objectives for cost anticipation / milestones)
	// ============================================
	// Health Insurance: $300 * 12 = $3,600/year - anticipate annual premium variation
	budget.NewTarget("Health Insurance", 2026, 3600.00, 0)

	// Savings: $500/month = $6,000/year milestone - emergency fund target
	budget.NewTarget("Savings", 2026, 6000.00, 0)

	return budget
}
