package start_service

import (
	"strconv"
    "fmt"
	"net/url"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vgrusdev/sap_system_exporter/collector"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"

	//"github.com/hooklift/gowsdl/soap"
)

func NewCollector(webService sapcontrol.WebService) (*startServiceCollector, error) {

	c := &startServiceCollector{
		collector.NewDefaultCollector("start_service"),
		webService,
	}

	c.SetDescriptor("processes", "The processes started by the SAP Start Service", 
					[]string{"name", "pid", "status", "instance_name", "instance_number", "SID", "instance_hostname"})
	//c.SetDescriptor("instances", "The SAP instances in the context of the whole SAP system", []string{"features", "start_priority", "instance_name", "instance_number", "SID", "instance_hostname"})
	c.SetDescriptor("instances", "The SAP instances in the context of the whole SAP system", 
					[]string{"features", "start_priority", "instance_name", "instance_number", "SID", "instance_hostname", "dispstatus"})
	
	return c, nil
}

type startServiceCollector struct {
	collector.DefaultCollector
	webService sapcontrol.WebService
}

func (c *startServiceCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debugln("Collecting SAP Start Service metrics")

	errs := collector.RecordConcurrently([]func(ch chan<- prometheus.Metric) error{
		c.recordProcesses,
		c.recordInstances,
	}, ch)

	for _, err := range errs {
		log.Warnf("Start Service Collector scrape failed: %s", err)
	}
}

func (c *startServiceCollector) recordProcesses(ch chan<- prometheus.Metric) error {
	// VG ++    loop on instances
	log.Debugln("SAP Processes collecting")
	instanceList, err := c.webService.GetSystemInstanceList()
	if err != nil {
		return errors.Wrap(err, "SAPControl web service error")
	}

	log.Debugf("Processes: Instances in the list: %d", len(instanceList.Instances) )

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

		processList, err := myWebService.GetProcessList()
		if err != nil {
			log.Warnf("SAPControl web service error: %s", err)
			continue
		}

		for _, process := range processList.Processes {
			state, err := sapcontrol.StateColorToFloat(process.Dispstatus)
			if err != nil {
				log.Warnf("SAPControl web service error, unable to process SAPControl OSProcess data %v: %s", *process, err)
				continue
			}
			ch <- c.MakeGaugeMetric(
				"processes",
				state,
				process.Name,
				strconv.Itoa(int(process.Pid)),
				process.Textstatus,
				currentSapInstance.Name,
				strconv.Itoa(int(currentSapInstance.Number)),
				currentSapInstance.SID,
				currentSapInstance.Hostname)
		}
	}

	return nil
}


func (c *startServiceCollector) recordInstances(ch chan<- prometheus.Metric) error {
	log.Debugln("SAP Instances collecting")
	instanceList, err := c.webService.GetSystemInstanceList()
	if err != nil {
		return errors.Wrap(err, "SAPControl web service error")
	}

	log.Debugf("Instances: Instances in the list: %d", len(instanceList.Instances) )

	client := c.webService.GetMyClient()
	config := client.Config.Viper
	sapControlUrl := config.GetString("sap-control-url")

	u, err := url.ParseRequestURI(sapControlUrl)
	if err != nil {
		return errors.Wrap(err, "could not parse uri: " + sapControlUrl)
	}
	useHTTPS := false
	if u.Scheme == "https" {
		useHTTPS = true
	} 

	myConfig, err := client.Config.Copy()
	if err != nil {
		return errors.Wrap(err, "SAPControl config Copy error")
	}

	for _, instance := range instanceList.Instances {

		if useHTTPS == true {
			url := fmt.Sprintf("https://%s:%d", instance.Hostname, instance.HttpsPort)
		} else {
			url := fmt.Sprintf("http://%s:%d", instance.Hostname, instance.HttpPort)
		}
		log.Debugf(" Instances: use url: %s", url)
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
		
		//log.Debugf("Collecting SAP Start Service metrics. \n url=%s\n currentSapInstance.Name=%s\n instance.InstanceNr=%d\n", url, currentSapInstance.Name, instance.InstanceNr)

		instanceStatus, err := sapcontrol.StateColorToFloat(instance.Dispstatus)
		if err != nil {
			log.Warnf("SAPControl web service error, unable to process SAPControl Instance data %v: %s", *instance, err)
			continue
		}
		ch <- c.MakeGaugeMetric(
			"instances",
			instanceStatus,
			instance.Features,
			instance.StartPriority,
			currentSapInstance.Name,
			strconv.Itoa(int(currentSapInstance.Number)),
			currentSapInstance.SID,
			currentSapInstance.Hostname,
			string(instance.Dispstatus))
	}
	return nil
}
