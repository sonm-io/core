## SONM Core

Official core client

[![Build Status](https://travis-ci.org/sonm-io/core.svg?branch=master)](https://travis-ci.org/sonm-io/core)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/sonm-io_core/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonm-io/core)](https://goreportcard.com/report/github.com/sonm-io/core)



How to build Ethereum contracts
------------------------

Prerequisites:

  - Install [truffle](https://github.com/trufflesuite/truffle):
      `yarn global add truffle` or `npm install -g truffle`
  - Install [testrpc](https://github.com/ethereumjs/testrpc):
      `yarn global add ethereumjs-testrpc` or `npm install -g ethereumjs-testrpc`
  - Install [OpenZeppelin](https://github.com/OpenZeppelin/zeppelin-solidity) contracts:
      `yarn install` or `npm install`

Usage:

  - `truffle compile` to build contracts
  - `testrpc` & `truffle console` to play with contracts
