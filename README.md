## SONM Core

Official core client

[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/sonm-io_core/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link)

# What is it here?

This repository contains code for Hub, Miner and CLI.

# Where can I get it?

A docker container contained every CLI, Miner, Hub can be found on public DockerHub: [sonm/insonmnia](https://hub.docker.com/r/sonm/insonmnia/)

```bash
docker pull sonm/insonmnia
```

If you want it's easy to build all the components. You need *golang > 1.8*:

```bash
make build
```

Also there is a Dockerfile to build a container:

```bash
docker build .
```

# How to run

## Hub

To start a hub it's needed to expose a couple of ports.
*10001* handles gRCP requests from CLI
*10002* is used to handle communication with miners

```bash
docker run --rm -p 10002:10002 -p 10001:10001  sonm/insonmnia sonmhub
```

## Miner

To run Miner from the container you need to pass docker socket inside and specify IP of the HUB

```bash
  docker run --net host -e DOCKER_API_VERSION=1.24 -v /run:/var/run sonm/insonmnia:alpha3 sonmminer -h <hubip:10002>
```

## CLI commands

CLI sends commands to a hub. A hub must be pointed via *--hub=<hubip:port>*. Port is usually *10001*.

### ping

Just check that a hub is reachable and alive.

```bash
sonmcli --hub <hubip:10001> ping
OK
```

### list

List shows a list of miners connected to a hub with tasks assigned to them.

**NOTE: later each miner will have a unique signed ID instead of host:port**

```bash
sonmcli --hub <hubip:port> list
Connected Miners
{
	"<minerip:port": {
		"values": [
			"2b845fcc-143a-400b-92c7-aac2867ab62f",
			"412dd411-96df-442a-a397-6a2eba9147f9"
		]
	}
}
```

### start a container

To start a container you have to pick a hub and miner connected to that hub.
You can pick a miner from output of List command. See above.

```bash
./sonmcli --hub <hubip:port> --timeout=3000s  start --image schturmfogel/sonm-q3:alpha  --miner=<minerhost:port>
```
The result would look like:
```
ID <jobid>, Endpoint [27960/tcp-><ip:port> 27960/udp-><ip:port>]
```
 + **jobid** is an unique name for the task. Later it can be used to specify a task for various operations.
 + **Endpoint** describes mapping of exposed ports (google for Docker EXPOSE) to the real ports of a miner

**NOTE**: later STUN will be used for UDP packets and LVS (ipvs) or userspace proxy (like SSH tunnel) for TCP. Miners who have a public IPv4 or can be reached via IPv6 would not need this proxy. The proxy is intended to get through NAT.

### stop a container

To stop the task just provide the *jobid*

```bash
sonmcli --hub <hubip:port> stop <jobid>
```

# How to cook a container

Dockerfile for the image should follow several requirements:
 + *ENTRYPOINT* or *CMD* or both must present
 + Network ports should be specified via *EXPOSE*
