package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"net"
)

func init() {
	flag.Parse()

	parsedLevel, err := log.ParseLevel(*rawLevel)
	if err != nil {
		log.Fatal(err)
	}
	logLevel = parsedLevel

	prometheus.MustRegister(version.NewCollector("zookeeper_exporter"))
}

var (
	logLevel      log.Level = log.ErrorLevel
	bindAddr                = flag.String("bind-addr", "/dev/shm/zookeeper.sock", "bind address for the metrics server")
	metricsPath             = flag.String("metrics-path", "/metrics", "path to metrics endpoint")
	zookeeperAddr           = flag.String("zookeeper", "localhost:2181", "host:port for zookeeper socket")
	rawLevel                = flag.String("log-level", "info", "log level")
	resetOnScrape           = flag.Bool("reset-on-scrape", true, "should a reset command be sent to zookeeper on each scrape")
	showVersion             = flag.Bool("version", false, "show version and exit")
)

func main() {
	log.SetLevel(logLevel)
	if *showVersion {
		return
	}

	go serveMetrics()

	exitChannel := make(chan os.Signal)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitSignal := <-exitChannel
	log.WithFields(log.Fields{"signal": exitSignal}).Infof("Caught %s signal, exiting", exitSignal)
}

func serveMetrics() {
	mux := http.NewServeMux()
	mux.Handle(*metricsPath, prometheus.Handler())
	mux.HandleFunc("/", rootHandler)
	server := http.Server{
		Handler: mux, // http.DefaultServeMux,
	}
	os.Remove(*bindAddr)

	listener, err := net.Listen("unix", *bindAddr)
	if err != nil {
		panic(err)
	}
	server.Serve(listener)
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html>
		<head><title>Zookeeper Exporter</title></head>
		<body>
		<h1>Zookeeper Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
}
