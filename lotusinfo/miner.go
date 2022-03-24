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

func Strval(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}
