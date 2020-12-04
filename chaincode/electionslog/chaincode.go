package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type BallotLog struct {
	VoterId         string    `json:"voterId"`
	DateTime        time.Time `json:"dateTime"`
	ConfirmationKey string    `json:"confirmationKey"`
}

func (bl *BallotLog) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (bl *BallotLog) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()

	switch fcn {
	case "initElection":
		return bl.InitElection(stub)
	case "endElection":
		return bl.EndElection(stub)
	case "register":
		return bl.RegisterLog(stub, params)
	case "getByVoterId":
		return bl.GetByVoterId(stub, params)
	case "getByRange":
		return bl.GetByRange(stub, params)
	}

	return shim.Error("Error: invalid function")
}

func (bl *BallotLog) InitElection(stub shim.ChaincodeStubInterface) sc.Response {
	value, err := stub.GetState("InitElection")

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if value != nil {
		return shim.Error("Error: election has already been started.")
	}

	err = stub.PutState("InitElection", []byte{0x00})

	if err != nil {
		return shim.Error("Error putting state: " + err.Error())
	}

	return shim.Success(nil)
}

func (bl *BallotLog) EndElection(stub shim.ChaincodeStubInterface) sc.Response {
	value, err := stub.GetState("InitElection")

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if value == nil {
		return shim.Error("Error: election hasn't been started yet.")
	}

	value, err = stub.GetState("EndElection")

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if value != nil {
		return shim.Error("Error: election has already been ended.")
	}

	err = stub.PutState("EndElection", []byte{0x00})

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (bl *BallotLog) RegisterLog(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Wrong number of arguments! Expected 3 but got " + strconv.Itoa(len(args)))
	}

	ballotLog, err := ReadAndValidateArgs(args)

	if err != nil {
		return shim.Error("Error validating data: " + err.Error())
	}

	ballotBytes, err := json.Marshal(ballotLog)

	if err != nil {
		return shim.Error("Error serializing data: " + err.Error())
	}

	status, err := CheckElectionStatus(stub)

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if status != "STARTED" {
		return shim.Error("Error: election on status " + status)
	}

	ballotLogBytes, err := stub.GetState(ballotLog.VoterId)

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if ballotLogBytes != nil {
		return shim.Error("Error: voter has already voted")
	}

	err = stub.PutState(ballotLog.VoterId, ballotBytes)

	if err != nil {
		return shim.Error("Error putting ballot log on state: " + err.Error())
	}

	return shim.Success([]byte(ballotLog.VoterId))
}

func (bl *BallotLog) GetByVoterId(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Wrong number of arguments! Expected 1 but got " + strconv.Itoa(len(args)))
	}

	ballotLogBytes, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if ballotLogBytes == nil {
		return shim.Error("Error getting state: ballot log not found.")
	}

	return shim.Success(ballotLogBytes)
}

func (bl *BallotLog) GetByRange(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Wrong number of arguments! Expected 2 but got " + strconv.Itoa(len(args)))
	}

	iterator, err := stub.GetStateByRange(args[0], args[1])

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	}

	var ballots []BallotLog

	for iterator.HasNext() {
		ballotLogEntry, err := iterator.Next()

		if err != nil {
			return shim.Error("Error getting state: " + err.Error())
		}

		if ballotLogEntry.Key != "InitElection" && ballotLogEntry.Key != "EndElection" {
			ballotLogBytes := ballotLogEntry.Value

			ballotLog := BallotLog{}

			err = json.Unmarshal(ballotLogBytes, &ballotLog)

			if err != nil {
				return shim.Error("Error getting state: " + err.Error())
			}

			ballots = append(ballots, ballotLog)
		}
	}

	ballotLogsAsBytes, err := json.Marshal(ballots)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ballotLogsAsBytes)
}

// Function to get the election status. Status can be NOT_STARTED, STARTED or ENDED (or ERROR)
func CheckElectionStatus(stub shim.ChaincodeStubInterface) (string, error) {
	value, err := stub.GetState("InitElection")

	if err != nil {
		return "ERROR", err
	} else if value == nil {
		return "NOT_STARTED", nil
	}

	value, err = stub.GetState("EndElection")

	if err != nil {
		return "ERROR", err
	} else if value != nil {
		return "ENDED", nil
	}

	return "STARTED", nil
}

func ReadAndValidateArgs(args []string) (*BallotLog, error) {
	var ballotLog BallotLog
	var err error

	ballotLog.VoterId = args[0]
	ballotLog.ConfirmationKey = args[1]
	ballotLog.DateTime, err = time.Parse(time.RFC3339, args[2])

	if err != nil {
		return nil, err
	}

	return &ballotLog, nil
}
