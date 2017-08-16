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

### Running in docker:

Start boootnode
```
docker run -it --rm --name bootnode -p 8092:8092  sonm/bootnode
```


Start hub, open port 10001 to connect with cli
```
docker run -it --rm --name hub -p 10001:10001 sonm/hub
```


Start miner, link it to hub, mount `/var/run` to access docker.sock
```
docker run -it --rm --link hub -v /var/run:/var/run sonm/miner:latest /sonmminer
```


Check all components using cli:

```
./sonmcli --hub 127.0.0.1:10001 ping
```


### How to build

If you want to build all the components by yourself:

```
make build
```

build docker containers locally:

```
make docker_all
```
