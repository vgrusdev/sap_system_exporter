package sapcontrol

import (
	"context"
	"encoding/xml"
	"fmt"

	"strings"

	//"github.com/hooklift/gowsdl/soap"
	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"
	"github.com/vgrusdev/promtail-client/promtail"
)

//go:generate mockgen -destination ../../test/mock_sapcontrol/webservice.go github.com/vgrusdev/sap_system_exporter/lib/sapcontrol WebService

// the main interface exposed by this package
type WebService interface {
	/* Returns a list of all processes directly started by the webservice according to the SAP start profile. */
	GetProcessList(context.Context, string) (*GetProcessListResponse, error)

	/* Returns enque statistic. */
	EnqGetStatistic(context.Context, string) (*EnqGetStatisticResponse, error)

	/* Returns a list of queue information of work processes and icm (similar to dpmon). */
	GetQueueStatistic(context.Context, string) (*GetQueueStatisticResponse, error)

	/* Returns a list of SAP instances of the SAP system. */
	GetSystemInstanceList(context.Context) (*GetSystemInstanceListResponse, error)

	/* Returns a list of available instance features and information how to get it. */
	GetInstanceProperties(context.Context, string) (*GetInstancePropertiesResponse, error)

	/* Custom method to get the current instance data. This is not something natively exposed by the webservice. */
	GetCurrentInstance(context.Context, string) (*InstanceProperties, error)
	GetCachedInstanceList(context.Context) ([]InstanceInfo, error)
	GetCachedProcessList(context.Context, string) ([]ProcessInfo, error)

	GetAlerts(context.Context, string) (*GetAlertsResponse, error)

	GetMyClient() *MyClient
	SetLokiClient(promtail.Client)
	GetLokiClient() promtail.Client
}

type STATECOLOR string
type STATECOLOR_CODE int

const (
	STATECOLOR_GRAY        STATECOLOR      = "SAPControl-GRAY"
	STATECOLOR_GREEN       STATECOLOR      = "SAPControl-GREEN"
	STATECOLOR_YELLOW      STATECOLOR      = "SAPControl-YELLOW"
	STATECOLOR_RED         STATECOLOR      = "SAPControl-RED"
	STATECOLOR_CODE_GRAY   STATECOLOR_CODE = 1
	STATECOLOR_CODE_GREEN  STATECOLOR_CODE = 2
	STATECOLOR_CODE_YELLOW STATECOLOR_CODE = 3
	STATECOLOR_CODE_RED    STATECOLOR_CODE = 4
)

type EnqGetStatistic struct {
	XMLName xml.Name `xml:"urn:SAPControl EnqGetStatistic"`
}

type EnqGetStatisticResponse struct {
	XMLName            xml.Name   `xml:"urn:SAPControl EnqStatistic"`
	OwnerNow           int32      `xml:"owner-now,omitempty" json:"owner-now,omitempty"`
	OwnerHigh          int32      `xml:"owner-high,omitempty" json:"owner-high,omitempty"`
	OwnerMax           int32      `xml:"owner-max,omitempty" json:"owner-max,omitempty"`
	OwnerState         STATECOLOR `xml:"owner-state,omitempty" json:"owner-state,omitempty"`
	ArgumentsNow       int32      `xml:"arguments-now,omitempty" json:"arguments-now,omitempty"`
	ArgumentsHigh      int32      `xml:"arguments-high,omitempty" json:"arguments-high,omitempty"`
	ArgumentsMax       int32      `xml:"arguments-max,omitempty" json:"arguments-max,omitempty"`
	ArgumentsState     STATECOLOR `xml:"arguments-state,omitempty" json:"arguments-state,omitempty"`
	LocksNow           int32      `xml:"locks-now,omitempty" json:"locks-now,omitempty"`
	LocksHigh          int32      `xml:"locks-high,omitempty" json:"locks-high,omitempty"`
	LocksMax           int32      `xml:"locks-max,omitempty" json:"locks-max,omitempty"`
	LocksState         STATECOLOR `xml:"locks-state,omitempty" json:"locks-state,omitempty"`
	EnqueueRequests    int64      `xml:"enqueue-requests,omitempty" json:"enqueue-requests,omitempty"`
	EnqueueRejects     int64      `xml:"enqueue-rejects,omitempty" json:"enqueue-rejects,omitempty"`
	EnqueueErrors      int64      `xml:"enqueue-errors,omitempty" json:"enqueue-errors,omitempty"`
	DequeueRequests    int64      `xml:"dequeue-requests,omitempty" json:"dequeue-requests,omitempty"`
	DequeueErrors      int64      `xml:"dequeue-errors,omitempty" json:"dequeue-errors,omitempty"`
	DequeueAllRequests int64      `xml:"dequeue-all-requests,omitempty" json:"dequeue-all-requests,omitempty"`
	CleanupRequests    int64      `xml:"cleanup-requests,omitempty" json:"cleanup-requests,omitempty"`
	BackupRequests     int64      `xml:"backup-requests,omitempty" json:"backup-requests,omitempty"`
	ReportingRequests  int64      `xml:"reporting-requests,omitempty" json:"reporting-requests,omitempty"`
	CompressRequests   int64      `xml:"compress-requests,omitempty" json:"compress-requests,omitempty"`
	VerifyRequests     int64      `xml:"verify-requests,omitempty" json:"verify-requests,omitempty"`
	LockTime           float64    `xml:"lock-time,omitempty" json:"lock-time,omitempty"`
	LockWaitTime       float64    `xml:"lock-wait-time,omitempty" json:"lock-wait-time,omitempty"`
	ServerTime         float64    `xml:"server-time,omitempty" json:"server-time,omitempty"`
	ReplicationState   STATECOLOR `xml:"replication-state,omitempty" json:"replication-state,omitempty"`
}

type GetInstanceProperties struct {
	XMLName xml.Name `xml:"urn:SAPControl GetInstanceProperties"`
}

type GetInstancePropertiesResponse struct {
	XMLName    xml.Name            `xml:"urn:SAPControl GetInstancePropertiesResponse"`
	Properties []*InstanceProperty `xml:"properties>item,omitempty" json:"properties>item,omitempty"`
}

type GetProcessList struct {
	XMLName xml.Name `xml:"urn:SAPControl GetProcessList"`
}

type GetProcessListResponse struct {
	XMLName   xml.Name     `xml:"urn:SAPControl GetProcessListResponse"`
	Processes []*OSProcess `xml:"process>item,omitempty" json:"process>item,omitempty"`
}

type GetQueueStatistic struct {
	XMLName xml.Name `xml:"urn:SAPControl GetQueueStatistic"`
}

type GetQueueStatisticResponse struct {
	XMLName xml.Name            `xml:"urn:SAPControl GetQueueStatisticResponse"`
	Queues  []*TaskHandlerQueue `xml:"queue>item,omitempty" json:"queue>item,omitempty"`
}

type GetSystemInstanceList struct {
	XMLName xml.Name `xml:"urn:SAPControl GetSystemInstanceList"`
	Timeout int32    `xml:"timeout,omitempty" json:"timeout,omitempty"`
}

type GetSystemInstanceListResponse struct {
	XMLName   xml.Name       `xml:"urn:SAPControl GetSystemInstanceListResponse"`
	Instances []*SAPInstance `xml:"instance>item,omitempty" json:"instance>item,omitempty"`
}

type OSProcess struct {
	Name        string     `xml:"name,omitempty" json:"name,omitempty"`
	Description string     `xml:"description,omitempty" json:"description,omitempty"`
	Dispstatus  STATECOLOR `xml:"dispstatus,omitempty" json:"dispstatus,omitempty"`
	Textstatus  string     `xml:"textstatus,omitempty" json:"textstatus,omitempty"`
	Starttime   string     `xml:"starttime,omitempty" json:"starttime,omitempty"`
	Elapsedtime string     `xml:"elapsedtime,omitempty" json:"elapsedtime,omitempty"`
	Pid         int32      `xml:"pid,omitempty" json:"pid,omitempty"`
}

type InstanceProperty struct {
	Property     string `xml:"property,omitempty" json:"property,omitempty"`
	Propertytype string `xml:"propertytype,omitempty" json:"propertytype,omitempty"`
	Value        string `xml:"value,omitempty" json:"value,omitempty"`
}

type SAPInstance struct {
	Hostname      string     `xml:"hostname,omitempty" json:"hostname,omitempty"`
	InstanceNr    int32      `xml:"instanceNr,omitempty" json:"instanceNr,omitempty"`
	HttpPort      int32      `xml:"httpPort,omitempty" json:"httpPort,omitempty"`
	HttpsPort     int32      `xml:"httpsPort,omitempty" json:"httpsPort,omitempty"`
	StartPriority string     `xml:"startPriority,omitempty" json:"startPriority,omitempty"`
	Features      string     `xml:"features,omitempty" json:"features,omitempty"`
	Dispstatus    STATECOLOR `xml:"dispstatus,omitempty" json:"dispstatus,omitempty"`
}

type TaskHandlerQueue struct {
	Type   string `xml:"Typ,omitempty" json:"Typ,omitempty"`
	Now    int32  `xml:"Now,omitempty" json:"Now,omitempty"`
	High   int32  `xml:"High,omitempty" json:"High,omitempty"`
	Max    int32  `xml:"Max,omitempty" json:"Max,omitempty"`
	Writes int32  `xml:"Writes,omitempty" json:"Writes,omitempty"`
	Reads  int32  `xml:"Reads,omitempty" json:"Reads,omitempty"`
}

type GetAlerts struct {
	XMLName xml.Name `xml:"urn:SAPControl GetAlerts"`
	RootTid string   `xml:"RootTid,omitempty" json:"RootTid,omitempty"`
}

type GetAlertsResponse struct {
	XMLName     xml.Name `xml:"urn:SAPControl GetAlertsResponse"`
	RootTidName string   `xml:"RootTidName,omitempty" json:"RootTidName,omitempty"`
	Alerts      []*Alert `xml:"alert>item,omitempty" json:"instance>item,omitempty"`
}

type Alert struct {
	Object      string     `xml:"Object,omitempty" json:"Object,omitempty"`
	Attribute   string     `xml:"Attribute,omitempty" json:"Attribute,omitempty"`
	Value       STATECOLOR `xml:"Value,omitempty" json:"Value,omitempty"`
	Description string     `xml:"Description,omitempty" json:"Description,omitempty"`
	ATime       string     `xml:"Time,omitempty" json:"Time,omitempty"`
	Tid         string     `xml:"Tid,omitempty" json:"Tid,omitempty"`
	Aid         string     `xml:"Aid,omitempty" json:"Aid,omitempty"`
}

type ABAPGetWPTable struct {
	XMLName xml.Name `xml:"urn:SAPControl ABAPGetWPTable"`
}
type ABAPGetWPTableResponse struct {
	XMLName     xml.Name       `xml:"urn:SAPControl ABAPGetWPTableResponse"`
	WorkProcess []*WorkProcess `xml:"workprocess>item,omitempty" json:"workprocess>item,omitempty"`
}
type WorkProcess struct {
	No      string `xml:"No,omitempty" json:"No,omitempty"`
	Type    string `xml:"Typ,omitempty" json:"Typ,omitempty"`
	Pid     string `xml:"Pid,omitempty" json:"Pid,omitempty"`
	Status  string `xml:"Status,omitempty" json:"Status,omitempty"`
	Reason  string `xml:"Reason,omitempty" json:"Reason,omitempty"`
	Start   string `xml:"Start,omitempty" json:"Start,omitempty"`
	Err     string `xml:"Err,omitempty" json:"Err,omitempty"`
	Sem     string `xml:"Sem,omitempty" json:"Sem,omitempty"`
	Cpu     string `xml:"Cpu,omitempty" json:"Cpu,omitempty"`
	Time    string `xml:"Time,omitempty" json:"Time,omitempty"`
	Program string `xml:"Program,omitempty" json:"Program,omitempty"`
	Client  string `xml:"Client,omitempty" json:"Client,omitempty"`
	User    string `xml:"User,omitempty" json:"User,omitempty"`
	Action  string `xml:"Action,omitempty" json:"Action,omitempty"`
	Table   string `xml:"Table,omitempty" json:"Table,omitempty"`
}

type webService struct {
	Client *MyClient
	//once               *sync.Once
	//currentSapInstance *CurrentSapInstance
	LokiClient promtail.Client
}

// constructor of a WebService interface
func NewWebService(myClient *MyClient) WebService {
	return &webService{
		Client: myClient,
		//once:       &sync.Once{},
		LokiClient: nil,
	}
}

func (s *webService) GetMyClient() *MyClient {
	return s.Client
}
func (s *webService) SetLokiClient(pClient promtail.Client) {
	s.LokiClient = pClient
}
func (s *webService) GetLokiClient() promtail.Client {
	return s.LokiClient
}

// implements WebService.GetSystemInstanceList(context.Context)
func (s *webService) GetSystemInstanceList(ctx context.Context) (*GetSystemInstanceListResponse, error) {

	c := s.Client
	log := s.Client.logger
	sapURL := c.config.Viper.GetString("sap_control_url")

	endpoints := []string{
		fmt.Sprintf("%s/sap/bc/soap/rfc", sapURL),
		fmt.Sprintf("%s/SAPControl.cgi", sapURL),
		fmt.Sprintf("%s/sap/bc/webdynpro/sap/dba_control", sapURL),
	}
	var lastErr []string
	for _, endpoint := range endpoints {
		client := c.CreateSoapClient(endpoint)

		request := &GetSystemInstanceList{}
		response := &GetSystemInstanceListResponse{}

		if err := client.CallContext(ctx, "GetSystemInstanceList", request, response); err != nil {
			lastErr = append(lastErr, fmt.Sprintf("%v", err))
			continue
		}
		if len(response.Instances) == 0 {
			lastErr = append(lastErr, fmt.Sprintf("no instances found at %s", endpoint))
			continue
		}
		log.Debugf("Got Instancelist from endpoint %s", endpoint)
		return response, nil
	}
	return nil, fmt.Errorf("GetSystemInstanceList: failed to get instances from any endpoint: %s.", strings.Join((lastErr), ", "))
}

// implements WebService.GetInstanceProperties(context.Context, string)
func (s *webService) GetInstanceProperties(ctx context.Context, endpoint string) (*GetInstancePropertiesResponse, error) {
	c := s.Client
	endpoint = fmt.Sprintf("%s/sap/bc/soap/rfc", endpoint)
	client := c.CreateSoapClient(endpoint)

	request := &GetInstanceProperties{}
	response := &GetInstancePropertiesResponse{}

	err := client.CallContext(ctx, "GetInstanceProperties", request, response)
	if err != nil {
		return nil, fmt.Errorf("GetInstanceProperties failed, endpoint=%s, err=%v", endpoint, err)
	}
	return response, nil
}

// implements WebService.GetProcessList(context.Context, string)
func (s *webService) GetProcessList(ctx context.Context, endpoint string) (*GetProcessListResponse, error) {
	c := s.Client
	endpoint = fmt.Sprintf("%s/sap/bc/soap/rfc", endpoint)
	client := c.CreateSoapClient(endpoint)

	request := &GetProcessList{}
	response := &GetProcessListResponse{}

	err := client.CallContext(ctx, "GetProcessList", request, response)
	if err != nil {
		return nil, fmt.Errorf("GetProcessList failed, endpoint=%s, err=%v", endpoint, err)
	}
	return response, nil
}

// implements WebService.EnqGetStatistic(context.Context, string)
func (s *webService) EnqGetStatistic(ctx context.Context, endpoint string) (*EnqGetStatisticResponse, error) {
	c := s.Client
	endpoint = fmt.Sprintf("%s/sap/bc/soap/rfc", endpoint)
	client := c.CreateSoapClient(endpoint)

	request := &EnqGetStatistic{}
	response := &EnqGetStatisticResponse{}

	err := client.CallContext(ctx, "EnqGetStatistic", request, response)
	if err != nil {
		return nil, fmt.Errorf("EnqGetStatistic failed, endpoint=%s, err=%v", endpoint, err)
	}
	return response, nil
}

// implements WebService.GetQueueStatistic(context.Context, string)
func (s *webService) GetQueueStatistic(ctx context.Context, endpoint string) (*GetQueueStatisticResponse, error) {
	c := s.Client
	endpoint = fmt.Sprintf("%s/sap/bc/soap/rfc", endpoint)
	client := c.CreateSoapClient(endpoint)

	request := &GetQueueStatistic{}
	response := &GetQueueStatisticResponse{}

	err := client.CallContext(ctx, "GetQueueStatistic", request, response)
	if err != nil {
		return nil, fmt.Errorf("GetQueueStatistic failed, endpoint=%s, err=%v", endpoint, err)
	}
	return response, nil
}

// implements WebService.GetAlerts(context.Context, string)
func (s *webService) GetAlerts(ctx context.Context, endpoint string) (*GetAlertsResponse, error) {
	c := s.Client
	endpoint = fmt.Sprintf("%s/sap/bc/soap/rfc", endpoint)
	client := c.CreateSoapClient(endpoint)

	request := &GetAlerts{}
	response := &GetAlertsResponse{}

	err := client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, fmt.Errorf("GetAlerts failed, endpoint=%s, err=%v", endpoint, err)
	}
	return response, nil
}

// makes the STATECOLOR values more metric friendly
func StateColorToFloat(statecolor STATECOLOR) (float64, error) {
	switch statecolor {
	case STATECOLOR_GRAY:
		return float64(STATECOLOR_CODE_GRAY), nil
	case STATECOLOR_GREEN:
		return float64(STATECOLOR_CODE_GREEN), nil
	case STATECOLOR_YELLOW:
		return float64(STATECOLOR_CODE_YELLOW), nil
	case STATECOLOR_RED:
		return float64(STATECOLOR_CODE_RED), nil
	default:
		return 0, errors.New("Invalid STATECOLOR value")
	}
}

// makes the STATECOLOR values more metric friendly
func StateColorToLevel(statecolor STATECOLOR) (string, error) {
	switch statecolor {
	case STATECOLOR_GRAY:
		return "unknown", nil
	case STATECOLOR_GREEN:
		return "info", nil
	case STATECOLOR_YELLOW:
		return "warning", nil
	case STATECOLOR_RED:
		return "error", nil
	default:
		return "alert", errors.New("Invalid STATECOLOR value")
	}
}

// removes any duplicates in the array of comparable elements, e.g. structs
//
//	parameter - array, returns same type array, but w/o duplicated elements
func RemoveDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// make map from the string in format "KEY1=VALUE1;KEY2=VALUE2;...;KEYx=VALUEx;"
func Make_string_map(s string) map[string]string {

	m := make(map[string]string)
	var s_arr []string

	s_arr = strings.Split(s, ";")
	for _, item := range s_arr {
		if before, after, found := strings.Cut(item, "="); found {
			m[before] = after
		}
	}
	return m
}
