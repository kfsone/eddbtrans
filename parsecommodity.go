package eddbtrans

import (
	"errors"
	"io"
	"io/ioutil"

	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"github.com/kfsone/gomenacing/pkg/parsing"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
)

func ParseCommodityJson(source io.Reader) (<-chan parsing.EntityPacket, error) {
	json, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}
	if !gjson.ValidBytes(json) {
		return nil, errors.New("malformed commodity-list json")
	}

	commodities := make(chan parsing.EntityPacket, 1)
	go func() {
		defer close(commodities)
		results := gjson.ParseBytes(json)
		results.ForEach(func(_, value gjson.Result) bool {
			if !value.IsObject() {
				return true
			}
			values := value.Map()
			id := uint32(values["id"].Uint())
			data, err := proto.Marshal(&gom.Commodity{
				Id:              id,
				Name:            values["name"].String(),
				CategoryId:      gom.Commodity_Category(values["category_id"].Uint()),
				IsRare:          values["is_rare"].Bool(),
				IsNonMarketable: values["is_non_marketable"].Bool(),
				AverageCr:       uint32(values["average_price"].Uint()),
			})
			if err != nil {
				panic(err)
			}
			commodities <- parsing.EntityPacket{ObjectId: id, Data: data}
			return true
		})
	}()

	return commodities, nil
}
