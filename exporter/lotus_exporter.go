package exporter

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/spark8899/lotus_exporter/lotus"
)

const (
	namespace = "lotus" //for Prometheus metrics.
)

// LotusOpt is for option
type LotusOpt struct {
	FullNodeApiInfo string
	MinerApiInfo    string
}

// setting collector
type lotusCollector struct {
	lotusInfo      *prometheus.Desc
	lotusLocalTime *prometheus.Desc

	ltOptions LotusOpt
}

//You must create a constructor for your collector that
//initializes every descriptor and returns a pointer to the collector
func newLotusCollector(opts *LotusOpt) *lotusCollector {
	return &lotusCollector{
		lotusLocalTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "local_time"),
			"lotus_local_time time on the node machine when last execution start in epoch",
			nil, nil,
		),
		lotusInfo: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "info"),
			"lotus daemon information like address version, value is set to network version number",
			[]string{"miner_id", "version", "network"}, nil,
		),

		ltOptions: *opts,
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *lotusCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.lotusInfo
	ch <- collector.lotusLocalTime
}

//Collect implements required collect function for all promehteus collectors
func (collector *lotusCollector) Collect(ch chan<- prometheus.Metric) {

	//Implement logic here to determine proper metric value to return to prometheus
	//for each descriptor or call other functions that do so.

	fullNodeApiInfo := collector.ltOptions.FullNodeApiInfo
	minerApiInfo := collector.ltOptions.MinerApiInfo

	minerId, err := lotus.GetMinerID(minerApiInfo)
	if err != nil {
		log.Fatal(err)
	}

	fullNodeInfo, err := lotus.GetInfo(fullNodeApiInfo)
	if err != nil {
		log.Fatal(err)
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.lotusLocalTime, prometheus.GaugeValue, float64(lotus.GetLocalTime()))
	ch <- prometheus.MustNewConstMetric(collector.lotusInfo, prometheus.GaugeValue, float64(fullNodeInfo.Value), minerId, fullNodeInfo.Network, fullNodeInfo.Version)
}

// Register registers the volume metrics
func Register(options *LotusOpt) {
	collector := newLotusCollector(options)
	prometheus.MustRegister(version.NewCollector("lotus_exporter"))
	prometheus.MustRegister(collector)
}
