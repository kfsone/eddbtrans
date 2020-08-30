package main

import (
	"github.com/kfsone/eddbtrans"
	"github.com/kfsone/gomenacing/pkg/parsing"
)

func convertCommodities(path string) {
	eddbtrans.ConvertFile(path, "commodities.json", "commodities.gom", eddbtrans.ParseCommodityJson, func (item ettudata.EntityPacket) {
		eddbtrans.RegisterCommodity(uint32(item.ObjectId))
	})
}

func convertListings(path string) {
	eddbtrans.ConvertFile(path, "listings.csv", "listings.gom", eddbtrans.ParseListingsCSV, func (item ettudata.EntityPacket) {
	})
}

func convertSystems(path string) {
	eddbtrans.ConvertFile(path, "systems_populated.jsonl", "systems.gom", eddbtrans.ParseSystemsPopulatedJSONL, nil)
}

func convertStations(path string) {
	eddbtrans.ConvertFile(path, "stations.jsonl", "stations.gom", eddbtrans.ParseStationJSONL, nil)
}

