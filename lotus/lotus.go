package lotus

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"log"
	"net/http"
	"strings"
	"time"
)

type DaemonInfo struct {
	Network string
	Version string
	Value   int64
}

func GetMinerID(minerApiInfo string) (id string, err error) {
	minerApiInfoS := parseApiInfo(strings.TrimSpace(minerApiInfo))
	minerApiHeaders := http.Header{"Authorization": []string{"Bearer " + string(minerApiInfoS.token)}}

	var mi lotusapi.StorageMinerStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), minerApiInfoS.addr, "Filecoin", []interface{}{&mi.Internal, &mi.CommonStruct.Internal}, minerApiHeaders)
	if err != nil {
		log.Fatalf("connecting with lotus-miner failed: %s", err)
		return "", err
	}
	defer closer()
	// Now you can call any API you're interested in.
	address, err := mi.ActorAddress(context.Background())
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

func GetInfo(fullNodeApiInfo string) (daemonInfo DaemonInfo, err error) {
	fullNodeApiInfoS := parseApiInfo(strings.TrimSpace(fullNodeApiInfo))
	fullNodeApiHeaders := http.Header{"Authorization": []string{"Bearer " + string(fullNodeApiInfoS.token)}}

	var api lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), fullNodeApiInfoS.addr, "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, fullNodeApiHeaders)
	if err != nil {
		log.Fatalf("connecting with lotus failed: %s", err)
	}
	defer closer()

	// get daemon_network.
	daemonNetwork, err := api.StateNetworkName(context.Background())
	if err != nil {
		log.Fatalf("calling daemonNetwork: %s", err)
	}

	// Now you can call any API you're interested in.
	tipset, err01 := api.ChainHead(context.Background())
	if err01 != nil {
		log.Fatalf("calling chain head: %s", err01)
	}
	fmt.Printf("Current chain head is: %s\n", tipset.String())

	// get daemon_network_version.
	// get networker_version
	daemonNetworkVersion, err01 := api.StateNetworkVersion(context.Background(), tipset.Key())
	if err01 != nil {
		log.Fatalf("calling networker version: %s", err01)
	}

	// get daemon_version.
	daemonVersion, err := api.Version(context.Background())
	if err != nil {
		log.Fatalf("calling daemonVersion: %s", err)
	}

	daemonInfo = DaemonInfo{
		Network: string(daemonNetwork),
		Version: daemonVersion.String(),
		Value:   int64(daemonNetworkVersion),
	}
	return daemonInfo, nil
}
