# lotus_exporter
monitor lotus daemon and miner

To run it:

    go build
    ./lotus_exporter [flags]

## Exported Metrics
| Metric | Description | Labels |
| ------ | ------- | ------ |
| mirth_up | Was the last Mirth CLI query successful | |
| mirth_messages_received_total | How many messages have been received | channel |
| mirth_messages_filtered_total  | How many messages have been filtered | channel |
| mirth_messages_queued | How many messages are currently queued | channel |
| mirth_messages_sent_total  | How many messages have been sent | channel |
| mirth_messages_errored_total  | How many messages have errored | channel |

## Flags
    ./lotus_exporter --help

| Flag | Description | Default |
| ---- | ----------- | ------- |
| log.level | Logging level | `info` |
| web.listen-address | Address to listen on for telemetry | `:9141` |
| web.telemetry-path | Path under which to expose metrics | `/metrics` |
| config.file-path | Optional environment file path | `None` |

## Env Variables

Use a .env file in the local folder, /etc/lotus_exporter/.env, or
use the --config.file-path command line flag to provide a path to your
environment file
```
MINER_API_INFO=xxxx-xxx-xx:/ip4/xxx.xx.xx.xx/tcp/2345/http
FULLNODE_API_INFO=xxxx-xxx-xx:/ip4/xxx.xx.xx.xx/tcp/1234/http
```

## Notice

This exporter is inspired by the [consul_exporter](https://github.com/prometheus/consul_exporter)
and has some common code. Any new code here is Copyright &copy; 2020 TeamZero, Inc. See the included
LICENSE file for terms and conditions.

## 引用
* https://github.com/teamzerolabs/mirth_channel_exporter
* https://github.com/mnadeem/volume_exporter