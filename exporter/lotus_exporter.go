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
	lotusInfo                *prometheus.Desc
	lotusLocalTime           *prometheus.Desc
	lotusChainBasefee        *prometheus.Desc
	lotusChainHeight         *prometheus.Desc
	lotusChainSyncDiff       *prometheus.Desc
	lotusChainSyncStatus     *prometheus.Desc
	lotusMpoolTotal          *prometheus.Desc
	lotusMpoolLocalTotal     *prometheus.Desc
	lotusMpoolLocalMessage   *prometheus.Desc
	lotusPower               *prometheus.Desc
	lotusPowerEligibility    *prometheus.Desc
	lotusWalletBalance       *prometheus.Desc
	lotusWalletLockedBalance *prometheus.Desc
	minerInfo                *prometheus.Desc
	minerInfoSectorSize      *prometheus.Desc
	minerWorkerCpu           *prometheus.Desc
	minerWorkerGpu           *prometheus.Desc
	minerWorkerRamTotal      *prometheus.Desc
	minerWorkerRamReserved   *prometheus.Desc
	minerWorkerRamTasks      *prometheus.Desc
	minerWorkerVmemTotal     *prometheus.Desc
	minerWorkerVmemReserved  *prometheus.Desc
	minerWorkerVmemTasks     *prometheus.Desc
	minerWorkerCpuUsed       *prometheus.Desc
	minerWorkerGpuUsed       *prometheus.Desc

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
		lotusMpoolTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "mpool_total"),
			"return number of message pending in mpool",
			[]string{"miner_id"}, nil,
		),
		lotusMpoolLocalTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "mpool_local_total"),
			"return number of messages pending in local mpool",
			[]string{"miner_id"}, nil,
		),
		lotusMpoolLocalMessage: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "mpool_local_message"),
			"local message details",
			[]string{"miner_id", "msg_from", "msg_to", "msg_nonce", "msg_value", "msg_gaslimit", "msg_gasfeecap", "msg_gaspremium",
				"msg_method", "msg_method_type", "msg_to_actor_type"}, nil,
		),
		lotusPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "power"),
			"return miner power",
			[]string{"miner_id", "scope", "power_type"}, nil,
		),
		lotusPowerEligibility: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "power_mining_eligibility"),
			"return miner mining eligibility",
			[]string{"miner_id"}, nil,
		),
		lotusWalletBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "wallet_balance"),
			"return wallet balance",
			[]string{"miner_id", "address", "name"}, nil,
		),
		lotusWalletLockedBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "wallet_locked_balance"),
			"return miner wallet locked funds",
			[]string{"miner_id", "address", "locked_type"}, nil,
		),
		minerInfo: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_info"),
			"lotus miner information like address version etc",
			[]string{"miner_id", "version", "owner", "owner_addr", "worker", "worker_addr", "control0", "control0_addr"}, nil,
		),
		minerInfoSectorSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_info_sector_size"),
			"lotus miner sector size",
			[]string{"miner_id"}, nil,
		),
		minerWorkerCpu: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_cpu"),
			"number of CPU used by lotus",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerGpu: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_gpu"),
			"is the GPU used by lotus",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerRamTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_ram_total"),
			"worker server RAM",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerRamReserved: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_ram_reserved"),
			"worker memory reserved by lotus",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerRamTasks: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_ram_tasks"),
			"worker minimal memory used",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerVmemTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_vmem_total"),
			"server Physical RAM + Swap",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerVmemReserved: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_vmem_reserved"),
			"worker VMEM used by on-going tasks",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerVmemTasks: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_vmem_tasks"),
			"worker VMEM reserved by lotus",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerCpuUsed: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_cpu_used"),
			"number of CPU used by lotused by lotus",
			[]string{"miner_id", "worker_host"}, nil,
		),
		minerWorkerGpuUsed: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "miner_worker_gpu_used"),
			"is the GPU used by lotus",
			[]string{"miner_id", "worker_host"}, nil,
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
	osOwnerID := os.Getenv("OWNER_ID")
	osOwnerADDR := os.Getenv("OWNER_ADDR")

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

	// get owner info
	var ownerID, ownerADDR string
	if osOwnerID != "" {
		ownerID = osOwnerID
	} else {
		ownerID = minerInfo.Owner
	}

	if osOwnerADDR != "" {
		ownerADDR = osOwnerADDR
	} else {
		ownerADDR = minerInfo.OwnerAddr
	}

	// get local wallet
	walletList := []string{ownerADDR, minerInfo.WorkerAddr, minerInfo.Control0Addr}

	// get chain sync info
	chainSyncStats := lotusinfo.GetChainSyncState(ctx, fuApi)

	// get chain height
	chainHeight := lotusinfo.GetChainHeight(chainTipSetKey)

	// get chain basefee
	basefee := lotusinfo.GetChainBasefee(chainTipSetKey)

	// get mpool total, local msg total, local msg list
	mpoolTotal, localMpollTotal, msgLst := lotusinfo.GetMpoolInfo(ctx, fuApi, chainTipSetKey, walletList)

	// get miner power
	mpRaw, mpQua, tpRaw, tpQua := lotusinfo.GetPowerList(ctx, fuApi, minerId, chainTipSetKey)

	// get miner power eligibility
	powerEligibility := lotusinfo.GetBaseInfo(ctx, fuApi, minerId, chainHeight, chainTipSetKey)

	// get Locked Balance
	lockedInfoS := lotusinfo.GetLockedFunds(ctx, fuApi, minerId, chainTipSetKey)

	// get miner version
	minerVersion, err := lotusinfo.GetMinerVersion(ctx, miApi)
	if err != nil {
		log.Fatal(err)
	}

	// get worker info
	workerGroupInfo := lotusinfo.GetWorkerInfo(ctx, miApi)

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

	ch <- prometheus.MustNewConstMetric(collector.lotusMpoolTotal, prometheus.GaugeValue, float64(mpoolTotal), minerId)
	ch <- prometheus.MustNewConstMetric(collector.lotusMpoolLocalTotal, prometheus.GaugeValue, float64(localMpollTotal), minerId)
	for _, mmsg := range msgLst {
		ch <- prometheus.MustNewConstMetric(collector.lotusMpoolLocalMessage, prometheus.GaugeValue, 1, minerId,
			mmsg.Mfrom, mmsg.Mto, string(mmsg.Mnonce), string(mmsg.Mvalue), string(mmsg.Mgaslimit), string(mmsg.Mgasfeecap),
			string(mmsg.Mgaspremium), string(mmsg.Mmethod), mmsg.Mmethodtype, mmsg.Mactortype)
	}

	ch <- prometheus.MustNewConstMetric(collector.lotusPower, prometheus.GaugeValue, float64(mpRaw), minerId, "miner", "RawBytePower")
	ch <- prometheus.MustNewConstMetric(collector.lotusPower, prometheus.GaugeValue, float64(mpQua), minerId, "miner", "QualityAdjPower")
	ch <- prometheus.MustNewConstMetric(collector.lotusPower, prometheus.GaugeValue, float64(tpRaw), minerId, "network", "RawBytePower")
	ch <- prometheus.MustNewConstMetric(collector.lotusPower, prometheus.GaugeValue, float64(tpQua), minerId, "network", "QualityAdjPower")
	ch <- prometheus.MustNewConstMetric(collector.lotusPowerEligibility, prometheus.GaugeValue, float64(powerEligibility), minerId)

	ch <- prometheus.MustNewConstMetric(collector.lotusWalletBalance, prometheus.GaugeValue, float64(lotusinfo.GetWalletBalance(ctx, fuApi, minerId)), minerId, minerId, minerId)
	ch <- prometheus.MustNewConstMetric(collector.lotusWalletBalance, prometheus.GaugeValue, float64(lotusinfo.GetWalletBalance(ctx, fuApi, ownerADDR)), minerId, ownerID, ownerADDR)
	ch <- prometheus.MustNewConstMetric(collector.lotusWalletBalance, prometheus.GaugeValue, float64(lotusinfo.GetWalletBalance(ctx, fuApi, minerInfo.WorkerAddr)), minerId, minerInfo.Worker, minerInfo.WorkerAddr)
	ch <- prometheus.MustNewConstMetric(collector.lotusWalletBalance, prometheus.GaugeValue, float64(lotusinfo.GetWalletBalance(ctx, fuApi, minerInfo.Control0Addr)), minerId, minerInfo.Control0, minerInfo.Control0Addr)

	for _, lockedI := range lockedInfoS {
		ch <- prometheus.MustNewConstMetric(collector.lotusWalletLockedBalance, prometheus.GaugeValue, float64(lockedI.Balance), minerId, minerId, lockedI.LockedType)
	}

	ch <- prometheus.MustNewConstMetric(collector.minerInfo, prometheus.GaugeValue, 1, minerId, minerVersion, ownerID, ownerADDR,
		minerInfo.Worker, minerInfo.WorkerAddr, minerInfo.Control0, minerInfo.Control0Addr)

	ch <- prometheus.MustNewConstMetric(collector.minerInfoSectorSize, prometheus.GaugeValue, float64(minerInfo.SectorSize), minerId)

	for _, worker0 := range workerGroupInfo {
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerCpu, prometheus.GaugeValue, float64(worker0.WCpu), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerGpu, prometheus.GaugeValue, float64(worker0.WGpu), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerRamTotal, prometheus.GaugeValue, float64(worker0.WRamTotal), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerRamReserved, prometheus.GaugeValue, float64(worker0.WRamReserved), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerRamTasks, prometheus.GaugeValue, float64(worker0.WRamTasks), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerVmemTotal, prometheus.GaugeValue, float64(worker0.WVmemTotal), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerVmemReserved, prometheus.GaugeValue, float64(worker0.WVmemReseved), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerVmemTasks, prometheus.GaugeValue, float64(worker0.WvmemTasks), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerCpuUsed, prometheus.GaugeValue, float64(worker0.WCpuUsed), minerId, worker0.WHost)
		ch <- prometheus.MustNewConstMetric(collector.minerWorkerGpuUsed, prometheus.GaugeValue, float64(worker0.WGpuUsed), minerId, worker0.WHost)
	}
}

// Register registers the volume metrics
func Register(options *LotusOpt) {
	collector := newLotusCollector(options)
	prometheus.MustRegister(version.NewCollector("lotus_exporter"))
	prometheus.MustRegister(collector)
}
