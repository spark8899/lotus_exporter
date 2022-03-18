package lotus

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DaemonInfo struct {
	Network string
	Version string
	Value   uint64
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
	daemonNetwork, err := api.Version(context.Background())
	if err != nil {
		log.Fatalf("calling daemonNetwork: %s", err)
	}

	// get daemon_network_version.
	daemonNetworkVersion, err := api.Version(context.Background())
	if err != nil {
		log.Fatalf("calling daemonNetworkVersion: %s", err)
	}

	// get daemon_network_version.
	daemonVersion, err := api.Version(context.Background())
	if err != nil {
		log.Fatalf("calling daemonVersion: %s", err)
	}
	daemonVersionNum, _ := strconv.Atoi(daemonVersion.Version)

	daemonInfo = DaemonInfo{
		Network: daemonNetwork.String(),
		Version: daemonNetworkVersion.String(),
		Value:   uint64(daemonVersionNum),
	}
	return daemonInfo, nil
}
