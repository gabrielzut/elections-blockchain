package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type Ballot struct {
	Id              string    `json:"id"`
	DateTime        time.Time `json:"dateTime"`
	CandidateNumber int       `json:"candidateNumber"`
}

func (bal *Ballot) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (bal *Ballot) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()

	switch fcn {
	case "initElection":
		return bal.InitElection(stub)
	case "endElection":
		return bal.EndElection(stub)
	case "vote":
		return bal.Vote(stub, params)
	case "auditById":
		return bal.AuditByID(stub, params)
	case "auditByRange":
		return bal.AuditByRange(stub, params)
	}

	return shim.Error("Error: invalid function")
}

func (bal *Ballot) Vote(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Wrong number of arguments! Expected 3 but got " + strconv.Itoa(len(args)))
	}

	ballot, err := ReadAndValidateArgs(args)

	if err != nil {
		return shim.Error("Error validating data: " + err.Error())
	}

	ballotBytes, err := json.Marshal(ballot)

	if err != nil {
		return shim.Error("Error serializing data: " + err.Error())
	}

	status, err := CheckElectionStatus(stub)

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if status != "STARTED" {
		return shim.Error("Error: election on status " + status)
	}

	err = stub.PutState(ballot.Id, ballotBytes)

	if err != nil {
		return shim.Error("Error putting ballot on state: " + err.Error())
	}

	return shim.Success([]byte(ballot.Id))
}

func (bal *Ballot) AuditByID(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Wrong number of arguments! Expected 1 but got " + strconv.Itoa(len(args)))
	}

	ballotBytes, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if ballotBytes == nil {
		return shim.Error("Error getting state: ballot not found.")
	}

	return shim.Success(ballotBytes)
}

func (bal *Ballot) InitElection(stub shim.ChaincodeStubInterface) sc.Response {
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

func (bal *Ballot) EndElection(stub shim.ChaincodeStubInterface) sc.Response {
	status, err := CheckElectionStatus(stub)

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	} else if status != "STARTED" {
		return shim.Error("Error: election on status " + status)
	}

	err = stub.PutState("EndElection", []byte{0x00})

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (bal *Ballot) AuditByRange(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Wrong number of arguments! Expected 2 but got " + strconv.Itoa(len(args)))
	}

	iterator, err := stub.GetStateByRange(args[0], args[1])

	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	}

	var ballots []Ballot

	for iterator.HasNext() {
		ballotEntry, err := iterator.Next()

		if err != nil {
			return shim.Error("Error getting state: " + err.Error())
		}

		ballotBytes := ballotEntry.Value

		ballot := Ballot{}

		err = json.Unmarshal(ballotBytes, &ballot)

		if err != nil {
			return shim.Error(err.Error())
		}

		ballots = append(ballots, ballot)
	}

	ballotsAsBytes, err := json.Marshal(ballots)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ballotsAsBytes)
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

func ReadAndValidateArgs(args []string) (*Ballot, error) {
	var ballot Ballot
	var err error

	ballot.Id = args[0]

	ballot.CandidateNumber, err = strconv.Atoi(args[1])

	if err != nil {
		return nil, err
	}

	ballot.DateTime, err = time.Parse(time.RFC3339, args[2])

	if err != nil {
		return nil, err
	}

	return &ballot, nil
}
