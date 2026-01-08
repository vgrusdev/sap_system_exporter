package sapcontrol

import (
	//"context"
	//"net"
	//"net/http"
	"crypto/tls"
	"strings"

	"github.com/hooklift/gowsdl/soap"
	//"github.com/spf13/viper"
	"github.com/vgrusdev/sap_system_exporter/cache"
	"github.com/vgrusdev/sap_system_exporter/internal/config"
)

type MyClient struct {
	//SoapClient *soap.Client
	config   *config.MyConfig
	cacheMgr *cache.CacheManager
	logger   *config.Logger
}

func NewSoapClient(myConfig *config.MyConfig, cacheMgr *cache.CacheManager) *MyClient {
	// creates new SOAP client struct
	// returns MyClient scruct, that contains MyConfig struct.
	//
	v := myConfig.Viper
	c := &MyClient{
		config:   myConfig,
		cacheMgr: cacheMgr,
		logger:   config.NewLogger("sapcontrol"),
	}
	c.logger.SetLevel(v.GetString("log_level"))
	return c
}

func (c *MyClient) CreateSoapClient(endpoint string) *soap.Client {
	// creates new SOAP client instance for provided url and opts (user:pwd, tls params).
	// params are exctracted from myConfig.Viper
	// returns MyClient scrict, that contains soap client and MyConfig struct.
	//
	v := c.config.Viper
	log := c.logger
	//log.SetLevel(v.GetString("log_level"))

	opts := []soap.Option{
		soap.WithBasicAuth(
			v.GetString("sap_control_user"),
			v.GetString("sap_control_password"),
		),
	}
	if v.GetBool("sap_use_ssl") {
		tlsSkipVfy := false
		if strings.ToUpper(v.GetString("tls_skip_verify")) == "YES" {
			tlsSkipVfy = true
		}
		tlsOpts := &tls.Config{
			InsecureSkipVerify: tlsSkipVfy,
		}
		opts = append(opts, soap.WithTLS(tlsOpts))
	}
	timeout := v.GetDuration("scrape_timeout")
	if timeout != 0 {
		log.Debugf("Soap client with Request timeout %v", timeout)
		opts = append(opts, soap.WithRequestTimeout(timeout))
		opts = append(opts, soap.WithTimeout(timeout))
	}

	log.Debugf("Creating new soap client with URL: %s", endpoint)
	client := soap.NewClient(endpoint, opts...)

	return client
}
func (c *MyClient) GetMyConfig() *config.MyConfig {
	return c.config
}
