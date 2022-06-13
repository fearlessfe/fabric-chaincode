package main

import (
	"encoding/json"
	"reflect"
	"regexp"
	"testing"

)

func TestReflectField(t *testing.T)  {
	order := Order{}
	orderMap := map[string]string{
		ORDER_ID: "test",
		"marketNo": "eeee",
		"marketName": "sdadasd",
	}
	t.Log(order)
	orderType := reflect.TypeOf(&order).Elem()
	orderValue := reflect.ValueOf(&order).Elem()
	for i := 0; i < orderType.NumField(); i++ {
		field := orderType.Field(i)
		jsonTag := field.Tag.Get("json")
		if val, ok := orderMap[jsonTag]; ok {
			orderValue.FieldByName(field.Name).SetString(val)
		}
	}
	t.Log(order)

	res, _ := json.Marshal(&order)
	t.Log(string(res))
}

func TestMergeStructAndMap(t *testing.T)  {
	order := Order{}
	orderMap := map[string]string{
		ORDER_ID: "test",
		"marketNo": "eeee",
		"marketName": "sdadasd",
	}
	res := mergeStructAndMap(&order, orderMap)
	update := res.(*Order)
	if update.OrderNo != "test" {
		t.Fatal()
	}
	if update.MarketNo != "eeee" {
		t.Fatal()
	}
	if update.MarketName != "sdadasd" {
		t.Fatal()
	}
}

func TestQueryString(t *testing.T) {
	shopId := "222"
	kindName := "name"
	selectorMap := make(map[string]map[string]string)
	selectorMap["shopId"] = map[string]string{
		"$eq": shopId,
	}

	queryMap := map[string]interface{}{
		"sort": []map[string]string{
			{"storageTime":"desc"},
		},
		"use_index": []string{"_design/goodsDoc", "goods"},
	}
	queryMap["selector"] = selectorMap
	if kindName != "" {
		selectorMap["kindName"] = map[string]string{
			"$regex": kindName,
		}
	}
	res, _ := json.Marshal(&queryMap)
	t.Log(string(res))
}

func TestGenerateQueryString(t *testing.T)  {
	equal := map[string]string{
		"shopId" : "222",
	}
	reg := map[string]string{
		"kindName" : "name",
	}
	sort := map[string]string{
		"storageTime":"desc",
	}
	index := []string{"_design/goodsDoc", "goods"}
	res, _ := generateQueryString(equal, reg, sort, index)
	expect := "{\"selector\":{\"kindName\":{\"$regex\":\"name\"},\"shopId\":{\"$eq\":\"222\"}},\"sort\":[{\"storageTime\":\"desc\"}],\"use_index\":[\"_design/goodsDoc\",\"goods\"]}"
	if res != expect {
		t.Fatal()
	}
}

func TestUnmarshal(t *testing.T)  {
	testMap := map[string]interface{}{
		"shopId": "shopId",
		"kindId": "kindId",
		"goodsId": "goodsId",
		"gsiStatus": "gsiStatus",
		"gcStatus": "gcStatus",
		"bookmark": "bookmark",
		"pageSize": 20,
	}
	str, _ := json.Marshal(&testMap)
	s := GoodsDetailList{}
	_ = json.Unmarshal(str, &s)
	t.Log(s)
}

func TestReg(t *testing.T)  {
	str := "{\"a\":\"1\",\"b\":\"2\",\"c\":\"3\"}"
	re := regexp.MustCompile("\"b\":\"(.*?)\\\"")
	t.Log(re.FindString(str))
}
