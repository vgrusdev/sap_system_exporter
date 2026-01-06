package config

import (
	"fmt"

	//"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func loadConfigurationFromFile(v *viper.Viper, log *Logger) error {
	//v := c.Viper
	//log := c.logger

	// try to get the "config" value from the bound "config" CLI flag
	configPath := v.GetString("config")

	if configPath != "" {
		v.SetConfigFile(configPath)
		// we hard-code the config type to yaml, otherwise ReadConfig will not load the values
		// see https://github.com/spf13/viper/issues/316
		//v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		// Look for config in default locations
		v.SetConfigName("sap_system_exporter") // name of config file (without extension)
		//v.SetConfigType("yaml")   				// or json, toml, etc.
		v.AddConfigPath("./")
		v.AddConfigPath("./config/")
		v.AddConfigPath("./conf/")
		v.AddConfigPath("/etc/")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Warn("No config file found, using defaults and environment variables")
			} else {
				return fmt.Errorf("error reading config file: %w", err)
			}
		}
	}
	log.Infof("Using config file: %s", v.ConfigFileUsed())
	return nil
}
