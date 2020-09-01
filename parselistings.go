package eddbtrans

import (
	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"github.com/kfsone/gomenacing/pkg/parsing"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"strconv"
)

func convertValues(from [][]byte, into []uint64) (err error) {
	for idx, value := range from {
		into[idx], err = strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseListingsCSV(source io.Reader) (<-chan parsing.EntityPacket, error) {
	// We'll marshal all the listings for a station together and writ t
	facilityListings := make(map[uint32]*gom.FacilityListing)

	// channel we'll use to ask daycare if stations are registered
	marshalling := make(chan parentCheck)
	go func() {
		defer close(marshalling)
		listingCount, badValuesCount, badCommodityCount := 0, 0, 0

		listings, err := parsing.ParseCSV(source, getListingFields())
		row := make([]uint64, len(getListingFields()))
		ErrIsBad(err)
		for listing := range listings {
			listingCount++
			err = convertValues(listing, row)
			if err != nil {
				badValuesCount++
				continue
			}
			stationID := uint32(row[0])
			id := uint32(row[1])
			if CommodityExists(id) == false {
				badCommodityCount++
			}
			commodityListing := &gom.CommodityListing{
				CommodityId:   id,
				SupplyUnits:   uint32(row[2]),
				SupplyCredits: uint32(row[3]),
				DemandUnits:   uint32(row[4]),
				DemandCredits: uint32(row[5]),
				TimestampUtc:  row[6],
			}
			facilityListing, exists := facilityListings[stationID]
			if exists == false {
				facilityListing = &gom.FacilityListing{
					Id:       stationID,
					Listings: make([]*gom.CommodityListing, 16),
				}
				facilityListings[stationID] = facilityListing
			}
			facilityListing.Listings = append(facilityListing.Listings, commodityListing)
		}

		for stationID, facilityListing := range facilityListings {
			marshalling <- parentCheck{
				parentID: stationID,
				entity:   facilityListing,
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
