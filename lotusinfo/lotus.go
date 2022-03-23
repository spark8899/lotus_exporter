package lotusinfo

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/multiformats/go-multibase"
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

type MpoolMsg struct {
	Mfrom       string
	Mto         string
	Mnonce      uint64
	Mvalue      int64
	Mgaslimit   int64
	Mgasfeecap  int64
	Mgaspremium int64
	Mmethod     int64
	Mmethodtype string
	Mactortype  string
}

type WalletInfo struct {
	Name    string
	Address string
	Balance uint64
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

func GetMpoolInfo(ctx context.Context, fu lotusapi.FullNodeStruct, chainTipSetKey *types.TipSet, addrList []string) (mpoolTotal int, localMpollTotal int, messageList []MpoolMsg) {
	mpoolPending, err := fu.MpoolPending(ctx, chainTipSetKey.Key())
	if err != nil {
		log.Fatalf("get mpool pending err: %s", err)
	}

	filter := map[address.Address]struct{}{}
	for _, a := range addrList {
		b, err := address.NewFromString(a)
		if err != nil {
			log.Fatalf("get addr err: %s", err)
		}
		filter[b] = struct{}{}
	}

	var msgList []MpoolMsg
	for _, msg := range mpoolPending {
		if filter != nil {
			if _, has := filter[msg.Message.From]; !has {
				continue
			}
		}

		actor, err01 := fu.StateGetActor(ctx, msg.Message.To, chainTipSetKey.Key())
		if err01 != nil {
			log.Fatalf("get actor err: %s", err01)
		}

		_, actorType, err01 := multibase.Decode(actor.Code.String())
		if err01 != nil {
			log.Fatalf("get actor type err: %s", err01)
		}

		messageType := MethodMessageType[string(actorType[10:])][msg.Message.Method]

		msgList = append(msgList, MpoolMsg{
			msg.Message.From.String(),
			msg.Message.To.String(),
			msg.Message.Nonce,
			msg.Message.Value.Int64(),
			msg.Message.GasLimit,
			msg.Message.GasFeeCap.Int64(),
			msg.Message.GasPremium.Int64(),
			int64(msg.Message.Method),
			messageType,
			string(actorType[10:])})
		//	msg.Message.From
	}

	return len(mpoolPending), len(msgList), msgList
}

func GetWalletlist(ctx context.Context, fu lotusapi.FullNodeStruct) (mpoolTotal int) {
	walletList, err := fu.WalletList(ctx)
	if err != nil {
		log.Fatalf("get wallet list err: %s", err)
	}

	return len(walletList)
}

func GetWalletBalance(ctx context.Context, fu lotusapi.FullNodeStruct, addrStg string) uint64 {
	addr, err := address.NewFromString(addrStg)
	if err != nil {
		log.Fatalf("convert miner id err: %s", err)
	}

	addrBlance, err := fu.WalletBalance(ctx, addr)
	if err != nil {
		log.Fatalf("get Blance err: %s", err)
	}

	return addrBlance.Uint64() / 1000000000000000000
}
