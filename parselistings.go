package eddbtrans

import (
	. "github.com/kfsone/gomenacing/pkg/gomschema"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"sort"
	"strconv"
)

// getBracket ensures a uint32 bracket value is mapped to a meaningful
// Listing_Capacity_Bracket enum. It's a simple 1:1 right now but that
// ensures we don't get out-of-bounds values (e.g. 4).
func getBracket(source uint64) Listing_Capacity_Bracket {
	switch source {
	case uint64(Listing_Capacity_Low):
		return Listing_Capacity_Low
	case uint64(Listing_Capacity_Med):
		return Listing_Capacity_Med
	case uint64(Listing_Capacity_High):
		return Listing_Capacity_High
	default:
		return Listing_Capacity_None
	}
}

func reportStats(facilityListings map[uint64]*FacilityListing) {
	// I have a hunch we may want to write out stations when they reach
	// a certain number of listings, so that some of the writes happen
	// before the end of the conversion process.
	//
	// We'll need to make the loader/import functions able to merge
	// records to support that, but it may alter performance by seconds.
	//
	// The data I experimented with from EDDB showed an average of 83
	// listings and a max of 169. If we were to use a "save point" of
	// 85, we would never have more than two writes for a station, and
	// a large number of stations would get written early.
	//
	// Adding a p50 suggested that we'd write more than 50% of the
	// stations if the value was 94. 90 might be a good start.
	//
	// Note to the future: Don't adopt this strategy without having
	// the client keep statistics for itself, so that it can adapt
	// going forward.

	if len(facilityListings) <= 0 {
		return
	}

	// assume the first listing is the least/most, to start with.
	var least *FacilityListing
	var most *FacilityListing
	for _, facility := range facilityListings {
		least, most = facility, facility
		break
	}

	var totalListings = 0
	var meanList = make(map[uint32]uint32, 200)
	var estAverage = 86
	var writesAtAverage = 0
	var counts = make([]uint32, 0, len(facilityListings))
	for _, facility := range facilityListings {
		listed := len(facility.Listings)
		if len(least.Listings) > listed {
			least = facility
		}
		if len(most.Listings) < listed {
			most = facility
		}
		totalListings += listed
		meanList[uint32(listed)]++
		if listed >= estAverage {
			writesAtAverage++
		}
		counts = append(counts, uint32(listed))
	}

	avgListings := 0
	if len(facilityListings) > 0 {
		avgListings = totalListings / len(facilityListings)
	}
	var mean uint32
	for idx, qty := range meanList {
		if qty > meanList[mean] {
			mean = idx
		}
	}

	// Calculate p50: the number of listings that would cause 50% of the stations
	// to have written more than once if we used that number as our "snapshot" count.
	// When the p50 index wouldn't be a whole number, you take the value half way
	// between the value before and the value after.
	sort.SliceStable(counts, func(lhs, rhs int) bool { return counts[rhs] < counts[lhs] })
	p50idx := len(counts) * 50 / 100 // aka * 0.50
	p50val := counts[p50idx]
	if totalListings%2 == 1 { // odd number will do it.
		nextVal := counts[p50idx+1]
		delta := nextVal - p50val
		p50val = p50val + (delta / 2)
	}

	log.Printf("%d listings for %d stations. min/avg/mean/max: %d/%d/%d/%d. p50: %d",
		totalListings, len(facilityListings),
		len(least.Listings), avgListings, mean, len(most.Listings),
		p50val)
}

func convertValues(from [][]byte, into []uint64) (err error) {
	for idx, value := range from {
		into[idx], err = strconv.ParseUint(string(value), 10, 64)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseListingsCSV(source io.Reader) (<-chan EntityPacket, error) {
	// We'll marshal all the listings for a station together and writ t
	facilityListings := make(map[uint64]*FacilityListing)

	// See stats for why this is 94 (based on p50)
	const earlySnapshotPoint = 94

	// channel we'll use to ask daycare if stations are registered
	marshalling := make(chan parentCheck)
	go func() {
		defer close(marshalling)
		listingCount, badValuesCount, badCommodityCount := 0, 0, 0

		listings, err := ParseCSV(source, getListingFields())
		row := make([]uint64, len(getListingFields()))
		ErrIsBad(err)
		for listing := range listings {
			listingCount++
			err = convertValues(listing, row)
			if err != nil {
				badValuesCount++
				continue
			}
			stationID := row[0]
			id := uint32(row[1])
			if CommodityExists(id) == false {
				badCommodityCount++
			}
			listing := &Listing{
				CommodityId: id,
				Supply: &Listing_Capacity{
					Bracket: getBracket(row[2]),
					Units:   uint32(row[3]),
					Credits: uint32(row[4]),
				},
				Demand: &Listing_Capacity{
					Bracket: getBracket(row[5]),
					Units:   uint32(row[6]),
					Credits: uint32(row[7]),
				},
				Collected: row[8],
			}
			facilityListing, exists := facilityListings[row[0]]
			if exists == false {
				facilityListing = &FacilityListing{
					FacilityId: 0,
					Listings:   make([]*Listing, 16),
				}
				facilityListings[row[0]] = facilityListing
			}
			facilityListing.Listings = append(facilityListing.Listings, listing)
			if earlySnapshotPoint > 0 && len(facilityListing.Listings) >= earlySnapshotPoint {
				marshalling <- parentCheck{
					parentID: stationID,
					entity:   facilityListing,
				}
				delete(facilityListings, stationID)
			}
		}

		if earlySnapshotPoint == 0 {
			reportStats(facilityListings)
		}

		// all the remaining facilities need to be checked for station validity.
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
			data, err := proto.Marshal(incoming.entity.(*FacilityListing))
			if err != nil {
				log.Printf("Marshalling err for station %d", incoming.parentID)
				continue
			}
			registry <- parentCheck{parentID: incoming.parentID, entity: EntityPacket{ObjectId: incoming.parentID, Data: data }}
		}
	}()

	channel := make(chan EntityPacket, 1)
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
				channel <- approved.(EntityPacket)
			}
		}()
	} else {
		go func() {
			defer close(channel)
			for check := range registry {
				channel <- check.entity.(EntityPacket)
			}
		}()
	}

	return channel, nil
}
