NVIDIA-docker v1
================

This repository contains source code of the nvidia-docker v1 plugin.
Forked from [oficial repo](https://github.com/NVIDIA/nvidia-docker). 

Build dependencies are included into  the repo. Use the `CGO_` flags to build code that use this library:

```
CGO_LDFLAGS=-L/path/to/nvidia-docker/build/lib/ \
CGO_CFLAGS=-I/path/to/nvidia-docker/build/include/ \
go build ...
```
