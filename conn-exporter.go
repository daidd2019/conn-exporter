package main

import (
"log"
"flag"
"net/http"

"github.com/daidd2019/conn-exporter/collector"
"github.com/prometheus/client_golang/prometheus"
"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Set during go build
	// version   string
	// gitCommit string

	// 命令行参数
	listenAddr  = flag.String("web.listen-port", "9001", "An port to listen on for web interface and telemetry.")
	metricsPath = flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	metricsNamespace = flag.String("metric.namespace", "gp", "Prometheus metrics namespace, as the prefix of metrics name")
	appName = flag.String("app.name", "mysql", "app name")
	portFlag = flag.String("port.flag", "10.247.32.250:3306", "port flag")
)


func main() {
	flag.Parse()

	apps := []*collector.PortData{collector.NewPortData(*appName, *portFlag)}

	metrics := collector.NewMetrics(*metricsNamespace, apps)
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>A Prometheus Exporter</title></head>
			<body>
			<h1>A Prometheus Exporter</h1>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Starting Server at http://localhost:%s%s", *listenAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(":"+*listenAddr, nil))
}

