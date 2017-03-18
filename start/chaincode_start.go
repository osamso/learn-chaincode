/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
//	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var votingIndexStr = "_votingindex"

type Member struct{
	Id string `json:"id"`				
	Name string `json:"name"`
	Category string `json:"category"`
	Office string `json:"office"`
}

type Vote struct{
	Voter Member `json:"voter"`
	Justification string `json:"justification"`
}

type Option struct{
	Id int `json:"id"`
	Description string `json:"description"`
	NumberOfVotes int `json:"number_of_votes"`
	Votes []Vote `json:"votes"`
}

type Voting struct{
	Id string `json:"id"`
	Description string `json:"description"`
	Voters []Member `json:"voters"`
	Options []Option `json:"options"`
	Status bool `json:"status"`
	VotingDeadlineInMinutes int `json:"voting_deadline_in_minutes"`
	StartVotingTimestamp int64 `json:"start_voting_timestamp"` 
}

type AllVotings struct{
	Votings []Voting `json:"votings"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Voting chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error
	var jsonAsBytes []byte 	

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}


	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var votings AllVotings
	jsonAsBytes, _ = json.Marshal(votings)
	err = stub.PutState(votingIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil

}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "add_voting" {										//deletes an entity from its state
		res, err := t.add_voting(stub, args)
		return res, err
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")

}

func (t *SimpleChaincode) add_voting(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var option Option
	
	//	0      1      2     3      4      * 
	//	["1", "Iniciativas Innovacion 2017", "30", "Arquitecturas IoT", "BlockChain", "ChatBots", "ELK", "Gestión entornos de Desarrollo", "Physical web"]
	if len(args) < 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 5?")
	}

	// 1st argument - Voting ID
	id1, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}

	// 2nd argument - Voting Description
	if len(args[1]) <= 0 {
                return nil, errors.New("2nd argument must be a non-empty string")
        }

	// 3rd argument - VotingDeadlineInMinutes 
        deadline1, err := strconv.Atoi(args[2])
        if err != nil {
                return nil, errors.New("3rd argument must be a numeric string")
        }

	fmt.Println("Creating new Voting...")
	voting := Voting{}
	voting.Id = strconv.Itoa(id1)
	fmt.Println("voting.Id: %s", voting.Id)
	voting.Description = args[1]
        fmt.Println("voting.Description: %s", voting.Description)
	voting.StartVotingTimestamp = makeTimestamp()
	fmt.Println("voting.StartVotingTimestamp: %s", strconv.FormatInt(voting.StartVotingTimestamp, 10))	
	voting.VotingDeadlineInMinutes = deadline1
	fmt.Println("voting.VotingDeadlineInMinutes: %s", strconv.Itoa(voting.VotingDeadlineInMinutes))	

	// voting.Voters
	jsonAsBytes, _ := json.Marshal(voting)
	
	fmt.Println("Add Voting - PutState")
	err = stub.PutState("_debug1", jsonAsBytes)
	
	optionId := 1	
	for i:=3; i < len(args); i++ {
		
		fmt.Println("Creating new voting Option...")
		fmt.Println("Option.Id: " + strconv.Itoa(optionId))
		option.Id = optionId
		optionId++;
		
		if len(args[i]) <= 0 {
			return nil, errors.New(strconv.Itoa(i) + " argument must be a non-empty string")
		}
		option.Description = args[i] 		
		fmt.Println("Option.Description: %s", option.Description)		

		option.NumberOfVotes = 0		
		fmt.Println("Option.NumberOfVotes: %s", strconv.Itoa(option.NumberOfVotes)) 
		
		// TODO: Pendent inicialitzar vots
		
		fmt.Println("! created option: " + strconv.Itoa(option.Id))
		jsonAsBytes, _ = json.Marshal(option)
		err = stub.PutState("_debug2", jsonAsBytes)
		
		voting.Options = append(voting.Options, option)
		fmt.Println("! appended option to voting")
		
	}
	
	//get the voting struct
	votingsAsBytes, err := stub.GetState(votingIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get voting")
	}
	var allvotings AllVotings
	json.Unmarshal(votingsAsBytes, &allvotings)										//un stringify it aka JSON.parse()
	
	allvotings.Votings = append(allvotings.Votings, voting);				
	fmt.Println("! appended voting to allvotings")
	jsonAsBytes, _ = json.Marshal(allvotings)
	err = stub.PutState(votingIndexStr, jsonAsBytes)					
	if err != nil {
		return nil, err
	}
	fmt.Println("- end init voting")
	return nil, nil
}

func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")

}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}
