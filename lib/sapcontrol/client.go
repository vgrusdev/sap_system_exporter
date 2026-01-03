package sapcontrol

import (
	//"context"
	//"net"
	//"net/http"
	"crypto/tls"
	"strings"

	"github.com/hooklift/gowsdl/soap"
	//"github.com/spf13/viper"
	"github.com/vgrusdev/sap_system_exporter/internal/config"

	log "github.com/sirupsen/logrus"
)

type MyClient struct {
	SoapClient *soap.Client
	Config     *config.MyConfig
}

func NewSoapClient(myConfig *config.MyConfig) *MyClient {
	// creates new SOAP client instance for provided url and opts (user:pwd, tls params).
	// params are exctracted from myConfig.Viper
	// returns MyClient scrict, that contains soap client and MyConfig struct.
	//
	c := &MyClient{}
	config := myConfig.Viper

	opts := []soap.Option{
		soap.WithBasicAuth(
			config.GetString("sap-control-user"),
			config.GetString("sap-control-password"),
		),
	}
	tlsIgnore := config.GetString("tls-skip-verify")
	if strings.ToUpper(tlsIgnore) == "YES" {
		opts = append(opts, soap.WithTLS(&tls.Config{InsecureSkipVerify: true}))
	}
	/*
		c.SoapClient = soap.NewClient(
			config.GetString("sap-control-url"),
			soap.WithBasicAuth(
				config.GetString("sap-control-user"),
				config.GetString("sap-control-password"),
			),
			soap.WithTLS(&tls.Config{InsecureSkipVerify: true}),
		)
	*/
	log.Debugf("Creating new soap client with URL: %s", config.GetString("sap-control-url"))
	c.SoapClient = soap.NewClient(config.GetString("sap-control-url"), opts...)
	c.Config = myConfig
	return c
}
