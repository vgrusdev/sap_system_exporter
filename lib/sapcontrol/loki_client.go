package sapcontrol

import (
	"time"
	//"regexp"
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

func NewLokiClient(myConfig *config.MyConfig) promtail.Client {

	v := myConfig.Viper

	lokiURL := v.GetString("loki_url")
	if lokiURL == "" {
		log.Infoln("loki_url option is empty, will not use LOKI to push Alerts")
		return nil
	}
	bw := v.GetInt("loki_batch_wait")
	if bw <= 0 {
		bw = 100
	}
	bn := v.GetInt("loki_batch_entries_number")
	if bn <= 0 {
		bn = 1
	}
	timeout := v.GetInt("loki_http_timeout")
	if timeout <= 0 {
		timeout = 1000
	}
	loc, err := time.LoadLocation(v.GetString("loki_time_location"))
	if err != nil {
		log.Errorf("Option loki_time_location incorrect: %s. Use UTC", err)
		loc, _ = time.LoadLocation("")
	}

	cfg := promtail.ClientConfig{
		Name:               v.GetString("loki_name"),
		PushURL:            lokiURL,
		TenantID:           v.GetString("loki_tenantid"),
		BatchWait:          time.Duration(bw) * time.Millisecond,
		BatchEntriesNumber: bn,
		Timeout:            time.Duration(timeout) * time.Millisecond,
		Location:           loc,
	}

	c, err := promtail.NewClientProto(&cfg)
	//c, err := promtail.NewClientJson(&cfg)
	if err != nil {
		log.Errorf("Will not use LOKI to push Alerts. Client Create Error: %s\n", err)
		return nil
	}
	return c
}
