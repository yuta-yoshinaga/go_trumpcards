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

func TestBoliviaActionLogKeysAllBranches(t *testing.T) {
	if got := boliviaMeldKindKey(BoliviaMeldEscalera); got != "bolivia.meldTypeSequence" {
		t.Errorf("boliviaMeldKindKey(Escalera) = %q", got)
	}
	if got := boliviaMeldKindKey(BoliviaMeldSet); got != "bolivia.meldTypeSet" {
		t.Errorf("boliviaMeldKindKey(Set) = %q", got)
	}
	if got := boliviaCanastaTypeKey(true); got != "bolivia.meldTypeNatural" {
		t.Errorf("boliviaCanastaTypeKey(true) = %q", got)
	}
	if got := boliviaCanastaTypeKey(false); got != "bolivia.meldTypeMixed" {
		t.Errorf("boliviaCanastaTypeKey(false) = %q", got)
	}
}

func TestSambaActionLogKeysAllBranches(t *testing.T) {
	if got := sambaMeldKindKey(SambaMeldSequence); got != "samba.meldTypeSequence" {
		t.Errorf("sambaMeldKindKey(Sequence) = %q", got)
	}
	if got := sambaMeldKindKey(SambaMeldSet); got != "samba.meldTypeSet" {
		t.Errorf("sambaMeldKindKey(Set) = %q", got)
	}
	if got := sambaCanastaTypeKey(true); got != "samba.meldTypeNatural" {
		t.Errorf("sambaCanastaTypeKey(true) = %q", got)
	}
	if got := sambaCanastaTypeKey(false); got != "samba.meldTypeMixed" {
		t.Errorf("sambaCanastaTypeKey(false) = %q", got)
	}
}

func TestCanastaTypeKeyAllBranches(t *testing.T) {
	if got := canastaTypeKey(true); got != "canasta.meldTypeNatural" {
		t.Errorf("canastaTypeKey(true) = %q", got)
	}
	if got := canastaTypeKey(false); got != "canasta.meldTypeMixed" {
		t.Errorf("canastaTypeKey(false) = %q", got)
	}
}

func TestHandAndFootCanastaTypeKeyAllBranches(t *testing.T) {
	if got := handAndFootCanastaTypeKey(true); got != "handandfoot.meldTypeNatural" {
		t.Errorf("handAndFootCanastaTypeKey(true) = %q", got)
	}
	if got := handAndFootCanastaTypeKey(false); got != "handandfoot.meldTypeMixed" {
		t.Errorf("handAndFootCanastaTypeKey(false) = %q", got)
	}
}

func TestBaccaratBetTypeKeyAllBranches(t *testing.T) {
	tests := []struct {
		betType int
		want    string
	}{
		{BaccaratBetPlayer, "baccarat.sidePlayer"},
		{BaccaratBetBanker, "baccarat.sideBanker"},
		{BaccaratBetTie, "baccarat.sideTie"},
		{99, "baccarat.sideUnknown"},
	}
	for _, tt := range tests {
		if got := betTypeKey(tt.betType); got != tt.want {
			t.Errorf("betTypeKey(%d) = %q, want %q", tt.betType, got, tt.want)
		}
	}
}

func TestAndarBaharColumnKeyAllBranches(t *testing.T) {
	tests := []struct {
		column int
		want   string
	}{
		{AndarBaharBetAndar, "andarbahar.columnAndar"},
		{AndarBaharBetBahar, "andarbahar.columnBahar"},
		{99, "andarbahar.columnUnknown"},
	}
	for _, tt := range tests {
		if got := andarBaharColumnKey(tt.column); got != tt.want {
			t.Errorf("andarBaharColumnKey(%d) = %q, want %q", tt.column, got, tt.want)
		}
	}
}

func TestMusRoundKeyAllBranches(t *testing.T) {
	tests := []struct {
		round int
		want  string
	}{
		{0, "mus.roundGrande"},
		{1, "mus.roundChica"},
		{2, "mus.roundPares"},
		{3, "mus.roundJuego"},
		{99, "mus.roundUnknown"},
	}
	for _, tt := range tests {
		if got := musRoundKey(tt.round); got != tt.want {
			t.Errorf("musRoundKey(%d) = %q, want %q", tt.round, got, tt.want)
		}
	}
}

func TestViraBidKeyAllBranches(t *testing.T) {
	tests := []struct {
		bid  int
		want string
	}{
		{int(ViraBidGask), "vira.bidShort.gask"},
		{int(ViraBidSolo), "vira.bidShort.solo"},
		{int(ViraBidMisere), "vira.bidShort.misere"},
		{int(ViraBidVira), "vira.bidShort.vira"},
		{int(ViraBidPass), "vira.bidShort.pass"},
		{99, "vira.bidShort.pass"},
	}
	for _, tt := range tests {
		if got := ViraBidKey(tt.bid); got != tt.want {
			t.Errorf("ViraBidKey(%d) = %q, want %q", tt.bid, got, tt.want)
		}
	}
}
