package ojologger

import (
	"os"
	"path"
)

type OjoLoggerConfig struct {
	Enabled       bool
	UseFile       bool
	UseConsole    bool
	LogFolderPath string
}

var (
	config = createDefaultConfig()
)

func createDefaultConfig() *OjoLoggerConfig {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return &OjoLoggerConfig{
		Enabled:       true,
		UseConsole:    true,
		UseFile:       true,
		LogFolderPath: path.Join(root, "logs"),
	}
}

func Config(config_ *OjoLoggerConfig) {
	config = config_
}
