package main

import (
	"StructData/Internal/Data"
	"StructData/Internal/Service"
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
			// TODO: INIT NEW BUDGET PROCESS TO BE IMPLEMENTED
			budget = Data.Budget{}
		} else {
			budget = *budgetTmp
		}
	}

	//USE BUDGET SECTION - LIVE APP

	// ON CLOSURE: SAVE BUDGET + SAVE CONFIG
	if cfg.LastBudgetPath == "" {
		budgetName := budget.Name
		if budgetName == "" {
			budgetName = "default" // or generate UUID-based name
		}
		cfg.LastBudgetPath = filepath.Join(appDir, budgetName+".json")
	}
	err = budget.SaveToFile(cfg.LastBudgetPath)
	if err != nil {
		panic(err)
	}
	err = cfg.Save(filepath.Join(appDir, configFile))
	if err != nil {
		panic(err)
	}

}
