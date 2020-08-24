//+build !test

package main

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
