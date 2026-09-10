//go:build test

package domain

import "testing"

func TestTappTarockDealUsesThreeHandsOfSixteenAndSixCardTalon(t *testing.T) {
	g := NewDefaultTappTarock()
	g.Reset()
	if g.GetPlayerCnt() != TappTarockPlayerCnt || TappTarockPlayerCnt != 3 {
		t.Fatalf("expected three players, got %d", g.GetPlayerCnt())
	}
	for i := 0; i < TappTarockPlayerCnt; i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != TappTarockHandSize {
			t.Errorf("player %d has %d cards", i, got)
		}
	}
	if got := g.GetTalonSize(); got != TappTarockTalonSize {
		t.Errorf("talon has %d cards", got)
	}
}

func TestTappTarockPlaysSixteenTricks(t *testing.T) {
	players := []*TappTarockPlayer{NewTappTarockPlayer(true), NewTappTarockPlayer(false), NewTappTarockPlayer(false)}
	g := NewTappTarock(players, DefaultTappTarockConfig())
	deck := buildKoenigrufenDeck()
	for i, c := range deck[:TappTarockPlayerCnt*TappTarockHandSize] {
		players[i%TappTarockPlayerCnt].AddCard(c)
	}
	g.startPlay()
	for trick := 0; trick < TappTarockTrickCount && g.GetPhase() == TappTarockPhasePlay; trick++ {
		for seat := 0; seat < TappTarockPlayerCnt; seat++ {
			idxs := g.GetValidPlayIndices(g.currentPlayerIdx)
			if len(idxs) == 0 {
				t.Fatalf("no legal card at trick %d seat %d", trick, seat)
			}
			g.playCard(g.currentPlayerIdx, idxs[0])
		}
		g.NextTrick()
	}
	totalTricks := 0
	for i := 0; i < TappTarockPlayerCnt; i++ {
		totalTricks += g.GetPlayer(i).GetTrickCount()
	}
	for i := 0; i < TappTarockPlayerCnt; i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != 0 {
			t.Fatalf("player %d still has %d cards after the round", i, got)
		}
	}
	if totalTricks != TappTarockTrickCount || g.GetPhase() != TappTarockPhaseRoundEnd {
		t.Fatalf("expected %d completed tricks and round end, got %d/%d", TappTarockTrickCount, totalTricks, g.GetPhase())
	}
}

func TestTappTarockTrullBonusRequiresAllThreeTrullCards(t *testing.T) {
	g := NewDefaultTappTarock()
	trull := []*Card{
		NewCard(KoenigrufenTrumpDesign, KoenigrufenPagatValue, false),
		NewCard(KoenigrufenTrumpDesign, KoenigrufenMaxTrump, false),
		NewCard(KoenigrufenSkusDesign, KoenigrufenSkusValue, false),
	}
	for _, c := range trull {
		g.players[0].AddTrick([]*Card{c})
	}
	if got, want := g.GetCardPoints(0), 15+TappTarockTrullBonus; got != want {
		t.Fatalf("trull score = %d, want %d", got, want)
	}
}

func TestTappTarockContractOutcomeChangesWithDeclarerMajority(t *testing.T) {
	winning := NewDefaultTappTarock()
	winning.contract = TappTarockBidDreier
	winning.declarerIdx = 0
	winning.players[0].AddTrick(buildKoenigrufenDeck())
	if breakdown := winning.scoreContract(); !breakdown.Won || breakdown.Seats[0] <= 0 {
		t.Fatalf("expected declarer majority to win, got %+v", breakdown)
	}

	losing := NewDefaultTappTarock()
	losing.contract = TappTarockBidDreier
	losing.declarerIdx = 0
	losing.players[1].AddTrick(buildKoenigrufenDeck())
	if breakdown := losing.scoreContract(); breakdown.Won || breakdown.Seats[0] >= 0 {
		t.Fatalf("expected declarer minority to lose, got %+v", breakdown)
	}
}
