# inSONMnia

It's an alpha platform for SONM.io project.

# What is it here?

This repository contains code for Hub, Miner and CLI. 

# Hub

Hub provides public gRPC-based API. proto files can be found in proto dir.

# Miner

Miner is expected to discover a Hub using Whisper. Later the miner connects to the hub via TCP. Right now a Hub must have a *public IP address*. Hub sends orders to the miner via gRPC on top of the connection. Hub pings the miner from time to time.

# Roadmap

