package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type AssetFactory struct {
}

type Asset struct {
	ID string `json:"id"` //资产id
	Location string `json:"location"` //所在位置
	AssetNo string `json:"assetNo"` //资产编号
	AssetName string `json:"assetName"` //资产名称
	AssetType string `json:"assetType"` //固定资产类别
	HouseJg string `json:"houseJg"`//房屋结构
	BuildYear string `json:"buildYear"`//构建年份（启用日期）
	BuildingType string `json:"buildingType"`//房屋类别
	Area string `json:"area"`//建筑面积
	FloorArea string `json:"floorArea"`//占地面积
	RentableArea string `json:"rentableArea"`//可出租面积
	UnrentableArea string `json:"unrentableArea"`//不可出租面积
	HouseCert string `json:"houseCert"`//房产证
	LandCert string `json:"landCert"`//土地证
	AssetsUsage string `json:"assetsUsage"`//房屋用途
	IsMortgage string `json:"isMortgage"`//是否抵押
	Account string `json:"account"`//账内或代管
	LocationDetail string `json:"locationDetail"`//坐落位置
	Longitude string `json:"longitude"`//经度
	Latitude string `json:"latitude"`//纬度
	Remark string `json:"remark"`//备注
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t AssetFactory) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t AssetFactory) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	switch fn {
	case "add":
		return add(stub, args)
	case "getById":
		return getById(stub, args)
	case "update":
		return update(stub, args)
	case "patch":
		return patch(stub, args)
	default:
		return shim.Error("unsupported method " + fn)
	}
}

func add(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("should have 2 args")
	}
	id, jsonValue := args[0], args[1]
	err := stub.PutState(id, []byte(jsonValue))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(id))
}

func getById(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) !=  1 {
		return shim.Error("should have 1 args")
	}
	id := args[0]
	jsonValue, err := stub.GetState(id)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(jsonValue)
}

func update(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("should have 2 args")
	}
	id, jsonValue := args[0], args[1]
	err := stub.PutState(id, []byte(jsonValue))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(id))
}

func patch(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("should have 2 args")
	}
	id, patchValue := args[0], args[1]

	asset := Asset{}

	jsonVal, err := stub.GetState(id)

	if err != nil {
		return shim.Error(err.Error())
	}

	err = json.Unmarshal(jsonVal, &asset)

	if err != nil {
		return shim.Error(err.Error())
	}

	patchMap := make(map[string]string)

	err = json.Unmarshal([]byte(patchValue), &patchMap)

	if err != nil {
		return shim.Error(err.Error())
	}

	point := mergeStructAndMap(&asset, patchMap).(*Asset)

	update, err := json.Marshal(point)

	err = stub.PutState(id, update)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(stub.GetTxID()))

}

func mergeStructAndMap(point interface{}, jsonMap map[string]string) interface{} {
	orderType := reflect.TypeOf(point).Elem()
	orderValue := reflect.ValueOf(point).Elem()
	for i := 0; i < orderType.NumField(); i++ {
		field := orderType.Field(i)
		jsonTag := field.Tag.Get("json")
		if val, ok := jsonMap[jsonTag]; ok {
			orderValue.FieldByName(field.Name).SetString(val)
		}
	}
	return point
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(AssetFactory)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
