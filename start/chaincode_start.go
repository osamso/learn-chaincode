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

const votingIndexStr string = "_votingindex"

const ERROR_CODE_VOTING_CLOSED string = "100"
const ERROR_DESCRIPTION_VOTING_CLOSED string = "Voting is closed"

const ERROR_CODE_MEMBER_ALREADY_VOTED string = "101"
const ERROR_DESCRIPTION_MEMBER_ALREADY_VOTED string = "Member already voted"


type Member struct{
	Id string `json:"id"`
	Name string `json:"name"`
	Channel string `json:"channel"`
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
	Status bool `json:"status"` // Open = true, Closed = false
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
	if function == "init" {
		return t.Init(stub, "init", args)

	} else if function == "add_voting" {
		res, err := t.add_voting(stub, args)
		return res, err

	} else if function == "vote" {
  	res, err := t.vote(stub, args)
    return res, err

	} else if function == "update_voting_status" {
		res, err := t.update_voting_status(stub, args)
		return res, err

	} else if function == "update_votings_status" {
	  res, err := t.update_votings_status(stub, args)
		return res, err
	}

	fmt.Println("invoke did not find func: " + function)
	// Error
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
	fmt.Printf("voting.Id: %s\n", voting.Id)
	voting.Description = args[1]
  fmt.Printf("voting.Description: %s\n", voting.Description)
	voting.Status = true
  fmt.Printf("voting.Status: %b\n", voting.Status)

	voting.StartVotingTimestamp = makeTimestamp()
	fmt.Printf("voting.StartVotingTimestamp: %s\n", strconv.FormatInt(voting.StartVotingTimestamp, 10))
	voting.VotingDeadlineInMinutes = deadline1
	fmt.Printf("voting.VotingDeadlineInMinutes: %s\n", strconv.Itoa(voting.VotingDeadlineInMinutes))

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
		fmt.Printf("Option.Description: %s\n", option.Description)

		option.NumberOfVotes = 0
		fmt.Printf("Option.NumberOfVotes: %s\n", strconv.Itoa(option.NumberOfVotes))

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
	json.Unmarshal(votingsAsBytes, &allvotings)

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

func (t *SimpleChaincode) vote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var currentTime, expirationTime int64
	var votingDeadlineInMilliseconds int

	// votingId, optionId, justification, id, name, category, office, channel
	//     0        1      	     2         3    4      5	    6	     7
  // ["1", "2", "Justification", "1", "Peter", "SKL", "Barcelona", "Twitter"]
  if len(args) < 8 {
  	return nil, errors.New("Incorrect number of arguments. Expecting like 8?")
  }

	// 1st argument - votingId
	votingId1, err := strconv.Atoi(args[0])
	if err != nil {
  	return nil, errors.New("1st argument must be a numeric string")
	}

	// 2nd argument - optionId
  optionId1, err := strconv.Atoi(args[1])
  if err != nil {
    return nil, errors.New("2nd argument must be a numeric string")
  }

	// 4th argument - memberId
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}

	// Get all votings
	votingsAsBytes, err := stub.GetState(votingIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get votings")
	}
	var allvotings AllVotings
  json.Unmarshal(votingsAsBytes, &allvotings)

	// Get current timestamp
	currentTime = makeTimestamp()

	// Iterate all votings
	for i := range allvotings.Votings{

		// Checking by voting ID
		if allvotings.Votings[i].Id == strconv.Itoa(votingId1){
			fmt.Println("Voting found!");

			// Checking if voting is closed
			if allvotings.Votings[i].Status == false {
				fmt.Println("Error - Voting is closed");
				t.error(ERROR_CODE_VOTING_CLOSED, ERROR_DESCRIPTION_VOTING_CLOSED)

			} else {

				votingDeadlineInMilliseconds = allvotings.Votings[i].VotingDeadlineInMinutes * 60 * 1000
				expirationTime = allvotings.Votings[i].StartVotingTimestamp + int64(votingDeadlineInMilliseconds)

				fmt.Printf("Checking voting expiration...\n")
				fmt.Printf("votingDeadlineInMinutes: %s\n", strconv.Itoa(allvotings.Votings[i].VotingDeadlineInMinutes))
				fmt.Printf("votingDeadlineInMilliseconds: %s\n", strconv.Itoa(votingDeadlineInMilliseconds))
				fmt.Printf("expirationTime: %s\n", strconv.FormatInt(expirationTime, 10))
				fmt.Printf("currentTime: %s\n", strconv.FormatInt(currentTime, 10))

				// Checking if voting has expired and closes it
				if currentTime > expirationTime {
					fmt.Printf("Error Voting identified by Id = '%s' should be closed!!\n", allvotings.Votings[i].Id)
					allvotings.Votings[i].Status = false

				} else {

					// Check if member already voted
					for j := range allvotings.Votings[i].Voters{
						if allvotings.Votings[i].Voters[j].Id == args[3]{
							fmt.Println("Error - Member already voted");
							return t.error(ERROR_CODE_MEMBER_ALREADY_VOTED, ERROR_DESCRIPTION_MEMBER_ALREADY_VOTED)
						}
					}

					// Add new member
					newMember := Member{}
					newMember.Id = args[3]
					newMember.Name = args[4]
					newMember.Category = args[5]
					newMember.Office = args[6]
					newMember.Channel = args[7]
					allvotings.Votings[i].Voters = append(allvotings.Votings[i].Voters, newMember)

					// Find option by ID
					for k := range allvotings.Votings[i].Options{

						if allvotings.Votings[i].Options[k].Id == optionId1{
							fmt.Println("Option found!");

							newVote := Vote{}
							newVote.Voter = newMember
							newVote.Justification = args[2]

							allvotings.Votings[i].Options[k].Votes = append(allvotings.Votings[i].Options[k].Votes, newVote)
							allvotings.Votings[i].Options[k].NumberOfVotes += 1
						}
					} // end for voting.options
				} // end if voting expired

				jsonAsBytes, _ := json.Marshal(allvotings)
				err = stub.PutState(votingIndexStr, jsonAsBytes)
				if err != nil {
					return nil, err
				}
				fmt.Println("- end vote")
	      return nil, nil

			} // end voting not closed

		} // end voting found

	} // end iterate all votings
	return nil, errors.New("Voting not found")

}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "get_votings" {
		return t.get_votings(stub, args)
	} else if function == "get_voting"{
		return t.get_voting(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error
	return nil, errors.New("Received unknown function query")

}

func (t *SimpleChaincode) get_votings(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	//get the var from chaincode state
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}
	return valAsbytes, nil
}

func (t *SimpleChaincode) get_voting(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	// 1st argument - votingId
	votingId1, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}

	// Get all votings
	votingsAsBytes, err := stub.GetState(votingIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get votings")
	}
	var allvotings AllVotings
  json.Unmarshal(votingsAsBytes, &allvotings)

	// Iterate all votings
	for i := range allvotings.Votings{

		// Checking by voting ID
		if allvotings.Votings[i].Id == strconv.Itoa(votingId1){
			fmt.Println("Voting '" + args[0] + "' found!");
			voting := allvotings.Votings[i]
			jsonAsBytes, _ := json.Marshal(voting)
			return jsonAsBytes, nil
		}
	}
	fmt.Println("Voting '" + args[0] + "' not found!");
	return nil, errors.New("Voting not found")
}

func (t *SimpleChaincode) update_voting_status(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	// 1st argument - votingId
	votingId1, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("1st argument must be a numeric string")
	}

	// Get all votings
	votingsAsBytes, err := stub.GetState(votingIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get votings")
	}
	var allvotings AllVotings
  json.Unmarshal(votingsAsBytes, &allvotings)

	// Iterate all votings
	for i := range allvotings.Votings{

		// Find by voting ID
		if allvotings.Votings[i].Id == strconv.Itoa(votingId1){
			fmt.Println("Voting '" + args[0] + "' found!");
			allvotings.Votings[i].Status = false

			jsonAsBytes, _ := json.Marshal(allvotings)
			err = stub.PutState(votingIndexStr, jsonAsBytes)
			if err != nil {
				return nil, err
			}
			fmt.Printf("Voting ''" + args[0] + "'' status updated to false ")
			return nil, nil
		}
	}
	return nil, errors.New("Voting not found")
}

func (t *SimpleChaincode) update_votings_status(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var currentTime int64
	var votingDeadlineInMilliseconds int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	//get the var from chaincode state
	allVotingsAsBytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	var allvotings AllVotings
  json.Unmarshal(allVotingsAsBytes, &allvotings)

	// Get current timestamp
	currentTime = makeTimestamp()

	// Iterate all votings
	votingUpdated := false
	for i := range allvotings.Votings{

		votingDeadlineInMilliseconds = allvotings.Votings[i].VotingDeadlineInMinutes * 60 * 1000
		expirationTime := allvotings.Votings[i].StartVotingTimestamp + int64(votingDeadlineInMilliseconds)

		fmt.Printf("Checking voting expiration...\n")
		fmt.Printf("votingDeadlineInMinutes: %s\n", strconv.Itoa(allvotings.Votings[i].VotingDeadlineInMinutes))
		fmt.Printf("votingDeadlineInMilliseconds: %s\n", strconv.Itoa(votingDeadlineInMilliseconds))
		fmt.Printf("expirationTime: %s\n", strconv.FormatInt(expirationTime, 10))
		fmt.Printf("currentTime: %s\n", strconv.FormatInt(currentTime, 10))

		// Checking if voting has expired for closing it
		if currentTime > expirationTime {
			fmt.Printf("Voting identified by Id = '%s' has expired!!\n", allvotings.Votings[i].Id)
			allvotings.Votings[i].Status = false
			votingUpdated = true
		}
  }
	if votingUpdated == true {
		fmt.Printf("Votings status updated")
		jsonAsBytes, _ := json.Marshal(allvotings)
		err = stub.PutState(votingIndexStr, jsonAsBytes)
	} else {
			fmt.Printf("Votings status not updated")
	}
	if err != nil {
		return nil, err
	}
	fmt.Println("- end update_votings")
	return nil, nil
}


func (t *SimpleChaincode) error(code, description string) ([]byte, error) {
	jsonResp := "{\"error\": {\"code\":" + code + "\", \"description\":" + description + "\"}}"
	return nil, errors.New(jsonResp)
}
