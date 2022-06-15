package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Goods struct {
	BatchNo     string `json:"batchNo"`     // 商品批次号
	StockId     string `json:"stockId"`     // 商品库存编号，主键
	GoodsName   string `json:"goodsName"`   // 商品名称
	GoodsOrigin string `json:"goodsOrigin"` // 产地
	MarketName  string `json:"marketName"`  // 市场名称
	GoodsId     string `json:"goodsId"`     // 商品编号
	GoodsPic    string `json:"goodsPic"`    // 商品图片
	KindId      string `json:"kindId"`      // 分类编号
	KindName    string `json:"kindName"`    // 分类名称
	Weight      string `json:"weight"`      // 重量
	Price       string `json:"price"`       // 单价
	ShopId      string `json:"shopId"`      // 摊位编号
	Amount      string `json:"amount"`      // 上架数量
	StockNum    string `json:"stockNum"`    // 库存数量
	IsSelf      string `json:"isSelf"`      // 0 非自产， 1 自产
	FileName    string `json:"fileName"`    // 进货单
	Desc        string `json:"desc"`        // 备注
	StorageTime string `json:"storageTime"` // 入链时间
	SubmitTime  string `json:"submitTime"`  // 提交时间
	GsiStatus   string `json:"gsiStatus"`   // 上架状态 0 上架， 1下架
	QcStatus    string `json:"qcstatus"`    // 0 未检测，1 合格， 2 不合格 3 复检合格 4 复检不合格
}

type PersonalGoodsRes struct {
	KindId   string `json:"kindId"`   // 分类编号
	KindName string `json:"kindName"` // 分类名称
	Amount   string `json:"amount"`   // 上架数量
	StockNum string `json:"stockNum"` // 库存数量
}

type Pagination struct {
	Bookmark string `json:"bookmark"`
	PageSize int32  `json:"pageSize"`
}

var ErrorNotFound = fmt.Sprint("record not found")

type GoodsContract struct {
}

var orderContractName = "order"

func (t GoodsContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 0 {
		orderContractName = args[0]
	}
	return shim.Success(nil)
}

func (t GoodsContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	switch fn {
	case "addGoods":
		return addGoods(stub, args)
	case "updateGoodStatus":
		return updateGoodStatus(stub, args)
	case "queryGoodsShopIdAndKindName":
		return queryGoodsShopIdAndKindName(stub, args)
	case "queryGoodsDetailByMap":
		return queryGoodsDetailByMap(stub, args)
	case "queryFriendGoodsListByMap":
		return queryFriendGoodsListByMap(stub, args)
	case "updateGoodsAmount":
		return updateGoodsAmount(stub, args)
	case "queryGoodsByStockId":
		return queryGoodsByStockId(stub, args)
	case "updateGoodsStockFileNameOrPrice":
		return updateGoodsStockFileNameOrPrice(stub, args)
	case "traceGoodsAndOrderByBatchNo":
		return traceGoodsAndOrderByBatchNo(stub, args)
	case "addOrder":
		return addOrder(stub, args)
	case "updateOrder":
		return updateOrder(stub, args)
	default:
		return shim.Error("unsupported method " + fn)
	}
}

// stockId 为主键
func addGoods(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("args length should be 1")
	}
	jsonValue := args[0]
	goods := Goods{}
	err := json.Unmarshal([]byte(jsonValue), &goods)
	if err != nil {
		return shim.Error("unmarshal goods failed" + err.Error())
	}
	id := goods.StockId
	if id == "" {
		return shim.Error("stockId is required")
	}
	err = stub.PutState(id, []byte(jsonValue))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(stub.GetTxID()))
}

// 根据 stockId 更新 gsiStatus 状态
// stockId gsiStatus
func updateGoodStatus(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("args length should be 2")
	}
	stockId, gsiStatus := args[0], args[1]
	goodsJson, err := stub.GetState(stockId)
	if err != nil {
		return shim.Error(err.Error())
	}
	goods := Goods{}
	err = json.Unmarshal(goodsJson, &goods)
	if err != nil {
		return shim.Error("failed to unmarshal goods:" + err.Error())
	}
	goods.GsiStatus = gsiStatus
	updateValue, err := json.Marshal(&goods)
	if err != nil {
		return shim.Error("failed to marshal goods:" + err.Error())
	}
	err = stub.PutState(stockId, updateValue)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(stub.GetTxID()))
}

// 根据 shopId 查 goods， kindName 为模糊匹配
// shopId string required
// kindName string
// bookmark string
// pageSize int
// res {"data": [PersonalGoodsRes], "bookmark": "bookmark"}
func queryGoodsShopIdAndKindName(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("args length should be 4")
	}
	shopId, kindName, bookmark, sizeString := args[0], args[1], args[2], args[3]
	pageSize, err := strconv.Atoi(sizeString)
	if err != nil {
		return shim.Error("failed to parse pageSize" + sizeString)
	}
	equal := map[string]string{
		"shopId": shopId,
	}
	reg := make(map[string]string)
	if kindName != "" {
		reg["kindName"] = kindName
	}
	sort := map[string]string{
		"storageTime": "desc",
	}
	index := []string{"_design/goodsShopIdDoc", "goodsShopId"}
	query, err := generateQueryString(equal, reg, sort, index)

	if err != nil {
		return shim.Error("failed to generate query string:" + err.Error())
	}

	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(query, int32(pageSize), bookmark)

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	data := make([]PersonalGoodsRes, 0)

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("failed get resultsIterator:" + err.Error())
		}
		val := PersonalGoodsRes{}
		err = json.Unmarshal(queryResponse.Value, &val)
		if err != nil {
			return shim.Error("failed to unmarshal PersonalGoodsRes:" + err.Error())
		}
		data = append(data, val)
	}
	if err != nil {
		return shim.Error("constructQueryResponse failed:" + err.Error())
	}
	nextBookMark := responseMetadata.Bookmark
	res := map[string]interface{}{
		"data":     data,
		"bookmark": nextBookMark,
	}
	resStr, err := json.Marshal(&res)
	if err != nil {
		return shim.Error("failed to marshal res" + err.Error())
	}
	return shim.Success(resStr)
}

type GoodsDetailList struct {
	Pagination
	ShopId    string `json:"shopId"`
	KindId    string `json:"kindId"`
	GoodsId   string `json:"goodsId"`
	GsiStatus string `json:"gsiStatus"`
	GcStatus  string `json:"gcStatus"`
}

// 根据 shopId 和 kindId 查 goods， 可选字段精准匹配
// shopId string required
// kindId string required
// goodsId string
// gsiState string
// qcstatus string
// bookmark string required
// pageSize int required
// res : {data:[Goods],"bookmark": "bookmark"}
func queryGoodsDetailByMap(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("args length should be 1")
	}
	argStruct := GoodsDetailList{}
	err := json.Unmarshal([]byte(args[0]), &argStruct)
	if err != nil {
		return shim.Error("failed to unmarshal argStruct:" + err.Error())
	}

	if argStruct.ShopId == "" || argStruct.KindId == "" {
		return shim.Error("shopId or kindId required")
	}

	equal := make(map[string]string)
	reg := make(map[string]string)
	sort := map[string]string{
		"storageTime": "desc",
	}
	index := []string{"_design/goodsShopIdKindIdDoc", "goodsShopIdKindId"}

	if argStruct.GsiStatus != "" {
		equal["gsiStatus"] = argStruct.GsiStatus
	}
	if argStruct.GoodsId != "" {
		equal["goodsId"] = argStruct.GoodsId
	}
	if argStruct.GcStatus != "" {
		equal["gcStatus"] = argStruct.GcStatus
	}

	query, err := generateQueryString(equal, reg, sort, index)

	if err != nil {
		return shim.Error("failed to generate query string:" + err.Error())
	}

	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(query, argStruct.PageSize, argStruct.Bookmark)

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buf, err := constructQueryResponseFromIterator(resultsIterator, responseMetadata.Bookmark)

	if err != nil {
		return shim.Error("failed to generate res" + err.Error())
	}

	return shim.Success(buf.Bytes())
}

type FriendGoodsParam struct {
	Pagination
	ShopIdList []string `json:"shopIdList"`
	GoodsName  string   `json:"goodsName"`
}

// shopIdList [] 好友shopId
// goodsName string  模糊匹配
// bookmark string 默认1
// pageSize int 默认20
// 根据上链时间 desc
// res : {data:[Goods],"bookmark": "bookmark"}
func queryFriendGoodsListByMap(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("args length should be 1")
	}

	argStruct := FriendGoodsParam{}
	err := json.Unmarshal([]byte(args[0]), &argStruct)
	if err != nil {
		return shim.Error("failed to unmarshal argStruct:" + err.Error())
	}

	selectMap := map[string]interface{}{
		"shopId": map[string]interface{}{
			"$in": argStruct.ShopIdList,
		},
	}

	if argStruct.GoodsName != "" {
		selectMap["kindName"] = map[string]string{
			"$regex": argStruct.GoodsName,
		}
	}

	queryMap := map[string]interface{}{
		"sort": []map[string]string{{"storageTime": "desc"}},
	}
	queryMap["use_index"] = []string{"_design/goodsShopIdDoc", "goodsShopId"}
	queryMap["selector"] = selectMap

	query, err := json.Marshal(&queryMap)

	if err != nil {
		return shim.Error("failed to marshal queryMap:" + err.Error())
	}

	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(string(query), argStruct.PageSize, argStruct.Bookmark)

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buf, err := constructQueryResponseFromIterator(resultsIterator, responseMetadata.Bookmark)

	if err != nil {
		return shim.Error("failed to generate res" + err.Error())
	}

	return shim.Success(buf.Bytes())

}

// 根据 stockId 更新 goods 库存
// stockId string required
// amount string required
// updateType string required 0减库存 1加库存
func updateGoodsAmount(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("args length should be 3")
	}
	stockId, amountStr, updateType := args[0], args[1], args[2]
	if stockId == "" || amountStr == "" || updateType == "" {
		return shim.Error("stockId, amount and updateType is required")
	}
	if updateType != "0" && updateType != "1" {
		return shim.Error("updateType should be 0 or 1, get " + updateType)
	}
	amount, err := strconv.Atoi(amountStr)

	if err != nil {
		return shim.Error("failed to get amount" + err.Error())
	}

	goodsStr, err := stub.GetState(stockId)

	if err != nil {
		return shim.Error(err.Error())
	}

	goods := Goods{}
	err = json.Unmarshal(goodsStr, &goods)

	if err != nil {
		return shim.Error("failed to unmarshal goods" + err.Error())
	}
	stockNum, err := strconv.Atoi(goods.StockNum)
	if err != nil {
		return shim.Error(err.Error())
	}

	if updateType == "0" {
		stockNum -= amount
	} else {
		stockNum += amount
	}

	stockNumStr := strconv.Itoa(stockNum)
	goods.StockNum = stockNumStr

	bytes, err := json.Marshal(&goods)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(stockId, bytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(stub.GetTxID()))
}

// 根据 stockId 获取 goods 详情
// stockId string
func queryGoodsByStockId(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("args length should be 1")
	}

	stockId := args[0]
	goodsStr, err := stub.GetState(stockId)
	if err != nil {
		return shim.Error(err.Error())
	}
	if goodsStr == nil {
		return shim.Error(ErrorNotFound)
	}
	return shim.Success(goodsStr)
}

// 更新 goods的进货单或价格
// stockId string
// fileName string
// price string
func updateGoodsStockFileNameOrPrice(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("args length should be 3")
	}
	stockId, fileName, price := args[0], args[1], args[2]
	if stockId == "" {
		return shim.Error("stockId is required")
	}
	goods := Goods{}
	jsonVal, err := stub.GetState(stockId)
	if err != nil {
		return shim.Error(err.Error())
	}
	if jsonVal == nil {
		return shim.Error(ErrorNotFound)
	}
	err = json.Unmarshal(jsonVal, &goods)
	if err != nil {
		return shim.Error(err.Error())
	}
	if fileName != "" {
		goods.FileName = fileName
	}
	if price != "" {
		goods.Price = price
	}

	update, err := json.Marshal(&goods)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(stockId, update)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(stub.GetTxID()))
}

// 根据 batchNo 查询 order 列表和 goods列表
// batchNo string
// res : {"good": Goods, "orders": [Order]}
func traceGoodsAndOrderByBatchNo(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("args length should be 1")
	}
	batchNo := args[0]
	if batchNo == "" {
		return shim.Error("batchNo required")
	}
	goodsQuery := fmt.Sprintf("{\"selector\":{\"batchNo\":{\"$eq\": \"%s\" }},\"sort\":[{\"storageTime\":\"desc\"}],\"use_index\":[\"_design/goodsStorageTimeDoc\",\"goodsStorageTime\"]}", batchNo)
	resultsIterator, err := stub.GetQueryResult(goodsQuery)

	var buffer bytes.Buffer
	buffer.WriteString("{\"good\":")

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {

		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("failed to get goods")
		}

		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString(",")
		break
	}

	buffer.WriteString("\"orders\":[")

	orderRes := stub.InvokeChaincode(orderContractName, [][]byte{[]byte("queryOrder"), []byte(batchNo)}, stub.GetChannelID())

	if orderRes.Status != shim.OK {
		return shim.Error("failed to invoke queryOrder method")
	}

	buffer.WriteString(string(orderRes.Payload))

	buffer.WriteString("]}")

	return shim.Success(buffer.Bytes())
}

func addOrder(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	res := stub.InvokeChaincode(orderContractName, [][]byte{[]byte("addOrder"), []byte(args[0])}, stub.GetChannelID())
	if res.Status != shim.OK {
		return shim.Error("failed to invoke addOrder")
	}
	return shim.Success(res.Payload)
}

func updateOrder(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	res := stub.InvokeChaincode(orderContractName, [][]byte{[]byte("updateOrder"), []byte(args[0])}, stub.GetChannelID())
	if res.Status != shim.OK {
		return shim.Error("failed to invoke addOrder")
	}
	return shim.Success(res.Payload)
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
	if bookmark == "" {
		buffer.WriteString("\"\"")
	} else {
		buffer.WriteString(bookmark)
	}
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
	if err := shim.Start(new(GoodsContract)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
