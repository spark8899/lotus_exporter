package main

import (
	"flag"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/spark8899/lotus_exporter/exporter"
	"log"
	"net/http"
	"os"
)

func main() {

	var (
		listenAddress = flag.String("web.listen-address", ":9141",
			"Address to listen on for telemetry")
		metricsPath = flag.String("web.telemetry-path", "/metrics",
			"Path under which to expose metrics")
		configPath = flag.String("config-path", "/etc/lotus_exporter/.env",
			"Path to environment file")
	)

	log.Println("Running lotus_exporter")
	flag.Parse()

	configFile := *configPath
	if configFile != "" {
		log.Printf("Loading %s env file.\n", configFile)
		err := godotenv.Load(configFile)
		if err != nil {
			log.Fatalf("Error loading %s env file.\n", configFile)
		}
	} else {
		err := godotenv.Load()
		if err != nil {
			log.Fatalln("Error loading .env file, assume env variables are set.")
		}
	}

	fullNodeApiInfo := os.Getenv("FULLNODE_API_INFO")
	minerApiInfo := os.Getenv("MINER_API_INFO")

	ltOpt := exporter.LotusOpt{fullNodeApiInfo, minerApiInfo}

	exporter.Register(&ltOpt)

	log.Printf("Starting lotus_exporter\n", version.Info())
	log.Printf("Build context\n", version.BuildContext())

	log.Fatal(serverMetrics(*listenAddress, *metricsPath))
}

func serverMetrics(listenAddress, metricsPath string) error {
	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>Lotus Exporter Metrics</title></head>
			<body>
			<h1>Volume Exporter Metrics</h1>
			<p><a href='` + metricsPath + `'>Metrics</a></p>
			</body>
			</html>
		`))
	})

	log.Printf("Starting Server: %s\n", listenAddress)
	return http.ListenAndServe(listenAddress, nil)
}
