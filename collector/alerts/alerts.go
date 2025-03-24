package alerts

import (
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vgrusdev/sap_system_exporter/collector"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
)

func NewCollector(webService sapcontrol.WebService) (*alertsCollector, error) {

	c := &alertsCollector{
		collector.NewDefaultCollector("sap_alerts"),
		webService,
	}

	c.SetDescriptor("alerts", "SAP System open Alerts", []string{"instance_name", "instance_number", "SID", "instance_hostname", "Object", "Attribute", "Description", "Time", "Tid", "Aid"})

	return c, nil
}

type alertsCollector struct {
	collector.DefaultCollector
	webService sapcontrol.WebService
}

func (c *alertsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting Alerts")

	err := c.recordAlerts(ch)
	if err != nil {
		log.Warnf("Alerts Collector scrape failed: %s", err)
		return
	}
}

func (c *alertsCollector) recordAlerts(ch chan<- prometheus.Metric) error {
	alertList, err := c.webService.GetAlerts()
	if err != nil {
		return errors.Wrap(err, "SAPControl web service error")
	}

	currentSapInstance, err := c.webService.GetCurrentInstance()
	if err != nil {
		return errors.Wrap(err, "SAPControl web service error")
	}


	for _, alert := range alertList.Alerts {
		state, err := sapcontrol.StateColorToFloat(alert.Value)
		if err != nil {
			log.Warnf("SAPControl web service error, unable to process SAPControl Alert Value data %v: %s", *alert, err)
			continue
		}
		ch <- c.MakeGaugeMetric(
			"alerts",
			state,
			currentSapInstance.Name,
			strconv.Itoa(int(currentSapInstance.Number)),
			currentSapInstance.SID,
			currentSapInstance.Hostname,
			alert.Object,
			alert.Attribute,
			alert.Description,
			alert.Time,
			alert.Tid,
			alert.Aid
		)
	}

	return nil
}
