package config

import (
	"net/url"
	"regexp"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func New(flagSet *flag.FlagSet) (*viper.Viper, error) {
	config := viper.New()

	var savedURL string = nil

	err := config.BindPFlags(flagSet)
	if err != nil {
		return nil, errors.Wrap(err, "could not bind config to CLI flags")
	}

	// try to get the "config" value from the bound "config" CLI flag
	path := config.GetString("config")
	if path != "" {
		// try to manually load the configuration from the given path
		err = loadConfigurationFromFile(config, path)
	} else {
		// otherwise try viper's auto-discovery
		err = loadConfigurationAutomatically(config)
	}

	if err != nil {
		return nil, errors.Wrap(err, "could not load configuration file")
	}

	setLogLevel(config.GetString("log-level"))

	sanitizeSapControlUrl(config)

	err = validateSapControlUrl(config)
	if err != nil {
		return nil, errors.Wrap(err, "invalid config value for sap-control-url")
	}

	return config, nil
}

// returns an error in case the sap-control-url config value cannot be parsed as URL
func validateSapControlUrl(config *viper.Viper) error {
	sapControlUrl := config.GetString("sap-control-url")
	if _, err := url.ParseRequestURI(sapControlUrl); err != nil {
		return errors.Wrap(err, "could not parse uri")
	}
	/* VG
	if u, err := url.ParseRequestURI(sapControlUrl); err != nil {
		return errors.Wrap(err, "could not parse uri")
	}
	port := u.Port()
	if len(port) == 0 {
		return errors.Wrap(err, "could not parse uri, port missing")
	}
	config.Set("sap-control-url-port", port)
	*/
	return nil
}

// automatically adds an http:// prefix in case it's missing from the value, to avoid the downstream consumer
// throw errors due to missing schema URL component
func sanitizeSapControlUrl(config *viper.Viper) {
	sapControlUrl := config.GetString("sap-control-url")
	hasScheme, _ := regexp.MatchString("^https?://", sapControlUrl)
	if !hasScheme {
		sapControlUrl = "http://" + sapControlUrl
		config.Set("sap-control-url", sapControlUrl)
	}
}
func SaveURL(config *viper.Viper) error {
	if config.savedURL != nil {
		return errors.Wrap(err, "URL already saved")
	}
	config.savedURL = config.GetString("sap-control-url")
	return nil
}
func RestoreURL(config *viper.Viper) error {
	if config.savedURL == nil {
		return errors.Wrap(err, "URL was, not saved")
	}
	config.Set("sap-control-url", config.savedURL)
	config.savedURL = nil
	return nil
}
func SetURL(config *viper.Viper, url string) error {

	err := config.SaveURL(config)
	if err != nil {
		return errors.Wrap(err, "always use config.RestoreURL() after config.SetURL()")
	}
	config.Set("sap-control-url", url)

	sanitizeSapControlUrl(config)

	err = validateSapControlUrl(config)
	if err != nil {
		return errors.Wrap(err, "invalid config value for sap-control-url")
	}
	return nil
}