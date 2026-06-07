//go:build test

package domain

import "testing"

// newTichuPlayState builds a 4-player game already in the play phase with empty
// hands, so tests can assign deterministic hands and exercise specific branches.
func newTichuPlayState() *Tichu {
	players := []*TichuPlayer{
		NewTichuPlayer(true),
		NewTichuPlayer(false),
		NewTichuPlayer(false),
		NewTichuPlayer(false),
	}
	g := NewTichu(NewTrumpCards(0), players, DefaultTichuConfig())
	g.round.phase = TichuPhasePlay
	g.round.lastPlayIdx = -1
	g.round.trickCards = make([]*Card, 0)
	return g
}

func TestTichuDogLeadPassesToPartner(t *testing.T) {
	g := newTichuPlayState()
	// give everyone two cards so the deal does not end mid-test
	for i := 0; i < TichuPlayerCnt; i++ {
		g.players[i].AddCard(tcNorm(7, CardDesignSpade))
		g.players[i].AddCard(tcNorm(8, CardDesignHeart))
	}
	// human (P0) leads the dog
	g.players[0].AddCard(tcDog())
	g.round.currentTurn = 0
	dogIdx := g.players[0].GetCardsSize() - 1
	if err := g.PlayerPlay([]int{dogIdx}); err != nil {
		t.Fatalf("playing dog failed: %v", err)
	}
	if g.round.currentTurn != 2 {
		t.Errorf("dog should pass lead to partner P2, got P%d", g.round.currentTurn)
	}
	if g.round.tableCombo != nil {
		t.Error("table should be cleared after a dog lead")
	}
}

func TestTichuDragonTrickGoesToOpponent(t *testing.T) {
	g := newTichuPlayState()
	for i := 0; i < TichuPlayerCnt; i++ {
		g.players[i].AddCard(tcNorm(3, CardDesignSpade))
		g.players[i].AddCard(tcNorm(4, CardDesignHeart))
	}
	g.players[0].AddCard(tcDragon())
	g.round.currentTurn = 0
	dragonIdx := g.players[0].GetCardsSize() - 1
	if err := g.PlayerPlay([]int{dragonIdx}); err != nil {
		t.Fatalf("playing dragon failed: %v", err)
	}
	// CPUs cannot beat the dragon (no bombs) and pass; the trick resolves to an opponent.
	for !g.GetGameEndFlag() && !g.IsHumanTurn() {
		g.CpuPlay()
	}
	if TichuCardsPoints(g.players[1].GetCollected()) < 25 {
		t.Errorf("dragon trick should go to opponent P1, P1 points=%d", TichuCardsPoints(g.players[1].GetCollected()))
	}
}

func TestTichuOneTwoScoring(t *testing.T) {
	g := newTichuPlayState()
	g.round.finishOrder = []int{0, 2} // partners take 1st and 2nd
	g.players[0].SetDeclType(TichuDeclTichu)
	g.players[0].SetRank(1)
	g.endDeal(true)
	if !g.GetIsOneTwo() {
		t.Error("expected one-two flag")
	}
	// 200 (one-two) + 100 (successful Tichu by the first finisher) = 300
	if g.round.scores[0] != 300 {
		t.Errorf("team A score = %d, want 300", g.round.scores[0])
	}
	if g.round.scores[1] != 0 {
		t.Errorf("team B score = %d, want 0", g.round.scores[1])
	}
}

func TestTichuNormalScoringSettlesLastPlayer(t *testing.T) {
	g := newTichuPlayState()
	g.players[0].AddCollected([]*Card{tcNorm(10, CardDesignSpade)}) // 10 to team A
	g.players[3].AddCollected([]*Card{tcNorm(13, CardDesignHeart)}) // 10, goes to first finisher
	g.players[3].AddCard(tcNorm(5, CardDesignSpade))                // last player's hand (5) goes to opponents
	g.round.finishOrder = []int{0, 1, 2}
	g.endDeal(false)
	// team A (P0/P2): 10 own + 10 (P3 tricks → P0) + 5 (P3 hand → opponent P0) = 25
	if g.round.scores[0] != 25 {
		t.Errorf("team A score = %d, want 25", g.round.scores[0])
	}
	if g.round.scores[1] != 0 {
		t.Errorf("team B score = %d, want 0", g.round.scores[1])
	}
}

func TestTichuFailedGrandTichuPenalty(t *testing.T) {
	g := newTichuPlayState()
	g.players[1].AddCollected([]*Card{tcNorm(5, CardDesignSpade)})
	g.players[1].SetDeclType(TichuDeclGrand)
	g.players[1].SetRank(3) // declared but did not finish first
	g.round.finishOrder = []int{0, 2, 3}
	g.endDeal(false)
	// team B = P1's 5 card points minus the failed grand tichu (-200) = -195
	if g.round.scores[1] >= 0 {
		t.Errorf("failed grand tichu should make team B negative, got %d", g.round.scores[1])
	}
}

func TestTichuCpuFindPlayBranches(t *testing.T) {
	g := newTichuPlayState()
	g.round.currentTurn = 1
	g.players[1].AddCard(tcNorm(5, CardDesignSpade))
	g.players[1].AddCard(tcNorm(9, CardDesignHeart))

	// leading: returns the weakest single
	if idx := g.cpuFindPlay(g.players[1]); len(idx) != 1 || idx[0] != 0 {
		t.Errorf("leading play = %v, want [0]", idx)
	}

	// following an opponent: beat the table single
	g.round.tableCombo = &TichuCombo{Type: TichuComboSingle, Rank: 4}
	g.round.lastPlayIdx = 0 // P0 is an opponent of P1
	if idx := g.cpuFindPlay(g.players[1]); len(idx) == 0 {
		t.Error("CPU should beat a low single it can top")
	}

	// table owned by the partner: pass
	g.round.lastPlayIdx = 3 // P3 is P1's partner
	if idx := g.cpuFindPlay(g.players[1]); len(idx) != 0 {
		t.Errorf("CPU should pass when the partner controls the table, got %v", idx)
	}
}

func TestTichuCpuBeatPairAndBomb(t *testing.T) {
	g := newTichuPlayState()
	g.round.currentTurn = 1
	g.players[1].AddCard(tcNorm(9, CardDesignSpade))
	g.players[1].AddCard(tcNorm(9, CardDesignHeart))
	g.round.tableCombo = &TichuCombo{Type: TichuComboPair, Rank: 5}
	g.round.lastPlayIdx = 0
	if idx := g.cpuFindPlay(g.players[1]); len(idx) != 2 {
		t.Errorf("CPU should beat a low pair with a higher pair, got %v", idx)
	}

	// bomb to capture a high-value trick
	g2 := newTichuPlayState()
	g2.round.currentTurn = 1
	for _, d := range []int{CardDesignSpade, CardDesignHeart, CardDesignClover, CardDesignDiamond} {
		g2.players[1].AddCard(tcNorm(8, d))
	}
	g2.round.tableCombo = &TichuCombo{Type: TichuComboSingle, Rank: tichuDragonRank}
	g2.round.lastPlayIdx = 0
	g2.round.trickCards = []*Card{tcNorm(10, CardDesignSpade), tcNorm(13, CardDesignHeart)} // 20 points
	if idx := g2.cpuFindPlay(g2.players[1]); len(idx) != 4 {
		t.Errorf("CPU should bomb a high-value trick, got %v", idx)
	}
}

func TestTichuCpuDeclareStrongHand(t *testing.T) {
	g := newTichuPlayState()
	g.players[1].AddCard(tcDragon())
	g.players[1].AddCard(tcNorm(1, CardDesignSpade)) // ace
	g.players[1].AddCard(tcNorm(1, CardDesignHeart)) // ace
	if g.cpuDeclare(1) != TichuDeclTichu {
		t.Error("CPU with dragon + 2 aces should declare Tichu on normal difficulty")
	}
	g.config.CpuDifficulty = TichuDifficultyEasy
	if g.cpuDeclare(1) != TichuDeclNone {
		t.Error("easy CPU should not declare")
	}
}
