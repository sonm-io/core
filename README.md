## SONM Core

Official core client

[![Build Status](https://travis-ci.org/sonm-io/core.svg?branch=master)](https://travis-ci.org/sonm-io/core)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/sonm-io_core/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonm-io/core)](https://goreportcard.com/report/github.com/sonm-io/core)


### Running in docker:

Start boootnode
```
d run -it --rm --name bootnode -p 8092:8092  sonm/bootnode
```


Start hub, open port 10001 to connect with cli
```
d run -it --rm --name hub --link  bootnode -p 10001:10001 sonm/hub
```


Start miner, link it to hub, mount `/var/run` to access docker.sock
```
d run -it --rm --link bootnode --link hub -v /var/run:/var/run sonm/miner:latest /sonmminer
```


Check all components using cli:

```
./sonmcli --hub 127.0.0.1:10001 ping
```
