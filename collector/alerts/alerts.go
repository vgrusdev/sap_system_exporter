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
	//c.SetDescriptor("Alert", "SAP System open Alerts", []string{"instance_name", "instance_number", "SID", "instance_hostname", "Object", "Attribute", "Description", "ATime", "Aluniqnum"})
	//c.SetDescriptor("Alert", "SAP System open Alerts", []string{"instance_name", "instance_number", "SID", "instance_hostname", "Object", "Attribute", "Description", "ATime", "State"})
	c.SetDescriptor("Alert", "SAP System open Alerts", []string{"Object", "Attribute", "Description", "ATime", "State", "instance_name", "instance_number", "SID", "instance_hostname"})
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
	Object      string
	Attribute   string
	Value		STATECOLOR
	Description string
	ATime       string
}

func (c *alertsCollector) recordAlerts(ch chan<- prometheus.Metric) error {

	alertList, err := c.webService.GetAlerts()
	if err != nil {
		return errors.Wrap(err, "SAPControl web service GetAlerts error")
	}

	currentSapInstance, err := c.webService.GetCurrentInstance()
	if err != nil {
		return errors.Wrap(err, "SAPControl web service error")
	}
	commonLabels := []string {
		currentSapInstance.Name,
		strconv.Itoa(int(currentSapInstance.Number)),
		currentSapInstance.SID,
		currentSapInstance.Hostname,
	}

	var alert_item_list []current_alert
	//var alert_item        current_alert

	for _, alert := range alertList.Alerts {

		alert_item := current_alert {
			Object:      alert.Object,
			Attribute:   alert.Attribute,
			Value:       alert.Value,
			Description: alert.Description,
			ATime:       alert.ATime,
		}
		alert_item_list = append(alert_item_list, alert_item)
	}
	
	log.Debugf("Alerts in the list before remove duplicates: %d\n", len(alert_item_list) )
	alert_item_list = sapcontrol.RemoveDuplicate(alert_item_list)
	log.Debugf("Alerts in the list AFTER remove duplicates: %d\n", len(alert_item_list) )

	for _, alert_item := range alert_item_list {

		state, err := sapcontrol.StateColorToFloat(alert_item.Value)
		if err != nil {
			log.Warnf("SAPControl web service error, unable to process SAPControl Alert Value data %v: %s", *alert_item.Value, err)
			continue
		}
		labels := append([]string{
								alert_item.Object, 
								alert_item.Attribute, 
								alert_item.Description, 
								alert_item.ATime, 
								string(alert_item.Value)
						}, 
						commonLabels...)

		ch <- c.MakeGaugeMetric("Alert", state, labels...)
	}

	return nil
}
