package start_service

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vgrusdev/sap_system_exporter/collector"
	"github.com/vgrusdev/sap_system_exporter/internal/config"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
	//"github.com/hooklift/gowsdl/soap"
)

type startServiceCollector struct {
	collector.DefaultCollector
	webService sapcontrol.WebService
	logger     *config.Logger
}

func NewCollector(webService sapcontrol.WebService) (*startServiceCollector, error) {

	c := &startServiceCollector{
		collector.NewDefaultCollector("start_service"),
		webService,
		config.NewLogger("start_service"),
	}
	c.logger.SetLevel(webService.GetMyClient().GetMyConfig().Viper.GetString("log_level"))

	//c.SetDescriptor("instances", "The SAP instances in the context of the whole SAP system", []string{"features", "start_priority", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("instances", "The SAP instances in the context of the whole SAP system",
		[]string{"features", "start_priority", "instance_name", "instance_number",
			"SID", "instance_hostname", "dispstatus"})

	c.SetDescriptor("processes", "The processes started by the SAP Start Service",
		[]string{"name", "pid", "status", "description", "starttime", "elapsedtime",
			"instance_name", "instance_number", "SID", "instance_hostname", "proc_dispstatus"})

	c.SetDescriptor("processesperinstance_gray", "Processes in state GRAY",
		[]string{"instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("processesperinstance_green", "Processes in state GREEN",
		[]string{"instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("processesperinstance_yellow", "Processes in state YELLOW",
		[]string{"instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("processesperinstance_red", "Processes in state RED",
		[]string{"instance_name", "instance_number", "SID", "instance_hostname"})

	return c, nil
}

func (c *startServiceCollector) Collect(ch chan<- prometheus.Metric) {
	log := c.logger
	log.Debug("Collecting SAP Start Service metrics")

	v := c.webService.GetMyClient().GetMyConfig().Viper
	timeout := v.GetDuration("scrape_timeout")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errs := collector.RecordConcurrently(ctx, []func(ctx context.Context, ch chan<- prometheus.Metric) error{
		c.recordInstances,
		c.recordProcesses,
		//c.recordProcessesPerInstance,
	}, ch)

	for _, err := range errs {
		log.Warnf("Start Service Collector scrape errors: %s", err)
	}
}

func (c *startServiceCollector) recordInstances(ctx context.Context, ch chan<- prometheus.Metric) error {
	log := c.logger
	log.Debug("  recordInstances collecting")

	instanceInfo, err := c.webService.GetCachedInstanceList(ctx)
	if err != nil {
		return errors.Wrap(err, "recordInstances collector error")
	}
	log.Debugf("recordInstances: Instances in the list: %d", len(instanceInfo))

	for _, instance := range instanceInfo {
		ch <- c.MakeGaugeMetric(
			"instances",
			instance.Status,
			instance.Features,
			instance.StartPriority,
			instance.Name,
			strconv.Itoa(int(instance.InstanceNr)),
			instance.SID,
			instance.Hostname,
			string(instance.Dispstatus))
	}
	return nil
}

func (c *startServiceCollector) recordProcesses(ctx context.Context, ch chan<- prometheus.Metric) error {
	// VG ++    loop on instances
	log := c.logger
	log.Debug("recordProcesses collecting")

	instanceInfo, err := c.webService.GetCachedInstanceList(ctx)
	if err != nil {
		return errors.Wrap(err, "recordProcesses collector error")
	}
	log.Debugf("recordProcesses: Instances in the list: %d", len(instanceInfo))

	processes := make(map[sapcontrol.STATECOLOR]int)
	processes[sapcontrol.STATECOLOR_GRAY] = 0
	processes[sapcontrol.STATECOLOR_GREEN] = 0
	processes[sapcontrol.STATECOLOR_YELLOW] = 0
	processes[sapcontrol.STATECOLOR_RED] = 0

	for _, instance := range instanceInfo {

		url := instance.Endpoint
		commonLabels := []string{
			instance.Name,
			strconv.Itoa(int(instance.InstanceNr)),
			instance.SID,
			instance.Hostname,
		}
		processList, err := c.webService.GetProcessList(ctx, url)
		if err != nil {
			log.Warnf("GetProcessList error: %s", err)
			continue
		}
		for _, process := range processList.Processes {

			if _, ok := processes[process.Dispstatus]; ok {
				processes[process.Dispstatus] += 1
			}
			state, err := sapcontrol.StateColorToFloat(process.Dispstatus)
			if err != nil {
				log.Warnf("Process status value error: %s", err)
				//continue
			}
			ch <- c.MakeGaugeMetric(
				"processes",
				state,
				process.Name,
				strconv.Itoa(int(process.Pid)),
				process.Textstatus,
				process.Description,
				process.Starttime,
				process.Elapsedtime,
				instance.Name,
				strconv.Itoa(int(instance.InstanceNr)),
				instance.SID,
				instance.Hostname,
				string(process.Dispstatus))
		}
		ch <- c.MakeGaugeMetric("processesperinstance_gray", float64(processes[sapcontrol.STATECOLOR_GRAY]), commonLabels...)
		ch <- c.MakeGaugeMetric("processesperinstance_green", float64(processes[sapcontrol.STATECOLOR_GREEN]), commonLabels...)
		ch <- c.MakeGaugeMetric("processesperinstance_yellow", float64(processes[sapcontrol.STATECOLOR_YELLOW]), commonLabels...)
		ch <- c.MakeGaugeMetric("processesperinstance_red", float64(processes[sapcontrol.STATECOLOR_RED]), commonLabels...)
	}
	return nil
}

func (c *startServiceCollector) recordProcessesPerInstance(ctx context.Context, ch chan<- prometheus.Metric) error {
	// VG ++    loop on instances
	log := c.logger
	log.Debug("recordProcessesPerInstance collecting")

	instanceInfo, err := c.webService.GetCachedInstanceList(ctx)
	if err != nil {
		return errors.Wrap(err, "recordProcessesPerInstance collector error")
	}
	log.Debugf("recordProcessesPerInstance: Instances in the list: %d", len(instanceInfo))

	processes := make(map[sapcontrol.STATECOLOR]int)
	processes[sapcontrol.STATECOLOR_GRAY] = 0
	processes[sapcontrol.STATECOLOR_GREEN] = 0
	processes[sapcontrol.STATECOLOR_YELLOW] = 0
	processes[sapcontrol.STATECOLOR_RED] = 0

	for _, instance := range instanceInfo {
		url := instance.Endpoint
		commonLabels := []string{
			instance.Name,
			strconv.Itoa(int(instance.InstanceNr)),
			instance.SID,
			instance.Hostname,
		}
		processList, err := c.webService.GetProcessList(ctx, url)
		if err != nil {
			log.Warnf("GetProcessList error: %s", err)
			continue
		}
		for _, process := range processList.Processes {
			if _, ok := processes[process.Dispstatus]; ok {
				processes[process.Dispstatus] += 1
			}
		}
		ch <- c.MakeGaugeMetric("processesperinstance_gray", float64(processes[sapcontrol.STATECOLOR_GRAY]), commonLabels...)
		ch <- c.MakeGaugeMetric("processesperinstance_green", float64(processes[sapcontrol.STATECOLOR_GREEN]), commonLabels...)
		ch <- c.MakeGaugeMetric("processesperinstance_yellow", float64(processes[sapcontrol.STATECOLOR_YELLOW]), commonLabels...)
		ch <- c.MakeGaugeMetric("processesperinstance_red", float64(processes[sapcontrol.STATECOLOR_RED]), commonLabels...)
	}

	return nil
}
