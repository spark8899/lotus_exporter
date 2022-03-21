package exporter

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/spark8899/lotus_exporter/lotusinfo"
	"log"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
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
	lotusInfo            *prometheus.Desc
	lotusLocalTime       *prometheus.Desc
	lotusChainBasefee    *prometheus.Desc
	lotusChainHeight     *prometheus.Desc
	lotusChainSyncDiff   *prometheus.Desc
	lotusChainSyncStatus *prometheus.Desc

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
		lotusChainHeight: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "chain_height"),
			"return current height",
			[]string{"miner_id"}, nil,
		),
		lotusChainBasefee: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "chain_basefee"),
			"return current basefee",
			[]string{"miner_id"}, nil,
		),
		lotusChainSyncDiff: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "chain_sync_diff"),
			"return daemon sync height diff with chainhead for each daemon worker",
			[]string{"miner_id", "worker_id"}, nil,
		),
		lotusChainSyncStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "chain_sync_status"),
			"return daemon sync status with chainhead for each daemon worker",
			[]string{"miner_id", "worker_id"}, nil,
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
	ctx := context.Background()

	// get fullApi
	fullNodeApiInfoS := lotusinfo.ParseApiInfo(strings.TrimSpace(fullNodeApiInfo))
	fullNodeApiHeaders := http.Header{"Authorization": []string{"Bearer " + string(fullNodeApiInfoS.Token)}}
	var fuApi lotusapi.FullNodeStruct
	closer01, err01 := jsonrpc.NewMergeClient(ctx, fullNodeApiInfoS.Addr, "Filecoin", []interface{}{&fuApi.Internal, &fuApi.CommonStruct.Internal}, fullNodeApiHeaders)
	if err01 != nil {
		log.Fatalf("connecting with lotus failed: %s", err01)
	}
	defer closer01()

	// get minerApi
	minerApiInfoS := lotusinfo.ParseApiInfo(strings.TrimSpace(minerApiInfo))
	minerApiHeaders := http.Header{"Authorization": []string{"Bearer " + string(minerApiInfoS.Token)}}
	var miApi lotusapi.StorageMinerStruct
	closer02, err02 := jsonrpc.NewMergeClient(context.Background(), minerApiInfoS.Addr, "Filecoin", []interface{}{&miApi.Internal, &miApi.CommonStruct.Internal}, minerApiHeaders)
	if err02 != nil {
		log.Fatalf("connecting with lotus-miner failed: %s", err02)
	}
	defer closer02()

	// get minerId
	minerId, err := lotusinfo.GetMinerID(ctx, miApi)
	if err != nil {
		log.Fatal(err)
	}

	// get chainHead
	chainHead, err := lotusinfo.GetChainHead(ctx, fuApi)
	if err != nil {
		log.Fatal(err)
	}

	// get lotusInfo
	fullNodeInfo, err := lotusinfo.GetInfo(ctx, fuApi, chainHead)
	if err != nil {
		log.Fatal(err)
	}

	// get chain sync info
	chainSyncStats := lotusinfo.GetChainSyncState(ctx, fuApi)

	// get chain height
	chainHeight := lotusinfo.GetChainHeight(chainHead)

	// get chain basefee
	basefee := lotusinfo.GetChainBasefee(chainHead)

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.lotusLocalTime, prometheus.GaugeValue, float64(lotusinfo.GetLocalTime()))
	ch <- prometheus.MustNewConstMetric(collector.lotusInfo, prometheus.GaugeValue, float64(fullNodeInfo.Value), minerId, fullNodeInfo.Network, fullNodeInfo.Version)
	ch <- prometheus.MustNewConstMetric(collector.lotusChainHeight, prometheus.GaugeValue, float64(chainHeight), minerId)
	ch <- prometheus.MustNewConstMetric(collector.lotusChainBasefee, prometheus.GaugeValue, float64(basefee), minerId)

	for _, i := range chainSyncStats {
		ch <- prometheus.MustNewConstMetric(collector.lotusChainSyncDiff, prometheus.GaugeValue, float64(i.CSDiff), i.CSWorkerID)
		ch <- prometheus.MustNewConstMetric(collector.lotusChainSyncStatus, prometheus.GaugeValue, float64(i.CSStatus), i.CSWorkerID)
	}
}

// Register registers the volume metrics
func Register(options *LotusOpt) {
	collector := newLotusCollector(options)
	prometheus.MustRegister(version.NewCollector("lotus_exporter"))
	prometheus.MustRegister(collector)
}
