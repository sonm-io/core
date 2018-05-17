// +build ignore
// This program generates wrappers for contracts.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

const (
	// `truffle compile` stores build artifacts here
	solidityArtifactsPath = "./build/contracts/*.json"
	// store generated contract wrappers here
	wrappersPath = "api"
	// generated wrappers belong to this package
	wrappersPackage = "api"
)

type SolidityArtifact struct {
	Name string        `json:"contractName"`
	ABI  []interface{} `json:"abi"`
	Bin  string        `json:"bytecode"`
}

func dieSoon(e error, msg string) {
	if e != nil {
		fmt.Printf(msg+": %v\n", e)
		os.Exit(-1)
	}
}

func main() {
	files, _ := filepath.Glob(solidityArtifactsPath)
	if files == nil {
		dieSoon(
			errors.New("No contract artifacts found (maybe run `truffle compile`)"),
			solidityArtifactsPath)
	}

	os.MkdirAll(wrappersPath, os.ModePerm)

	for _, jsonPath := range files {
		if "build/contracts/IterableMapping.json" == jsonPath {
			continue
		}
		jsonData, err := ioutil.ReadFile(jsonPath)
		dieSoon(err, "Failed to read file "+jsonPath)

		var ctr = SolidityArtifact{}
		err = json.Unmarshal(jsonData, &ctr)
		dieSoon(err, "Failed to parse json "+jsonPath)

		abi, _ := json.Marshal(ctr.ABI)

		goCode, err := bind.Bind(
			[]string{ctr.Name},
			[]string{string(abi)},
			[]string{ctr.Bin},
			wrappersPackage,
			bind.LangGo)
		dieSoon(err, "Failed to generate ABI binding")

		var resPath = path.Join(wrappersPath, ctr.Name+".go")
		err = ioutil.WriteFile(resPath, []byte(goCode), 0600)
		dieSoon(err, "Failed to write ABI binding")
	}
}
