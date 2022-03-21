package lotusinfo

import (
	"context"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"log"
	"time"
)

type DaemonInfo struct {
	Network string
	Version string
	Value   int64
}

func GetMinerID(ctx context.Context, mi lotusapi.StorageMinerStruct) (id string, err error) {
	address, err := mi.ActorAddress(ctx)
	if err != nil {
		log.Fatalf("get actor address: %s", err)
		return "", err
	}
	// fmt.Printf("miner actor address: %s\n", address.String())
	return address.String(), nil
}

func GetLocalTime() (localTime int64) {
	return time.Now().Unix()
}

func GetChainHead(ctx context.Context, fu lotusapi.FullNodeStruct) (tipSet *types.TipSet, err error) {
	chainHead, err := fu.ChainHead(ctx)
	if err != nil {
		log.Fatalf("Get chain head: %s", err)
		return nil, err
	}

	return chainHead, nil
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

func GetchainBasefee(chainHead *types.TipSet) int64 {
	return chainHead.Blocks()[0].ParentBaseFee.Int64()
}

//func GetchainHeight(chainHead *types.TipSet) int64 {
//	height := chainHead.Height().String()
//	return chainHead.Height()
//}
