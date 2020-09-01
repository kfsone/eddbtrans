package eddbtrans

import (
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

func getPadSize(padSize string) gom.PadSize {
	switch strings.ToUpper(padSize) {
	case "S":
		return gom.PadSize_PadSmall
	case "M":
		return gom.PadSize_PadMedium
	case "L":
		return gom.PadSize_PadLarge
	default:
		return gom.PadSize_PadNone
	}
}

// FacilityRegistry will provide facility-id checking for listings.
var FacilityRegistry *Daycare

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
				Services: &gom.Services{
					HasMarket:      station[5].Bool(),
					HasBlackMarket: station[6].Bool(),
					HasRefuel:      station[7].Bool(),
					HasRepair:      station[8].Bool(),
					HasRearm:       station[9].Bool(),
					HasOutfitting:  station[10].Bool(),
					HasShipyard:    station[11].Bool(),
					HasDocking:     station[12].Bool(),
					HasCommodities: station[13].Bool(),
					IsPlanetary:    station[14].Bool(),
				},
				PadSize:    getPadSize(station[15].String()),
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
