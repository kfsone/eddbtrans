package eddbtrans

import (
	"io"

	"google.golang.org/protobuf/proto"
)

// Parse the systems_populated.jsonl file

func getAllegianceType(jsonId uint64) Allegiance_Type {
	switch jsonId {
	case 1:
		return Allegiance_Alliance

	case 2:
		return Allegiance_Empire

	case 3:
		return Allegiance_Federation

	case 4:
		return Allegiance_Independent

	case 7:
		return Allegiance_PilotsFederation

	default:
		return Allegiance_None
	}
}

func getGovernmentType(jsonId uint64) Government_Type {
	switch jsonId {
	case 16:
		return Government_Anarchy

	case 32:
		return Government_Communism

	case 48:
		return Government_Confederacy

	case 64:
		return Government_Corporate

	case 80:
		return Government_Cooperative

	case 96:
		return Government_Democracy

	case 112:
		return Government_Dictatorship

	case 128:
		return Government_Feudal

	case 144:
		return Government_Patronage

	case 150:
		return Government_PrisonColony

	case 160:
		return Government_Theocracy

	case 208:
		return Government_Prison

	default:
		return Government_None
	}
}

func getPowerState(jsonId uint64) System_Power_State {
	switch jsonId {
	case 16:
		return System_Power_Control
	case 32:
		return System_Power_Exploited
	case 48:
		return System_Power_Contested
	case 64:
		return System_Power_Expansion
	default:
		return System_Power_None
	}
}

func getSecurityType(jsonId uint64) System_Security_Type {
	switch jsonId {
	case 16:
		return System_Security_Low
	case 32:
		return System_Security_Medium
	case 48:
		return System_Security_High
	case 64:
		return System_Security_Anarchy
	default:
		return System_Security_None
	}
}

func ParseSystemsPopulatedJsonl(source io.Reader) (<-chan Message, error) {
	channel := make(chan Message, 2)
	go func() {
		defer close(channel)
		systems := ParseJSONLines(source, getSystemFields())
		for systemJson := range systems {
			data, err := proto.Marshal(&System{
				Id:              systemJson[0].Uint(),
				Name:            systemJson[1].String(),
				Updated:         systemJson[2].Uint(),
				Position:        &System_Coordinates{X: systemJson[3].Float(), Y: systemJson[4].Float(), Z: systemJson[5].Float()},
				IsPopulated:     systemJson[6].Bool(),
				NeedsPermit:     systemJson[7].Bool(),
				Security:        &System_Security{Type: getSecurityType(systemJson[8].Uint())},
				Power:           &System_Power{State: getPowerState(systemJson[9].Uint())},
				Government:      &Government{Type: getGovernmentType(systemJson[10].Uint())},
				Allegiance:      &Allegiance{Type: getAllegianceType(systemJson[11].Uint())},
				EdsmId:          systemJson[12].Uint(),
				EdSystemAddress: systemJson[13].Uint(),
			})
			if err != nil {
				panic(err)
			} else {
				channel <- Message{systemJson[0].Uint(), data}
			}
		}
	}()

	return channel, nil
}
