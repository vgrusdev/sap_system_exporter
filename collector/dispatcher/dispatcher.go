package dispatcher

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vgrusdev/sap_system_exporter/collector"
	"github.com/vgrusdev/sap_system_exporter/internal/config"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
)

type dispatcherCollector struct {
	collector.DefaultCollector
	webService sapcontrol.WebService
	logger     *config.Logger
}

func NewCollector(webService sapcontrol.WebService) (*dispatcherCollector, error) {

	c := &dispatcherCollector{
		collector.NewDefaultCollector("dispatcher"),
		webService,
		config.NewLogger("dispatcher"),
	}

	c.SetDescriptor("queue_now", "Work process current queue length", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("queue_high", "Work process peak queue length", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("queue_max", "Work process maximum queue length", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("queue_writes", "Work process queue writes", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("queue_reads", "Work process queue reads", []string{"type", "instance_name", "instance_number", "SID", "instance_hostname"})

	return c, nil
}

func (c *dispatcherCollector) Collect(ch chan<- prometheus.Metric) {
	log := c.logger
	log.Debug("Collecting Dispatcher metrics")

	v := c.webService.GetMyClient().GetMyConfig().Viper
	timeout := v.GetDuration("scrape_timeout")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := c.recordWorkProcessQueueStats(ctx, ch)
	if err != nil {
		log.Warnf("Dispatcher Collector scrape error: %s", err)
		return
	}
}

func (c *dispatcherCollector) recordWorkProcessQueueStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	// VG ++    loop on instances
	log := c.logger
	log.Debug("recordWorkProcessQueueStats collecting")

	instanceInfo, err := c.webService.GetCachedInstanceList(ctx)
	if err != nil {
		return errors.Wrap(err, "recordWorkProcessQueueStats collector error")
	}
	log.Debugf("recordWorkProcessQueueStats: Instances in the list: %d", len(instanceInfo))

	for _, instance := range instanceInfo {

		url := instance.Endpoint

		dispatcherFound := false
		processList, err := c.webService.GetProcessList(ctx, url)
		if err != nil {
			return errors.Wrap(err, "GetProcessList error")
		}

		for _, process := range processList.Processes {
			if strings.Contains(process.Name, "disp+work") {
				dispatcherFound = true
				break
			}
		}
		// if we found msg_server on process name we Collect the Dispatcher Stats
		if dispatcherFound != true {
			continue
		}

		commonLabels := []string{
			instance.Name,
			strconv.Itoa(int(instance.InstanceNr)),
			instance.SID,
			instance.Hostname,
		}

		queueStatistic, err := c.webService.GetQueueStatistic(ctx, url)
		if err != nil {
			return errors.Wrap(err, "GetQueueStatistic error")
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
