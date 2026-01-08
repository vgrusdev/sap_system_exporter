package registry

import (
	//"strings"
	//"fmt"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vgrusdev/sap_system_exporter/collector/alerts"
	"github.com/vgrusdev/sap_system_exporter/collector/dispatcher"
	"github.com/vgrusdev/sap_system_exporter/collector/enqueue_server"
	"github.com/vgrusdev/sap_system_exporter/collector/workprocess"
	"github.com/vgrusdev/sap_system_exporter/internal/config"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
	//log "github.com/sirupsen/logrus"
)

// RegisterOptionalCollectors register depending on the system where the exporter run the additional collectors
func RegisterOptionalCollectors(webService sapcontrol.WebService) error {

	log := config.NewLogger("registry")
	log.SetLevel(webService.GetMyClient().GetMyConfig().Viper.GetString("log_level"))
	v := webService.GetMyClient().GetMyConfig().Viper

	if v.GetBool("collect_enqueueserver") {
		enqueueServerCollector, err := enqueue_server.NewCollector(webService)
		if err != nil {
			return errors.Wrap(err, "RegisterOptionalCollectors: Enqueue Server")
		} else {
			prometheus.MustRegister(enqueueServerCollector)
			log.Info("Enqueue Server optional collector registered")
		}
	} else {
		log.Debug("Enqueue Server optional collector is not registered")
	}
	if v.GetBool("collect_dispatcher") {
		dispatcherCollector, err := dispatcher.NewCollector(webService)
		if err != nil {
			return errors.Wrap(err, "RegisterOptionalCollectors: Dispatcher")
		} else {
			prometheus.MustRegister(dispatcherCollector)
			log.Info("Dispatcher optional collector registered")
		}
	} else {
		log.Debug("Dispatcher optional collector is not registered")
	}
	if v.GetBool("collect_workprocess") {
		workprocessCollector, err := workprocess.NewCollector(webService)
		if err != nil {
			return errors.Wrap(err, "RegisterOptionalCollectors: WorkProcess")
		} else {
			prometheus.MustRegister(workprocessCollector)
			log.Info("WorkProcess optional collector registered")
		}
	} else {
		log.Debug("WorkProcess optional collector is not registered")
	}
	if v.GetBool("collect_alerts") {
		alertsCollector, err := alerts.NewCollector(webService)
		if err != nil {
			return errors.Wrap(err, "RegisterOptionalCollectors: Alerts")
		} else {
			prometheus.MustRegister(alertsCollector)
			log.Info("Alerts optional collector registered")
		}
	} else {
		log.Debug("Alerts optional collector is not registered")
	}
	return nil
}
