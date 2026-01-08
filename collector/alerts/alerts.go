package alerts

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vgrusdev/promtail-client/promtail"
	"github.com/vgrusdev/sap_system_exporter/collector"
	"github.com/vgrusdev/sap_system_exporter/internal/config"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
)

type alertsCollector struct {
	collector.DefaultCollector
	webService sapcontrol.WebService
	logger     *config.Logger
}

func NewCollector(webService sapcontrol.WebService) (*alertsCollector, error) {

	c := &alertsCollector{
		collector.NewDefaultCollector("alerts"),
		webService,
		config.NewLogger("alerts"),
	}
	c.logger.SetLevel(webService.GetMyClient().GetMyConfig().Viper.GetString("log_level"))

	//c.SetDescriptor("Alert", "SAP System open Alerts", []string{"instance_name", "instance_number", "SID", "instance_hostname", "Object", "Attribute", "Description", "ATime", "Tid", "Aid"})
	//c.SetDescriptor("Alert", "SAP System open Alerts", []string{"instance_name", "instance_number", "SID", "instance_hostname", "Object", "Attribute", "Description", "ATime", "Aluniqnum"})
	//c.SetDescriptor("Alert", "SAP System open Alerts", []string{"instance_name", "instance_number", "SID", "instance_hostname", "Object", "Attribute", "Description", "ATime", "State"})
	//c.SetDescriptor("Alert", "SAP System open Alerts", []string{"Object", "Attribute", "Message", "ATime", "Level", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("Alert", "SAP System open Alerts", []string{"Object", "Attribute", "Message", "ATime", "State",
		"instance_name", "instance_number", "SID", "instance_hostname"})

	return c, nil
}

func (c *alertsCollector) Collect(ch chan<- prometheus.Metric) {
	log := c.logger
	log.Debug("Collecting Alerts")

	v := c.webService.GetMyClient().GetMyConfig().Viper
	timeout := v.GetDuration("scrape_timeout")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := c.recordAlerts(ctx, ch)
	if err != nil {
		log.Errorf("Alerts Collector: %s", err)
		return
	}
}

type current_alert struct {
	Object      string
	Attribute   string
	Value       sapcontrol.STATECOLOR
	Description string
	ATime       string
}

func labelSetFromArrays(keys []string, values []string) (map[string]string, error) {
	m := make(map[string]string)
	if len(keys) != len(values) {
		return m, errors.New("labelSetFromArrays: Arrays should be the same length")
	}
	for i, key := range keys {
		m[key] = values[i]
	}
	return m, nil
}

func (c *alertsCollector) recordAlerts(ctx context.Context, ch chan<- prometheus.Metric) error {
	// VG ++    loop on instances
	log := c.logger
	log.Debug("recordAlerts start")

	instanceInfo, err := c.webService.GetCachedInstanceList(ctx)
	if err != nil {
		return errors.Wrap(err, "recordAlerts")
	}
	log.Debugf("recordAlerts: Instances in the list: %d", len(instanceInfo))

	v := c.webService.GetMyClient().GetMyConfig().Viper

	send_to_prom := v.GetBool("send_alerts_to_prom") // send_alerts_to_prom = bool in config file
	//samples_max_age_str := v.GetString("alert_samples_max_age") // Oldest accespted timestamp for alert, golang Duration string, to avoid errors/wrnings in Loki
	var samples_max_age time.Duration
	//if samples_max_age_str == "" {
	//	samples_max_age = 0 * time.Second
	//} else {
	//	samples_max_age, _ = time.ParseDuration(samples_max_age_str)
	//}
	samples_max_age = v.GetDuration("alert_samples_max_age")
	if send_to_prom {
		log.Debug("Will send Alerts to Prom")
	} else {
		log.Debug("Will not send Alerts to Prom")
	}
	const timeFormat = "2006 01 02 15:04:05"
	var labelNames []string
	var timeLocation *time.Location

	loki_client := c.webService.GetLokiClient()

	if loki_client != nil {
		promDesc := c.GetDescriptor("Alert")
		promDescString := promDesc.String()

		log.Debugf("promDescriptorString = %s", promDescString)

		_, after, _ := strings.Cut(promDescString, "variableLabels: {")
		after = strings.TrimSuffix(after, "}}")
		labelNames = strings.Split(after, ",")
		//log.Debugf("alerts.go: Labels: %s\n", labelNames)
		timeLocation = loki_client.GetLocation()
	}

	for _, instance := range instanceInfo {

		url := instance.Endpoint

		commonLabels := []string{
			instance.Name,
			strconv.Itoa(int(instance.InstanceNr)),
			instance.SID,
			instance.Hostname,
		}
		log.Debugf("commonLabels: %v", commonLabels)

		alertList, err := c.webService.GetAlerts(ctx, url)
		if err != nil {
			log.Warnf("GetAlerts: %s", err)
			continue
		}

		alert_item_list := []current_alert{}

		for _, alert := range alertList.Alerts {

			alert_item := current_alert{
				Object:      alert.Object,
				Attribute:   alert.Attribute,
				Value:       alert.Value,
				Description: alert.Description,
				ATime:       alert.ATime,
			}
			alert_item_list = append(alert_item_list, alert_item)
		}
		if send_to_prom {
			// Remove duplicates for Prometheus
			log.Debugf("Alerts in the list BEFORE remove duplicates: %d", len(alert_item_list))
			alert_item_list = sapcontrol.RemoveDuplicate(alert_item_list)
			log.Debugf("Alerts in the list AFTER remove duplicates: %d", len(alert_item_list))
		} else {
			log.Debugf("Alerts in the list: %d", len(alert_item_list))
		}
		num_sent_to_loki := 0
		for _, alert_item := range alert_item_list {

			state, err := sapcontrol.StateColorToFloat(alert_item.Value)
			if err != nil {
				log.Warnf("SrecordAlerts: Alert State conversion %v: %s", alert_item.Value, err)
				continue
			}
			labels := append([]string{alert_item.Object,
				alert_item.Attribute,
				alert_item.Description,
				alert_item.ATime,
				string(alert_item.Value)},
				commonLabels...)

			// Response to Prometheus request (may be to setup IF...  TODO )
			// need to check what will be if I will not send anything to prom.
			// or should I send something small....
			if send_to_prom {
				ch <- c.MakeGaugeMetric("Alert", state, labels...)
			}

			// Push to LOKI ====================================================================
			if loki_client != nil {
				labelSet, err := labelSetFromArrays(labelNames, labels)
				if err != nil {
					log.Warnf("Alert metrics LabelSet mismatch: %s", err)
					continue
				}
				message := string(labelSet["Message"])
				aTime := string(labelSet["ATime"])

				t, err := time.ParseInLocation(timeFormat, aTime, timeLocation)
				if err != nil {
					log.Warnf("Alert ATime parsing: %s", err)
					t = time.Now()
				}
				if (samples_max_age >= 0) && (time.Since(t) > samples_max_age) {
					log.Infof("Alert entry too far behind, ts=%v", t)
					continue
				}
				delete(labelSet, "Message")
				delete(labelSet, "ATime")
				labelSet["level"], _ = sapcontrol.StateColorToLevel(alert_item.Value)

				sInputEntry := promtail.SingleEntry{
					Labels: labelSet,
					Ts:     t,
					Line:   message,
				}
				loki_client.Single() <- &sInputEntry
				num_sent_to_loki += 1
			} // if loki_client != nil
		} // for _, alert_item := range alert_item_list
		log.Debugf("Alerts sent to loki: %d", num_sent_to_loki)
	} // for _, instance := range instanceList.Instances
	log.Debug("recordAlerts success")
	return nil
}
