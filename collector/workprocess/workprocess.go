package workprocess

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	//"strings"

	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vgrusdev/sap_system_exporter/collector"
	"github.com/vgrusdev/sap_system_exporter/internal/config"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
)

type workprocessCollector struct {
	collector.DefaultCollector
	webService sapcontrol.WebService
	logger     *config.Logger
}

func NewCollector(webService sapcontrol.WebService) (*workprocessCollector, error) {

	c := &workprocessCollector{
		collector.NewDefaultCollector("workprocess"),
		webService,
		config.NewLogger("workprocess"),
	}
	c.logger.SetLevel(webService.GetMyClient().GetMyConfig().Viper.GetString("log_level"))

	c.SetDescriptor("dispatcher_work_processes", "Dispatcher work process counts by type and status",
		[]string{"wp_type", "status", "instance_name", "instance_number", "SID", "instance_hostname"})

	c.SetDescriptor("dispatcher_work_processes_status", "Status of SAP process",
		[]string{"wp_type", "status", "pid", "name", "description", "client", "user", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("dispatcher_work_processes_cpu", "CPU usage percentage of SAP process",
		[]string{"wp_type", "status", "pid", "name", "description", "client", "user", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("dispatcher_work_processes_elapsed", "Elapsed time of SAP process in seconds",
		[]string{"wp_type", "status", "pid", "name", "description", "client", "user", "instance_name", "instance_number", "SID", "instance_hostname"})

	return c, nil
}

func (c *workprocessCollector) Collect(ch chan<- prometheus.Metric) {
	log := c.logger
	log.Debug("Collecting WP Dispatcher metrics")

	v := c.webService.GetMyClient().GetMyConfig().Viper
	timeout := v.GetDuration("scrape_timeout")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := c.recordWorkProcessStats(ctx, ch)
	if err != nil {
		log.Warnf("Dispatcher Collector scrape error: %s", err)
		return
	}
}

func (c *workprocessCollector) recordWorkProcessStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	// VG ++    loop on instances
	log := c.logger
	log.Debug("recordWorkProcessStats collecting")

	instanceInfo, err := c.webService.GetCachedInstanceList(ctx)
	if err != nil {
		return errors.Wrap(err, "recordWorkProcessStats collector error")
	}
	log.Debugf("recordWorkProcessStats: Instances in the list: %d", len(instanceInfo))

	for _, instance := range instanceInfo {

		if !strings.Contains(strings.ToUpper(instance.Features), "ABAP") {
			continue
		}
		url := instance.Endpoint

		wpTable, err := c.webService.ABAPGetWPTable(ctx, url)
		if err != nil {
			log.Errorf("ABAPGetWPTable error %s", err)
			continue
		}

		commonLabels := []string{
			instance.Name,
			strconv.Itoa(int(instance.InstanceNr)),
			instance.SID,
			instance.Hostname,
		}

		// Process work processes
		wpCounts := make(map[string]map[string]int)
		for _, wp := range wpTable.WorkProcess {
			wpType := wp.Type
			status := wp.Status

			if wpCounts[wpType] == nil {
				wpCounts[wpType] = make(map[string]int)
			}
			wpCounts[wpType][status]++

			// Send detailed process metrics
			c.sendWorkProcessMetrics(ch, commonLabels, wp)

		}

		// Send aggregated work process metrics
		for wpType, statusCounts := range wpCounts {
			for status, count := range statusCounts {
				labels := append([]string{wpType, status}, commonLabels...)
				ch <- c.MakeGaugeMetric("dispatcher_work_processes", float64(count), labels...)
			}
		}
	}
	return nil
}

func (c *workprocessCollector) sendWorkProcessMetrics(ch chan<- prometheus.Metric, commonLabels []string, wp *sapcontrol.WorkProcess) {
	//log := c.logger

	// Work process status
	statusValue := 0.0
	statusUp := strings.ToUpper(wp.Status)
	if strings.Contains(statusUp, "RUN") {
		statusValue = 1
	} else if strings.Contains(statusUp, "WAIT") {
		statusValue = 0.5
	} else {
		statusValue = 0
	}

	labels := append([]string{wp.Type, wp.Status, wp.Pid, fmt.Sprintf("WP-%s", wp.No), wp.Reason, wp.Client, wp.User}, commonLabels...)

	ch <- c.MakeGaugeMetric("dispatcher_work_processes_status", float64(statusValue), labels...)
	if cpu, err := strconv.ParseFloat(wp.Cpu, 64); err == nil {
		ch <- c.MakeGaugeMetric("dispatcher_work_processes_cpu", cpu, labels...)
	}
	if elapsed, err := strconv.ParseFloat(wp.Time, 64); err == nil {
		ch <- c.MakeGaugeMetric("dispatcher_work_processes_elapsed", elapsed, labels...)
	}
}
