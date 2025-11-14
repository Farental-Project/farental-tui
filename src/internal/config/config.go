package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	VERSION = "beta 0.1.0"
)

func Init() {
	var configPath string
	var err error

	cfgDir, err := os.UserConfigDir()

	if err != nil {
		log.Panic("Failed to get user config dir")
	}

	configPath = filepath.Join(cfgDir, "farental_tui")

	err = os.MkdirAll(configPath, os.ModePerm)

	if err != nil {
		log.Panic("Failed to create the config directory : ", err)
	}

	// Set defaults
	viper.SetDefault("baseurl", "http://127.0.0.1:3000")
	viper.SetDefault("language", "en")
	viper.SetDefault("lastusedemail", "")
	viper.SetDefault("logintoken", "")
	viper.SetDefault("datetimeformat", "02.01.2006 15:04")
	viper.SetDefault("theme", "dark")

	viper.SetConfigName("farental")
	viper.SetConfigType("toml")

	viper.AddConfigPath(filepath.Join(cfgDir, "farental_tui"))

	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			err = viper.SafeWriteConfig()

			if err != nil {
				log.Panic("error while writing the config file : ", err)
			}
		} else {
			log.Panic("error while reading the config file : ", err)
		}
	}
}
