package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/vgrusdev/sap_system_exporter/cache"
	"github.com/vgrusdev/sap_system_exporter/collector/registry"
	"github.com/vgrusdev/sap_system_exporter/collector/start_service"
	"github.com/vgrusdev/sap_system_exporter/internal"
	"github.com/vgrusdev/sap_system_exporter/internal/config"
	"github.com/vgrusdev/sap_system_exporter/lib/sapcontrol"
)

var (
	// the released version
	version string = "development"
	// the time the binary was built
	buildDate string = "March 2025"
	// global --help flag
	helpFlag *bool
	// global --version flag
	versionFlag *bool
)

func init() {
	flag.String("port", "9680", "The port number to listen on for HTTP requests")
	flag.String("address", "0.0.0.0", "The address to listen on for HTTP requests")
	flag.String("log-level", "info", "The minimum logging level; levels are, in ascending order: debug, info, warn, error")
	flag.String("sap-control-url", "localhost:50013", "The URL of the SAPControl SOAP web service, e.g. [https://]$HOST:$PORT. Port: 5xx13(http) or 5xx14(https). Recommendation to connect to Central Instance.")
	flag.String("host-domain", "", "Optional Domain name to make FQDN together with hostname, Recommended in case of SAP hostname is a sigle-word hostname.")
	flag.String("tls-skip-verify", "no", "For HTTPS scheme, should certificates signed by unknown authority being ignored")
	flag.String("alert_samples_max_age", "24h", "Oldest acceptable timestamp for Alert item (back since now()). Use \"-1s\" for unlim.")
	flag.StringP("config", "c", "", "The path to a custom configuration file. NOTE: it must be in yaml format.")
	flag.CommandLine.SortFlags = false

	helpFlag = flag.BoolP("help", "h", false, "show this help message")
	versionFlag = flag.Bool("version", false, "show version and build information")
}

func main() {
	flag.Parse()

	switch {
	case *helpFlag:
		showHelp()
	case *versionFlag:
		showVersion()
	default:
		run()
	}
}

func run() {
	var err error

	// Initialize logger
	logger := config.NewLogger("main")

	myConfig, err := config.New(flag.CommandLine)
	if err != nil {
		log.Fatalf("Could not initialize config: %s", err)
	}
	v := myConfig.Viper

	logger.SetLevel(v.GetString("log_level"))
	//logger.Debug("Config %s", )
	logger.Info("Starting SAP System Exporter",
		"version", version,
		"sap_control_url", v.GetString("sap_control_url"),
		"loki_url", v.GetString("loki_url"),
		//"primary_instance", cfg.PrimaryInstance,
		//"host", cfg.Host,
		//"port", cfg.Port,
	)

	// Initialize cache manager
	cacheMgr := cache.NewManager(cfg.CacheTTL)

	myClient := sapcontrol.NewSoapClient(myConfig)
	loki_client := sapcontrol.NewLokiClient(myConfig)

	if loki_client != nil {
		defer loki_client.Shutdown()
	}

	webService := sapcontrol.NewWebService(myClient)
	webService.SetLokiClient(loki_client)

	// VG ++
	/*
		client = webService.GetMyClient()
		myConfig, _ := client.Config.Copy()
		_ = myConfig.SetURL("http://abc:3456")
		myClient := sapcontrol.NewSoapClient(myConfig)
		myWebService := sapcontrol.NewWebService(myClient)

		_, _ = myWebService.GetCurrentInstance()
	*/
	// VG --

	currentSapInstance, err := webService.GetCurrentInstance()
	if err != nil {
		log.Fatal(errors.Wrap(err, "SAPControl web service error"))
	}

	log.Infof("Monitoring SAP Instance %s", currentSapInstance)

	startServiceCollector, err := start_service.NewCollector(webService)
	if err != nil {
		log.Warn(err)
	} else {
		prometheus.MustRegister(startServiceCollector)
		log.Info("Start Service collector registered")
	}

	err = registry.RegisterOptionalCollectors(webService)
	if err != nil {
		log.Fatal(err)
	}

	// if we're not in debug log level, we unregister the Go runtime metrics collector that gets registered by default
	if !log.IsLevelEnabled(log.DebugLevel) {
		// prometheus.Unregister(prometheus.NewGoCollector())
		prometheus.Unregister(collectors.NewGoCollector())
	}

	fullListenAddress := fmt.Sprintf("%s:%s", myConfig.Viper.Get("address"), myConfig.Viper.Get("port"))

	http.HandleFunc("/", internal.Landing)
	http.Handle("/metrics", promhttp.Handler())

	log.Infof("Serving metrics on %s", fullListenAddress)
	log.Fatal(http.ListenAndServe(fullListenAddress, nil))
}

func showHelp() {
	flag.Usage()
	os.Exit(0)
}

func showVersion() {
	fmt.Printf("%s version\nbuilt with %s %s/%s %s\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH, buildDate)
	os.Exit(0)
}
