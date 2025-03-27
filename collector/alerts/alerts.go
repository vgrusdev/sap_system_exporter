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
	Value		sapcontrol.STATECOLOR
	Description string
	ATime       string
}

func (c *alertsCollector) recordAlerts(ch chan<- prometheus.Metric) error {

	// VG ++    loop on instances
	log.Debugln("SAP Alerts collecting")
	instanceList, err := c.webService.GetSystemInstanceList()
	if err != nil {
		return errors.Wrap(err, "SAPControl web service Alerts error")
	}

	log.Debugf("Alerts: Instances in the list: %d\n", len(instanceList.Instances) )

	client := c.webService.GetMyClient()
	myConfig, err := client.Config.Copy()
	if err != nil {
		return errors.Wrap(err, "SAPControl config Copy error")
	}

	for _, instance := range instanceList.Instances {

		url := fmt.Sprintf("http://%s:%d", instance.Hostname, instance.HttpPort)

		err := myConfig.SetURL(url)
		if err != nil {
			log.Warnf("SAPControl URL error (%s): %s", url, err)
			continue
		}
		myClient := sapcontrol.NewSoapClient(myConfig)
		myWebService := sapcontrol.NewWebService(myClient)

		currentSapInstance, err := myWebService.GetCurrentInstance()
		if err != nil {
			log.Warnf("SAPControl web service error: %s", err)
			continue
		}
		commonLabels := []string {
			currentSapInstance.Name,
			strconv.Itoa(int(currentSapInstance.Number)),
			currentSapInstance.SID,
			currentSapInstance.Hostname,
		}

		alertList, err := c.myWebService.GetAlerts()
		if err != nil {
			log.Warnf("SAPControl web service GetAlerts error: %s", err)
			continue
		}

		alert_item_list := []current_alert{}

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
	
		log.Debugf("Alerts in the list BEFORE remove duplicates: %d\n", len(alert_item_list) )
		alert_item_list = sapcontrol.RemoveDuplicate(alert_item_list)
		log.Debugf("Alerts in the list AFTER remove duplicates: %d\n", len(alert_item_list) )

		for _, alert_item := range alert_item_list {

			state, err := sapcontrol.StateColorToFloat(alert_item.Value)
			if err != nil {
				log.Warnf("SAPControl web service error, unable to process SAPControl Alert Value data %v: %s", alert_item.Value, err)
				continue
			}
			labels := append([]string{	alert_item.Object, 
										alert_item.Attribute, 
										alert_item.Description, 
										alert_item.ATime, 
										string(alert_item.Value) },
								commonLabels...)

			ch <- c.MakeGaugeMetric("Alert", state, labels...)
		}
	}
	return nil
}
