package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Financial-Times/api-endpoint"
	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/http-handlers-go/v2/httphandlers"
	"github.com/Financial-Times/relations-api/v3/relations"
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
		Value:  "bolt://localhost:7687",
		Desc:   "neo-url value must use the bolt protocol",
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
	apiYml := app.String(cli.StringOpt{
		Name:   "api-yml",
		Value:  "./api.yml",
		Desc:   "Location of the API Swagger YML file.",
		EnvVar: "API_YML",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "log-level",
		Value:  "INFO",
		Desc:   "Logging level (DEBUG, INFO, WARN, ERROR)",
		EnvVar: "LOG_LEVEL",
	})
	dbDriverLogLevel := app.String(cli.StringOpt{
		Name:   "db-driver-log-level",
		Value:  "WARN",
		Desc:   "Db's driver log level (DEBUG, INFO, WARN, ERROR)",
		EnvVar: "DB_DRIVER_LOG_LEVEL",
	})

	app.Action = func() {
		log := logger.NewUPPLogger(serviceName, *logLevel)
		dbDriverLog := logger.NewUPPLogger(serviceName+"-cm-neo4j-driver", *dbDriverLogLevel)

		log.WithField("args", os.Args).Info("Application started")
		log.Infof("relations-api will listen on port: %s, connecting to: %s", *port, *neoURL)

		runServer(*neoURL, *port, *cacheDuration, *apiYml, log, dbDriverLog)
	}
	app.Run(os.Args)
}

func runServer(neoURL, port, cacheDuration, apiYml string, log, dbDriverLog *logger.UPPLogger) {
	var cacheControlHeader string
	if duration, durationErr := time.ParseDuration(cacheDuration); durationErr != nil {
		log.WithError(durationErr).Fatal("Failed to parse cache duration string")
	} else {
		cacheControlHeader = fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))
	}

	driver, err := cmneo4j.NewDefaultDriver(neoURL, dbDriverLog)
	if err != nil {
		log.WithError(err).Fatal("Failed to create new cmneo4j driver")
	}

	httpHandlers := relations.NewHttpHandlers(relations.NewCypherDriver(driver), cacheControlHeader)
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

	http.Handle("/", router(httpHandlers, apiYml, log))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.WithError(err).Fatal("Unable to start server")
	}
}

func router(hh relations.HttpHandlers, apiYml string, log *logger.UPPLogger) http.Handler {
	servicesRouter := mux.NewRouter()

	servicesRouter.HandleFunc("/content/{uuid}/relations", hh.GetContentRelations).Methods("GET")
	servicesRouter.HandleFunc("/contentcollection/{uuid}/relations", hh.GetContentCollectionRelations).Methods("GET")
	if apiYml != "" {
		if endpoint, err := api.NewAPIEndpointForFile(apiYml); err == nil {
			servicesRouter.HandleFunc(api.DefaultPath, endpoint.ServeHTTP).Methods("GET")
		}
	}

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log, monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	return monitoringRouter
}
