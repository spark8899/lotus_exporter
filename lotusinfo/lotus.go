package lotusinfo

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"log"
	"strconv"
	"time"
)

type DaemonInfo struct {
	Network string
	Version string
	Value   int64
}

type ChainSyncState struct {
	CSWorkerID string
	CSDiff     int64
	CSStatus   int
}

func GetLocalTime() (localTime int64) {
	return time.Now().Unix()
}

func GetTipsetKey(ctx context.Context, fu lotusapi.FullNodeStruct) (tipSet *types.TipSet, err error) {
	GetTipsetKey, err := fu.ChainHead(ctx)
	if err != nil {
		log.Fatalf("Get chain head: %s", err)
		return nil, err
	}

	return GetTipsetKey, nil
}

func GetInfo(ctx context.Context, fu lotusapi.FullNodeStruct, chainHead *types.TipSet) (daemonInfo DaemonInfo, err error) {
	// get daemon_network.
	daemonNetwork, err := fu.StateNetworkName(ctx)
	if err != nil {
		log.Fatalf("calling daemonNetwork: %s", err)
	}

	// get daemon_network_version.
	daemonNetworkVersion, err01 := fu.StateNetworkVersion(ctx, chainHead.Key())
	if err01 != nil {
		log.Fatalf("calling networker version: %s", err01)
	}

	// get daemon_version.
	daemonVersion, err := fu.Version(ctx)
	if err != nil {
		log.Fatalf("calling daemonVersion: %s", err)
	}

	daemonInfo = DaemonInfo{
		Network: string(daemonNetwork),
		Version: daemonVersion.Version,
		Value:   int64(daemonNetworkVersion),
	}
	return daemonInfo, nil
}

func GetChainBasefee(chainTipSetKey *types.TipSet) int64 {
	return chainTipSetKey.Blocks()[0].ParentBaseFee.Int64()
}

func GetChainHeight(chainTipSetKey *types.TipSet) int64 {
	heightStr := chainTipSetKey.Height().String()
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		log.Fatalf("get chainHeight err: %s", err)
		return 0
	}
	return height
}

func GetChainSyncState(ctx context.Context, fu lotusapi.FullNodeStruct) []ChainSyncState {
	syncStat, err := fu.SyncState(ctx)
	if err != nil {
		// return err
	}
	var reSS []ChainSyncState

	for i, ss := range syncStat.ActiveSyncs {
		var heightDiff int64

		if int64(ss.Target.Height()) >= int64(ss.Base.Height()) {
			heightDiff = int64(ss.Target.Height()) - int64(ss.Base.Height())
		} else {
			heightDiff = -1
		}
		reSS = append(reSS, ChainSyncState{strconv.Itoa(i), heightDiff, int(ss.Stage)})
	}

	return reSS
}

func GetPowerList(ctx context.Context, fu lotusapi.FullNodeStruct, minerId string, chainTipSetKey *types.TipSet) (mpRW int64, mpQw int64, tpRw int64, tpQw int64) {
	addr, err := address.NewFromString(minerId)
	if err != nil {
		log.Fatalf("convert miner id err: %s", err)
	}

	power, err := fu.StateMinerPower(ctx, addr, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get miner power err: %s", err)
	}

	mp := power.MinerPower
	tp := power.TotalPower

	return mp.RawBytePower.Int64(), mp.QualityAdjPower.Int64(), tp.QualityAdjPower.Int64(), tp.RawBytePower.Int64()
}

func GetBaseInfo(ctx context.Context, fu lotusapi.FullNodeStruct, minerId string, chainHeight int64, chainTipSetKey *types.TipSet) (eligibility int) {
	addr, err := address.NewFromString(minerId)
	if err != nil {
		log.Fatalf("convert miner id err: %s", err)
	}

	baseInfo, err := fu.MinerGetBaseInfo(ctx, addr, abi.ChainEpoch(chainHeight), chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("convert miner id err: %s", err)
	}

	if baseInfo.EligibleForMining {
		return 1
	} else {
		return 0
	}
}

func GetMpoolTotal(ctx context.Context, fu lotusapi.FullNodeStruct, chainTipSetKey *types.TipSet) (mpoolTotal int) {
	mpoolPending, err := fu.MpoolPending(ctx, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get mpool pending err: %s", err)
	}

	//for _, msg := range mpoolPending {
	//	msg.Message.From
	//}

	return len(mpoolPending)
}

func GetLocalMpool(ctx context.Context, fu lotusapi.FullNodeStruct, chainTipSetKey *types.TipSet) (mpoolTotal int) {
	mpoolPending, err := fu.MpoolPending(ctx, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get mpool pending err: %s", err)
	}

	return len(mpoolPending)
}

func Getwalletlist(ctx context.Context, fu lotusapi.FullNodeStruct) (mpoolTotal int) {
	walletList, err := fu.WalletList(ctx)
	if err != nil {
		log.Fatalf("get mpool pending err: %s", err)
	}

	return len(walletList)
}
