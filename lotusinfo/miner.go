package lotusinfo

import (
	"context"
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

func GetMinerID(ctx context.Context, mi lotusapi.StorageMinerStruct) (id string, err error) {
	minerId, err := mi.ActorAddress(ctx)
	if err != nil {
		log.Fatalf("get actor address: %s", err)
		return "", err
	}
	// fmt.Printf("miner actor address: %s\n", address.String())
	return minerId.String(), nil
}

func GetMinerVersion(ctx context.Context, mi lotusapi.StorageMinerStruct) (miVer string, err error) {
	miVersion, err := mi.Version(ctx)
	if err != nil {
		log.Fatalf("get actor version: %s", err)
		return "", err
	}
	// fmt.Printf("miner actor address: %s\n", address.String())
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

	// fmt.Printf("miner actor address: %s\n", address.String())
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
