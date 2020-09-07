package eddbtrans

import (
	"github.com/tidwall/gjson"
	"io"
	"strings"

	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"github.com/kfsone/gomenacing/pkg/parsing"

	"google.golang.org/protobuf/proto"
)

func getFacilityType(typeId uint64) gom.FacilityType {
	if _, ok := gom.FacilityType_name[int32(typeId)]; ok {
		return gom.FacilityType(typeId)
	}
	return gom.FacilityType_FTNone
}

func conditionalOr()

func getPadSize(padSize string) gom.FeatureMasks {
	switch strings.ToUpper(padSize) {
	case "S":
		return gom.FeatureMasks_FeatSmallPad
	case "M":
		return gom.FeatureMasks_FeatMediumPad
	case "L":
		return gom.FeatureMasks_FeatLargePad
	default:
		return 0
	}
}

// FacilityRegistry will provide facility-id checking for listings.
var FacilityRegistry *Daycare

func getFeatures(row []*gjson.Result, hasMarket, hasBlackMarket, hasRefuel, hasRepair, hasRearm, hasOutfitting, hasShipyard, hasDocking, hasCommodities, isPlanetary, padSize int) uint32 {
}

func ParseStationJSONL(source io.Reader) (<-chan parsing.EntityPacket, error) {
	registry := make(chan parentCheck, 1)

	go func() {
		defer close(registry)
		if FacilityRegistry != nil {
			defer FacilityRegistry.CloseRegistration()
		}
		stations := parsing.ParseJSONLines(source, getStationFields())
		for station := range stations {
			facilityID := uint32(station[0].Uint())
			systemID := uint32(station[3].Uint())
			data, err := proto.Marshal(&gom.Facility{
				Id:           facilityID,
				Name:         station[1].String(),
				TimestampUtc: station[2].Uint(),
				SystemId:     systemID,
				FacilityType: getFacilityType(station[4].Uint()),
				Features:     getFeatures(station, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15),
				LsFromStar: uint32(station[16].Float()),
				Government: getGovernmentType(station[17].Uint()),
				Allegiance: getAllegianceType(station[18].Uint()),
			})
			if err != nil {
				panic(err)
			} else {
				registry <- parentCheck{parentID: systemID, entity: parsing.EntityPacket{ObjectId: facilityID, Data: data}}
				FacilityRegistry.Register(facilityID)
			}
		}
	}()

	channel := make(chan parsing.EntityPacket, 1)
	if SystemRegistry != nil {
		// Schedule the lookups
		go func() {
			defer SystemRegistry.CloseLookups()
			for check := range registry {
				SystemRegistry.Lookup(check.parentID, check.entity)
			}
		}()
		// Consume the approvals and forward them to channel
		go func() {
			defer close(channel)
			for approved := range SystemRegistry.Approvals() {
				channel <- approved.(parsing.EntityPacket)
			}
		}()
	} else {
		go func() {
			defer close(channel)
			for check := range registry {
				channel <- check.entity.(parsing.EntityPacket)
			}
		}()
	}

	return channel, nil
}
