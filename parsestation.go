package eddbtrans

import (
	"io"
	"strings"

	"google.golang.org/protobuf/proto"
)

func getStationType(typeId uint64) Station_Type {
	if _, ok := Station_Type_name[int32(typeId)]; ok {
		return Station_Type(typeId)
	}
	return Station_None
}

func getPadSize(padSize string) Station_Features_PadSize {
	switch strings.ToUpper(padSize) {
	case "S":
		return Station_Features_Small
	case "M":
		return Station_Features_Medium
	case "L":
		return Station_Features_Large
	default:
		return Station_Features_None
	}
}

func ParseStationsJsonl(source io.Reader) (<-chan []byte, error) {
	channel := make(chan []byte, 4)

	go func() {
		defer close(channel)
		stations := ParseJsonLines(source, getStationFields())
		for station := range stations {
			data, err := proto.Marshal(&Station{
				Id:       station[0].Uint(),
				Name:     station[1].String(),
				Updated:  station[2].Uint(),
				SystemId: station[3].Uint(),
				Type:     getStationType(station[4].Uint()),
				Features: &Station_Features{
					HasMarket:      station[5].Bool(),
					HasBlackmarket: station[6].Bool(),
					HasRefuel:      station[7].Bool(),
					HasRepair:      station[8].Bool(),
					HasRearm:       station[9].Bool(),
					HasOutfitting:  station[10].Bool(),
					HasShipyard:    station[11].Bool(),
					HasDocking:     station[12].Bool(),
					HasCommodities: station[13].Bool(),
					IsPlanetary:    station[14].Bool(),
					Pad:            getPadSize(station[15].String()),
				},
				LsToStar:   float32(station[16].Float()),
				Government: &Government{Type: getGovernmentType(station[17].Uint())},
				Allegiance: &Allegiance{Type: getAllegianceType(station[18].Uint())},
				EdMarketId: station[19].Uint(),
			})
			if err != nil {
				panic(err)
			} else {
				channel <- data
			}
		}
	}()

	return channel, nil
}
