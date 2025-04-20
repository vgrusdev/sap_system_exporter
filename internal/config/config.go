package config

import (
	"net/url"
	"regexp"
	"strings"

	//"log/slog"
	//"context"
	//"fmt"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
)

type MyConfig struct {
	flagSet *flag.FlagSet
	Viper   *viper.Viper
}

func New(flagSet *flag.FlagSet) (*MyConfig, error) {
    
	config := viper.New()

	c := &MyConfig {
		flagSet: flagSet,
		Viper:   config,
	}

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
	return c, nil
}

// returns an error in case the sap-control-url config value cannot be parsed as URL
func validateSapControlUrl(config *viper.Viper) error {
	sapControlUrl := config.GetString("sap-control-url")
	u, err := url.ParseRequestURI(sapControlUrl)
	if err != nil {
		return errors.Wrap(err, "could not parse uri: " + sapControlUrl)
	}

	host, domain, f := strings.Cut(u.Hostname(), ".")
	if f == true {
		// hostname contains domain after dot.
		config.Set("host-domain", domain)
		//return nil
	} else {
		// no domain part in the hostname
		if domain = config.GetString("host-domain"); domain != "" {
			// Add domain to hostname
			u.Host = host + "." + domain + ":" + u.Port()
			sapControlUrl = u.String()
			// double checking url validity after adding domain
			if _, err := url.ParseRequestURI(sapControlUrl); err != nil {
				return errors.Wrap(err, "could not parse uri after adding domain: " + sapControlUrl)
			}
			config.Set("sap-control-url", sapControlUrl)
		} else {
			log.Debugf("No domain part in the hostname, and host-domain parameter is empty: %s", u.Hostname())
		}
	}
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

func (c *MyConfig) Copy() (*MyConfig, error) {
	return New(c.flagSet)
}

func (c *MyConfig) SetURL(url string) error {

	config := c.Viper
	config.Set("sap-control-url", url)

	sanitizeSapControlUrl(config)

	err := validateSapControlUrl(config)
	if err != nil {
		return errors.Wrap(err, "invalid SetURL value for sap-control-url")
	}
	return nil
}

func (c *MyConfig) UseHTTPS() bool {
	config := c.Viper
	sapControlUrl := config.GetString("sap-control-url")
	u, err := url.ParseRequestURI(sapControlUrl)
	if err != nil {
		log.Warnf("could not parse uri (%s): %s", sapControlUrl, err)
		return false
	}
	if u.Scheme == "https" {
		return true
	}
	return false
}
