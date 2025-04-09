package sapcontrol

import (
	"time"
	//"context"
	//"net"
	//"net/http"
	//"crypto/tls"
	//"strings"

	//"github.com/hooklift/gowsdl/soap"
	//"github.com/spf13/viper"
	"github.com/vgrusdev/sap_system_exporter/internal/config"

	"github.com/vgrusdev/promtail-client/promtail"

	log "github.com/sirupsen/logrus"
)

func NewLokiClient( myConfig *config.MyConfig ) (promtail.Client) {

	config := myConfig.Viper

	url := config.GetString("loki-url")
	if url == "" {
		log.Infoln("loki-url option is empty, will not use LOKI to push Alerts")
		return nil
	}
	bw := config.GetInt("loki-batch-wait")
	if bw <= 0 { bw = 100 }
	bn := config.GetInt("loki-batch-entries-number")
	if bn <= 0 { bn = 1 }
	timeout := config.GetInt("loki-http-timeout")
	if timeout <= 0 { timeout = 1000 }
	loc, err := time.LoadLocation(config.GetString("loki-time-location"))
	if err != nil {
		log.Errorf("Option loki-time-location incorrect: %s. Use UTC", err)
		loc, _ = time.LoadLocation("")
	}

	cfg := promtail.ClientConfig {
		Name:                   config.GetString("loki-name"),
		PushURL:                url,
		BatchWait:              time.Duration(bw) * time.Millisecond,
		BatchEntriesNumber:     bn,
		Timeout:                time.Duration(timeout) * time.Millisecond,
		Location:               loc,
	}

	//c, err := promtail.NewClientProto(&cfg)
	c, err := promtail.NewClientJson(&cfg)
	if err != nil {
		log.Errorf("Client Create Error: %s", err)
		return nil
	}
	return c
}


