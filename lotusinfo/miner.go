package lotusinfo

import (
	"context"
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/extern/sector-storage/sealtasks"
	"log"
	"strconv"
	"time"
)

type MinerInfoStruct struct {
	Owner        string
	OwnerAddr    string
	Worker       string
	WorkerAddr   string
	Control0     string
	Control0Addr string
	SectorSize   uint64
}

type LockedInfoStruct struct {
	LockedType string
	Balance    float64
}

type WorkerInfoStuct struct {
	WHost        string
	WCpu         uint64
	WGpu         int
	WRamTotal    uint64
	WRamReserved int
	WRamTasks    uint64
	WVmemTotal   uint64
	WVmemReseved int
	WvmemTasks   uint64
	WCpuUsed     uint64
	WGpuUsed     int
}

type JobInfoStruct struct {
	JjobId         string
	Jsector        string
	Jhost          string
	Jtask          string
	JjobStartTime  string
	JrunWait       string
	JjobStartEpoch int
}

type SchedInfoStruct struct {
	Sector string
	Task   string
}

type SchedDiagRequestInfo struct {
	Sector   abi.SectorID
	TaskType sealtasks.TaskType
	Priority int
}

type SchedDiagInfo struct {
	Requests    []SchedDiagRequestInfo
	OpenWindows []string
}

type SchedParse struct {
	SchedInfo    SchedDiagInfo
	ReturnedWork []string
	Waiting      []string
	CallToWork   map[string]string
	EarlyRet     []string
}

func GetMinerID(ctx context.Context, mi lotusapi.StorageMinerStruct) (id string, err error) {
	minerId, err := mi.ActorAddress(ctx)
	if err != nil {
		log.Fatalf("get miner id: %s", err)
		return "", err
	}

	return minerId.String(), nil
}

func GetMinerVersion(ctx context.Context, mi lotusapi.StorageMinerStruct) (miVer string, err error) {
	miVersion, err := mi.Version(ctx)
	if err != nil {
		log.Fatalf("get miner version: %s", err)
		return "", err
	}

	return miVersion.String(), nil
}

func GetMinerInfo(ctx context.Context, fu lotusapi.FullNodeStruct, minerId string, chainTipSetKey *types.TipSet) (info MinerInfoStruct, err error) {
	addr, err := address.NewFromString(minerId)
	if err != nil {
		log.Fatalf("convert miner id err: %s", err)
		return MinerInfoStruct{}, err
	}

	minerStats, err := fu.StateMinerInfo(ctx, addr, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get miner stats: %s", err)
		return MinerInfoStruct{}, err
	}

	owner := minerStats.Owner
	worker := minerStats.Worker
	control0 := minerStats.ControlAddresses[0]
	sectorSize := minerStats.SectorSize.String()

	sectorSizeNum, err := strconv.ParseUint(sectorSize, 10, 64)
	if err != nil {
		log.Fatalf("Conversion sector size error: %s", err)
	}

	ownerAddr, err := fu.StateAccountKey(ctx, owner, chainTipSetKey.Key())
	if err != nil {
		log.Printf("get miner own address: %s", err)
		ownerAddr = owner
	}

	workerAddr, err := fu.StateAccountKey(ctx, worker, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get miner worker address: %s", err)
		return MinerInfoStruct{}, err
	}

	control0Addr, err := fu.StateAccountKey(ctx, control0, chainTipSetKey.Key())
	if err != nil {
		log.Printf("get miner control0 address: %s", err)
	}

	return MinerInfoStruct{
		Owner:        owner.String(),
		OwnerAddr:    ownerAddr.String(),
		Worker:       worker.String(),
		WorkerAddr:   workerAddr.String(),
		Control0:     control0.String(),
		Control0Addr: control0Addr.String(),
		SectorSize:   sectorSizeNum,
	}, nil
}

func GetLockedFunds(ctx context.Context, fu lotusapi.FullNodeStruct, minerId string, chainTipSetKey *types.TipSet) (linfo []LockedInfoStruct) {
	addr, err := address.NewFromString(minerId)
	if err != nil {
		log.Fatalf("convert miner id err: %s", err)
		return []LockedInfoStruct{}
	}

	lockedFunds, err := fu.StateReadState(ctx, addr, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get miner actor stats: %s", err)
		return []LockedInfoStruct{}
	}

	data1, err := json.Marshal(lockedFunds.State)
	if err != nil {
		log.Fatalf("convert interface to json: %s", err)
		return []LockedInfoStruct{}
	}

	m1 := make(map[string]interface{})
	err = json.Unmarshal(data1, &m1)
	if err != nil {
		log.Fatalf("convert json to map: %s", err)
		return []LockedInfoStruct{}
	}

	var lockedInfoG []LockedInfoStruct

	for _, i := range []string{"PreCommitDeposits", "LockedFunds", "FeeDebt", "InitialPledge"} {
		value := Strval(m1[i])
		valueInt, err := types.BigFromString(value)
		if err != nil {
			log.Fatalf("convert string to uint64: %s", err)
			return []LockedInfoStruct{}
		}

		balanceFl, err := strconv.ParseFloat(types.FIL(valueInt).Unitless(), 64)
		if err != nil {
			log.Fatalf("convert blance float err: %s", err)
			return []LockedInfoStruct{}
		}
		lockedInfoG = append(lockedInfoG, LockedInfoStruct{
			LockedType: i,
			Balance:    balanceFl,
		})
	}

	return lockedInfoG
}

func GetWorkerInfo(ctx context.Context, mi lotusapi.StorageMinerStruct) (workers []WorkerInfoStuct) {
	workerStats, err := mi.WorkerStats(ctx)
	if err != nil {
		log.Fatalf("get miner actor address: %s", err)
	}

	var WorkerGroups []WorkerInfoStuct
	for _, worker := range workerStats {
		workerHost := worker.Info.Hostname
		cpus := worker.Info.Resources.CPUs
		gpus := len(worker.Info.Resources.GPUs)

		ramTotal := worker.Info.Resources.MemPhysical
		ramTasks := worker.MemUsedMin
		ramUsed := worker.Info.Resources.MemUsed
		ramReserved := 0
		if ramUsed > ramTasks {
			ramReserved = int(ramUsed) - int(ramTasks)
		}

		vmemTotal := ramTotal + worker.Info.Resources.MemSwap
		vmemTasks := worker.MemUsedMax
		vmemUsed := ramUsed + worker.Info.Resources.MemSwapUsed
		vmemReserved := 0
		if vmemUsed > vmemTasks {
			vmemReserved = int(vmemUsed) - int(vmemTasks)
		}

		var gpuUsed int
		if worker.GpuUsed > 0 {
			gpuUsed = 1
		} else {
			gpuUsed = 0
		}
		cpuUsed := worker.CpuUse

		WorkerGroups = append(WorkerGroups, WorkerInfoStuct{
			workerHost,
			cpus,
			gpus,
			ramTotal,
			ramReserved,
			ramTasks,
			vmemTotal,
			vmemReserved,
			vmemTasks,
			cpuUsed,
			gpuUsed,
		})
	}

	return WorkerGroups
}

func GetWorkerJobs(ctx context.Context, mi lotusapi.StorageMinerStruct) (jobInfo []JobInfoStruct) {
	workerJobs, err := mi.WorkerJobs(ctx)
	if err != nil {
		log.Fatalf("get miner worker jobs: %s", err)
	}

	var reJobInfo []JobInfoStruct
	for _, jobList := range workerJobs {
		for _, job := range jobList {
			jobId := job.ID.String()
			sector := job.Sector.Number.String()
			workerHost := job.Hostname
			task := string(job.Task)
			jobStartTime := job.Start.Format("2006-01-02T15:04:05Z07:00")
			runWait := strconv.Itoa(job.RunWait)
			jobStartEpoch := int(time.Now().Unix() - job.Start.Unix())
			reJobInfo = append(reJobInfo, JobInfoStruct{
				jobId,
				sector,
				workerHost,
				task,
				jobStartTime,
				runWait,
				jobStartEpoch,
			})
		}
	}
	return reJobInfo
}

func GetSchedDiag(ctx context.Context, mi lotusapi.StorageMinerStruct) []SchedInfoStruct {
	schedDiag, err := mi.SealingSchedDiag(ctx, true)
	if err != nil {
		log.Fatalf("get miner sched diag: %s", err)
	}
	log.Printf("schedDiag: %s", schedDiag)

	data1, err := json.Marshal(schedDiag)
	if err != nil {
		log.Fatalf("convert interface to json: %s", err)
	}

	var schedParseInfo SchedParse
	err = json.Unmarshal(data1, &schedParseInfo)
	if err != nil {
		log.Fatalf("convert json to map: %s", err)
	}

	var reSchedInfo []SchedInfoStruct
	for _, req := range schedParseInfo.SchedInfo.Requests {
		sector := req.Sector.Number.String()
		task := string(req.TaskType)
		reSchedInfo = append(reSchedInfo, SchedInfoStruct{
			sector,
			task,
		})
	}

	return reSchedInfo
}
