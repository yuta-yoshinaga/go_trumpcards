package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// helperContractRummyHand 任意の手札・スコア状態を持つ ContractRummy を構築する。
// テスト対象のフェーズへ素早く到達するためのヘルパー。
func helperContractRummyHand(t *testing.T) *ContractRummy {
	t.Helper()
	g := NewDefaultContractRummy()
	g.Reset()
	return g
}

// makeCard ヘルパー: 指定スート・値のカードを生成する
func crCard(design, value int) *Card {
	return NewCard(design, value, true)
}

// setHand プレイヤーの手札を差し替える
func setHand(p *ContractRummyPlayer, cards []*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNewDefaultContractRummy_构造(t *testing.T) {
	g := NewDefaultContractRummy()
	if g == nil {
		t.Fatal("NewDefaultContractRummy returned nil")
	}
	if g.GetPlayerCnt() != ContractRummyPlayerCnt {
		t.Errorf("player count = %d, want %d", g.GetPlayerCnt(), ContractRummyPlayerCnt)
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

func TestContractRummy_Reset_DealsHand(t *testing.T) {
	g := NewDefaultContractRummy()
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		got := g.GetPlayer(i).GetCardsSize()
		if got != ContractRummyHandSize {
			t.Errorf("player %d hand size = %d, want %d", i, got, ContractRummyHandSize)
		}
	}
	if g.GetDiscardTop() == nil {
		t.Error("discard top should be set after Reset")
	}
	if g.GetRoundNumber() != 1 {
		t.Errorf("roundNumber = %d, want 1", g.GetRoundNumber())
	}
	if g.GetPhase() != ContractRummyPhaseDraw {
		t.Errorf("phase = %d, want PhaseDraw", g.GetPhase())
	}
}

func TestContractRummy_NextRound_AdvancesAndDeals(t *testing.T) {
	g := NewDefaultContractRummy()
	g.Reset()
	g.SetPhase(ContractRummyPhaseRoundEnd)
	g.NextRound()
	if g.GetRoundNumber() != 2 {
		t.Errorf("roundNumber after NextRound = %d, want 2", g.GetRoundNumber())
	}
	if g.GetPhase() != ContractRummyPhaseDraw {
		t.Errorf("phase = %d, want PhaseDraw", g.GetPhase())
	}
}

func TestContractRummy_NextRound_NoOpWhenNotInRoundEnd(t *testing.T) {
	g := NewDefaultContractRummy()
	g.Reset()
	original := g.GetRoundNumber()
	g.NextRound() // phase is Draw, should noop
	if g.GetRoundNumber() != original {
		t.Errorf("NextRound during Draw must not advance round")
	}
}

func TestContractRummy_NextRound_AfterFinalRoundEndsGame(t *testing.T) {
	g := NewDefaultContractRummy()
	g.Reset()
	g.SetRoundNumber(ContractRummyTotalRounds)
	g.SetPhase(ContractRummyPhaseRoundEnd)
	g.NextRound()
	if !g.GetGameEndFlag() {
		t.Error("game should end after final round")
	}
	if g.GetWinnerIdx() < 0 {
		t.Error("winner should be set")
	}
}

func TestContractRummy_PlayerDrawFromStock_ProgressesPhase(t *testing.T) {
	g := helperContractRummyHand(t)
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("PlayerDrawFromStock error: %v", err)
	}
	if g.GetPhase() != ContractRummyPhasePlay {
		t.Errorf("phase = %d, want PhasePlay", g.GetPhase())
	}
	if g.GetPlayer(0).GetCardsSize() != ContractRummyHandSize+1 {
		t.Errorf("hand should grow by 1 after draw")
	}
}

func TestContractRummy_PlayerDrawFromStock_RejectsWrongPhase(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetPhase(ContractRummyPhasePlay)
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("expected ErrWrongPhase, got %v", err)
	}
}

func TestContractRummy_PlayerDrawFromStock_RejectsOnGameEnd(t *testing.T) {
	g := helperContractRummyHand(t)
	g.gameEndFlag = true
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrGameEnded) {
		t.Errorf("expected ErrGameEnded, got %v", err)
	}
}

func TestContractRummy_PlayerDrawFromStock_RejectsNotHumanTurn(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(1) // CPU
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("expected ErrNotHumanTurn, got %v", err)
	}
}

func TestContractRummy_PlayerDrawFromDiscard_Success(t *testing.T) {
	g := helperContractRummyHand(t)
	if err := g.PlayerDrawFromDiscard(); err != nil {
		t.Fatalf("PlayerDrawFromDiscard error: %v", err)
	}
	if g.GetPhase() != ContractRummyPhasePlay {
		t.Errorf("phase = %d, want PhasePlay", g.GetPhase())
	}
}

func TestContractRummy_PlayerDrawFromDiscard_EmptyError(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetDiscardPile(nil)
	if err := g.PlayerDrawFromDiscard(); err == nil {
		t.Error("expected error for empty discard")
	}
}

func TestContractRummy_PlayerDrawFromDiscard_WrongPhase(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetPhase(ContractRummyPhasePlay)
	if err := g.PlayerDrawFromDiscard(); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("expected ErrWrongPhase, got %v", err)
	}
}

func TestContractRummy_DrawFromStock_RecyclesDiscardWhenEmpty(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)})
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("expected recycle to succeed, got %v", err)
	}
	if g.GetDrawPileCount() == 0 {
		t.Error("draw pile should be replenished from discard")
	}
}

func TestContractRummy_DrawFromStock_StockOutEndsRound(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{crCard(0, 5)}) // only top, can't recycle
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if g.GetPhase() != ContractRummyPhaseRoundEnd {
		t.Errorf("expected round end on stock-out, got %d", g.GetPhase())
	}
}

func TestContractRummy_PlayerMeldContract_Round1Success(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	p := g.GetPlayer(0)
	// 2 sets of 3: hearts/diamonds/spades 5 + clubs/hearts/diamonds K
	hand := []*Card{
		crCard(0, 5), crCard(1, 5), crCard(2, 5), // set of 5s
		crCard(3, 13), crCard(0, 13), crCard(1, 13), // set of Ks
		crCard(0, 2), crCard(1, 3), crCard(2, 4), crCard(3, 6), crCard(0, 7),
	}
	setHand(p, hand)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)

	err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}})
	if err != nil {
		t.Fatalf("PlayerMeldContract error: %v", err)
	}
	if !p.IsContractMet() {
		t.Error("contract should be met")
	}
	if p.GetMeldCount() != 2 {
		t.Errorf("meld count = %d, want 2", p.GetMeldCount())
	}
	if p.GetCardsSize() != 5 {
		t.Errorf("hand size = %d, want 5", p.GetCardsSize())
	}
}

func TestContractRummy_PlayerMeldContract_RejectsWrongCount(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(ContractRummyPhasePlay)
	err := g.PlayerMeldContract([][]int{{0, 1, 2}}) // only 1 slot, need 2
	if err == nil {
		t.Error("expected error for wrong slot count")
	}
}

func TestContractRummy_PlayerMeldContract_RejectsInvalidSet(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	p := g.GetPlayer(0)
	// not actually a valid set (mixed ranks)
	hand := []*Card{
		crCard(0, 5), crCard(1, 6), crCard(2, 7),
		crCard(3, 13), crCard(0, 13), crCard(1, 13),
		crCard(0, 2), crCard(1, 3), crCard(2, 4), crCard(3, 6), crCard(0, 7),
	}
	setHand(p, hand)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)

	err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}})
	if err == nil {
		t.Error("expected error for invalid first set")
	}
}

func TestContractRummy_PlayerMeldContract_Round3RunsSuccess(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(3)
	p := g.GetPlayer(0)
	// 2 runs of 4: spades 2-3-4-5 + diamonds 7-8-9-10
	hand := []*Card{
		crCard(0, 2), crCard(0, 3), crCard(0, 4), crCard(0, 5),
		crCard(1, 7), crCard(1, 8), crCard(1, 9), crCard(1, 10),
		crCard(2, 12), crCard(3, 13), crCard(0, 1),
	}
	setHand(p, hand)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)

	err := g.PlayerMeldContract([][]int{{0, 1, 2, 3}, {4, 5, 6, 7}})
	if err != nil {
		t.Fatalf("PlayerMeldContract error: %v", err)
	}
	if !p.IsContractMet() {
		t.Error("contract should be met")
	}
}

func TestContractRummy_PlayerMeldContract_RejectsDuplicateIndices(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(ContractRummyPhasePlay)
	err := g.PlayerMeldContract([][]int{{0, 1, 2}, {2, 3, 4}})
	if err == nil {
		t.Error("expected error for duplicate index across slots")
	}
}

func TestContractRummy_PlayerMeldContract_RejectsAfterMet(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(ContractRummyPhasePlay)
	g.GetPlayer(0).SetContractMet(true)
	err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}})
	if err == nil {
		t.Error("expected error when contract already met")
	}
}

func TestContractRummy_PlayerLayoff_Success(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	p := g.GetPlayer(0)
	setHand(p, []*Card{crCard(3, 5)}) // only one card, can layoff
	p.SetContractMet(true)
	p.AppendMeld([]*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)

	err := g.PlayerLayoff(0, 0, 0)
	if err != nil {
		t.Fatalf("PlayerLayoff error: %v", err)
	}
	if p.GetCardsSize() != 0 {
		t.Errorf("hand should be empty after layoff, got %d", p.GetCardsSize())
	}
	if p.GetMeld(0) == nil || len(p.GetMeld(0)) != 4 {
		t.Errorf("meld should have grown to 4, got %v", p.GetMeld(0))
	}
}

func TestContractRummy_PlayerLayoff_BeforeContractMet(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(ContractRummyPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(1).SetContractMet(true)
	g.GetPlayer(1).AppendMeld([]*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)})
	setHand(g.GetPlayer(0), []*Card{crCard(3, 5)})
	err := g.PlayerLayoff(1, 0, 0)
	if err == nil {
		t.Error("layoff should fail before player meets contract")
	}
}

func TestContractRummy_PlayerLayoff_TargetNotMet(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	g.SetPhase(ContractRummyPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).SetContractMet(true)
	g.GetPlayer(0).AppendMeld([]*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)})
	setHand(g.GetPlayer(0), []*Card{crCard(3, 5)})
	err := g.PlayerLayoff(1, 0, 0) // target player 1 has no melds
	if err == nil {
		t.Error("layoff should fail when target player not met")
	}
}

func TestContractRummy_PlayerDiscard_AdvancesTurn(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	cardsBefore := g.GetPlayer(0).GetCardsSize()
	err := g.PlayerDiscard(0)
	if err != nil {
		t.Fatalf("PlayerDiscard error: %v", err)
	}
	if g.GetCurrentPlayerIdx() != 1 {
		t.Errorf("currentPlayerIdx = %d, want 1", g.GetCurrentPlayerIdx())
	}
	if g.GetPhase() != ContractRummyPhaseDraw {
		t.Errorf("phase = %d, want PhaseDraw", g.GetPhase())
	}
	if g.GetPlayer(0).GetCardsSize() != cardsBefore-1 {
		t.Errorf("hand should shrink by 1")
	}
}

func TestContractRummy_PlayerDiscard_LastCardWithoutContractFails(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	setHand(g.GetPlayer(0), []*Card{crCard(0, 5)})
	g.GetPlayer(0).SetContractMet(false)
	err := g.PlayerDiscard(0)
	if err == nil {
		t.Error("discarding last card before meeting contract should fail")
	}
}

func TestContractRummy_PlayerDiscard_LastCardAfterContractWins(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	setHand(g.GetPlayer(0), []*Card{crCard(0, 5)})
	g.GetPlayer(0).SetContractMet(true)
	g.GetPlayer(0).AppendMeld([]*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)})
	g.GetPlayer(0).AppendMeld([]*Card{crCard(0, 13), crCard(1, 13), crCard(2, 13)})
	err := g.PlayerDiscard(0)
	if err != nil {
		t.Fatalf("PlayerDiscard error: %v", err)
	}
	if g.GetPhase() != ContractRummyPhaseRoundEnd {
		t.Errorf("phase = %d, want PhaseRoundEnd", g.GetPhase())
	}
	if g.GetRoundWinnerIdx() != 0 {
		t.Errorf("roundWinnerIdx = %d, want 0", g.GetRoundWinnerIdx())
	}
}

func TestContractRummy_PlayerMeldExtra_BeforeContractFails(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	err := g.PlayerMeldExtra([]int{0, 1, 2})
	if err == nil {
		t.Error("expected failure: extra meld before contract met")
	}
}

func TestContractRummy_PlayerMeldExtra_Success(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	p := g.GetPlayer(0)
	p.SetContractMet(true)
	setHand(p, []*Card{crCard(0, 9), crCard(1, 9), crCard(2, 9), crCard(3, 11)})
	err := g.PlayerMeldExtra([]int{0, 1, 2})
	if err != nil {
		t.Fatalf("PlayerMeldExtra error: %v", err)
	}
	if p.GetMeldCount() != 1 || len(p.GetMeld(0)) != 3 {
		t.Errorf("meld not added correctly: %v", p.GetMeld(0))
	}
	if p.GetCardsSize() != 1 {
		t.Errorf("hand size = %d, want 1", p.GetCardsSize())
	}
}

func TestContractRummy_PlayerMeldExtra_InvalidMeldFails(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	p := g.GetPlayer(0)
	p.SetContractMet(true)
	setHand(p, []*Card{crCard(0, 9), crCard(1, 4), crCard(2, 11)})
	err := g.PlayerMeldExtra([]int{0, 1, 2})
	if err == nil {
		t.Error("expected error: invalid meld")
	}
}

func TestContractRummy_FinishRound_PenalizesNonWinners(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		setHand(g.GetPlayer(i), []*Card{crCard(0, 13), crCard(1, 13)}) // 20 each
	}
	g.GetPlayer(0).SetContractMet(true)
	g.finishRound(0)
	if g.GetPlayer(0).GetCumulativeScore() != 0 {
		t.Errorf("winner score = %d, want 0", g.GetPlayer(0).GetCumulativeScore())
	}
	for i := 1; i < g.GetPlayerCnt(); i++ {
		got := g.GetPlayer(i).GetCumulativeScore()
		if got <= 0 {
			t.Errorf("player %d should be penalized, got %d", i, got)
		}
	}
}

func TestContractRummy_FinishRound_FailedContractAddsPenalty(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		setHand(g.GetPlayer(i), []*Card{crCard(0, 5)}) // 5 each
		g.GetPlayer(i).SetContractMet(false)
	}
	g.GetPlayer(0).SetContractMet(true) // winner met contract
	g.finishRound(0)
	expectedPenalty := 5 + g.GetConfig().FailContractPenalty
	if g.GetPlayer(1).GetCumulativeScore() != expectedPenalty {
		t.Errorf("player 1 score = %d, want %d", g.GetPlayer(1).GetCumulativeScore(), expectedPenalty)
	}
}

func TestContractRummy_FinishRound_FinalRoundEndsGame(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(ContractRummyTotalRounds)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		setHand(g.GetPlayer(i), []*Card{crCard(0, 5)})
	}
	g.GetPlayer(0).SetContractMet(true)
	g.finishRound(0)
	if !g.GetGameEndFlag() {
		t.Error("game should end after final round")
	}
}

func TestContractRummy_FinishRound_DrawNoWinnerStockOut(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	g.endRoundStockOut()
	if g.GetRoundWinnerIdx() != -1 {
		t.Errorf("roundWinnerIdx = %d, want -1", g.GetRoundWinnerIdx())
	}
}

func TestContractRummy_CpuPlay_NoOpForHumanTurn(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(0) // human
	g.CpuPlay()
	if g.GetPhase() != ContractRummyPhaseDraw {
		t.Error("CpuPlay should be a no-op for human turn")
	}
}

func TestContractRummy_CpuPlay_DrawAndDiscard(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(1) // CPU
	g.CpuPlay()              // draw
	g.CpuPlay()              // play+discard
	if g.GetCurrentPlayerIdx() != 2 {
		t.Errorf("after CPU 1 turn, currentPlayerIdx = %d, want 2", g.GetCurrentPlayerIdx())
	}
}

func TestContractRummy_CpuPlay_NoActionAfterGameEnd(t *testing.T) {
	g := helperContractRummyHand(t)
	g.gameEndFlag = true
	g.SetCurrentPlayerIdx(1)
	g.CpuPlay()
	// nothing should happen (no panic)
}

func TestContractRummy_FullRound_HumanWins(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhaseDraw)

	// Set human hand to satisfy R1 contract immediately upon meld
	hand := []*Card{
		crCard(0, 5), crCard(1, 5), crCard(2, 5),
		crCard(3, 13), crCard(0, 13), crCard(1, 13),
		crCard(0, 2), crCard(1, 3), crCard(2, 4), crCard(3, 6), crCard(0, 7),
	}
	setHand(g.GetPlayer(0), hand)
	g.SetDrawPile([]*Card{crCard(2, 1)})
	g.SetDiscardPile([]*Card{crCard(3, 7)})

	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("draw: %v", err)
	}
	// drew card → 12 cards in hand. The drawn card is added to end (or sorted in), so re-find indices.
	// Re-set hand to a known state and meld:
	setHand(g.GetPlayer(0), hand)
	if err := g.PlayerMeldContract([][]int{{0, 1, 2}, {3, 4, 5}}); err != nil {
		t.Fatalf("meld: %v", err)
	}
	if !g.GetPlayer(0).IsContractMet() {
		t.Error("contract should be met")
	}
}

func TestContractRummy_GetCurrentContract_ReturnsCorrectSlots(t *testing.T) {
	g := NewDefaultContractRummy()
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

func TestContractForRound_OutOfRange(t *testing.T) {
	if c := ContractForRound(0); len(c.Slots) != 0 {
		t.Error("R0 should be empty")
	}
	if c := ContractForRound(99); len(c.Slots) != 0 {
		t.Error("R99 should be empty")
	}
}

func TestValidateContractSlot_Variants(t *testing.T) {
	setSlot := ContractSlot{Kind: ContractSlotSet, Size: 3}
	runSlot := ContractSlot{Kind: ContractSlotRun, Size: 4}

	tests := []struct {
		name  string
		slot  ContractSlot
		cards []*Card
		want  bool
	}{
		{"valid set", setSlot, []*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)}, true},
		{"set wrong size", setSlot, []*Card{crCard(0, 5), crCard(1, 5)}, false},
		// 2 デッキ運用ではスート重複を許容するため、同ランクなら 3 枚で成立
		{"set with duplicate suit (2-deck)", setSlot, []*Card{crCard(0, 5), crCard(0, 5), crCard(2, 5)}, true},
		{"set wrong rank", setSlot, []*Card{crCard(0, 5), crCard(1, 6), crCard(2, 5)}, false},
		// Ace-high run
		{"valid Ace-high run", runSlot, []*Card{crCard(0, 11), crCard(0, 12), crCard(0, 13), crCard(0, 1)}, true},
		{"valid run", runSlot, []*Card{crCard(0, 2), crCard(0, 3), crCard(0, 4), crCard(0, 5)}, true},
		{"run wrong suit", runSlot, []*Card{crCard(0, 2), crCard(1, 3), crCard(0, 4), crCard(0, 5)}, false},
		{"run not consecutive", runSlot, []*Card{crCard(0, 2), crCard(0, 3), crCard(0, 5), crCard(0, 6)}, false},
		{"unknown kind", ContractSlot{Kind: 99, Size: 3}, []*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateContractSlot(tt.slot, tt.cards); got != tt.want {
				t.Errorf("ValidateContractSlot(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsContractRummyMeld_Cases(t *testing.T) {
	tests := []struct {
		name  string
		cards []*Card
		want  bool
	}{
		{"set of 3", []*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)}, true},
		{"set of 4", []*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5), crCard(3, 5)}, true},
		{"too small", []*Card{crCard(0, 5), crCard(1, 5)}, false},
		{"run of 3", []*Card{crCard(0, 5), crCard(0, 6), crCard(0, 7)}, true},
		{"run of 5", []*Card{crCard(0, 5), crCard(0, 6), crCard(0, 7), crCard(0, 8), crCard(0, 9)}, true},
		{"run J-Q-K-A (Ace high)", []*Card{crCard(0, 11), crCard(0, 12), crCard(0, 13), crCard(0, 1)}, true},
		{"run A-2-3 (Ace low)", []*Card{crCard(0, 1), crCard(0, 2), crCard(0, 3)}, true},
		// 2 デッキでは同スート同ランクの重複を許容（同ランクなのでセット成立）
		{"set with duplicate suit (2-deck)", []*Card{crCard(0, 5), crCard(0, 5), crCard(1, 5)}, true},
		{"invalid mixed", []*Card{crCard(0, 5), crCard(0, 6), crCard(2, 8)}, false},
		{"wraparound K-A-2 not allowed", []*Card{crCard(0, 13), crCard(0, 1), crCard(0, 2)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsContractRummyMeld(tt.cards); got != tt.want {
				t.Errorf("IsContractRummyMeld(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFindContractMeld_Round1(t *testing.T) {
	contract := ContractForRound(1)
	cards := []*Card{
		crCard(0, 5), crCard(1, 5), crCard(2, 5),
		crCard(0, 13), crCard(1, 13), crCard(2, 13),
		crCard(3, 7),
	}
	groups, ok := FindContractMeld(contract, cards)
	if !ok {
		t.Fatal("expected contract to be findable")
	}
	if len(groups) != 2 {
		t.Errorf("groups = %d, want 2", len(groups))
	}
}

func TestFindContractMeld_NotPossible(t *testing.T) {
	contract := ContractForRound(1)
	cards := []*Card{
		crCard(0, 5), crCard(1, 6), crCard(2, 7),
	}
	if _, ok := FindContractMeld(contract, cards); ok {
		t.Error("expected contract to be impossible")
	}
}

func TestFindContractMeld_EmptyContract(t *testing.T) {
	if _, ok := FindContractMeld(Contract{}, []*Card{crCard(0, 5)}); ok {
		t.Error("empty contract should return false")
	}
}

func TestContractRummyConfig_Validate(t *testing.T) {
	c := DefaultContractRummyConfig()
	if err := c.Validate(); err != nil {
		t.Errorf("default config should validate: %v", err)
	}
	bad := DefaultContractRummyConfig()
	bad.CpuDifficulty = -1
	if err := bad.Validate(); err == nil {
		t.Error("expected validation to fail")
	}
	bad2 := DefaultContractRummyConfig()
	bad2.FailContractPenalty = -1
	if err := bad2.Validate(); err == nil {
		t.Error("expected fail penalty validation to fail")
	}
}

func TestContractRummy_JSONRoundTrip(t *testing.T) {
	g := NewDefaultContractRummy()
	g.Reset()
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored ContractRummy
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

func TestContractRummy_JSONRejectsOversized(t *testing.T) {
	// craft a JSON with too many players
	bigPlayers := make([]map[string]any, contractRummyMaxSliceLen+1)
	for i := range bigPlayers {
		bigPlayers[i] = map[string]any{}
	}
	doc := map[string]any{"pl": bigPlayers}
	data, _ := json.Marshal(doc)
	var g ContractRummy
	if err := json.Unmarshal(data, &g); err == nil {
		t.Error("expected oversize rejection")
	}
}

func TestContractRummyPlayer_JSONRoundTrip(t *testing.T) {
	p := NewContractRummyPlayer(true)
	p.AddCard(crCard(0, 5))
	p.AppendMeld([]*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)})
	p.SetContractMet(true)
	p.SetContractIndex([]int{0})

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored ContractRummyPlayer
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

func TestContractRummyPlayer_ResetRoundClearsState(t *testing.T) {
	p := NewContractRummyPlayer(false)
	p.AddCard(crCard(0, 5))
	p.AppendMeld([]*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)})
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

func TestContractRummyPlayer_AddCardToMeldOutOfRange(t *testing.T) {
	p := NewContractRummyPlayer(false)
	if p.AddCardToMeld(99, crCard(0, 5)) {
		t.Error("AddCardToMeld should reject out-of-range index")
	}
	if p.GetMeld(99) != nil {
		t.Error("GetMeld should return nil for out-of-range")
	}
}

func TestContractRummyPlayer_SetContractIndexNilCopy(t *testing.T) {
	p := NewContractRummyPlayer(false)
	p.SetContractIndex(nil)
	if p.GetContractIndex() != nil {
		t.Error("SetContractIndex(nil) should clear")
	}
	src := []int{1, 2, 3}
	p.SetContractIndex(src)
	src[0] = 99 // mutate source
	if p.GetContractIndex()[0] == 99 {
		t.Error("SetContractIndex should defensively copy")
	}
}

func TestContractRummy_Penalty(t *testing.T) {
	if got := contractRummyCardPenalty(crCard(0, 1)); got != 15 {
		t.Errorf("Ace penalty = %d, want 15", got)
	}
	if got := contractRummyCardPenalty(crCard(0, 13)); got != 10 {
		t.Errorf("King penalty = %d, want 10", got)
	}
	if got := contractRummyCardPenalty(crCard(0, 5)); got != 5 {
		t.Errorf("5 penalty = %d, want 5", got)
	}
}

func TestContractRummy_PlayerNameFallback(t *testing.T) {
	g := helperContractRummyHand(t)
	if got := playerName(g.players, -1); got == "" {
		t.Error("playerName(-1) should return non-empty fallback")
	}
	if got := playerName(g.players, 99); got == "" {
		t.Error("playerName(99) should return non-empty fallback")
	}
}

func TestContractRummy_CanAddToContractRummyMeld(t *testing.T) {
	tests := []struct {
		name string
		meld []*Card
		card *Card
		want bool
	}{
		{"empty meld", nil, crCard(0, 5), false},
		{"nil card", []*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)}, nil, false},
		{"set: same rank, any suit allowed (2-deck)", []*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)}, crCard(0, 5), true},
		{"set: different rank rejected", []*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)}, crCard(0, 6), false},
		{"run: extend low end", []*Card{crCard(0, 5), crCard(0, 6), crCard(0, 7)}, crCard(0, 4), true},
		{"run: extend high end", []*Card{crCard(0, 5), crCard(0, 6), crCard(0, 7)}, crCard(0, 8), true},
		{"run: wrong suit rejected", []*Card{crCard(0, 5), crCard(0, 6), crCard(0, 7)}, crCard(1, 8), false},
		{"run: gap rejected", []*Card{crCard(0, 5), crCard(0, 6), crCard(0, 7)}, crCard(0, 9), false},
		{"run: Ace-high on K-end", []*Card{crCard(0, 11), crCard(0, 12), crCard(0, 13)}, crCard(0, 1), true},
		{"run: Ace-low extension", []*Card{crCard(0, 2), crCard(0, 3), crCard(0, 4)}, crCard(0, 1), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canAddToContractRummyMeld(tt.meld, tt.card); got != tt.want {
				t.Errorf("canAddToContractRummyMeld(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestContractRummy_AceHighRun_Round3CpuPlanning(t *testing.T) {
	// Verify FindContractMeld can pick a J-Q-K-A run for round 3 (2 runs of 4).
	contract := ContractForRound(3)
	cards := []*Card{
		crCard(0, 11), crCard(0, 12), crCard(0, 13), crCard(0, 1), // ♠ J-Q-K-A
		crCard(1, 2), crCard(1, 3), crCard(1, 4), crCard(1, 5), // ♥ 2-3-4-5
		crCard(2, 8),
	}
	groups, ok := FindContractMeld(contract, cards)
	if !ok {
		t.Fatal("expected J-Q-K-A + 2-3-4-5 to satisfy R3 contract")
	}
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestContractRummy_FindExtraMeld_AceHigh(t *testing.T) {
	cards := []*Card{
		crCard(0, 11), crCard(0, 12), crCard(0, 13), crCard(0, 1), // ♠ J-Q-K-A
		crCard(1, 7),
	}
	got := findExtraMeld(cards)
	if got == nil {
		t.Fatal("expected an extra meld of size >=3 from J-Q-K-A")
	}
	if len(got) < 3 {
		t.Errorf("expected meld len >= 3, got %d", len(got))
	}
}

func TestContractRummy_FindExtraMeld_DuplicateSuitSet(t *testing.T) {
	// Two ♠5s + one ♥5 — duplicate-suit set is allowed in 2-deck game.
	cards := []*Card{crCard(0, 5), crCard(0, 5), crCard(1, 5), crCard(2, 7)}
	got := findExtraMeld(cards)
	if got == nil {
		t.Fatal("expected duplicate-suit set to be valid in 2-deck game")
	}
	if len(got) != 3 {
		t.Errorf("expected size 3, got %d", len(got))
	}
}

func TestContractRummy_LongestRun_AceHigh(t *testing.T) {
	// Ace alone should yield 1; J-Q-K-A should yield 4 (Ace-high path).
	if got := longestRun([]int{11, 12, 13, 1}); got != 4 {
		t.Errorf("longestRun(J,Q,K,A) = %d, want 4", got)
	}
	// A-2-3 should yield 3 (Ace-low path).
	if got := longestRun([]int{1, 2, 3}); got != 3 {
		t.Errorf("longestRun(A,2,3) = %d, want 3", got)
	}
	// Disjoint: 2-3 and 7-8-9 -> 3.
	if got := longestRun([]int{2, 3, 7, 8, 9}); got != 3 {
		t.Errorf("longestRun(2,3,7,8,9) = %d, want 3", got)
	}
	// Empty -> 0.
	if got := longestRun(nil); got != 0 {
		t.Errorf("longestRun(nil) = %d, want 0", got)
	}
	// Duplicates do not extend the run.
	if got := longestRun([]int{5, 5, 6}); got != 2 {
		t.Errorf("longestRun(5,5,6) = %d, want 2", got)
	}
}

func TestContractRummy_AceVariants(t *testing.T) {
	// No Ace -> single variant.
	got := aceVariants([]int{2, 5, 3})
	if len(got) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(got))
	}
	if got[0][0] != 2 || got[0][1] != 3 || got[0][2] != 5 {
		t.Errorf("expected sorted, got %v", got[0])
	}
	// Ace -> 2 variants (low + high).
	got = aceVariants([]int{1, 12, 13})
	if len(got) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(got))
	}
	// Empty -> 1 variant (empty).
	got = aceVariants(nil)
	if len(got) != 1 {
		t.Fatalf("expected 1 variant for empty, got %d", len(got))
	}
}

func TestContractRummy_PickRunOf3(t *testing.T) {
	cardLookup := func(v int) *Card { return crCard(0, v) }
	got := pickRunOf3([]int{5, 6, 7}, cardLookup)
	if len(got) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(got))
	}
	if got := pickRunOf3([]int{5, 7, 9}, cardLookup); got != nil {
		t.Errorf("expected nil for non-consecutive")
	}
	if got := pickRunOf3([]int{5}, cardLookup); got != nil {
		t.Errorf("expected nil for short")
	}
}

func TestContractRummy_ScoreContractProgress_DoubleSet(t *testing.T) {
	contract := Contract{Slots: []ContractSlot{
		{Kind: ContractSlotSet, Size: 3},
		{Kind: ContractSlotSet, Size: 3},
	}}
	// Six 5s should count as 2 sets, capped to setSlots=2.
	cards := []*Card{
		crCard(0, 5), crCard(1, 5), crCard(2, 5),
		crCard(0, 5), crCard(1, 5), crCard(2, 5),
		crCard(3, 9),
	}
	got := scoreContractProgress(contract, cards)
	if got < 20 {
		t.Errorf("expected at least 20 (2 sets * 10), got %d", got)
	}
}

func TestContractRummy_CpuFullRound_Smoke(t *testing.T) {
	g := NewDefaultContractRummy()
	g.Reset()
	// Drive the CPUs through a few turns; just assert no panic and progress.
	g.SetCurrentPlayerIdx(1) // start with CPU
	for i := 0; i < 50 && !g.GetGameEndFlag(); i++ {
		if g.IsHumanTurn() || g.GetPhase() == ContractRummyPhaseRoundEnd {
			break
		}
		g.CpuPlay()
	}
}

func TestContractRummy_CpuShouldTakeDiscard_HardLayoff(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(1)
	cfg := g.GetConfig()
	cfg.CpuDifficulty = ContractRummyCpuDifficultyHard
	g.SetConfig(cfg)
	cpu := g.GetPlayer(1)
	cpu.SetContractMet(true)
	cpu.AppendMeld([]*Card{crCard(0, 5), crCard(1, 5), crCard(2, 5)})
	g.SetDiscardPile([]*Card{crCard(3, 5)})
	if !g.cpuShouldTakeDiscard(crCard(3, 5)) {
		t.Error("Hard CPU with completed contract should take a layoff-able discard")
	}
}

func TestContractRummy_CpuShouldTakeDiscard_NoLayoff(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetCurrentPlayerIdx(1)
	cpu := g.GetPlayer(1)
	cpu.SetContractMet(true) // already opened, no layoff target available
	if g.cpuShouldTakeDiscard(crCard(3, 5)) {
		t.Error("Opened CPU with no layoff target should not take discard")
	}
}

func TestContractRummy_GetterSetterCoverage(t *testing.T) {
	g := helperContractRummyHand(t)
	g.SetConfig(DefaultContractRummyConfig())
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

// #5588: 難易度が実際に CPU の拾い方を変えることを固定する。Web に選択肢を出す
// にあたり、**選んでも何も変わらない**のでは意味がない (受け入れ条件3)。
//
// 拾うかどうかは乱数を含むので 1 回では測れない。同じ局面を何度も打たせて
// 「拾った割合」で比べる。
func TestContractRummy_CpuDifficultyChangesTheDiscardPickup(t *testing.T) {
	// コントラクト未達で、拾っても進捗が上がらない札を出す。ここで初めて
	// 難易度ごとの無作為性が効く。
	// **手札を固定する。**配りごとに変えると、測っているものが試行ごとに変わり、
	// 「進捗の上がらない札」が見つからない。バラバラのランクを持たせて、
	// プローブ (♦2) がどのセットにも列にも寄与しないようにする。
	hand := []*Card{
		crCard(0, 5), crCard(1, 7), crCard(2, 9), crCard(3, 11), crCard(0, 13),
	}
	rate := func(difficulty ContractRummyCpuDifficulty) float64 {
		took := 0
		const trials = 600
		for range trials {
			g := helperContractRummyHand(t)
			g.SetCurrentPlayerIdx(1)
			cfg := g.GetConfig()
			cfg.CpuDifficulty = difficulty
			g.SetConfig(cfg)
			cpu := g.GetPlayer(1)
			cpu.SetContractMet(false)
			setHand(cpu, hand)
			if g.cpuShouldTakeDiscard(crCard(3, 2)) {
				took++
			}
		}
		return float64(took) / float64(trials)
	}

	hard := rate(ContractRummyCpuDifficultyHard)
	normal := rate(ContractRummyCpuDifficultyNormal)
	easy := rate(ContractRummyCpuDifficultyEasy)

	// Hard は進捗の上がらない札を拾わない。
	assert.Zero(t, hard, "hard never takes a useless discard")
	// Easy ほど無駄拾いが多い。**順序を固定する**ので、3 値が同じ実装では通らない。
	assert.Greater(t, easy, normal, "easy takes useless discards more often than normal")
	assert.Greater(t, normal, hard, "normal takes them more often than hard")
}
