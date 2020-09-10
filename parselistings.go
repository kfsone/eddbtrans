package eddbtrans

import (
	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"github.com/kfsone/gomenacing/pkg/parsing"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
)

type FacilityListings map[uint32]*gom.FacilityListing

func aggregateListings(rows <-chan []uint64) <-chan []*[]uint64 {
	channel := make(chan []*[]uint64, 2)
	go func() {
		defer close(channel)
		seen := make(map[uint32]bool, 80000)
		last := uint32(0)

		var aggregate []*[]uint64

		for columns := range rows {
			stationID := uint32(columns[0])
			if stationID == 0 || uint64(stationID) != columns[0] {
				log.Printf("invalid facility id: %d", stationID)
				continue
			}

			id := uint32(columns[1])
			if CommodityExists(id) == false {
				log.Printf("facility %d: unrecognized commodity: %d", stationID, id)
			}

			if stationID != last {
				if seen[stationID] == true {
					log.Fatalf("multiple listings for facility: %d", stationID)
				}
				seen[stationID] = true
				if aggregate != nil {
					channel <- aggregate
				}
				aggregate = make([]*[]uint64, 0, 32)
				last = stationID
			}

			aggregate = append(aggregate, &columns)
		}

		if aggregate != nil {
			channel <- aggregate
		}
	}()
	return channel
}

func convertToListingMessages(aggregates <-chan []*[]uint64) <-chan parentCheck {
	channel := make(chan parentCheck, 1)
	go func() {
		defer close(channel)

		for aggregate := range aggregates {
			stationID := (*aggregate[0])[0]
			// Try and add this commodity to the existing FacilityListing, or generate a new one
			// if we've moved to a new facility.

			listing := &gom.FacilityListing{Id: uint32(stationID), Listings: make([]*gom.CommodityListing, len(aggregate))}
			listings := make([]gom.CommodityListing, len(aggregate))
			for idx, values := range aggregate {
				row := *values
				listings[idx].CommodityId = uint32(row[1])
				listings[idx].SupplyUnits = uint32(row[2])
				listings[idx].SupplyCredits = uint32(row[3])
				listings[idx].DemandUnits =   uint32(row[4])
				listings[idx].DemandCredits = uint32(row[5])
				listings[idx].TimestampUtc =  row[6]
				listing.Listings[idx] = &listings[idx]
			}

			data, err := proto.Marshal(listing)
			if err != nil {
				log.Printf("unable to marshal facility %d: %s", listing.Id, err)
				continue
			}

			channel <- parentCheck{ parentID: listing.Id,
				entity:   parsing.EntityPacket{ObjectId: listing.Id, Data: data},
			}
		}
	}()

	return channel
}

func ParseListingsCSV(source io.Reader) (<-chan parsing.EntityPacket, error) {
	// Convert incoming values to uint64s
	uint64Listings, err := parsing.ParseCSVToUint64s(source, getListingFields())
	if err != nil {
		return nil, err
	}

	// Group commodities together into FacilityListing objects
	aggregates := aggregateListings(uint64Listings)

	// Convert listing aggregates to facility listings
	messages := convertToListingMessages(aggregates)

	channel := make(chan parsing.EntityPacket)
	if FacilityRegistry != nil {
		// Schedule the lookups
		go func() {
			defer FacilityRegistry.CloseLookups()
			for check := range messages {
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
			for check := range messages {
				channel <- check.entity.(parsing.EntityPacket)
			}
		}()
	}

	return channel, nil
}
