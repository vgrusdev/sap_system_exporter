package dispatcher

import (
	"strconv"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vgrusdev/sap_system_exporter/collector"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
)

func NewCollector(webService sapcontrol.WebService) (*dispatcherCollector, error) {

	c := &dispatcherCollector{
		collector.NewDefaultCollector("dispatcher"),
		webService,
	}

	c.SetDescriptor("queue_now", "Work process current queue length", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("queue_high", "Work process peak queue length", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("queue_max", "Work process maximum queue length", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("queue_writes", "Work process queue writes", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("queue_reads", "Work process queue reads", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})

	return c, nil
}

type dispatcherCollector struct {
	collector.DefaultCollector
	webService sapcontrol.WebService
}

func (c *dispatcherCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting Dispatcher metrics")

	err := c.recordWorkProcessQueueStats(ch)
	if err != nil {
		log.Warnf("Dispatcher Collector scrape failed: %s", err)
		return
	}
}

func (c *dispatcherCollector) recordWorkProcessQueueStats(ch chan<- prometheus.Metric) error {

	// VG ++    loop on instances
	log.Debugln("SAP WorkProcessQueueStats collecting")
	instanceList, err := c.webService.GetSystemInstanceList()
	if err != nil {
		return errors.Wrap(err, "SAPControl web service error")
	}

	log.Debugf("WorkProcessesQueueStats: Instances in the list: %d\n", len(instanceList.Instances) )

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
		commonLabels := []string{
			currentSapInstance.Name,
			strconv.Itoa(int(currentSapInstance.Number)),
			currentSapInstance.SID,
			currentSapInstance.Hostname,
		}

		queueStatistic, err := myWebService.GetQueueStatistic()
		if err != nil {
			return errors.Wrap(err, "SAPControl web service error")
		}

		// for each work queue, we record a different line for each stat of that queue, with the type as a common label
		for _, queue := range queueStatistic.Queues {
			labels := append([]string{queue.Type}, commonLabels...)
			ch <- c.MakeGaugeMetric("queue_now", float64(queue.Now), labels...)
			ch <- c.MakeCounterMetric("queue_high", float64(queue.High), labels...)
			ch <- c.MakeGaugeMetric("queue_max", float64(queue.Max), labels...)
			ch <- c.MakeCounterMetric("queue_writes", float64(queue.Writes), labels...)
			ch <- c.MakeCounterMetric("queue_reads", float64(queue.Reads), labels...)
		}
	}
	return nil
}
