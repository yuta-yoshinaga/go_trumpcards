//go:build test

package domain

import (
	"testing"
)

func TestActionLogNewKeyHelpers(t *testing.T) {
	contracts := []struct {
		contract SchafkopfContract
		want     string
	}{
		{SchafkopfContractWenz, "schafkopf.contractShort.wenz"},
		{SchafkopfContractSolo, "schafkopf.contractShort.solo"},
		{SchafkopfContractRufspiel, "schafkopf.contractShort.rufspiel"},
	}
	for _, tt := range contracts {
		if got := schafkopfContractKey(tt.contract); got != tt.want {
			t.Errorf("schafkopfContractKey(%d) = %q, want %q", tt.contract, got, tt.want)
		}
	}

	directions := []struct {
		dir  int
		want string
	}{
		{BidWhistDirectionUptown, "bidwhist.dirUptown"},
		{BidWhistDirectionDowntown, "bidwhist.dirDowntown"},
		{BidWhistDirectionNoTrump, "bidwhist.dirNoTrump"},
		{-1, "bidwhist.dirUnknown"},
	}
	for _, tt := range directions {
		if got := bidWhistDirectionKey(tt.dir); got != tt.want {
			t.Errorf("bidWhistDirectionKey(%d) = %q, want %q", tt.dir, got, tt.want)
		}
	}

	sources := []struct {
		src  RussianBankSource
		want string
	}{
		{RussianBankSource{Zone: RussianBankZoneReserve}, "russianbank.srcReserve"},
		{RussianBankSource{Zone: RussianBankZoneReserve, FromOpponent: true}, "russianbank.srcOppReserve"},
		{RussianBankSource{Zone: RussianBankZoneWaste}, "russianbank.srcWaste"},
		{RussianBankSource{Zone: RussianBankZoneWaste, FromOpponent: true}, "russianbank.srcOppWaste"},
	}
	for _, tt := range sources {
		if got := rbSourceKey(tt.src); got != tt.want {
			t.Errorf("rbSourceKey(%+v) = %q, want %q", tt.src, got, tt.want)
		}
	}
	if got := rbSourceTableauCode(RussianBankSource{Zone: RussianBankZoneTableau}, "foundation"); got != "russianbank.log.toFoundationTableau" {
		t.Errorf("tableau foundation code = %q", got)
	}
	if got := rbSourceTableauCode(RussianBankSource{Zone: RussianBankZoneTableau}, "tableau"); got != "russianbank.log.toTableauTableau" {
		t.Errorf("tableau tableau code = %q", got)
	}
}

func TestQuodlibetPointsStrUsesPlayerNames(t *testing.T) {
	q := NewDefaultQuodlibet()
	got := quodlibetPointsStr(q, map[int]int{0: 3, 1: -1, 2: 0, 3: 5})
	if got != "You 3 / CPU 1 -1 / CPU 2 0 / CPU 3 5" {
		t.Fatalf("quodlibetPointsStr() = %q", got)
	}
	if got == "[p0=3 p1=-1 p2=0 p3=5]" {
		t.Fatal("debug point identifiers must not be emitted")
	}
}
