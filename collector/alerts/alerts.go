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
		collector.NewDefaultCollector("alerts"),
		webService,
	}

	//c.SetDescriptor("Alert", "SAP System open Alerts", []string{"instance_name", "instance_number", "SID", "instance_hostname", "Object", "Attribute", "Description", "ATime", "Tid", "Aid"})
	c.SetDescriptor("Alert", "SAP System open Alerts", []string{"instance_name", "instance_number", "SID", "instance_hostname", "Object", "Attribute", "Description", "ATime"})

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

type current_alert  struct {
	//State       float64
	Object      string
	Attribute   string
	Description string
	ATime       string
	//Tid         string
	//Aid         string
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

	var alert_item_list []current_alert
	var alert_item        current_alert

	for _, alert := range alertList.Alerts {
		state, err := sapcontrol.StateColorToFloat(alert.Value)
		if err != nil {
			log.Warnf("SAPControl web service error, unable to process SAPControl Alert Value data %v: %s", *alert, err)
			continue
		}
		alert_item = current_alert {
			//State:       state,
			Object:      alert.Object,
			Attribute:   alert.Attribute,
			Description: alert.Description,
			ATime:       alert.ATime,
			//Tid:         alert.Tid,
			//Aid:         alert.Aid,
		}
		alert_item_list = append(alert_item_list, alert_item)
	}
	
	log.Debugf("Alerts in the list before remove duplicates: %d\n", len(alert_item_list) )


	alert_item_list = sapcontrol.RemoveDuplicate(alert_item_list)

	log.Debugf("Alerts in the list AFTER remove duplicates: %d\n", len(alert_item_list) )

	for _, alert_item := range alert_item_list {

		ch <- c.MakeGaugeMetric(
			"Alert",
			4.0,
			currentSapInstance.Name,
			strconv.Itoa(int(currentSapInstance.Number)),
			currentSapInstance.SID,
			currentSapInstance.Hostname,
			alert_item.Object,
			alert_item.Attribute,
			alert_item.Description,
			alert_item.ATime)
			//alert.Tid,
			//alert.Aid)
	}

	return nil
}
