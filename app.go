package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/relations-api/relations"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
	metrics "github.com/rcrowley/go-metrics"
)

const (
	serviceName        = "relations-api-neo4j"
	serviceDescription = "A public RESTful API for accessing Relations in neo4j"
)

func main() {
	app := cli.App(serviceName, serviceDescription)
	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://localhost:7474/db/data",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL"})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})
	cacheDuration := app.String(cli.StringOpt{
		Name:   "cache-duration",
		Value:  "30s",
		Desc:   "Duration Get requests should be cached for. e.g. 2h45m would set the max-age value to '7440' seconds",
		EnvVar: "CACHE_DURATION",
	})

	logger.InitDefaultLogger(serviceName)

	app.Action = func() {
		logger.Infof("relations-api will listen on port: %s, connecting to: %s", *port, *neoURL)
		runServer(*neoURL, *port, *cacheDuration)
	}
	logger.Infof("Application started with args %s", os.Args)
	app.Run(os.Args)
}

func runServer(neoURL string, port string, cacheDuration string) {
	var cacheControlHeader string
	if duration, durationErr := time.ParseDuration(cacheDuration); durationErr != nil {
		logger.Fatalf("Failed to parse cache duration string, %v", durationErr)
	} else {
		cacheControlHeader = fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))
	}

	conf := neoutils.ConnectionConfig{
		BatchSize:     1024,
		Transactional: false,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
			},
			Timeout: 1 * time.Minute,
		},
		BackgroundConnect: true,
	}
	conn, err := neoutils.Connect(neoURL, &conf)

	if err != nil {
		logger.Fatalf("Error connecting to neo4j %s", err)
	}

	httpHandlers := relations.NewHttpHandlers(relations.NewCypherDriver(conn), cacheControlHeader)
	// The following endpoints should not be monitored or logged (varnish calls one of these every second, depending on config)
	// The top one of these build info endpoints feels more correct, but the lower one matches what we have in Dropwizard,
	// so it's what apps expect currently same as ping, the content of build-info needs more definition
	healthCheck := fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  "upp-relations-api",
			Name:        "RelationsApi Healthchecks",
			Description: "Checks for accessing neo4j",
			Checks:      []fthealth.Check{httpHandlers.HealthCheck(neoURL)},
		},
		Timeout: 10 * time.Second,
	}
	http.HandleFunc("/__health", fthealth.Handler(healthCheck))
	http.HandleFunc(status.PingPath, status.PingHandler)
	http.HandleFunc(status.PingPathDW, status.PingHandler)
	http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	http.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)
	http.HandleFunc("/__gtg", status.NewGoodToGoHandler(httpHandlers.GTG))

	http.Handle("/", router(httpHandlers))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatalf("Unable to start server: %v", err)
	}
}

func router(hh relations.HttpHandlers) http.Handler {
	servicesRouter := mux.NewRouter()

	servicesRouter.HandleFunc("/content/{uuid}/relations", hh.GetContentRelations).Methods("GET")
	servicesRouter.HandleFunc("/contentcollection/{uuid}/relations", hh.GetContentCollectionRelations).Methods("GET")

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(logger.Logger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	return monitoringRouter
}
