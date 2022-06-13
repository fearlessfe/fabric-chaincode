package main

import (
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func TestAdd(t *testing.T) {
	args := []string{"key", "value"}
	chaincode := AssetFactory{}
	mockStub := shim.NewMockStub("asset", chaincode)
	res := add(mockStub, args)
	t.Log(res)
}