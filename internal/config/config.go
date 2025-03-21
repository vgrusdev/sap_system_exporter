package config

import (
	"net/url"
	"regexp"

	"log/slog"
	"context"
	"fmt"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type MyConfig struct {
	flagSet *flag.FlagSet
	Viper   *viper.Viper
}

type discardHandler struct{
	Level slog.Level
}

func (n *discardHandler) Enabled(_ context.Context, level slog.Level) bool {
	//return false
	fmt.Printf("logHandler-Enables: n.Level=%d, level=%d, return=%v\n", n.Level, level, level >= n.Level)
	return level >= n.Level
}
func (n *discardHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}
func (n *discardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return n
}
func (n *discardHandler) WithGroup(_ string) slog.Handler {
	return n
}
func (n *discardHandler) SetLevel(level slog.Level) {
	n.Level = level
}

func New(flagSet *flag.FlagSet) (*MyConfig, error) {
    
	myHandler := &discardHandler {
		Level: slog.LevelInfo,
	}
	viperLogger := slog.New(myHandler)

	//config := viper.New()
	config := viper.NewWithOptions(viper.WithLogger(viperLogger))

	c := &MyConfig {
		flagSet: flagSet,
		Viper:   config,
	}

	err := config.BindPFlags(flagSet)
	if err != nil {
		return nil, errors.Wrap(err, "could not bind config to CLI flags")
	}

	fmt.Printf("====> Point 1\n")

	// try to get the "config" value from the bound "config" CLI flag
	path := config.GetString("config")
	if path != "" {
		// try to manually load the configuration from the given path
		err = loadConfigurationFromFile(config, path)
	} else {
		// otherwise try viper's auto-discovery
		err = loadConfigurationAutomatically(config)
	}

	fmt.Printf("====> Point 2\n")

	if err != nil {
		return nil, errors.Wrap(err, "could not load configuration file")
	}

	fmt.Printf("====> Point 3\n")
	setLogLevel(config.GetString("log-level"))
	myHandler.Level = slog.LevelError
	myHandler.SetLevel(slog.LevelError)
	fmt.Printf("Setting new Log level: slog.LevelError=%d, myHandler.Level=%d\n", slog.LevelError, myHandler.Level)

	fmt.Printf("====> Point 4\n")

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

func (c *MyConfig) Copy() (*MyConfig, error) {

	return New(c.flagSet)
}
func (c *MyConfig) SetURL(url string) error {

	config := c.Viper
	config.Set("sap-control-url", url)

	sanitizeSapControlUrl(config)

	err := validateSapControlUrl(config)
	if err != nil {
		return errors.Wrap(err, "invalid config value for sap-control-url")
	}
	return nil
}
