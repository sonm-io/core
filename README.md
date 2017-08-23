<p align="center">
  <img alt="SONM-IO CORE Logo" src="https://wiki.sonm.io/lib/exe/fetch.php?w=300&tok=2448ef&media=sonm-logo_no-text.png" height="140" />
  <h3 align="center">SONM Core</h3>
  <p align="center">Official core client</p>
</p>

---

[![Build Status](https://travis-ci.org/sonm-io/core.svg?branch=master)](https://travis-ci.org/sonm-io/core)
[![Build status](https://ci.appveyor.com/api/projects/status/01d7cpccwi8scwqp/branch/master?svg=true)](https://ci.appveyor.com/project/Sokel/core/branch/master)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/sonm-io_core/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonm-io/core)](https://goreportcard.com/report/github.com/sonm-io/core)


### How to build

If you want to build all the components by yourself:

```
make build
```

You can build each component separately:
```
make build/hub
make build/miner
make build/cli
```

### How to run

Download binaries for your platform on [release](https://github.com/sonm-io/core/releases) page. Package includes Hub, Miner and cli, that’s what you need to start working with SONM Network.

#### Hub

As first, you need to review `hub.yaml` config file. There are two configured endpoints: the first one allows miners to connect to your hub, and the second one allows you to connect to the hub using cli and perform management operations;

Start the Hub on the machine with publicly-accessible IP address:
```
./sonmhub
```

Hub must show configured endpoint to connect to:
```
2017-08-23T10:37:41.254Z	INFO	hub/server.go:288	listening for connections from Miners	{"address": "[::]:10002"}
2017-08-23T10:37:41.254Z	INFO	hub/server.go:297	listening for gRPC API connections	{"address": "[::]:10001"}
```


#### Miner:

Similar to the Hub, the Miner has it own config file `miner.yaml`.

If the Hub has no publicly-available IP address or if you start all the SONM components into a local network, you must explicitly set Hub's IP via endpoint parameter into the config file.

Ensure that docker daemon is started.

Miner must be started with escalated privileges because it needs to perform operations with docker socket, cgroups, etc

In the new terminal window start the Miner:
```
sudo ./sonmminer
```

Miner must show successful handshake with the Hub.
```2017-08-23T10:37:59.967Z	INFO	miner/server.go:156	handling Handshake request	{"req": ""}
2017-08-23T10:37:59.971Z	INFO	miner/server.go:415	starting tasks status server
2017-08-23T10:37:59.971Z	DEBUG	miner/server.go:368	handling tasks status request
2017-08-23T10:37:59.972Z	INFO	miner/server.go:343	sending result	{"info": {}, "statuses": {}}
2017-08-23T10:38:04.967Z	INFO	miner/server.go:512	yamux.Ping OK	{"rtt": "211.715µs"}
2017-08-23T10:38:09.968Z	INFO	miner/server.go:512	yamux.Ping OK	{"rtt": "253.23µs"}
```


#### CLI

Use cli to connect to the Hub and check that it is available. You need to specify Hub IP address and port. In the new terminal run the following command:
```
./sonmcli --addr 127.0.0.1:10001 hub status
```

You must see the number of connected miners and Hub's uptime:
```
Connected miners: 1
Uptime:           51s
```

Discover available Miners to start task:
```
./sonmcli --addr 127.0.0.1:10001 miner list
```

There is a list of connected miners:
```
Miner: 127.0.0.1:42572	      	Idle
```

Use miner's address as id to retrieve detailed statistic:
```
./sonmcli --addr 127.0.0.1:10001 miner status 127.0.0.1:42572
```

There are some details about Miner's hardware resources and running tasks:
```
Miner: "127.0.0.1:42572" (b8ca1ffc-ff8c-46a7-9ec1-2a2e19cf0636):
  Hardware:
    CPU0: 1 x Intel(R) Core(TM) i5-5257U CPU @ 2.70GHz
    CPU1: 1 x Intel(R) Core(TM) i5-5257U CPU @ 2.70GHz
    GPU: None
    RAM:
      Total: 992.2 MB
      Used:  274.5 MB
  No active tasks
```

As you can see, there is no running tasks. Let's run something!

To start a task you need to write the description file. For demo purposes we will use minimal allowed task file, full version with comments can be found [there](https://github.com/sonm-io/core/blob/release/0.2.1.1/task.yaml).

You need to specify docker image name and resource requirements, then save task definition into file named `task.yaml`
```yaml
task:
  container:
    # image name to start, requried
    name: httpd:latest
  resources:
      # number of CPU cores required by task, required param
      CPU: 1
      # amount of memory required by task, required param
      # You may use Kb, Mb and Gb suffixes
      RAM: 20Mb
```


Then just pass task file and miner address as arguments to `task start` command:
```
./sonmcli --addr 127.0.0.1:10001 task start 127.0.0.1:42572 task.yaml
```

An output shows ID assigned to task and exposed container ports:
```
Starting "httpd:latest" on miner 127.0.0.1:42572...
ID eb081f7d-bd1c-4d9b-976f-7674a293b199
Endpoint [80/tcp->1.2.3.4:32768]
```

Now you can connect to the container using given IP:port pair and receive a response from container running in SONM network:
```
> curl -XGET http://1.2.3.4:32768
<html><body><h1>It works!</h1></body></html>
```

Now you can get some info about running tasks by its ID and Miner address:
```
./sonmcli --addr 127.0.0.1:10001 task status 127.0.0.1:42572 eb081f7d-bd1c-4d9b-976f-7674a293b199
```

Detailed task statistics shows container status and resources consumption:
```
Task eb081f7d-bd1c-4d9b-976f-7674a293b199 (on 127.0.0.1:42572):
  Image:  httpd:latest
  Status: RUNNING
  Uptime: 1m5.049056944s
  Resources:
    CPU: 153670361
    MEM: 11.8 MB
    NET:
      eth0:
        Tx/Rx bytes: 584/1821
        Tx/Rx packets: 5/23
        Tx/Rx errors: 0/0
        Tx/Rx dropped: 0/0
  Ports:
    80/tcp: 0.0.0.0:32768
```
