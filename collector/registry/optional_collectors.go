package registry

import (
	//"strings"
	//"fmt"

	"github.com/vgrusdev/sap_system_exporter/collector/dispatcher"
	"github.com/vgrusdev/sap_system_exporter/collector/enqueue_server"
	"github.com/vgrusdev/sap_system_exporter/collector/alerts"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// RegisterOptionalCollectors register depending on the system where the exporter run the additional collectors
func RegisterOptionalCollectors(webService sapcontrol.WebService) error {

	enqueueServerCollector, err := enqueue_server.NewCollector(webService)
	if err != nil {
		return errors.Wrap(err, "error registering Enqueue Server collector")
	} else {
		prometheus.MustRegister(enqueueServerCollector)
		log.Info("Enqueue Server optional collector registered")
	}

	dispatcherCollector, err := dispatcher.NewCollector(webService)
	if err != nil {
		return errors.Wrap(err, "error registering Dispatcher collector")
	} else {
		prometheus.MustRegister(dispatcherCollector)
		log.Info("Dispatcher optional collector registered")
	}

	alertsCollector, err := alerts.NewCollector(webService)
	if err != nil {
		return errors.Wrap(err, "error registering Alerts collector")
	} else {
		prometheus.MustRegister(alertsCollector)
		log.Info("Alerts optional collector registered")
	}

	return nil
}
