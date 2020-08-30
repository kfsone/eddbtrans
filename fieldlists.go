// Specifies which EDDB fields we want to lift from each source,
// and the order we want them to be in when parsing returns them.
//
//+build !test

package eddbtrans

func getSystemFields() []string {
	return []string{
		"id", "name", "updated_at",
		"x", "y", "z",
		"is_populated", "needs_permit",
		"government_id", "allegiance_id",
		"security_id", "power_state_id",
		"edsm_id", "ed_system_address",
	}
}

func getStationFields() []string {
	return []string{
		"id", "name", "updated_at", "system_id",

		"type_id",

		"has_blackmarket",
		"has_market",
		"has_refuel",
		"has_repair",
		"has_rearm",
		"has_outfitting",
		"has_shipyard",
		"has_docking",
		"has_commodities",
		"is_planetary",
		"max_landing_pad_size",

		"distance_to_star",
		"government_id", "allegiance_id",

		"ed_market_id",
	}
}

func getListingFields() []string {
	return []string{
		"station_id", "commodity_id",
		"supply_bracket", "supply", "sell_price",
		"demand_bracket", "demand", "buy_price",
		"collected_at",
	}
}
