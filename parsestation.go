package eddbtrans

import (
	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"github.com/kfsone/gomenacing/pkg/parsing"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
	"io"
)

func getFacilityType(typeId uint64) gom.FacilityType {
	switch typeId {
	case 1:
		return gom.FacilityType_FTCivilianOutpost
	case 2:
		return gom.FacilityType_FTCommercialOutpost
	case 3:
		return gom.FacilityType_FTCoriolisStarport
	case 4:
		return gom.FacilityType_FTIndustrialOutpost
	case 5:
		return gom.FacilityType_FTMilitaryOutpost
	case 6:
		return gom.FacilityType_FTMiningOutpost
	case 7:
		return gom.FacilityType_FTOcellusStarport
	case 8:
		return gom.FacilityType_FTOrbisStarport
	case 9:
		return gom.FacilityType_FTScientificOutpost
	case 13:
		return gom.FacilityType_FTPlanetaryOutpost
	case 14:
		return gom.FacilityType_FTPlanetaryPort
	case 19:
		return gom.FacilityType_FTMegaship
	case 20:
		return gom.FacilityType_FTAsteroidBase
	case 24:
		return gom.FacilityType_FTFleetCarrier
	default:
		return gom.FacilityType_FTNone
	}
}

// FacilityRegistry will provide facility-id checking for listings.
var FacilityRegistry *Daycare

func maskForBit(bit gom.FeatureBit, basedOn bool) uint32 {
	var value uint32
	if basedOn {
		value = 1
	}
	return value << bit
}

func getFeatures(row []*gjson.Result) uint32 {
	var mask uint32
	mask |= maskForBit(gom.FeatureBit_Market, row[5].Bool())
	mask |= maskForBit(gom.FeatureBit_BlackMarket, row[6].Bool())
	mask |= maskForBit(gom.FeatureBit_Refuel, row[7].Bool())
	mask |= maskForBit(gom.FeatureBit_Repair, row[8].Bool())
	mask |= maskForBit(gom.FeatureBit_Rearm, row[9].Bool())
	mask |= maskForBit(gom.FeatureBit_Outfitting, row[10].Bool())
	mask |= maskForBit(gom.FeatureBit_Shipyard, row[11].Bool())
	mask |= maskForBit(gom.FeatureBit_Docking, row[12].Bool())
	mask |= maskForBit(gom.FeatureBit_Commodities, row[13].Bool())
	mask |= maskForBit(gom.FeatureBit_Planetary, row[14].Bool())
	if len(row[15].String()) != 0 {
		pad := row[15].String()[0]
		mask |= maskForBit(gom.FeatureBit_SmallPad, pad == 'S')
		mask |= maskForBit(gom.FeatureBit_MediumPad, pad == 'M')
		mask |= maskForBit(gom.FeatureBit_LargePad, pad == 'L')
	}
	return mask
}

func ParseStationJSONL(source io.Reader) (<-chan parsing.EntityPacket, error) {
	registry := make(chan parentCheck, 1)

	go func() {
		defer close(registry)
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
				Features:     getFeatures(station),
				LsFromStar: uint32(station[16].Float()),
				Government: getGovernmentType(station[17].Uint()),
				Allegiance: getAllegianceType(station[18].Uint()),
			})
			if err != nil {
				panic(err)
			} else {
				registry <- parentCheck{parentID: systemID, entity: parsing.EntityPacket{ObjectId: facilityID, Data: data}}
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
			defer FacilityRegistry.CloseRegistration()
			for approved := range SystemRegistry.Approvals() {
				channel <- approved.(parsing.EntityPacket)
				FacilityRegistry.Register(approved.(parsing.EntityPacket).ObjectId)
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
