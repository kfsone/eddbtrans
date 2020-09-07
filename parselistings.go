package eddbtrans

import (
	"io"
	"log"
	"strconv"

	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"github.com/kfsone/gomenacing/pkg/parsing"
	"google.golang.org/protobuf/proto"
)

type FacilityListings map[uint32]*gom.FacilityListing

func convertValues(from [][]byte, into []uint64) (err error) {
	for idx, value := range from {
		into[idx], err = strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return err
		}
	}
	return nil
}

func registerListing(stationID uint32, listing *gom.CommodityListing, facilityListings FacilityListings) {
	// The current maximum number of commodities any station has listed.
	const maxCommodities = 131
	listings, exists := facilityListings[stationID]
	if exists == false {
		listings = &gom.FacilityListing{
			Id:       stationID,
			Listings: make([]*gom.CommodityListing, 0, maxCommodities),
		}
		facilityListings[stationID] = listings
	}

	listings.Listings = append(listings.Listings, listing)
}

func registerCommodityListing(row []uint64, facilityListings FacilityListings) bool {
	// Check station ID for 0 and truncation
	stationID := uint32(row[0])
	if stationID == 0 || uint64(stationID) != row[0] {
		return false
	}

	id := uint32(row[1])
	if CommodityExists(id) == false {
		return false
	}

	listing := &gom.CommodityListing{
		CommodityId:   id,
		SupplyUnits:   uint32(row[2]),
		SupplyCredits: uint32(row[3]),
		DemandUnits:   uint32(row[4]),
		DemandCredits: uint32(row[5]),
		TimestampUtc:  row[6],
	}

	registerListing(stationID, listing, facilityListings)

	return true
}

func ParseListingsCSV(source io.Reader) (<-chan parsing.EntityPacket, error) {
	// We'll marshal all the listings for a station together and writ t
	listings := make(FacilityListings, 80000)

	// channel we'll use to ask daycare if stations are registered
	marshalling := make(chan parentCheck)
	go func() {
		defer close(marshalling)

		count := 0
		rows, err := parsing.ParseCSV(source, getListingFields())
		row := make([]uint64, len(getListingFields()))
		ErrIsBad(err)
		for columns := range rows {
			err = convertValues(columns, row)
			if err != nil {
				continue
			}
			if registerCommodityListing(row, listings) {
				count++
			}
		}

		log.Printf("Parsed %d commodity listings, processing.", count)

		for stationID, listing := range listings {
			marshalling <- parentCheck{
				parentID: stationID,
				entity:   listing,
			}
		}
	}()

	// Marshall incoming stations. It's likely to take a few cycles so worth concurrency.
	registry := make(chan parentCheck)
	go func() {
		defer close(registry)
		for incoming := range marshalling {
			data, err := proto.Marshal(incoming.entity.(*gom.FacilityListing))
			if err != nil {
				log.Printf("Marshalling err for station %d", incoming.parentID)
				continue
			}
			registry <- parentCheck{parentID: incoming.parentID, entity: parsing.EntityPacket{ObjectId: incoming.parentID, Data: data}}
		}
	}()

	channel := make(chan parsing.EntityPacket, 1)
	if FacilityRegistry != nil {
		// Schedule the lookups
		go func() {
			defer FacilityRegistry.CloseLookups()
			for check := range registry {
				FacilityRegistry.Lookup(check.parentID, check.entity)
			}
		}()
		// Consume the approvals and forward them to channel
		go func() {
			defer close(channel)
			for approved := range FacilityRegistry.Approvals() {
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
