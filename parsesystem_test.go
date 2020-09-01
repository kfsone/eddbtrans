package eddbtrans

// import (
// 	. "github.com/kfsone/gomenacing/pkg/gomschema"
// 	"testing"
// )

// func TestParseSystemsPopulatedJsonl(t *testing.T) {
// }

// func Test_getAllegianceType(t *testing.T) {
// 	type args struct {
// 		jsonId uint64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want Allegiance_Type
// 	}{
// 		{"Default", args{0}, Allegiance_None},
// 		{"Alliance", args{1}, Allegiance_Alliance},
// 		{"Empire", args{2}, Allegiance_Empire},
// 		{"Federation", args{3}, Allegiance_Federation},
// 		{"Independent", args{4}, Allegiance_Independent},
// 		{"Unused 5", args{5}, Allegiance_None},
// 		{"Pilot's Federation", args{7}, Allegiance_PilotsFederation},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := getAllegianceType(tt.args.jsonId); got != tt.want {
// 				t.Errorf("getAllegianceType() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_getGovernmentType(t *testing.T) {
// 	type args struct {
// 		jsonId uint64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want Government_Type
// 	}{
// 		{"Default", args{0}, Government_None},
// 		{"Anarchy", args{16}, Government_Anarchy},
// 		{"Communism", args{32}, Government_Communism},
// 		{"Confederacy", args{48}, Government_Confederacy},
// 		{"Corporate", args{64}, Government_Corporate},
// 		{"Cooperative", args{80}, Government_Cooperative},
// 		{"Democracy", args{96}, Government_Democracy},
// 		{"Dictatorship", args{112}, Government_Dictatorship},
// 		{"Feudal", args{128}, Government_Feudal},
// 		{"Patronage", args{144}, Government_Patronage},
// 		{"Prison Colony", args{150}, Government_PrisonColony},
// 		{"Theocracy", args{160}, Government_Theocracy},
// 		{"Prison", args{208}, Government_Prison},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := getGovernmentType(tt.args.jsonId); got != tt.want {
// 				t.Errorf("getGovernmentType() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_getPowerState(t *testing.T) {
// 	type args struct {
// 		jsonId uint64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want System_Power_State
// 	}{
// 		{"Default", args{0}, System_Power_None},
// 		{"Invalid", args{1}, System_Power_None},
// 		{"Control", args{16}, System_Power_Control},
// 		{"Exploited", args{32}, System_Power_Exploited},
// 		{"Contested", args{48}, System_Power_Contested},
// 		{"Expansion", args{64}, System_Power_Expansion},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := getPowerState(tt.args.jsonId); got != tt.want {
// 				t.Errorf("getPowerState() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_getSecurityType(t *testing.T) {
// 	type args struct {
// 		jsonId uint64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want System_Security_Type
// 	}{
// 		{"Default", args{0}, System_Security_None},
// 		{"Invalid", args{1}, System_Security_None},
// 		{"Low", args{16}, System_Security_Low},
// 		{"Medium", args{32}, System_Security_Medium},
// 		{"High", args{48}, System_Security_High},
// 		{"Anarchy", args{64}, System_Security_Anarchy},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := getSecurityType(tt.args.jsonId); got != tt.want {
// 				t.Errorf("getSecurityType() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
