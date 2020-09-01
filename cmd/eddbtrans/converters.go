package main

import (
	"github.com/kfsone/eddbtrans"
	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"github.com/kfsone/gomenacing/pkg/parsing"
)

func convertCommodities(path string) {
	eddbtrans.ConvertFile(path, "commodities.json", "commodities.gom", gom.Header_CCommodity, eddbtrans.ParseCommodityJson, func(item parsing.EntityPacket) {
		eddbtrans.RegisterCommodity(uint32(item.ObjectId))
	})
}

func convertListings(path string) {
	eddbtrans.ConvertFile(path, "listings.csv", "listings.gom", gom.Header_CListing, eddbtrans.ParseListingsCSV, func(item parsing.EntityPacket) {
	})
}

func convertSystems(path string) {
	eddbtrans.ConvertFile(path, "systems_populated.jsonl", "systems.gom", gom.Header_CSystem, eddbtrans.ParseSystemsPopulatedJSONL, nil)
}

func convertStations(path string) {
	eddbtrans.ConvertFile(path, "stations.jsonl", "stations.gom", gom.Header_CFacility, eddbtrans.ParseStationJSONL, nil)
}
