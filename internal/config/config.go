package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	//"log/slog"
	//"context"
	//"fmt"

	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type MyConfig struct {
	//flagSet *flag.FlagSet
	Viper  *viper.Viper
	logger *Logger
}

// LoadConfig loads configuration from file and environment variables
func New(flagSet *flag.FlagSet) (*MyConfig, error) {
	// Initialize Viper
	v := viper.New()

	c := &MyConfig{
		//flagSet: flagSet,
		Viper: v,
	}
	logger := NewLogger("config")
	//logger.SetLevel(v.GetString("log_level")). set below...
	c.logger = logger

	// Viper binds pflafs (command-line flags to Viper struct)
	err := v.BindPFlags(flagSet)
	if err != nil {
		//return nil, errors.Wrap(err, "could not bind config to CLI flags")
		return nil, fmt.Errorf("could not bind config to CLI flags: %w", err)
	}

	// Set default values
	setDefaults(v)

	// Enable environment variable support
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Bind environment variables
	bindEnvVars(v)

	// try to get the "config" value from the bound "config" CLI flag
	// and load to Viper.
	err = loadConfigurationFromFile(v, logger)
	if err != nil {
		return nil, err
	}

	logger.SetLevel(v.GetString("log_level"))

	sanitizeSapControlUrl(v)
	err = validateSapControlUrl(v, logger)
	if err != nil {
		return nil, errors.Wrap(err, "invalid config value for sap_control_url")
	}
	logger.Debug("Viper all settings", "config", v.AllSettings())

	return c, nil
}

// returns an error in case the sap_control_url config value cannot be parsed as URL
func validateSapControlUrl(v *viper.Viper, log *Logger) error {
	sapControlUrl := v.GetString("sap_control_url")
	hostDomain := v.GetString("host_domain")

	u, err := url.ParseRequestURI(sapControlUrl)
	if err != nil {
		return errors.Wrap(err, "could not parse url: "+sapControlUrl)
	}

	if u.Port() == "" {
		return fmt.Errorf("port must be provided for sap_control_url: %s", sapControlUrl)
	}
	v.Set("sap_port", u.Port())

	if u.Scheme == "https" {
		v.Set("sap_use_ssl", true)
	} else {
		v.Set("sap_use_ssl", false)
	}

	host, domain, f := strings.Cut(u.Hostname(), ".")
	if f == true { // hostname provided in sap_control_url( after 1st dot).
		if hostDomain != "" {
			log.Warn("host_domain parameter is overwritten by sap_contril_url", "host_domain", domain)
		}
		v.Set("sap_host", u.Hostname())
		v.Set("host_domain", domain)
		//return nil
	} else { // no domain provided in sap_control_url
		if hostDomain != "" {
			// Add domain to hostname
			u.Host = host + "." + hostDomain + ":" + u.Port()
			sapControlUrl = u.String()
			// double checking url validity after adding domain
			if uu, err := url.ParseRequestURI(sapControlUrl); err != nil {
				return errors.Wrap(err, "could not parse sap_control_url after merge with host_domain: "+sapControlUrl)
			} else {
				v.Set("sap_control_url", sapControlUrl)
				v.Set("sap_host", uu.Hostname)
			}
		} else {
			log.Warnf("host_domain parameter is empty and no domain part in the sap_contril_url: %s", sapControlUrl)
			v.Set("sap_host", u.Hostname())
		}
	}
	return nil
}

// automatically adds an http:// prefix in case it's missing from the value, to avoid the downstream consumer
// throw errors due to missing schema URL component
func sanitizeSapControlUrl(v *viper.Viper) {
	sapControlUrl := v.GetString("sap_control_url")
	hasScheme, _ := regexp.MatchString("^https?://", sapControlUrl)
	if !hasScheme {
		sapControlUrl = "http://" + sapControlUrl
		v.Set("sap_control_url", sapControlUrl)
	}
}

/*
func (c *MyConfig) Copy() (*MyConfig, error) {
	return New(c.flagSet)
}

func (c *MyConfig) SetURL(url string) error {

	v := c.Viper
	v.Set("sap_control_url", url)

	sanitizeSapControlUrl(v)

	err := validateSapControlUrl(v)
	if err != nil {
		return errors.Wrap(err, "invalid SetURL value for sap_control_url")
	}
	return nil
}

func (c *MyConfig) UseHTTPS() bool {
	v := c.Viper
	log := c.logger
	sapControlUrl := v.GetString("sap_control_url")
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
*/

func setDefaults(v *viper.Viper) {
	v.SetDefault("address", "0.0.0.0")
	v.SetDefault("port", "9680")
	v.SetDefault("log_level", "info")
	v.SetDefault("sap_control_url", "https://localhost:50014")
	v.SetDefault("sap_control_access_point", "/sap/bc/soap/rfc")
	v.SetDefault("sap_control_domain", "")
	v.SetDefault("tls_skip_verify", false)
	v.SetDefault("sap_control_user", "")
	v.SetDefault("sap_control_password", "")
	v.SetDefault("sap_cache_ttl", "30s")
	v.SetDefault("scrape_timeout", "30s")
	v.SetDefault("send_alerts_to_prom", false)
	v.SetDefault("loki_url", "")
	v.SetDefault("loki_name", "sap_alerts")
	v.SetDefault("loki_tenantid", "fake")
	v.SetDefault("loki_batch_wait", 100)
	v.SetDefault("loki_batch_entries_number", 32)
	v.SetDefault("loki_http_timeout", 1000)
	v.SetDefault("loki_time_location", "Europe/Moscow")
	v.SetDefault("collect_enqueueserver", true)
	v.SetDefault("collect_dispatcher", true)
	v.SetDefault("collect_workprocess", true)
	v.SetDefault("collect_alerts", true)
}

func bindEnvVars(v *viper.Viper) {
	// Bind each field to environment variable
	v.BindEnv("address", "ADDRESS")
	v.BindEnv("port", "PORT")
	v.BindEnv("sap_control_url", "SAP_CONTROL_URL")
	v.BindEnv("sap_control_domain", "SAP_CONTROL_DOMAIN")
	v.BindEnv("log_level", "LOG_LEVEL")
}
