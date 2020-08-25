package eddbtrans

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
)

func ParseCommodityJson(source io.Reader) (<-chan []byte, error) {
	json, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}
	if !gjson.ValidBytes(json) {
		return nil, errors.New("malformed commodity-list json")
	}

	commodities := make(chan []byte, 8)
	go func() {
		defer close(commodities)
		results := gjson.ParseBytes(json)
		results.ForEach(func(_, value gjson.Result) bool {
			if !value.IsObject() {
				return true
			}
			values := value.Map()
			data, err := proto.Marshal(&Commodity{
				Id:           uint32(values["id"].Uint()),
				Name:         values["name"].String(),
				Category:     Commodity_Category(values["category_id"].Uint()),
				IsRare:       values["is_rare"].Bool(),
				IsMarketable: !values["is_non_marketable"].Bool(),
				AveragePrice: uint32(values["average_price"].Uint()),
				EdId:         values["ed_id"].Uint(),
			})
			if err != nil {
				panic(err)
			}
			commodities <- data
			return true
		})
	}()

	return commodities, nil
}
