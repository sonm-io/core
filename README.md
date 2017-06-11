# inSONMnia

It's an alpha platform for SONM.io project.

# What is it here?

This repository contains code for Hub, Miner and CLI.

# Hub

Hub provides public gRPC-based API. proto files can be found in proto dir.

# Miner

Miner is expected to discover a Hub using Whisper. Later the miner connects to the hub via TCP. Right now a Hub must have a *public IP address*. Hub sends orders to the miner via gRPC on top of the connection. Hub pings the miner from time to time.

# Roadmap

Look at milestone https://github.com/sonm-io/insonmnia/milestones

# How to run

## Hub

```bash
docker run --rm -p 10002:10002 -p 10001:10001  sonm/insonmnia sonmhub
```

## Miner

To run Miner from the container you need to pass docker socket inside and specify IP of the HUB

```bash
docker run --rm -v /var/run/:/var/run/  sonm/insonmnia sonmminer --hubaddress=<yourip>:10002
```

## CLI
