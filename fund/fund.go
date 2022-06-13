package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Fund struct {
}

type FundBill struct {
	ID string `json:"id,omitempty"` //资产id
	District string `json:"district,omitempty"` //所在位置
	AssetNo string `json:"assetNo,omitempty"` //资产编号
	AssetName string `json:"assetName,omitempty"` //资产名称
	AssetType string `json:"assetType,omitempty"` //固定资产类别
	HouseJg string `json:"houseJg,omitempty"`//房屋结构
	BuildYear string `json:"buildYear"`//构建年份（启用日期）
	BuildingType string `json:"buildingType,omitempty"`//房屋类别
	Area string `json:"area,omitempty"`//建筑面积
	FloorArea string `json:"floorArea,omitempty"`//占地面积
	RentableArea string `json:"rentableArea,omitempty"`//可出租面积
	UnrentableArea string `json:"unrentableArea,omitempty"`//不可出租面积
	HouseCert string `json:"houseCert"`//房产证
	LandCert string `json:"landCert"`//土地证
	AssetsUserd string `json:"assetsUserd"`//房屋用途
	IsMortage string `json:"is_mortage"`//是否抵押
	Account string `json:"account,omitempty"`//账内或代管
	LocationDetail string `json:"locationDetail"`//坐落位置
	Longitude string `json:"longitude"`//经度
	Latitude string `json:"latitude"`//纬度
	Remark string `json:"remark,omitempty"`//备注
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *Fund) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *Fund) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	switch fn {
	case "add":
		return t.add(stub, args)
	case "getById":
		return t.getById(stub, args)
	case "update":
		return t.update(stub, args)
	default:
		return shim.Error("unsupported method " + fn)
	}
}

func (t *Fund) add(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	id, jsonValue := args[0], args[1]
	err := stub.PutState(id, []byte(jsonValue))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(id))
}

func (t *Fund) getById(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	id := args[0]
	jsonValue, err := stub.GetState(id)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(jsonValue)
}

func (t *Fund) update(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	id, jsonValue := args[0], args[1]
	err := stub.PutState(id, []byte(jsonValue))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(id))
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(Fund)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}

