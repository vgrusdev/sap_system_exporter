package sapcontrol

import (
	"context"
	"net"
	"net/http"

	"github.com/hooklift/gowsdl/soap"
	"github.com/spf13/viper"
	"github.com/vgrusdev/sap_system_exporter/internal/config"
)

type myClient struct {
	soapClient *soap.Client
	Config     *config.myConfig
}

func NewSoapClient(myConfig *config.myConfig) *myClient {

	c := &myClient{}
    config := myConfig.viper

	c.soapClient = soap.NewClient(
		config.GetString("sap-control-url"),
		soap.WithBasicAuth(
			config.GetString("sap-control-user"),
			config.GetString("sap-control-password"),
		),
	)
	c.Config = myConfig
	return c
}

