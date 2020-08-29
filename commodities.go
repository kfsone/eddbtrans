package eddbtrans

// commodities is a list of known commodity ids.
var commodities = make(map[uint32]bool, 500)

// Register commodity registers that a given id is valid.
func RegisterCommodity(id uint32) {
	commodities[id] = true
}

// CommodityExists checks that the id given is valid for a known Commodity.
func CommodityExists(id uint32) (exists bool) {
	_, exists = commodities[id]
	return
}
