package eddbtrans

import (
	"io"

	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"github.com/kfsone/gomenacing/pkg/parsing"

	"google.golang.org/protobuf/proto"
)

// Parse the systems_populated.jsonl file

func getAllegianceType(jsonId uint64) gom.AllegianceType {
	switch jsonId {
	case 1:
		return gom.AllegianceType_AllegAlliance

	case 2:
		return gom.AllegianceType_AllegEmpire

	case 3:
		return gom.AllegianceType_AllegFederation

	case 4:
		return gom.AllegianceType_AllegIndependent

	case 7:
		return gom.AllegianceType_AllegPilotsFederation

	default:
		return gom.AllegianceType_AllegNone
	}
}

func getGovernmentType(jsonId uint64) gom.GovernmentType {
	switch jsonId {
	case 16:
		return gom.GovernmentType_GovAnarchy

	case 32:
		return gom.GovernmentType_GovCommunism

	case 48:
		return gom.GovernmentType_GovConfederacy

	case 64:
		return gom.GovernmentType_GovCorporate

	case 80:
		return gom.GovernmentType_GovCooperative

	case 96:
		return gom.GovernmentType_GovDemocracy

	case 112:
		return gom.GovernmentType_GovDictatorship

	case 128:
		return gom.GovernmentType_GovFeudal

	case 144:
		return gom.GovernmentType_GovPatronage

	case 150:
		return gom.GovernmentType_GovPrisonColony

	case 160:
		return gom.GovernmentType_GovTheocracy

	case 208:
		return gom.GovernmentType_GovPrison

	default:
		return gom.GovernmentType_GovNone
	}
}

func getSecurityType(jsonId uint64) gom.SecurityLevel {
	switch jsonId {
	case 16:
		return gom.SecurityLevel_SecurityLow
	case 32:
		return gom.SecurityLevel_SecurityMedium
	case 48:
		return gom.SecurityLevel_SecurityHigh
	case 64:
		return gom.SecurityLevel_SecurityAnarchy
	default:
		return gom.SecurityLevel_SecurityNone
	}
}

// SystemRegistry will provide system-id checking for facilities.
var SystemRegistry *Daycare

func ParseSystemsPopulatedJSONL(source io.Reader) (<-chan parsing.EntityPacket, error) {
	channel := make(chan parsing.EntityPacket, 1)
	go func() {
		defer close(channel)
		if SystemRegistry != nil {
			defer SystemRegistry.CloseRegistration()
		}

		systems := parsing.ParseJSONLines(source, getSystemFields())
		for systemJson := range systems {
			systemId := uint32(systemJson[0].Uint())
			data, err := proto.Marshal(&gom.System{
				Id:            systemId,
				Name:          systemJson[1].String(),
				TimestampUtc:  systemJson[2].Uint(),
				Position:      &gom.Coordinate{X: systemJson[3].Float(), Y: systemJson[4].Float(), Z: systemJson[5].Float()},
				Populated:     systemJson[6].Bool(),
				NeedsPermit:   systemJson[7].Bool(),
				SecurityLevel: getSecurityType(systemJson[8].Uint()),
				Government:    getGovernmentType(systemJson[9].Uint()),
				Allegiance:    getAllegianceType(systemJson[10].Uint()),
			})
			if err != nil {
				panic(err)
			} else {
				channel <- parsing.EntityPacket{ObjectId: systemId, Data: data}
				if SystemRegistry != nil {
					SystemRegistry.Register(systemId)
				}
			}
		}
	}()

	return channel, nil
}
