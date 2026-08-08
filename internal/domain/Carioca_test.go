package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

// helperCariocaHand 既定構成の Carioca を Reset 済みで返す。
func helperCariocaHand(t *testing.T) *Carioca {
	t.Helper()
	g := NewDefaultCarioca()
	g.Reset()
	return g
}

// cariocaCard 指定スート・値のカードを生成する
func cariocaCard(design, value int) *Card {
	return NewCard(design, value, true)
}

// cariocaJoker ジョーカー（ワイルド）を生成する（i は 1..4）
func cariocaJoker(i int) *Card {
	return NewCard(CardDesignJoker, i, true)
}

// cariocaSetHand プレイヤーの手札を差し替える
func cariocaSetHand(p *CariocaPlayer, cards []*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNewDefaultCarioca_Construct(t *testing.T) {
	g := NewDefaultCarioca()
	if g == nil {
		t.Fatal("NewDefaultCarioca returned nil")
	}
	if g.GetPlayerCnt() != CariocaDefaultPlayerCount {
		t.Errorf("player count = %d, want %d", g.GetPlayerCnt(), CariocaDefaultPlayerCount)
	}
	if g.GetWinnerIdx() != -1 {
		t.Errorf("initial winnerIdx = %d, want -1", g.GetWinnerIdx())
	}
	if !g.GetPlayer(0).GetIsHuman() {
		t.Error("player 0 should be human")
	}
	if g.GetPlayer(1).GetIsHuman() {
		t.Error("player 1 should be CPU")
	}
}

func TestCarioca_Reset_DealsHand(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		got := g.GetPlayer(i).GetCardsSize()
		if got != CariocaHandSize {
			t.Errorf("player %d hand size = %d, want %d", i, got, CariocaHandSize)
		}
	}
	if g.GetDiscardTop() == nil {
		t.Error("discard top should be set after Reset")
	}
	if g.GetRoundNumber() != 1 {
		t.Errorf("roundNumber = %d, want 1", g.GetRoundNumber())
	}
	if g.GetPhase() != CariocaPhaseDraw {
		t.Errorf("phase = %d, want PhaseDraw", g.GetPhase())
	}
}

func TestCarioca_Reset_RebuildsPlayersFromConfig(t *testing.T) {
	g := NewDefaultCarioca()
	cfg := g.GetConfig()
	cfg.PlayerCount = 6
	g.SetConfig(cfg)
	g.Reset()
	if g.GetPlayerCnt() != 6 {
		t.Errorf("player count after reset = %d, want 6", g.GetPlayerCnt())
	}
	cfg.PlayerCount = 3
	g.SetConfig(cfg)
	g.Reset()
	if g.GetPlayerCnt() != 3 {
		t.Errorf("player count after reset = %d, want 3", g.GetPlayerCnt())
	}
}

func TestCarioca_DeckHas108Cards(t *testing.T) {
	deck := newCariocaDeck()
	if got := deck.GetTotalCount(); got != 108 {
		t.Errorf("carioca deck size = %d, want 108", got)
	}
}

func TestCarioca_NextRound_AdvancesAndDeals(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	g.SetPhase(CariocaPhaseRoundEnd)
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("roundNumber after NextRound = %d, want 2", g.GetRoundNumber())
	}
	if g.GetPhase() != CariocaPhaseDraw {
		t.Errorf("phase = %d, want PhaseDraw", g.GetPhase())
	}
}

func TestCarioca_NextRound_NoOpWhenNotInRoundEnd(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	original := g.GetRoundNumber()
	g.NextRound()
	if g.GetRoundNumber() != original {
		t.Errorf("NextRound during Draw must not advance round")
	}
}

func TestCarioca_NextRound_AfterFinalRoundEndsGame(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	g.SetRoundNumber(CariocaTotalRounds)
	g.SetPhase(CariocaPhaseRoundEnd)
	g.NextRound()
	if !g.GetGameEndFlag() {
		t.Error("game should end after final round")
	}
	if g.GetWinnerIdx() < 0 {
		t.Error("winner should be set")
	}
}

func TestCarioca_PlayerDrawFromStock_ProgressesPhase(t *testing.T) {
	g := helperCariocaHand(t)
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("PlayerDrawFromStock error: %v", err)
	}
	if g.GetPhase() != CariocaPhasePlay {
		t.Errorf("phase = %d, want PhasePlay", g.GetPhase())
	}
	if g.GetPlayer(0).GetCardsSize() != CariocaHandSize+1 {
		t.Errorf("hand should grow by 1 after draw")
	}
}

func TestCarioca_PlayerDrawFromStock_RejectsWrongPhase(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetPhase(CariocaPhasePlay)
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("expected ErrWrongPhase, got %v", err)
	}
}

func TestCarioca_PlayerDrawFromStock_RejectsOnGameEnd(t *testing.T) {
	g := helperCariocaHand(t)
	g.gameEndFlag = true
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrGameEnded) {
		t.Errorf("expected ErrGameEnded, got %v", err)
	}
}

func TestCarioca_PlayerDrawFromStock_RejectsNotHumanTurn(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(1)
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("expected ErrNotHumanTurn, got %v", err)
	}
}

func TestCarioca_PlayerDrawFromDiscard_Success(t *testing.T) {
	g := helperCariocaHand(t)
	if err := g.PlayerDrawFromDiscard(); err != nil {
		t.Fatalf("PlayerDrawFromDiscard error: %v", err)
	}
	if g.GetPhase() != CariocaPhasePlay {
		t.Errorf("phase = %d, want PhasePlay", g.GetPhase())
	}
}

func TestCarioca_PlayerDrawFromDiscard_EmptyError(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetDiscardPile(nil)
	if err := g.PlayerDrawFromDiscard(); err == nil {
		t.Error("expected error for empty discard")
	}
}

func TestCarioca_PlayerDrawFromDiscard_WrongPhase(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetPhase(CariocaPhasePlay)
	if err := g.PlayerDrawFromDiscard(); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("expected ErrWrongPhase, got %v", err)
	}
}

func TestCarioca_DrawFromStock_RecyclesDiscardWhenEmpty(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("expected recycle to succeed, got %v", err)
	}
	if g.GetDrawPileCount() == 0 {
		t.Error("draw pile should be replenished from discard")
	}
}

func TestCarioca_DrawFromStock_StockOutEndsRound(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{cariocaCard(1, 5)})
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if g.GetPhase() != CariocaPhaseRoundEnd {
		t.Errorf("expected round end on stock-out, got %d", g.GetPhase())
	}
}

func TestCarioca_PlayerMeldContract_Round1Success(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	p := g.GetPlayer(0)
	hand := []*Card{
		cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5),
		cariocaCard(4, 13), cariocaCard(1, 13), cariocaCard(2, 13),
		cariocaCard(1, 2), cariocaCard(2, 3), cariocaCard(3, 4), cariocaCard(4, 6), cariocaCard(1, 7), cariocaCard(2, 8),
	}
	cariocaSetHand(p, hand)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)

	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err != nil {
		t.Fatalf("PlayerMeldContract error: %v", err)
	}
	if !p.IsContractMet() {
		t.Error("contract should be met")
	}
	if p.GetMeldCount() != 2 {
		t.Errorf("meld count = %d, want 2", p.GetMeldCount())
	}
	if p.GetCardsSize() != 6 {
		t.Errorf("hand size = %d, want 6", p.GetCardsSize())
	}
}

func TestCarioca_PlayerMeldContract_Round1WithJoker(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	p := g.GetPlayer(0)
	// First set uses a joker as wild.
	hand := []*Card{
		cariocaCard(1, 5), cariocaCard(2, 5), cariocaJoker(1),
		cariocaCard(4, 13), cariocaCard(1, 13), cariocaCard(2, 13),
		cariocaCard(1, 2), cariocaCard(2, 3), cariocaCard(3, 4), cariocaCard(4, 6), cariocaCard(1, 7), cariocaCard(2, 8),
	}
	cariocaSetHand(p, hand)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)

	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err != nil {
		t.Fatalf("PlayerMeldContract with joker error: %v", err)
	}
	if !p.IsContractMet() {
		t.Error("contract with wild joker should be met")
	}
}

func TestCarioca_PlayerMeldContract_RejectsWrongCount(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(CariocaPhasePlay)
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}}); err == nil {
		t.Error("expected error for wrong slot count")
	}
}

func TestCarioca_PlayerMeldContract_RejectsInvalidSet(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	p := g.GetPlayer(0)
	hand := []*Card{
		cariocaCard(1, 5), cariocaCard(2, 6), cariocaCard(3, 7),
		cariocaCard(4, 13), cariocaCard(1, 13), cariocaCard(2, 13),
		cariocaCard(1, 2), cariocaCard(2, 3), cariocaCard(3, 4), cariocaCard(4, 6), cariocaCard(1, 7), cariocaCard(2, 8),
	}
	cariocaSetHand(p, hand)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)

	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err == nil {
		t.Error("expected error for invalid first set")
	}
}

func TestCarioca_PlayerMeldContract_Round3RunsSuccess(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(3)
	p := g.GetPlayer(0)
	hand := []*Card{
		cariocaCard(1, 2), cariocaCard(1, 3), cariocaCard(1, 4), cariocaCard(1, 5),
		cariocaCard(2, 7), cariocaCard(2, 8), cariocaCard(2, 9), cariocaCard(2, 10),
		cariocaCard(3, 12), cariocaCard(4, 13), cariocaCard(1, 1), cariocaCard(2, 1),
	}
	cariocaSetHand(p, hand)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)

	if err := g.PlayerMeldContract([][]int{{0, 1, 2, 3}, {4, 5, 6, 7}}); err != nil {
		t.Fatalf("PlayerMeldContract error: %v", err)
	}
	if !p.IsContractMet() {
		t.Error("contract should be met")
	}
}

func TestCarioca_PlayerMeldContract_RejectsDuplicateIndices(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(CariocaPhasePlay)
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {2, 3, 4}}); err == nil {
		t.Error("expected error for duplicate index across slots")
	}
}

func TestCarioca_PlayerMeldContract_RejectsAfterMet(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(CariocaPhasePlay)
	g.GetPlayer(0).SetContractMet(true)
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err == nil {
		t.Error("expected error when contract already met")
	}
}

func TestCarioca_PlayerLayoff_Success(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	p := g.GetPlayer(0)
	cariocaSetHand(p, []*Card{cariocaCard(4, 5)})
	p.SetContractMet(true)
	p.AppendMeld([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)

	if err := g.PlayerLayoff(0, 0, 0); err != nil {
		t.Fatalf("PlayerLayoff error: %v", err)
	}
	if p.GetCardsSize() != 0 {
		t.Errorf("hand should be empty after layoff, got %d", p.GetCardsSize())
	}
	if p.GetMeld(0) == nil || len(p.GetMeld(0)) != 4 {
		t.Errorf("meld should have grown to 4, got %v", p.GetMeld(0))
	}
}

func TestCarioca_PlayerLayoff_BeforeContractMet(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(CariocaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(1).SetContractMet(true)
	g.GetPlayer(1).AppendMeld([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	cariocaSetHand(g.GetPlayer(0), []*Card{cariocaCard(4, 5)})
	if err := g.PlayerLayoff(1, 0, 0); err == nil {
		t.Error("layoff should fail before player meets contract")
	}
}

func TestCarioca_PlayerLayoff_TargetNotMet(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(CariocaPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).SetContractMet(true)
	g.GetPlayer(0).AppendMeld([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	cariocaSetHand(g.GetPlayer(0), []*Card{cariocaCard(4, 5)})
	if err := g.PlayerLayoff(1, 0, 0); err == nil {
		t.Error("layoff should fail when target player not met")
	}
}

func TestCarioca_PlayerDiscard_AdvancesTurn(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	cardsBefore := g.GetPlayer(0).GetCardsSize()
	if err := g.PlayerDiscard(0); err != nil {
		t.Fatalf("PlayerDiscard error: %v", err)
	}
	if g.GetCurrentPlayerIdx() != 1 {
		t.Errorf("currentPlayerIdx = %d, want 1", g.GetCurrentPlayerIdx())
	}
	if g.GetPhase() != CariocaPhaseDraw {
		t.Errorf("phase = %d, want PhaseDraw", g.GetPhase())
	}
	if g.GetPlayer(0).GetCardsSize() != cardsBefore-1 {
		t.Errorf("hand should shrink by 1")
	}
}

func TestCarioca_PlayerDiscard_LastCardWithoutContractFails(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	cariocaSetHand(g.GetPlayer(0), []*Card{cariocaCard(1, 5)})
	g.GetPlayer(0).SetContractMet(false)
	if err := g.PlayerDiscard(0); err == nil {
		t.Error("discarding last card before meeting contract should fail")
	}
}

func TestCarioca_PlayerDiscard_LastCardAfterContractWins(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	cariocaSetHand(g.GetPlayer(0), []*Card{cariocaCard(1, 5)})
	g.GetPlayer(0).SetContractMet(true)
	g.GetPlayer(0).AppendMeld([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	g.GetPlayer(0).AppendMeld([]*Card{cariocaCard(1, 13), cariocaCard(2, 13), cariocaCard(3, 13)})
	if err := g.PlayerDiscard(0); err != nil {
		t.Fatalf("PlayerDiscard error: %v", err)
	}
	if g.GetPhase() != CariocaPhaseRoundEnd {
		t.Errorf("phase = %d, want PhaseRoundEnd", g.GetPhase())
	}
	if g.GetRoundWinnerIdx() != 0 {
		t.Errorf("roundWinnerIdx = %d, want 0", g.GetRoundWinnerIdx())
	}
}

func TestCarioca_PlayerMeldExtra_BeforeContractFails(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	if err := g.PlayerMeldExtra([]int{0, 1, 2}); err == nil {
		t.Error("expected failure: extra meld before contract met")
	}
}

func TestCarioca_PlayerMeldExtra_Success(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	p := g.GetPlayer(0)
	p.SetContractMet(true)
	cariocaSetHand(p, []*Card{cariocaCard(1, 9), cariocaCard(2, 9), cariocaCard(3, 9), cariocaCard(4, 11)})
	if err := g.PlayerMeldExtra([]int{0, 1, 2}); err != nil {
		t.Fatalf("PlayerMeldExtra error: %v", err)
	}
	if p.GetMeldCount() != 1 || len(p.GetMeld(0)) != 3 {
		t.Errorf("meld not added correctly: %v", p.GetMeld(0))
	}
	if p.GetCardsSize() != 1 {
		t.Errorf("hand size = %d, want 1", p.GetCardsSize())
	}
}

func TestCarioca_PlayerMeldExtra_InvalidMeldFails(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	p := g.GetPlayer(0)
	p.SetContractMet(true)
	cariocaSetHand(p, []*Card{cariocaCard(1, 9), cariocaCard(2, 4), cariocaCard(3, 11)})
	if err := g.PlayerMeldExtra([]int{0, 1, 2}); err == nil {
		t.Error("expected error: invalid meld")
	}
}

func TestCarioca_FinishRound_PenalizesNonWinners(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		cariocaSetHand(g.GetPlayer(i), []*Card{cariocaCard(1, 13), cariocaCard(2, 13)})
	}
	g.GetPlayer(0).SetContractMet(true)
	g.finishRound(0)
	if g.GetPlayer(0).GetCumulativeScore() != 0 {
		t.Errorf("winner score = %d, want 0", g.GetPlayer(0).GetCumulativeScore())
	}
	for i := 1; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCumulativeScore(); got <= 0 {
			t.Errorf("player %d should be penalized, got %d", i, got)
		}
	}
}

func TestCarioca_FinishRound_FailedContractAddsPenalty(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		cariocaSetHand(g.GetPlayer(i), []*Card{cariocaCard(1, 5)})
		g.GetPlayer(i).SetContractMet(false)
	}
	g.GetPlayer(0).SetContractMet(true)
	g.finishRound(0)
	expectedPenalty := 5 + g.GetConfig().FailContractPenalty
	if g.GetPlayer(1).GetCumulativeScore() != expectedPenalty {
		t.Errorf("player 1 score = %d, want %d", g.GetPlayer(1).GetCumulativeScore(), expectedPenalty)
	}
}

func TestCarioca_FinishRound_JokerPenalty(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		cariocaSetHand(g.GetPlayer(i), []*Card{cariocaJoker(1)})
		g.GetPlayer(i).SetContractMet(true) // no fail penalty, only pip penalty
	}
	g.finishRound(0) // player 0 wins
	if got := g.GetPlayer(1).GetRoundScore(); got != CariocaJokerPenalty {
		t.Errorf("joker round penalty = %d, want %d", got, CariocaJokerPenalty)
	}
}

func TestCarioca_FinishRound_FinalRoundEndsGame(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetRoundNumber(CariocaTotalRounds)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		cariocaSetHand(g.GetPlayer(i), []*Card{cariocaCard(1, 5)})
	}
	g.GetPlayer(0).SetContractMet(true)
	g.finishRound(0)
	if !g.GetGameEndFlag() {
		t.Error("game should end after final round")
	}
}

func TestCarioca_FinishRound_DrawNoWinnerStockOut(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(CariocaPhasePlay)
	g.endRoundStockOut()
	if g.GetRoundWinnerIdx() != -1 {
		t.Errorf("roundWinnerIdx = %d, want -1", g.GetRoundWinnerIdx())
	}
}

func TestCarioca_CpuPlay_NoOpForHumanTurn(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(0)
	g.CpuPlay()
	if g.GetPhase() != CariocaPhaseDraw {
		t.Error("CpuPlay should be a no-op for human turn")
	}
}

func TestCarioca_CpuPlay_DrawAndDiscard(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay() // draw
	g.CpuPlay() // play+discard
	if g.GetCurrentPlayerIdx() != 2 {
		t.Errorf("after CPU 1 turn, currentPlayerIdx = %d, want 2", g.GetCurrentPlayerIdx())
	}
}

func TestCarioca_CpuPlay_NoActionAfterGameEnd(t *testing.T) {
	g := helperCariocaHand(t)
	g.gameEndFlag = true
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
}

func TestCarioca_GetCurrentContract_ReturnsCorrectSlots(t *testing.T) {
	g := NewDefaultCarioca()
	g.SetRoundNumber(1)
	c := g.GetCurrentContract()
	if len(c.Slots) != 2 {
		t.Errorf("R1 slots = %d, want 2", len(c.Slots))
	}
	g.SetRoundNumber(7)
	c = g.GetCurrentContract()
	if len(c.Slots) != 3 {
		t.Errorf("R7 slots = %d, want 3", len(c.Slots))
	}
	for _, s := range c.Slots {
		if s.Kind != ContractSlotRun || s.Size != 4 {
			t.Errorf("R7 slot kind/size mismatch: %+v", s)
		}
	}
}

func TestCariocaContractForRound_OutOfRange(t *testing.T) {
	if c := CariocaContractForRound(0); len(c.Slots) != 0 {
		t.Error("R0 should be empty")
	}
	if c := CariocaContractForRound(99); len(c.Slots) != 0 {
		t.Error("R99 should be empty")
	}
}

func TestCariocaValidateContractSlot_Variants(t *testing.T) {
	setSlot := ContractSlot{Kind: ContractSlotSet, Size: 3}
	runSlot := ContractSlot{Kind: ContractSlotRun, Size: 4}

	tests := []struct {
		name  string
		slot  ContractSlot
		cards []*Card
		want  bool
	}{
		{"valid set", setSlot, []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}, true},
		{"set with joker", setSlot, []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaJoker(1)}, true},
		{"set wrong size", setSlot, []*Card{cariocaCard(1, 5), cariocaCard(2, 5)}, false},
		{"set duplicate suit", setSlot, []*Card{cariocaCard(1, 5), cariocaCard(1, 5), cariocaCard(2, 5)}, true},
		{"set wrong rank", setSlot, []*Card{cariocaCard(1, 5), cariocaCard(2, 6), cariocaCard(3, 5)}, false},
		{"set two jokers", setSlot, []*Card{cariocaCard(1, 5), cariocaJoker(1), cariocaJoker(2)}, false},
		{"valid Ace-high run", runSlot, []*Card{cariocaCard(1, 11), cariocaCard(1, 12), cariocaCard(1, 13), cariocaCard(1, 1)}, true},
		{"valid run", runSlot, []*Card{cariocaCard(1, 2), cariocaCard(1, 3), cariocaCard(1, 4), cariocaCard(1, 5)}, true},
		{"run with joker", runSlot, []*Card{cariocaCard(1, 2), cariocaCard(1, 3), cariocaJoker(1), cariocaCard(1, 5)}, true},
		{"run wrong suit", runSlot, []*Card{cariocaCard(1, 2), cariocaCard(2, 3), cariocaCard(1, 4), cariocaCard(1, 5)}, false},
		{"run not consecutive", runSlot, []*Card{cariocaCard(1, 2), cariocaCard(1, 3), cariocaCard(1, 6), cariocaCard(1, 7)}, false},
		{"unknown kind", ContractSlot{Kind: 99, Size: 3}, []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cariocaValidateContractSlot(tt.slot, tt.cards); got != tt.want {
				t.Errorf("cariocaValidateContractSlot(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCariocaIsMeld_Cases(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  bool
	}{
		{"set of 3", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}, true},
		{"set of 4", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5), cariocaCard(4, 5)}, true},
		{"set with joker", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaJoker(1)}, true},
		{"too small", []*Card{cariocaCard(1, 5), cariocaCard(2, 5)}, false},
		{"run of 4", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7), cariocaCard(1, 8)}, true},
		{"run of 3 too short", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7)}, false},
		{"run J-Q-K-A (Ace high)", []*Card{cariocaCard(1, 11), cariocaCard(1, 12), cariocaCard(1, 13), cariocaCard(1, 1)}, true},
		{"run A-2-3-4 (Ace low)", []*Card{cariocaCard(1, 1), cariocaCard(1, 2), cariocaCard(1, 3), cariocaCard(1, 4)}, true},
		{"run with joker", []*Card{cariocaCard(1, 5), cariocaJoker(1), cariocaCard(1, 7), cariocaCard(1, 8)}, true},
		{"invalid mixed", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(3, 8), cariocaCard(2, 2)}, false},
		{"wraparound K-A-2-3 not allowed", []*Card{cariocaCard(1, 13), cariocaCard(1, 1), cariocaCard(1, 2), cariocaCard(1, 3)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cariocaIsMeld(tt.cards); got != tt.want {
				t.Errorf("cariocaIsMeld(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCariocaConfig_Validate(t *testing.T) {
	c := DefaultCariocaConfig()
	if err := c.Validate(); err != nil {
		t.Errorf("default config should validate: %v", err)
	}
	badPlayers := DefaultCariocaConfig()
	badPlayers.PlayerCount = 2
	if err := badPlayers.Validate(); err == nil {
		t.Error("expected player-count validation to fail (2 < 3)")
	}
	badPlayers2 := DefaultCariocaConfig()
	badPlayers2.PlayerCount = 7
	if err := badPlayers2.Validate(); err == nil {
		t.Error("expected player-count validation to fail (7 > 6)")
	}
	bad := DefaultCariocaConfig()
	bad.CpuDifficulty = -1
	if err := bad.Validate(); err == nil {
		t.Error("expected difficulty validation to fail")
	}
	bad2 := DefaultCariocaConfig()
	bad2.FailContractPenalty = -1
	if err := bad2.Validate(); err == nil {
		t.Error("expected fail penalty validation to fail")
	}
}

func TestCarioca_JSONRoundTrip(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored Carioca
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.GetPlayerCnt() != g.GetPlayerCnt() {
		t.Errorf("player count differs: %d vs %d", restored.GetPlayerCnt(), g.GetPlayerCnt())
	}
	if restored.GetRoundNumber() != g.GetRoundNumber() {
		t.Errorf("round number differs: %d vs %d", restored.GetRoundNumber(), g.GetRoundNumber())
	}
}

func TestCarioca_JSONRejectsOversized(t *testing.T) {
	bigPlayers := make([]map[string]any, cariocaMaxSliceLen+1)
	for i := range bigPlayers {
		bigPlayers[i] = map[string]any{}
	}
	doc := map[string]any{"pl": bigPlayers}
	data, _ := json.Marshal(doc)
	var g Carioca
	if err := json.Unmarshal(data, &g); err == nil {
		t.Error("expected oversize rejection")
	}
}

func TestCarioca_JSONRejectsOutOfRangeIndex(t *testing.T) {
	// 4 players but currentPlayerIdx = 9 → must be rejected.
	doc := map[string]any{
		"pl": []map[string]any{{}, {}, {}, {}},
		"ci": 9,
		"ps": int(CariocaPhaseDraw),
		"rn": 1,
	}
	data, _ := json.Marshal(doc)
	var g Carioca
	if err := json.Unmarshal(data, &g); err == nil {
		t.Error("expected out-of-range currentPlayerIdx rejection")
	}
}

func TestCarioca_JSONRejectsBadPhase(t *testing.T) {
	doc := map[string]any{
		"pl": []map[string]any{{}, {}, {}, {}},
		"ps": 99,
	}
	data, _ := json.Marshal(doc)
	var g Carioca
	if err := json.Unmarshal(data, &g); err == nil {
		t.Error("expected invalid phase rejection")
	}
}

func TestCariocaPlayer_JSONRoundTrip(t *testing.T) {
	p := NewCariocaPlayer(true)
	p.AddCard(cariocaCard(1, 5))
	p.AppendMeld([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	p.SetContractMet(true)
	p.SetContractIndex([]int{0})

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored CariocaPlayer
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !restored.IsContractMet() {
		t.Error("contract met should round-trip")
	}
	if restored.GetMeldCount() != 1 {
		t.Errorf("meld count = %d, want 1", restored.GetMeldCount())
	}
}

func TestCariocaPlayer_ResetRoundClearsState(t *testing.T) {
	p := NewCariocaPlayer(false)
	p.AddCard(cariocaCard(1, 5))
	p.AppendMeld([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	p.SetContractMet(true)
	p.SetRoundScore(50)
	p.ResetRound()
	if p.GetCardsSize() != 0 {
		t.Error("cards should be cleared")
	}
	if p.GetMeldCount() != 0 {
		t.Error("melds should be cleared")
	}
	if p.IsContractMet() {
		t.Error("contractMet should be false")
	}
	if p.GetRoundScore() != 0 {
		t.Error("round score should be reset")
	}
}

func TestCariocaPlayer_AddCardToMeldOutOfRange(t *testing.T) {
	p := NewCariocaPlayer(false)
	if p.AddCardToMeld(99, cariocaCard(1, 5)) {
		t.Error("AddCardToMeld should reject out-of-range index")
	}
	if p.GetMeld(99) != nil {
		t.Error("GetMeld should return nil for out-of-range")
	}
}

func TestCariocaPlayer_SetContractIndexNilCopy(t *testing.T) {
	p := NewCariocaPlayer(false)
	p.SetContractIndex(nil)
	if p.GetContractIndex() != nil {
		t.Error("SetContractIndex(nil) should clear")
	}
	src := []int{1, 2, 3}
	p.SetContractIndex(src)
	src[0] = 99
	if p.GetContractIndex()[0] == 99 {
		t.Error("SetContractIndex should defensively copy")
	}
}

func TestCarioca_Penalty(t *testing.T) {
	if got := cariocaCardPenalty(cariocaJoker(1)); got != CariocaJokerPenalty {
		t.Errorf("Joker penalty = %d, want %d", got, CariocaJokerPenalty)
	}
	if got := cariocaCardPenalty(cariocaCard(1, 1)); got != 15 {
		t.Errorf("Ace penalty = %d, want 15", got)
	}
	if got := cariocaCardPenalty(cariocaCard(1, 13)); got != 10 {
		t.Errorf("King penalty = %d, want 10", got)
	}
	if got := cariocaCardPenalty(cariocaCard(1, 5)); got != 5 {
		t.Errorf("5 penalty = %d, want 5", got)
	}
}

func TestCarioca_PlayerNameFallback(t *testing.T) {
	g := helperCariocaHand(t)
	if got := playerName(g.players, -1); got == "" {
		t.Error("playerName(-1) should return non-empty fallback")
	}
	if got := playerName(g.players, 99); got == "" {
		t.Error("playerName(99) should return non-empty fallback")
	}
}

func TestCarioca_CanAddToCariocaMeld(t *testing.T) {
	tests := []struct {
		name string
		meld []*Card
		card *Card
		want bool
	}{
		{"empty meld", nil, cariocaCard(1, 5), false},
		{"nil card", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}, nil, false},
		{"set: same rank", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}, cariocaCard(4, 5), true},
		{"set: joker onto joker-less set", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}, cariocaJoker(1), true},
		{"set: joker onto set that already has joker", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaJoker(1)}, cariocaJoker(2), false},
		{"set: different rank rejected", []*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)}, cariocaCard(1, 6), false},
		{"run: extend low end", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7), cariocaCard(1, 8)}, cariocaCard(1, 4), true},
		{"run: extend high end", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7), cariocaCard(1, 8)}, cariocaCard(1, 9), true},
		{"run: wrong suit rejected", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7), cariocaCard(1, 8)}, cariocaCard(2, 9), false},
		{"run: gap rejected", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7), cariocaCard(1, 8)}, cariocaCard(1, 11), false},
		{"run: joker fills", []*Card{cariocaCard(1, 5), cariocaCard(1, 6), cariocaCard(1, 7), cariocaCard(1, 8)}, cariocaJoker(1), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canAddToCariocaMeld(tt.meld, tt.card); got != tt.want {
				t.Errorf("canAddToCariocaMeld(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCarioca_CpuFullRound_Smoke(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	g.SetCurrentPlayerIdx(1)
	for i := 0; i < 80 && !g.GetGameEndFlag(); i++ {
		if g.IsHumanTurn() || g.GetPhase() == CariocaPhaseRoundEnd {
			break
		}
		g.CpuPlay()
	}
}

func TestCarioca_CpuShouldTakeDiscard_HardLayoff(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(1)
	cfg := g.GetConfig()
	cfg.CpuDifficulty = CariocaCpuDifficultyHard
	g.SetConfig(cfg)
	cpu := g.GetPlayer(1)
	cpu.SetContractMet(true)
	cpu.AppendMeld([]*Card{cariocaCard(1, 5), cariocaCard(2, 5), cariocaCard(3, 5)})
	g.SetDiscardPile([]*Card{cariocaCard(4, 5)})
	if !g.cpuShouldTakeDiscard(cariocaCard(4, 5)) {
		t.Error("Hard CPU with completed contract should take a layoff-able discard")
	}
}

func TestCarioca_CpuShouldTakeDiscard_NoLayoff(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetCurrentPlayerIdx(1)
	cpu := g.GetPlayer(1)
	cpu.SetContractMet(true)
	if g.cpuShouldTakeDiscard(cariocaCard(4, 5)) {
		t.Error("Opened CPU with no layoff target should not take discard")
	}
}

func TestCarioca_GetterSetterCoverage(t *testing.T) {
	g := helperCariocaHand(t)
	g.SetConfig(DefaultCariocaConfig())
	_ = g.GetConfig()
	_ = g.GetActionLog()
	_ = g.GetDiscardPile()
	_ = g.GetDrawPileCount()
	_ = g.GetGameEndFlag()
	_ = g.GetWinnerIdx()
	if g.GetPlayer(99) != nil {
		t.Error("GetPlayer out-of-range should return nil")
	}
	if g.GetDiscardTop() == nil {
		t.Error("discard top should exist")
	}
	if !g.IsHumanTurn() {
		t.Error("should start with human turn")
	}
	g.SetCurrentPlayerIdx(99)
	if g.IsHumanTurn() {
		t.Error("out-of-range index should not be human")
	}
}

// --- Joker-aware CPU helper tests (FIX A) ---

// TestCariocaFindContractMeld_SetUsesJoker はセットスロットでジョーカーをワイルドとして使えることを検証する。
func TestCariocaFindContractMeld_SetUsesJoker(t *testing.T) {
	contract := CariocaContractForRound(1) // 2 セット
	// 7 のトリオ + 5 のペア + ジョーカー1枚（ジョーカー無しでは 2 セット目を作れない）。
	cards := []*Card{
		cariocaCard(CardDesignSpade, 7), cariocaCard(CardDesignHeart, 7), cariocaCard(CardDesignDiamond, 7),
		cariocaCard(CardDesignSpade, 5), cariocaCard(CardDesignHeart, 5), cariocaJoker(1),
	}
	// ジョーカー非対応の探索では 2 セット目を作れない。
	if _, ok := FindContractMeld(contract, cards); ok {
		t.Fatal("non-joker FindContractMeld should NOT satisfy the contract")
	}
	groups, ok := cariocaFindContractMeld(contract, cards)
	if !ok {
		t.Fatal("cariocaFindContractMeld should satisfy the contract using a joker")
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	usedJoker := false
	for _, g := range groups {
		if !cariocaIsSet(g) {
			t.Errorf("group %v is not a valid set", g)
		}
		for _, c := range g {
			if cariocaIsJoker(c) {
				usedJoker = true
			}
		}
	}
	if !usedJoker {
		t.Error("expected the contract meld to consume the joker as a wildcard")
	}
}

// TestCariocaFindContractMeld_RunUsesJoker はランスロットでジョーカーが隙間を埋められることを検証する。
func TestCariocaFindContractMeld_RunUsesJoker(t *testing.T) {
	contract := CariocaContractForRound(3) // 2 ラン
	cards := []*Card{
		// 完成ラン: 2-3-4-5 スペード
		cariocaCard(CardDesignSpade, 2), cariocaCard(CardDesignSpade, 3),
		cariocaCard(CardDesignSpade, 4), cariocaCard(CardDesignSpade, 5),
		// 隙間ラン: 7-8-[9=joker]-10 ハート
		cariocaCard(CardDesignHeart, 7), cariocaCard(CardDesignHeart, 8),
		cariocaCard(CardDesignHeart, 10), cariocaJoker(1),
	}
	if _, ok := FindContractMeld(contract, cards); ok {
		t.Fatal("non-joker FindContractMeld should NOT satisfy the two-run contract")
	}
	groups, ok := cariocaFindContractMeld(contract, cards)
	if !ok {
		t.Fatal("cariocaFindContractMeld should satisfy the two-run contract using a joker")
	}
	for _, g := range groups {
		if !cariocaIsRun(g) {
			t.Errorf("group %v is not a valid run", g)
		}
	}
}

// TestCarioca_CpuPlay_MeldsContractWithJoker は CPU がジョーカーを使ってコントラクトを達成することを検証する。
func TestCarioca_CpuPlay_MeldsContractWithJoker(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	g.SetRoundNumber(1) // 2 セット
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(CariocaPhasePlay)
	cpu := g.GetPlayer(1)
	cariocaSetHand(cpu, []*Card{
		cariocaCard(CardDesignSpade, 7), cariocaCard(CardDesignHeart, 7), cariocaCard(CardDesignDiamond, 7),
		cariocaCard(CardDesignSpade, 5), cariocaCard(CardDesignHeart, 5), cariocaJoker(1),
		cariocaCard(CardDesignSpade, 2), cariocaCard(CardDesignHeart, 9),
		cariocaCard(CardDesignDiamond, 11), cariocaCard(CardDesignClover, 13),
	})
	if cpu.IsContractMet() {
		t.Fatal("precondition: contract must not be met yet")
	}
	g.CpuPlay()
	if !cpu.IsContractMet() {
		t.Error("CPU should meet the contract using the joker as a wildcard")
	}
}

// TestCariocaScoreContractProgress_JokerHelps はジョーカーが進捗スコアを引き上げることを検証する。
func TestCariocaScoreContractProgress_JokerHelps(t *testing.T) {
	contract := CariocaContractForRound(1) // 2 セット
	base := []*Card{
		cariocaCard(CardDesignSpade, 5), cariocaCard(CardDesignHeart, 5),
		cariocaCard(CardDesignSpade, 2), cariocaCard(CardDesignHeart, 9),
	}
	before := cariocaScoreContractProgress(contract, base)
	after := cariocaScoreContractProgress(contract, append(append([]*Card(nil), base...), cariocaJoker(1)))
	if after <= before {
		t.Errorf("joker should raise progress score: before=%d after=%d", before, after)
	}
	if after < 10 {
		t.Errorf("joker-completed set should score at least 10, got %d", after)
	}

	// ランスロットでも同様（3 連 → 4 連へ完成）。
	runContract := CariocaContractForRound(3) // 2 ラン
	runBase := []*Card{
		cariocaCard(CardDesignSpade, 2), cariocaCard(CardDesignSpade, 3), cariocaCard(CardDesignSpade, 4),
	}
	rb := cariocaScoreContractProgress(runContract, runBase)
	ra := cariocaScoreContractProgress(runContract, append(append([]*Card(nil), runBase...), cariocaJoker(1)))
	if ra <= rb {
		t.Errorf("joker should raise run progress: before=%d after=%d", rb, ra)
	}
}

// TestCariocaCpuShouldTakeDiscard_PicksUsefulJoker は CPU が役に立つジョーカーを拾うことを検証する。
func TestCariocaCpuShouldTakeDiscard_PicksUsefulJoker(t *testing.T) {
	g := NewDefaultCarioca()
	g.Reset()
	g.SetRoundNumber(1) // 2 セット
	g.SetCurrentPlayerIdx(1)
	cpu := g.GetPlayer(1)
	cariocaSetHand(cpu, []*Card{
		cariocaCard(CardDesignSpade, 5), cariocaCard(CardDesignHeart, 5),
		cariocaCard(CardDesignSpade, 2), cariocaCard(CardDesignHeart, 9),
		cariocaCard(CardDesignDiamond, 11),
	})
	if !g.cpuShouldTakeDiscard(cariocaJoker(1)) {
		t.Error("CPU should take a joker that completes a set")
	}
}

// TestCariocaFindExtraMeld_UsesJoker は追加メルド探索がジョーカーを活用することを検証する。
func TestCariocaFindExtraMeld_UsesJoker(t *testing.T) {
	// セット: 5-5 + ジョーカー
	if meld, ok := cariocaFindExtraMeld([]*Card{
		cariocaCard(CardDesignSpade, 5), cariocaCard(CardDesignHeart, 5), cariocaJoker(1),
	}); !ok || len(meld) != 3 || !cariocaIsSet(meld) {
		t.Errorf("expected joker set extra meld, got ok=%v meld=%v", ok, meld)
	}
	// ラン（隙間埋め）: 6-7-[8]-9 スペード
	if meld, ok := cariocaFindExtraMeld([]*Card{
		cariocaCard(CardDesignSpade, 6), cariocaCard(CardDesignSpade, 7),
		cariocaCard(CardDesignSpade, 9), cariocaJoker(1),
	}); !ok || !cariocaIsRun(meld) {
		t.Errorf("expected joker gap-fill run, got ok=%v meld=%v", ok, meld)
	}
	// ラン（端延長）: 6-7-8-[9] スペード
	if meld, ok := cariocaFindExtraMeld([]*Card{
		cariocaCard(CardDesignSpade, 6), cariocaCard(CardDesignSpade, 7),
		cariocaCard(CardDesignSpade, 8), cariocaJoker(1),
	}); !ok || !cariocaIsRun(meld) {
		t.Errorf("expected joker end-extend run, got ok=%v meld=%v", ok, meld)
	}
	// メルド無し
	if _, ok := cariocaFindExtraMeld([]*Card{
		cariocaCard(CardDesignSpade, 2), cariocaCard(CardDesignHeart, 9),
	}); ok {
		t.Error("no meld should be found in a junk hand")
	}
}

// TestCariocaIsSet_RejectsTwoJokers はセットが 2 枚のジョーカーを拒否することを検証する。
func TestCariocaIsSet_RejectsTwoJokers(t *testing.T) {
	if cariocaIsSet([]*Card{cariocaCard(CardDesignSpade, 5), cariocaJoker(1), cariocaJoker(2)}) {
		t.Error("a set may contain at most one joker")
	}
}
