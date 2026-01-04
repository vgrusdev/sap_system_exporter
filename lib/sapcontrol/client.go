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
	logger   *config.Logger
	cacheMgr *cache.Manager
}

func NewSoapClient(myConfig *config.MyConfig) *MyClient {
	// creates new SOAP client struct
	// returns MyClient scruct, that contains MyConfig struct.
	//
	c := &MyClient{
		config: myConfig,
		logger: config.NewLogger("sapcontrol"),
	}
	return c
}

func (c *MyClient) CreateSoapClient(endpoint string) *soap.Client {
	// creates new SOAP client instance for provided url and opts (user:pwd, tls params).
	// params are exctracted from myConfig.Viper
	// returns MyClient scrict, that contains soap client and MyConfig struct.
	//
	v := c.config.Viper
	log := c.logger

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
	log.Debugf("Creating new soap client with URL: %s", endpoint)
	client := soap.NewClient(endpoint, opts...)

	return client
}
