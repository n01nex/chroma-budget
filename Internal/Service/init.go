package Service

import (
	"StructData/Internal/Data"
	"fmt"
	"os"
	"path/filepath"
)

func Init(configFile string, confDir string) (cfg *Data.Config, appDir string, err error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, "", err
	}
	// Full path to app directory
	appDir = filepath.Join(homeDir, confDir)

	// First time setup
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		fmt.Println("Creating app directory, none was found")
		err = os.Mkdir(appDir, 0755)
		if err != nil {
			return nil, "", err
		}
	}

	//Load config or create default
	cfg, err = Data.LoadConfig(filepath.Join(appDir, configFile))
	if err != nil {
		fmt.Println("Config file not found, creating new one")
		cfg = &Data.Config{
			LastBudgetPath: "",
		}
		err = nil

	}
	return cfg, appDir, err
}
