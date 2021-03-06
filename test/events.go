package test

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/Gearbox-protocol/sdk-go/core"
	"github.com/Gearbox-protocol/sdk-go/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (c *TestEvent) Process(contractName string) types.Log {
	topic0 := core.Topic(c.Topics[0])
	c.Topics[0] = topic0.Hex()
	var topics []common.Hash
	for _, value := range c.Topics {
		splits := strings.Split(value, ":")
		var newTopic string
		if len(splits) == 1 {
			newTopic = value
		} else {
			switch splits[0] {
			case "bigint":
				arg, ok := new(big.Int).SetString(splits[1], 10)
				if !ok {
					log.Fatalf("bigint parsing failed for %s", value)
				}
				newTopic = fmt.Sprintf("%x", arg)
			}
		}
		topics = append(topics, common.HexToHash(newTopic))
	}
	data, err := c.ParseData([]string{contractName}, topic0)
	log.CheckFatal(err)
	return types.Log{
		Data:    data,
		Topics:  topics,
		Address: common.HexToAddress(c.Address),
		TxHash:  common.HexToHash(c.TxHash),
	}
}

func (c *TestEvent) ParseData(contractName []string, topic0 common.Hash) ([]byte, error) {
	if len(c.Data) == 0 {
		return []byte{}, nil
	}
	if contractName[0] == "ACL" {
		contractName = append(contractName, "ACLTrait")
	}
	var event *abi.Event
	var err error
	for _, name := range contractName {
		abi := core.GetAbi(name)
		event, err = abi.EventByID(topic0)
		if err == nil {
			break
		}
	}
	log.CheckFatal(err)
	var args []interface{}
	for _, entry := range c.Data {
		var arg interface{}
		splits := strings.Split(entry, ":")
		if len(splits) == 2 {
			var ok bool
			switch splits[0] {
			case "bigint":
				arg, ok = new(big.Int).SetString(splits[1], 10)
				if !ok {
					log.Fatalf("bigint parsing failed for %s", entry)
				}
			case "addr":
				arg = common.HexToAddress(entry).Hex()
			case "bool":
				if splits[1] == "1" {
					arg = true
				} else {
					arg = false
				}
			}
		} else {
			arg = common.HexToAddress(entry)
		}
		args = append(args, arg)
	}
	return event.Inputs.NonIndexed().Pack(args...)
}

type TestEvent struct {
	Address string   `json:"address"`
	Data    []string `json:"data"`
	Topics  []string `json:"topics"`
	TxHash  string   `json:"txHash"`
}
