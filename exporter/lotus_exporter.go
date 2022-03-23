package exporter

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/spark8899/lotus_exporter/lotusinfo"
	"log"
	"net/http"
	"os"
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
	lotusInfo             *prometheus.Desc
	lotusLocalTime        *prometheus.Desc
	lotusChainBasefee     *prometheus.Desc
	lotusChainHeight      *prometheus.Desc
	lotusChainSyncDiff    *prometheus.Desc
	lotusChainSyncStatus  *prometheus.Desc
	lotusPower            *prometheus.Desc
	lotusPowerEligibility *prometheus.Desc
	minerInfo             *prometheus.Desc
	minerInfoSectorSize   *prometheus.Desc

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
		lotusPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "power"),
			"return miner power",
			[]string{"miner_id", "scope", "power_type"}, nil,
		),
		lotusPowerEligibility: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "power_mining_eligibility"),
			"return miner mining eligibility",
			[]string{"miner_id"}, nil,
		),
		minerInfo: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_info"),
			"lotus miner information like address version etc",
			[]string{"miner_id", "version", "owner", "owner_addr", "worker", "worker_addr", "control0", "control0_addr"}, nil,
		),
		minerInfoSectorSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_info_sector_size"),
			"lotus miner sector size",
			[]string{"miner_id"}, nil,
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

	// get owner from env
	ownerID := os.Getenv("OWNER_ID")
	ownerADDR := os.Getenv("OWNER_ADDR")
	if ownerID != "" {
		log.Printf("owner id: %s", ownerID)
	}
	if ownerADDR != "" {
		log.Printf("owner addr: %s", ownerADDR)
	}

	// get minerId
	minerId, err := lotusinfo.GetMinerID(ctx, miApi)
	if err != nil {
		log.Fatal(err)
	}

	// get chainHead
	chainTipSetKey, err := lotusinfo.GetTipsetKey(ctx, fuApi)
	if err != nil {
		log.Fatal(err)
	}

	// get lotusInfo
	fullNodeInfo, err := lotusinfo.GetInfo(ctx, fuApi, chainTipSetKey)
	if err != nil {
		log.Fatal(err)
	}

	// get miner info
	minerInfo, err := lotusinfo.GetMinerInfo(ctx, fuApi, minerId, chainTipSetKey)
	if err != nil {
		log.Fatal(err)
	}

	// get local wallet
	walletList := []string{minerInfo.OwnerAddr, minerInfo.WorkerAddr, minerInfo.Control0Addr}
	log.Printf("wallet: %s", walletList)

	// get chain sync info
	chainSyncStats := lotusinfo.GetChainSyncState(ctx, fuApi)

	// get chain height
	chainHeight := lotusinfo.GetChainHeight(chainTipSetKey)

	// get chain basefee
	basefee := lotusinfo.GetChainBasefee(chainTipSetKey)

	// get miner power
	mpRaw, mpQua, tpRaw, tpQua := lotusinfo.GetPowerList(ctx, fuApi, minerId, chainTipSetKey)

	// get miner power eligibility
	powerEligibility := lotusinfo.GetBaseInfo(ctx, fuApi, minerId, chainHeight, chainTipSetKey)

	// get miner version
	minerVersion, err := lotusinfo.GetMinerVersion(ctx, miApi)
	if err != nil {
		log.Fatal(err)
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.lotusLocalTime, prometheus.GaugeValue, float64(lotusinfo.GetLocalTime()))
	ch <- prometheus.MustNewConstMetric(collector.lotusInfo, prometheus.GaugeValue, float64(fullNodeInfo.Value), minerId, fullNodeInfo.Network, fullNodeInfo.Version)
	ch <- prometheus.MustNewConstMetric(collector.lotusChainHeight, prometheus.GaugeValue, float64(chainHeight), minerId)
	ch <- prometheus.MustNewConstMetric(collector.lotusChainBasefee, prometheus.GaugeValue, float64(basefee), minerId)

	for _, i := range chainSyncStats {
		ch <- prometheus.MustNewConstMetric(collector.lotusChainSyncDiff, prometheus.GaugeValue, float64(i.CSDiff), minerId, i.CSWorkerID)
		ch <- prometheus.MustNewConstMetric(collector.lotusChainSyncStatus, prometheus.GaugeValue, float64(i.CSStatus), minerId, i.CSWorkerID)
	}

	ch <- prometheus.MustNewConstMetric(collector.lotusPower, prometheus.GaugeValue, float64(mpRaw), minerId, "miner", "RawBytePower")
	ch <- prometheus.MustNewConstMetric(collector.lotusPower, prometheus.GaugeValue, float64(mpQua), minerId, "miner", "QualityAdjPower")
	ch <- prometheus.MustNewConstMetric(collector.lotusPower, prometheus.GaugeValue, float64(tpRaw), minerId, "network", "RawBytePower")
	ch <- prometheus.MustNewConstMetric(collector.lotusPower, prometheus.GaugeValue, float64(tpQua), minerId, "network", "QualityAdjPower")
	ch <- prometheus.MustNewConstMetric(collector.lotusPowerEligibility, prometheus.GaugeValue, float64(powerEligibility), minerId)

	ch <- prometheus.MustNewConstMetric(collector.minerInfo, prometheus.GaugeValue, 1, minerId, minerVersion, minerInfo.Owner, minerInfo.OwnerAddr,
		minerInfo.Worker, minerInfo.WorkerAddr, minerInfo.Control0, minerInfo.Control0Addr)

	ch <- prometheus.MustNewConstMetric(collector.minerInfoSectorSize, prometheus.GaugeValue, float64(minerInfo.SectorSize), minerId)
}

// Register registers the volume metrics
func Register(options *LotusOpt) {
	collector := newLotusCollector(options)
	prometheus.MustRegister(version.NewCollector("lotus_exporter"))
	prometheus.MustRegister(collector)
}
