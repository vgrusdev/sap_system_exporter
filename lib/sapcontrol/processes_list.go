package sapcontrol

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type ProcessInfo struct { // this will keep all Processes properties
	OSProcess         // Embedded base instance
	Status    float64 `json:"status"`
}

// Returns list of All instances properties, uses memory cache to reduce system calls.
// Calls GetProcessList, url for SAP instance.
func (s *webService) GetCachedProcessList(ctx context.Context, url string) ([]ProcessInfo, error) {

	client := s.Client
	cacheMgr := client.cacheMgr
	log := client.logger

	log.Debug("Starting webService.GetCachedProcessList")
	// Call cache function with callback in case of cache missed
	value := cacheMgr.GetOrSet(fmt.Sprintf("ProcessList_%s", url),
		func() (interface{}, time.Duration) {
			ttl := client.config.Viper.GetDuration("cache_ttl")
			if ttl == 0 {
				ttl = 30 * time.Second
			}
			newProcessList, err := s.GetProcesses(ctx, url)
			if err != nil {
				log.Error("GetProcesses error", err)
				return nil, ttl // return nil, and do not repeat ttl duration.
			}
			return newProcessList, ttl
		})
	if value != nil { // if value is not nil
		if instances, ok := value.([]ProcessInfo); ok { // convert interface{} to []InstanceInfo
			return instances, nil // return instances if everything OK.
		}
	}
	return []ProcessInfo{}, errors.New("GetProcesses error")
}

func (s *webService) GetProcesses(ctx context.Context, url string) ([]ProcessInfo, error) {

	log := s.Client.logger
	log.Debugf("Starting GetProcesses func (cache callback). url = %s", url)

	//v := s.Client.config.Viper

	// Get Instance list from the central Instance
	processList, err := s.GetProcessList(ctx, url)
	if err != nil {
		return []ProcessInfo{}, errors.Wrapf(err, "GetProcesses can not get Process List for url %s", url)
	}

	n := len(processList.Processes)
	log.Debugf("Processes in the list: %d", n)

	processes := make([]ProcessInfo, 0, n)

	for _, process := range processList.Processes {

		singleProcess := &ProcessInfo{
			OSProcess: *process,
		}
		status, err := StateColorToFloat(process.Dispstatus)
		if err != nil {
			log.Warnf("Process status error. Url %s: %s", url, err)
		}
		singleProcess.Status = status

		processes = append(processes, *singleProcess)
	}
	return processes, nil
}
