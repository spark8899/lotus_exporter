package lotusinfo

import (
	"context"
	"encoding/json"
	"github.com/filecoin-project/go-address"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"log"
	"strconv"
)

type MinerInfo struct {
	Owner        string
	OwnerAddr    string
	Worker       string
	WorkerAddr   string
	Control0     string
	Control0Addr string
	SectorSize   uint64
}

type LockedInfo struct {
	LockedType string
	Balance    float64
}

type WorkerInfo struct {
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

func GetMinerID(ctx context.Context, mi lotusapi.StorageMinerStruct) (id string, err error) {
	minerId, err := mi.ActorAddress(ctx)
	if err != nil {
		log.Fatalf("get actor address: %s", err)
		return "", err
	}

	return minerId.String(), nil
}

func GetMinerVersion(ctx context.Context, mi lotusapi.StorageMinerStruct) (miVer string, err error) {
	miVersion, err := mi.Version(ctx)
	if err != nil {
		log.Fatalf("get actor version: %s", err)
		return "", err
	}

	return miVersion.String(), nil
}

func GetMinerInfo(ctx context.Context, fu lotusapi.FullNodeStruct, minerId string, chainTipSetKey *types.TipSet) (info MinerInfo, err error) {
	addr, err := address.NewFromString(minerId)
	if err != nil {
		log.Fatalf("convert miner id err: %s", err)
		return MinerInfo{}, err
	}

	minerStats, err := fu.StateMinerInfo(ctx, addr, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get actor stats: %s", err)
		return MinerInfo{}, err
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
		ownerAddr = owner
	}

	workerAddr, err := fu.StateAccountKey(ctx, worker, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get worker address: %s", err)
		return MinerInfo{}, err
	}

	control0Addr, err := fu.StateAccountKey(ctx, control0, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get control0 address: %s", err)
		return MinerInfo{}, err
	}

	return MinerInfo{
		Owner:        owner.String(),
		OwnerAddr:    ownerAddr.String(),
		Worker:       worker.String(),
		WorkerAddr:   workerAddr.String(),
		Control0:     control0.String(),
		Control0Addr: control0Addr.String(),
		SectorSize:   sectorSizeNum,
	}, nil
}

func GetLockedFunds(ctx context.Context, fu lotusapi.FullNodeStruct, minerId string, chainTipSetKey *types.TipSet) (linfo []LockedInfo) {
	addr, err := address.NewFromString(minerId)
	if err != nil {
		log.Fatalf("convert miner id err: %s", err)
		return []LockedInfo{}
	}

	lockedFunds, err := fu.StateReadState(ctx, addr, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get actor stats: %s", err)
		return []LockedInfo{}
	}

	data1, err := json.Marshal(lockedFunds.State)
	if err != nil {
		log.Fatalf("convert interface to json: %s", err)
		return []LockedInfo{}
	}

	m1 := make(map[string]interface{})
	err = json.Unmarshal(data1, &m1)
	if err != nil {
		log.Fatalf("convert json to map: %s", err)
		return []LockedInfo{}
	}

	var lockedInfoG []LockedInfo

	for _, i := range []string{"PreCommitDeposits", "LockedFunds", "FeeDebt", "InitialPledge"} {
		value := Strval(m1[i])
		valueInt, err := types.BigFromString(value)
		if err != nil {
			log.Fatalf("convert string to uint64: %s", err)
			return []LockedInfo{}
		}

		balanceFl, err := strconv.ParseFloat(types.FIL(valueInt).Unitless(), 64)
		if err != nil {
			log.Fatalf("convert blance float err: %s", err)
			return []LockedInfo{}
		}
		lockedInfoG = append(lockedInfoG, LockedInfo{
			LockedType: i,
			Balance:    balanceFl,
		})
	}

	return lockedInfoG
}

func GetWorkerInfo(ctx context.Context, mi lotusapi.StorageMinerStruct) (workers []WorkerInfo) {
	workerStats, err := mi.WorkerStats(ctx)
	if err != nil {
		log.Fatalf("get actor address: %s", err)
		// return "", err
	}

	var WorkerGroups []WorkerInfo
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

		WorkerGroups = append(WorkerGroups, WorkerInfo{
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
