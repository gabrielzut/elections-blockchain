package main

import "github.com/hyperledger/fabric/core/chaincode/shim"

func main() {
	err := shim.Start(new(Ballot))
	if err != nil {
		panic(err)
	}
}
