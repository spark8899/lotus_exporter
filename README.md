# lotus_exporter
monitor lotus daemon and miner

To run it:

    go build -v -ldflags "-X main.version=`cat VERSION` -X main.branch=`git rev-parse --abbrev-ref HEAD` -X main.revision=`git rev-parse --short HEAD`"
    ./lotus_exporter [flags]

## Exported Metrics
| Metric                       | Description                                      | Labels  |
|------------------------------|--------------------------------------------------|---------|
| up                           | Was the last lotus_exporter CLI query successful |         |
| lotus_chain_basefee          | return current basefee                           | lotus   |
| lotus_chain_height           | return current height                            | lotus   |
| lotus_local_time             | time on the node machine when last execution start in epoch         | lotus   |
| lotus_info                   |  lotus daemon information like address version, value is set to network version number              | lotus   |

## Flags
    ./lotus_exporter --help

| Flag                | Description | Default |
|---------------------| ----------- | ------- |
| -config-path        | Path to environment file | `/etc/lotus_exporter/.env` |
| -web.listen-address | Address to listen on for telemetry | `:9141` |
| -web.telemetry-path | Path under which to expose metrics | `/metrics` |

## Env Variables

Use a .env file in the local folder, /etc/lotus_exporter/.env, or
use the -config.file-path command line flag to provide a path to your
environment file
```
MINER_API_INFO=xxxx-xxx-xx:/ip4/xxx.xx.xx.xx/tcp/2345/http
FULLNODE_API_INFO=xxxx-xxx-xx:/ip4/xxx.xx.xx.xx/tcp/1234/http
```