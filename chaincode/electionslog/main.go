package main

import "github.com/hyperledger/fabric/core/chaincode/shim"

func main() {
	err := shim.Start(new(BallotLog))
	if err != nil {
		panic(err)
	}
}
