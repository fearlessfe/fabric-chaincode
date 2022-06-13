package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

const (
	ORDER_ID = "orderNo"
	GOODS_ID = "stockId"

)

var ErrorNotFound = fmt.Sprint("record not found")

type Order struct {
	OrderNo string `json:"orderNo"`  // 订单号，主键
	BatchNo string `json:"batchNo"`  // 批次
	MarketNo string `json:"marketNo"`  // 市场编号
	MarketName string `json:"marketName"` // 市场名称
	GoodsName string `json:"goodsName"` // 商品名称
	GoodsOrigin string `json:"goodsOrigin"` // 产地
	GoodsId string `json:"goodsId"`  // 商品编号
	Amount string `json:"amount"`  // 金额
	Weight string `json:"weight"`  // 重量
	SellerShopId string `json:"sellerShopId"`  // 商家摊位号
	SellerShop string `json:"sellerShop"` // 摊位名称
	SellerId string `json:"sellerId"` // 卖家id
	Seller string `json:"seller"`  // 卖家名称
	BuyerShopId string `json:"buyerShopId"` // 买家摊位id
	BuyerShop string `json:"buyerShop"` // 买家名称
	BuyerId string `json:"buyerId"` // 买家编号
	Buyer string `json:"buyer"`  // 买家名称
	GoodsStockId string `json:"goodsStockId"` // 商品库存编号
	TranTime string `json:"tranTime"` // 交易时间
	SubmitTime string `json:"submitTime"` // 提交时间
}

type Pagination struct {
	Bookmark string `json:"bookmark"`
	PageSize int32 `json:"pageSize"`
}

type Chaincode struct {

}


func (t Chaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t Chaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	switch fn {
	case "addOrder":
		return addOrder(stub, args)
	case "updateOrder":
		return updateOrder(stub, args)
	case "queryOrder":
		return queryOrder(stub, args)
	default:
		return shim.Error("unsupported method " + fn)
	}
}

func addOrder(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("args length should be 1")
	}
	jsonValue := args[0]
	order := Order{}
	err := json.Unmarshal([]byte(jsonValue), &order)
	if err != nil {
		return shim.Error(err.Error())
	}
	id := order.OrderNo
	if id == "" {
		return shim.Error("orderNo is required")
	}
	err = stub.PutState(id, []byte(jsonValue))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(stub.GetTxID()))
}
// orderNo 为主键来更新
func updateOrder(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("args length should be 1")
	}
	jsonValue := args[0]
	orderMap := make(map[string]string)
	err := json.Unmarshal([]byte(jsonValue), &orderMap)
	if err != nil {
		return shim.Error("failed to unmarshal map:" + err.Error())
	}
	order := Order{}
	if id, ok := orderMap[ORDER_ID]; !ok {
		return shim.Error("orderNo is required")
	} else {
		orderJson, err := stub.GetState(id)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = json.Unmarshal(orderJson, &order)
		if err != nil {
			return shim.Error("failed to unmarshal order:" + err.Error())
		}
	}

	update := mergeStructAndMap(&order, orderMap).(*Order)
	updateOrder, err := json.Marshal(update)
	if err != nil {
		return shim.Error("failed to marshal order:" + err.Error())
	}
	err = stub.PutState(order.OrderNo, updateOrder)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(stub.GetTxID()))
}

// 根据 batchNo 获取所有的 order
func queryOrder(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("should have only 1 args")
	}

	batchNo := args[0]

	orderQuery := fmt.Sprintf("{\"selector\":{\"batchNo\":{\"$eq\": %s },\"use_index\":[\"_design/orderBatchNoDoc\",\"batchNo\"]}", batchNo)

	orderIterator, err := stub.GetQueryResult(orderQuery)


	if err != nil {
		return shim.Error(err.Error())
	}
	defer orderIterator.Close()
	var buffer bytes.Buffer

	bArrayMemberAlreadyWritten := false
	for orderIterator.HasNext() {
		queryResponse, err := orderIterator.Next()
		if err != nil {
			return shim.Error("failed to get order")
		}
		// 首次不用加 "，"
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	return shim.Success(buffer.Bytes())
}


func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface, bookmark string) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	buffer.WriteString("{\"data\":[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// 首次不用加 "，"
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("],")
	buffer.WriteString("\"bookmark\":")
	buffer.WriteString(bookmark)
	buffer.WriteString("}")

	return &buffer, nil
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

func generateQueryString(equal map[string]string, regex map[string]string, sort map[string]string, index []string) (string, error) {
	selectorMap := make(map[string]map[string]string)
	for key, val := range equal {
		selectorMap[key] = map[string]string{
			"$eq": val,
		}
	}

	for key, val := range regex {
		selectorMap[key] = map[string]string{
			"$regex": val,
		}
	}

	queryMap := map[string]interface{}{
		"selector": selectorMap,
		"sort": []map[string]string{
			sort,
		},
		"use_index": index,
	}

	query, err := json.Marshal(&queryMap)


	if err != nil {
		return "", err
	}
	return string(query), nil
}

func main() {
	if err := shim.Start(new(Chaincode)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
