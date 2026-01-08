package sapcontrol

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type InstanceInfo struct { // this will keep all Instance properties
	SAPInstance         // Embedded base instance
	Name        string  `json:"instance_name"`
	SID         string  `json:"SID"`
	Endpoint    string  `json:"endpoint"`
	Status      float64 `json:"status"`
}

// Returns list of All instances properties, uses memory cache to reduce system calls.
// Calls GetSystemInstanceList, and then GetInstanceProperties for each instance.
func (s *webService) GetCachedInstanceList(ctx context.Context) ([]InstanceInfo, error) {

	client := s.Client
	cacheMgr := client.cacheMgr
	log := client.logger

	log.Debug("GetCachedInstanceList start.")
	// Call cache function with callback in case of cache missed
	value := cacheMgr.GetOrSet("InstanceInfo",
		func() (interface{}, time.Duration) {
			ttl := client.config.Viper.GetDuration("cache_ttl")
			if ttl == 0 {
				ttl = 30 * time.Second
			}
			newInstances, err := s.GetAllInstances(ctx)
			if err != nil {
				log.Errorf("GetCachedInstanceList: %v", err)
				return nil, ttl // return nil, and do not repeat ttl duration.
			}
			return newInstances, ttl
		})
	if value != nil { // if value is not nil
		if instances, ok := value.([]InstanceInfo); ok { // convert interface{} to []InstanceInfo
			log.Debug(("GetCachedInstanceList success."))
			return instances, nil // return instances if everything OK.
		}
	}
	return []InstanceInfo{}, errors.New("GetAllInstances error")
}

func (s *webService) GetAllInstances(ctx context.Context) ([]InstanceInfo, error) {

	log := s.Client.logger
	log.Debug("GetAllInstances (cache callback) start")

	v := s.Client.config.Viper
	useHTTPS := v.GetBool("sap_use_ssl")

	// Get Instance list from the central Instance
	instanceList, err := s.GetSystemInstanceList(ctx)
	if err != nil {
		return []InstanceInfo{}, errors.Wrap(err, "GetAllInstances")
	}

	n := len(instanceList.Instances)
	log.Debugf("Instances in the list: %d", n)

	// Channel to collect Instances Properties results
	type result struct {
		prop *InstanceInfo
		err  error
	}
	results := make(chan result, n) // make channel with size - number of instances
	var wg sync.WaitGroup

	log.Debug("getting instances properties")
	for _, instance := range instanceList.Instances {
		var url string

		hostname := instance.Hostname
		if !strings.Contains(hostname, ".") {
			hostname = fmt.Sprintf("%s.%s", hostname, v.GetString("host_domain"))
		}
		if useHTTPS == true {
			url = fmt.Sprintf("https://%s:%d", hostname, instance.HttpsPort)
		} else {
			url = fmt.Sprintf("http://%s:%d", hostname, instance.HttpPort)
		}
		log.Debugf(" Instance props url: %s", url)
		singleInstance := &InstanceInfo{
			SAPInstance: SAPInstance{
				Hostname:      instance.Hostname,
				InstanceNr:    instance.InstanceNr,
				HttpPort:      instance.HttpPort,
				HttpsPort:     instance.HttpPort,
				StartPriority: instance.StartPriority,
				Features:      instance.Features,
				Dispstatus:    instance.Dispstatus,
			},
			Endpoint: url,
		}
		status, err := StateColorToFloat(instance.Dispstatus)
		if err != nil {
			log.Warnf("Instance Url %s: %s", url, err)
		}
		singleInstance.Status = status

		wg.Add(1)
		go func() {
			defer wg.Done()

			prop, err := s.GetSingleInstance(ctx, singleInstance)
			results <- result{prop, err}

		}()
	}
	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	instances := make([]InstanceInfo, 0, n)

	// Collect results
	for result := range results {
		sid := v.GetString("sap_sid")
		if result.err != nil {
			result.prop.SID = sid
			log.Errorf("GetAllInstances: properties for instance %d: %s", result.prop.InstanceNr, result.err)
		} else {
			if sid == "" {
				v.Set("sap_sid", result.prop.SID)
			}
		}
		instances = append(instances, *result.prop)
	}
	log.Debug("GetAllInstances (cache callback) success")
	return instances, nil
}

func (s *webService) GetSingleInstance(ctx context.Context, singleInstance *InstanceInfo) (*InstanceInfo, error) {

	endpoint := singleInstance.Endpoint
	response, err := s.GetInstanceProperties(ctx, endpoint)
	if err != nil {
		err = errors.Wrapf(err, "GetSingleInstance")
		singleInstance.Name = fmt.Sprintf("%d", singleInstance.InstanceNr) // instance Nr instead of Name
		return singleInstance, err
	}
	for _, prop := range response.Properties {
		switch prop.Property {
		case "SAPSYSTEMNAME":
			singleInstance.SID = prop.Value
		case "INSTANCE_NAME":
			singleInstance.Name = prop.Value
		}
	}
	return singleInstance, nil
}

// outdated functions below, Use GetCachedInstanceList for all cases !!!

// Instance properties from GetCurrentInstance (GetInstanceProperties)
type InstanceProperties struct {
	SID      string
	Number   int32
	Name     string
	Hostname string
}

func (i *InstanceProperties) String() string {
	return fmt.Sprintf("SID: %s, Name: %s, Number: %d, Hostname: %s", i.SID, i.Name, i.Number, i.Hostname)
}

// this structure will be used for static labels, common to all the metrics
//
//	type CurrentSapInstance struct {
//		SID      string
//		Number   int32
//		Name     string
//		Hostname string
//		URL      string
//	}
func (s *webService) GetCurrentInstance(ctx context.Context, endpoint string) (*InstanceProperties, error) {

	response, err := s.GetInstanceProperties(ctx, endpoint)
	if err != nil {
		err = errors.Wrap(err, "GetCurrentInstance")
		return nil, err
	}

	instanceProps := &InstanceProperties{}

	for _, prop := range response.Properties {
		switch prop.Property {
		case "SAPSYSTEM":
			var num int64
			num, err = strconv.ParseInt(prop.Value, 10, 32)
			if err != nil {
				err = errors.Wrapf(err, "GetCurrentInstance: instance number parse: %s", prop.Value)
				return nil, err
			}
			if num < math.MinInt32 || num > math.MaxInt32 {
				err = errors.New("GetCurrentInstance: parsed instance number out of int32 range")
				return nil, err
			}
			instanceProps.Number = int32(num)
		case "SAPSYSTEMNAME":
			instanceProps.SID = prop.Value
		case "INSTANCE_NAME":
			instanceProps.Name = prop.Value
		case "SAPLOCALHOST":
			instanceProps.Hostname = prop.Value
		}
	}
	return instanceProps, nil
}
