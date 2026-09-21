//go:build test

package domain

import "testing"

func TestFrenchTarotBidKeyAllBranches(t *testing.T) {
	tests := []struct {
		bid  FrenchTarotBid
		want string
	}{
		{FrenchTarotBidPetite, "frenchtarot.bidPetite"},
		{FrenchTarotBidGarde, "frenchtarot.bidGarde"},
		{FrenchTarotBidGardeSans, "frenchtarot.bidGardeSans"},
		{FrenchTarotBidGardeContre, "frenchtarot.bidGardeContre"},
		{FrenchTarotBidPass, "frenchtarot.bidPass"},
	}
	for _, tt := range tests {
		if got := frenchTarotBidKey(tt.bid); got != tt.want {
			t.Errorf("frenchTarotBidKey(%d) = %q, want %q", tt.bid, got, tt.want)
		}
	}
}

func TestCalabresellaBidKeyAllBranches(t *testing.T) {
	tests := []struct {
		bid  CalabresellaBid
		want string
	}{
		{CalabresellaBidChiamo, "calabresella.bidChiamo"},
		{CalabresellaBidSolo, "calabresella.bidSolo"},
		{CalabresellaBidNone, "calabresella.bidPass"},
	}
	for _, tt := range tests {
		if got := calabresellaBidKey(tt.bid); got != tt.want {
			t.Errorf("calabresellaBidKey(%d) = %q, want %q", tt.bid, got, tt.want)
		}
	}
}

func TestKoenigrufenBidKeyAllBranches(t *testing.T) {
	if got := koenigrufenBidKey(KoenigrufenBidRufer); got != "koenigrufen.bidRufer" {
		t.Errorf("koenigrufenBidKey(Rufer) = %q", got)
	}
	if got := koenigrufenBidKey(KoenigrufenBidPass); got != "koenigrufen.bidPass" {
		t.Errorf("koenigrufenBidKey(Pass) = %q", got)
	}
}

func TestCegoBidAndContractKeysAllBranches(t *testing.T) {
	if got := cegoBidKey(CegoBidPlay); got != "cego.bidPlay" {
		t.Errorf("cegoBidKey(Play) = %q", got)
	}
	if got := cegoBidKey(CegoBidPass); got != "cego.bidPass" {
		t.Errorf("cegoBidKey(Pass) = %q", got)
	}

	tests := []struct {
		contract CegoContract
		want     string
	}{
		{CegoContractCego, "cego.contractCego"},
		{CegoContractHandspiel, "cego.contractHandspiel"},
		{CegoContractNone, "cego.contractNone"},
	}
	for _, tt := range tests {
		if got := cegoContractKey(tt.contract); got != tt.want {
			t.Errorf("cegoContractKey(%d) = %q, want %q", tt.contract, got, tt.want)
		}
	}
}

func TestColourWhistContractKeyAllBranches(t *testing.T) {
	tests := []struct {
		contract int
		want     string
	}{
		{ColourWhistContractSamen, "colourwhist.contractShort.samen"},
		{ColourWhistContractAlleen, "colourwhist.contractShort.alleen"},
		{ColourWhistContractMiserie, "colourwhist.contractShort.miserie"},
		{ColourWhistContractTroel, "colourwhist.contractShort.troel"},
		{ColourWhistContractNone, "colourwhist.contractShort.none"},
	}
	for _, tt := range tests {
		if got := ColourWhistContractKey(tt.contract); got != tt.want {
			t.Errorf("ColourWhistContractKey(%d) = %q, want %q", tt.contract, got, tt.want)
		}
	}
}
