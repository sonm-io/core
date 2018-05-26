# Retag [![TravisCI](https://api.travis-ci.org/sevlyar/retag.svg)](https://travis-ci.org/sevlyar/retag) [![GoDoc](https://godoc.org/github.com/sevlyar/retag?status.svg)](https://godoc.org/github.com/sevlyar/retag) [![Go Report Card](https://goreportcard.com/badge/github.com/sevlyar/retag)](https://goreportcard.com/report/github.com/sevlyar/retag) [![codecov](https://codecov.io/gh/sevlyar/retag/branch/master/graph/badge.svg)](https://codecov.io/gh/sevlyar/retag)

Package retag provides an ability to change tags of structures' fields in runtime
without copying of the data. It may be helpful in next cases:

* Automatic tags generation;
* Different views of the one data;
* Fixing of leaky abstractions with minimal boilerplate code
when application has layers of abstractions and model is
separated from storages and presentation layers.

Please see [examples in documentation](https://godoc.org/github.com/sevlyar/retag#example-package--Snaker) for details.

Features:

* No memory allocations (for cached types);
* Fast converting (lookup in table and pointer creation for cached types);
* Works with complex and nested types (e.g. `map[struct]*struct`).

The package requires go1.7+.

## Installation

    go get github.com/sevlyar/retag

You can use [gopkg.in](http://labix.org/gopkg.in):

    go get gopkg.in/sevlyar/retag.v0

## Documentation

Please see [godoc.org/github.com/sevlyar/retag](https://godoc.org/github.com/sevlyar/retag)
