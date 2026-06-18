package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

// klCard ヘルパー: 指定スート・値のカードを生成する
func klCard(d, v int) *Card {
	return NewCard(d, v, true)
}

// klJoker ヘルパー: ジョーカーを生成する
func klJoker() *Card {
	return NewCard(CardDesignJoker, CardValueJoker, true)
}

// klSetHand プレイヤーの手札を差し替える
func klSetHand(p *KalookiPlayer, cards []*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// newTestKalooki デフォルト構成の Kalooki を返す（テスト用）
func newTestKalooki() *Kalooki {
	g := NewDefaultKalooki()
	g.Reset()
	return g
}

func TestNewDefaultKalooki_Construct(t *testing.T) {
	g := NewDefaultKalooki()
	if g == nil {
		t.Fatal("NewDefaultKalooki returned nil")
	}
	if g.GetPlayerCnt() != KalookiDefaultPlayers {
		t.Errorf("player count = %d, want %d", g.GetPlayerCnt(), KalookiDefaultPlayers)
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
	if g.GetPlayer(99) != nil {
		t.Error("out-of-range player should be nil")
	}
}

func TestKalooki_DeckIs106(t *testing.T) {
	tc := NewTrumpCardsWithDecks(2, 2)
	if got := tc.GetTotalCount(); got != 106 {
		t.Errorf("deck total = %d, want 106 (2 decks + 2 jokers)", got)
	}
}

func TestKalooki_Reset_DealsHand(t *testing.T) {
	g := newTestKalooki()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if got := g.GetPlayer(i).GetCardsSize(); got != KalookiHandSize {
			t.Errorf("player %d hand size = %d, want %d", i, got, KalookiHandSize)
		}
		if g.GetPlayer(i).HasOpened() {
			t.Errorf("player %d should not be opened at reset", i)
		}
	}
	if g.GetDiscardTop() == nil {
		t.Error("discard top should be set after Reset")
	}
	if g.GetPhase() != KalookiPhaseDraw {
		t.Errorf("phase = %d, want PhaseDraw", g.GetPhase())
	}
	// 山札 = 106 - (13*4) - 1 = 53
	if got := g.GetDrawPileCount(); got != 106-KalookiHandSize*4-1 {
		t.Errorf("draw pile count = %d, want %d", got, 106-KalookiHandSize*4-1)
	}
}

func TestKalooki_Reset_RebuildsPlayersOnConfigChange(t *testing.T) {
	g := NewDefaultKalooki()
	cfg := DefaultKalookiConfig()
	cfg.PlayerCount = 2
	g.SetConfig(cfg)
	g.Reset()
	if g.GetPlayerCnt() != 2 {
		t.Errorf("player count = %d, want 2 after config change", g.GetPlayerCnt())
	}
}

func TestKalookiConfig_Validate(t *testing.T) {
	if err := DefaultKalookiConfig().Validate(); err != nil {
		t.Errorf("default config should validate: %v", err)
	}
	bad := []KalookiConfig{
		{CpuDifficulty: KalookiCpuDifficulty(-1), PlayerCount: 4, OpeningThreshold: 51},
		{CpuDifficulty: KalookiCpuDifficultyNormal, PlayerCount: 1, OpeningThreshold: 51},
		{CpuDifficulty: KalookiCpuDifficultyNormal, PlayerCount: 5, OpeningThreshold: 51},
		{CpuDifficulty: KalookiCpuDifficultyNormal, PlayerCount: 4, OpeningThreshold: -1},
	}
	for i, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("config[%d] should be invalid", i)
		}
	}
}

func TestKalooki_CardValue(t *testing.T) {
	cases := []struct {
		card *Card
		want int
	}{
		{klCard(CardDesignSpade, 1), 15},  // Ace
		{klCard(CardDesignSpade, 13), 10}, // King
		{klCard(CardDesignSpade, 10), 10}, // Ten
		{klCard(CardDesignSpade, 9), 9},   // pip
		{klCard(CardDesignSpade, 5), 5},
		{klJoker(), KalookiJokerValue},
	}
	for _, c := range cases {
		if got := kalookiCardValue(c.card); got != c.want {
			t.Errorf("value of %v = %d, want %d", c.card, got, c.want)
		}
	}
}

func TestKalooki_MeldValue_JokerMultiplier(t *testing.T) {
	// Set of 7-7-7 = 21 base, no joker
	plain := []*Card{klCard(CardDesignSpade, 7), klCard(CardDesignHeart, 7), klCard(CardDesignDiamond, 7)}
	if got := kalookiMeldValue(plain); got != 21 {
		t.Errorf("plain meld value = %d, want 21", got)
	}
	// 7-7-Joker: base 7+7+15 = 29, *1.5 floor = 43
	withJoker := []*Card{klCard(CardDesignSpade, 7), klCard(CardDesignHeart, 7), klJoker()}
	if got := kalookiMeldValue(withJoker); got != 43 {
		t.Errorf("joker meld value = %d, want 43 (29*1.5 floor)", got)
	}
}

func TestKalooki_IsValidMeld(t *testing.T) {
	// valid set
	if !kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 5), klCard(CardDesignClover, 5)}) {
		t.Error("5-5-5 should be a valid set")
	}
	// valid run
	if !kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 4), klCard(CardDesignSpade, 5), klCard(CardDesignSpade, 6)}) {
		t.Error("S4-S5-S6 should be a valid run")
	}
	// set with joker
	if !kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 5), klJoker()}) {
		t.Error("5-5-Joker should be a valid set")
	}
	// run with joker filling a gap
	if !kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 4), klJoker(), klCard(CardDesignSpade, 6)}) {
		t.Error("S4-Joker-S6 should be a valid run")
	}
	// too short
	if kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 5)}) {
		t.Error("2 cards should not be a valid meld")
	}
	// mixed ranks / suits
	if kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 6), klCard(CardDesignClover, 7)}) {
		t.Error("5-6-7 of different suits is not valid")
	}
	// run wrong suit
	if kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 4), klCard(CardDesignHeart, 5), klCard(CardDesignSpade, 6)}) {
		t.Error("mixed-suit run is invalid")
	}
	// all jokers invalid
	if kalookiIsValidMeld([]*Card{klJoker(), klJoker(), klJoker()}) {
		t.Error("all-joker meld is invalid")
	}
	// run with duplicate values invalid
	if kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 5), klCard(CardDesignSpade, 5), klCard(CardDesignSpade, 6)}) {
		t.Error("run with duplicate value is invalid")
	}
	// Ace-high run Q-K-A
	if !kalookiIsValidMeld([]*Card{klCard(CardDesignSpade, 12), klCard(CardDesignSpade, 13), klCard(CardDesignSpade, 1)}) {
		t.Error("Q-K-A should be a valid run")
	}
}

func TestKalooki_DrawFromStock_ProgressesPhase(t *testing.T) {
	g := newTestKalooki()
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("PlayerDrawFromStock error: %v", err)
	}
	if g.GetPhase() != KalookiPhaseMeld {
		t.Errorf("phase = %d, want PhaseMeld", g.GetPhase())
	}
	if g.GetPlayer(0).GetCardsSize() != KalookiHandSize+1 {
		t.Error("hand should grow by 1 after draw")
	}
}

func TestKalooki_DrawFromStock_RejectsWrongPhase(t *testing.T) {
	g := newTestKalooki()
	g.SetPhase(KalookiPhaseMeld)
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
}

func TestKalooki_DrawFromStock_RejectsNotHuman(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(1)
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
}

func TestKalooki_DrawFromDiscard(t *testing.T) {
	g := newTestKalooki()
	top := g.GetDiscardTop()
	if err := g.PlayerDrawFromDiscard(); err != nil {
		t.Fatalf("PlayerDrawFromDiscard error: %v", err)
	}
	if g.GetPhase() != KalookiPhaseMeld {
		t.Errorf("phase = %d, want PhaseMeld", g.GetPhase())
	}
	// the drawn card is now in hand
	found := false
	for i := 0; i < g.GetPlayer(0).GetCardsSize(); i++ {
		if g.GetPlayer(0).GetCard(i) == top {
			found = true
		}
	}
	if !found {
		t.Error("discard top should be in hand after draw")
	}
}

func TestKalooki_DrawFromDiscard_EmptyPile(t *testing.T) {
	g := newTestKalooki()
	g.SetDiscardPile(nil)
	if err := g.PlayerDrawFromDiscard(); err == nil {
		t.Error("expected error when discard pile is empty")
	}
}

// klOpeningHand builds a hand that opens with two runs >= 51 points and leaves filler.
func klOpeningHand() []*Card {
	return []*Card{
		// Run 1: S-10,J,Q,K = 10+10+10+10 = 40
		klCard(CardDesignSpade, 10), klCard(CardDesignSpade, 11), klCard(CardDesignSpade, 12), klCard(CardDesignSpade, 13),
		// Set: A,A,A = 45 → total 85 >= 51
		klCard(CardDesignSpade, 1), klCard(CardDesignHeart, 1), klCard(CardDesignClover, 1),
		// filler
		klCard(CardDesignDiamond, 2), klCard(CardDesignDiamond, 4),
	}
}

func TestKalooki_Meld_OpeningRejectedBelowThreshold(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	// low-value set 2-2-2 = 6 points < 51
	klSetHand(g.GetPlayer(0), []*Card{
		klCard(CardDesignSpade, 2), klCard(CardDesignHeart, 2), klCard(CardDesignClover, 2),
		klCard(CardDesignDiamond, 9),
	})
	err := g.PlayerMeld([][]int{{0, 1, 2}})
	if err == nil {
		t.Fatal("expected opening below threshold to be rejected")
	}
	if g.GetPlayer(0).HasOpened() {
		t.Error("player should not be opened after a rejected meld")
	}
	if g.GetPlayer(0).GetMeldCount() != 0 {
		t.Error("no meld should be placed when opening rejected")
	}
}

func TestKalooki_Meld_OpeningAcceptedAtThreshold(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	klSetHand(g.GetPlayer(0), klOpeningHand())
	// meld the run (idx 0-3) and the set (idx 4-6): 40 + 45 = 85 >= 51
	if err := g.PlayerMeld([][]int{{0, 1, 2, 3}, {4, 5, 6}}); err != nil {
		t.Fatalf("opening meld should succeed: %v", err)
	}
	if !g.GetPlayer(0).HasOpened() {
		t.Error("player should be opened after a valid >=51 meld")
	}
	if g.GetPlayer(0).GetMeldCount() != 2 {
		t.Errorf("meld count = %d, want 2", g.GetPlayer(0).GetMeldCount())
	}
	if g.GetPlayer(0).GetCardsSize() != 2 {
		t.Errorf("hand size = %d, want 2 after melding 7 cards", g.GetPlayer(0).GetCardsSize())
	}
}

func TestKalooki_Meld_InvalidGroupRejected(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	klSetHand(g.GetPlayer(0), []*Card{
		klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 6), klCard(CardDesignClover, 7),
	})
	if err := g.PlayerMeld([][]int{{0, 1, 2}}); err == nil {
		t.Error("invalid (non-meld) group should be rejected")
	}
	// duplicate index
	klSetHand(g.GetPlayer(0), []*Card{klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 5), klCard(CardDesignClover, 5)})
	if err := g.PlayerMeld([][]int{{0, 0, 1}}); err == nil {
		t.Error("duplicate index should be rejected")
	}
	// out of range
	if err := g.PlayerMeld([][]int{{0, 1, 99}}); err == nil {
		t.Error("out-of-range index should be rejected")
	}
	// too few cards
	if err := g.PlayerMeld([][]int{{0, 1}}); err == nil {
		t.Error("too few cards should be rejected")
	}
	// empty groups
	if err := g.PlayerMeld([][]int{}); err == nil {
		t.Error("empty meld groups should be rejected")
	}
}

func TestKalooki_Meld_JokerMeldOpens(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	// A-A-Joker set: base 15+15+15=45, *1.5 floor = 67 >= 51
	klSetHand(g.GetPlayer(0), []*Card{
		klCard(CardDesignSpade, 1), klCard(CardDesignHeart, 1), klJoker(),
		klCard(CardDesignDiamond, 3),
	})
	if err := g.PlayerMeld([][]int{{0, 1, 2}}); err != nil {
		t.Fatalf("joker meld opening should succeed: %v", err)
	}
	if !g.GetPlayer(0).HasOpened() {
		t.Error("player should be opened after joker meld worth 67")
	}
}

func TestKalooki_Layoff_RequiresOpened(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	if err := g.PlayerLayoff(0, 0, 0); err == nil {
		t.Error("layoff should require opening first")
	}
}

func TestKalooki_Layoff_AddsCardToMeld(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	p := g.GetPlayer(0)
	p.SetHasOpened(true)
	p.SetMelds([][]*Card{{klCard(CardDesignSpade, 5), klCard(CardDesignSpade, 6), klCard(CardDesignSpade, 7)}})
	klSetHand(p, []*Card{klCard(CardDesignSpade, 8), klCard(CardDesignDiamond, 2)})
	if err := g.PlayerLayoff(0, 0, 0); err != nil {
		t.Fatalf("layoff should succeed: %v", err)
	}
	if len(p.GetMeld(0)) != 4 {
		t.Errorf("meld should grow to 4, got %d", len(p.GetMeld(0)))
	}
	if p.GetCardsSize() != 1 {
		t.Errorf("hand should shrink to 1, got %d", p.GetCardsSize())
	}
}

func TestKalooki_Layoff_RejectsInvalidCard(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	p := g.GetPlayer(0)
	p.SetHasOpened(true)
	p.SetMelds([][]*Card{{klCard(CardDesignSpade, 5), klCard(CardDesignSpade, 6), klCard(CardDesignSpade, 7)}})
	klSetHand(p, []*Card{klCard(CardDesignHeart, 13)})
	if err := g.PlayerLayoff(0, 0, 0); err == nil {
		t.Error("layoff of incompatible card should be rejected")
	}
	// bad target player
	if err := g.PlayerLayoff(99, 0, 0); err == nil {
		t.Error("bad target player should be rejected")
	}
	// target not opened
	g.GetPlayer(1).SetHasOpened(false)
	if err := g.PlayerLayoff(1, 0, 0); err == nil {
		t.Error("layoff to unopened player should be rejected")
	}
}

func TestKalooki_Discard_AdvancesTurn(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	klSetHand(g.GetPlayer(0), []*Card{klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 9)})
	if err := g.PlayerDiscard(0); err != nil {
		t.Fatalf("discard error: %v", err)
	}
	if g.GetDiscardTop().GetValue() != 5 {
		t.Error("discard top should be the discarded card")
	}
	if g.GetCurrentPlayerIdx() != 1 {
		t.Errorf("turn should advance to player 1, got %d", g.GetCurrentPlayerIdx())
	}
	if g.GetPhase() != KalookiPhaseDraw {
		t.Error("phase should reset to Draw after discard")
	}
}

func TestKalooki_Discard_GoOutWins(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	p := g.GetPlayer(0)
	p.SetHasOpened(true)
	klSetHand(p, []*Card{klCard(CardDesignSpade, 5)})
	// give opponents some deadwood
	klSetHand(g.GetPlayer(1), []*Card{klCard(CardDesignSpade, 1), klCard(CardDesignHeart, 13)}) // 15 + 10 = 25
	if err := g.PlayerDiscard(0); err != nil {
		t.Fatalf("discard error: %v", err)
	}
	if g.GetPhase() != KalookiPhaseRoundEnd {
		t.Errorf("phase = %d, want RoundEnd after going out", g.GetPhase())
	}
	if g.GetRoundWinnerIdx() != 0 {
		t.Errorf("round winner = %d, want 0", g.GetRoundWinnerIdx())
	}
	if g.GetPlayer(0).GetRoundScore() != 0 {
		t.Error("winner should have 0 round score")
	}
	if g.GetPlayer(1).GetRoundScore() != 25 {
		t.Errorf("player 1 deadwood = %d, want 25", g.GetPlayer(1).GetRoundScore())
	}
}

func TestKalooki_Discard_RejectsOutOfRange(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseMeld)
	klSetHand(g.GetPlayer(0), []*Card{klCard(CardDesignSpade, 5)})
	if err := g.PlayerDiscard(99); err == nil {
		t.Error("out-of-range discard should be rejected")
	}
}

func TestKalooki_NextRound_EndsGame(t *testing.T) {
	g := newTestKalooki()
	g.SetPhase(KalookiPhaseRoundEnd)
	g.NextRound()
	if !g.GetGameEndFlag() {
		t.Error("NextRound from RoundEnd should end the game")
	}
	if g.GetPhase() != KalookiPhaseGameEnd {
		t.Error("phase should be GameEnd")
	}
}

func TestKalooki_NextRound_NoOpWhenNotRoundEnd(t *testing.T) {
	g := newTestKalooki()
	g.NextRound() // phase is Draw
	if g.GetGameEndFlag() {
		t.Error("NextRound during Draw must not end game")
	}
}

func TestKalooki_FinalizeGameEnd_WinnerIsRoundWinner(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(2)
	g.SetPhase(KalookiPhaseMeld)
	p := g.GetPlayer(2)
	p.SetHasOpened(true)
	klSetHand(p, []*Card{klCard(CardDesignSpade, 5)})
	if err := g.PlayerDiscard(0); err != nil {
		t.Fatalf("discard error: %v", err)
	}
	g.NextRound()
	if g.GetWinnerIdx() != 2 {
		t.Errorf("winner = %d, want 2 (round winner)", g.GetWinnerIdx())
	}
}

func TestKalooki_StockOutEndsRound(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseDraw)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{klCard(CardDesignSpade, 5)}) // only 1 card → cannot recycle
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("draw error: %v", err)
	}
	if g.GetPhase() != KalookiPhaseRoundEnd {
		t.Errorf("phase = %d, want RoundEnd on stock-out", g.GetPhase())
	}
	if g.GetRoundWinnerIdx() != -1 {
		t.Error("stock-out should produce no round winner (-1)")
	}
}

func TestKalooki_StockRecycle(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(KalookiPhaseDraw)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{klCard(CardDesignSpade, 5), klCard(CardDesignHeart, 6), klCard(CardDesignClover, 7)})
	if err := g.PlayerDrawFromStock(); err != nil {
		t.Fatalf("draw error: %v", err)
	}
	if g.GetPhase() != KalookiPhaseMeld {
		t.Error("recycle should allow drawing and progress to Meld phase")
	}
}

func TestKalooki_CpuPlay_FullGameTerminates(t *testing.T) {
	g := NewDefaultKalooki()
	g.Reset()
	// Make all players CPU to force a fully automatic game.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).GamePlayer = NewGamePlayer(false)
	}
	for iter := 0; iter < 200000; iter++ {
		phase := g.GetPhase()
		if g.GetGameEndFlag() || phase == KalookiPhaseRoundEnd || phase == KalookiPhaseGameEnd {
			break
		}
		g.CpuPlay()
	}
	phase := g.GetPhase()
	if phase != KalookiPhaseRoundEnd && phase != KalookiPhaseGameEnd {
		t.Errorf("full-CPU game did not terminate, phase = %d", phase)
	}
}

func TestKalooki_CpuPlay_NoOpWhenHuman(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	before := g.GetPlayer(0).GetCardsSize()
	g.CpuPlay() // player 0 is human → no-op
	if g.GetPlayer(0).GetCardsSize() != before {
		t.Error("CpuPlay should be a no-op on a human turn")
	}
}

func TestKalooki_GameEndedGuards(t *testing.T) {
	g := newTestKalooki()
	g.SetPhase(KalookiPhaseGameEnd)
	g.finalizeGameEnd()
	if !g.GetGameEndFlag() {
		t.Fatal("game should be ended")
	}
	if err := g.PlayerDrawFromStock(); !errors.Is(err, ErrGameEnded) {
		t.Errorf("draw stock err = %v, want ErrGameEnded", err)
	}
	if err := g.PlayerDrawFromDiscard(); !errors.Is(err, ErrGameEnded) {
		t.Errorf("draw discard err = %v, want ErrGameEnded", err)
	}
	if err := g.PlayerMeld([][]int{{0, 1, 2}}); !errors.Is(err, ErrGameEnded) {
		t.Errorf("meld err = %v, want ErrGameEnded", err)
	}
	if err := g.PlayerLayoff(0, 0, 0); !errors.Is(err, ErrGameEnded) {
		t.Errorf("layoff err = %v, want ErrGameEnded", err)
	}
	if err := g.PlayerDiscard(0); !errors.Is(err, ErrGameEnded) {
		t.Errorf("discard err = %v, want ErrGameEnded", err)
	}
}

func TestKalooki_JSONRoundTrip(t *testing.T) {
	g := NewDefaultKalooki()
	g.Reset()
	data, err := g.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var g2 Kalooki
	if err := g2.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if g2.GetPlayerCnt() != g.GetPlayerCnt() {
		t.Errorf("player count mismatch after round-trip")
	}
	if g2.GetPhase() != g.GetPhase() {
		t.Error("phase mismatch after round-trip")
	}
}

func TestKalooki_UnmarshalJSON_RejectsInvalid(t *testing.T) {
	// bad json
	var g Kalooki
	if err := g.UnmarshalJSON([]byte("not json")); err == nil {
		t.Error("malformed JSON should error")
	}
	// invalid config (player count 1)
	bad := kalookiJSON{
		Config:           KalookiConfig{CpuDifficulty: KalookiCpuDifficultyNormal, PlayerCount: 1, OpeningThreshold: 51},
		Players:          []*KalookiPlayer{NewKalookiPlayer(true)},
		Phase:            KalookiPhaseDraw,
		CurrentPlayerIdx: 0,
	}
	data, _ := json.Marshal(bad)
	if err := (&Kalooki{}).UnmarshalJSON(data); err == nil {
		t.Error("invalid config should be rejected")
	}
	// nil player element
	bad2 := kalookiJSON{
		Config:           DefaultKalookiConfig(),
		Players:          []*KalookiPlayer{NewKalookiPlayer(true), nil, NewKalookiPlayer(false), NewKalookiPlayer(false)},
		Phase:            KalookiPhaseDraw,
		CurrentPlayerIdx: 0,
	}
	data2, _ := json.Marshal(bad2)
	if err := (&Kalooki{}).UnmarshalJSON(data2); err == nil {
		t.Error("nil player element should be rejected")
	}
	// out of range current player
	bad3 := kalookiJSON{
		Config:           DefaultKalookiConfig(),
		Players:          []*KalookiPlayer{NewKalookiPlayer(true), NewKalookiPlayer(false), NewKalookiPlayer(false), NewKalookiPlayer(false)},
		Phase:            KalookiPhaseDraw,
		CurrentPlayerIdx: 99,
	}
	data3, _ := json.Marshal(bad3)
	if err := (&Kalooki{}).UnmarshalJSON(data3); err == nil {
		t.Error("out-of-range current player should be rejected")
	}
	// bad phase
	bad4 := kalookiJSON{
		Config:           DefaultKalookiConfig(),
		Players:          []*KalookiPlayer{NewKalookiPlayer(true), NewKalookiPlayer(false), NewKalookiPlayer(false), NewKalookiPlayer(false)},
		Phase:            KalookiPhase(99),
		CurrentPlayerIdx: 0,
	}
	data4, _ := json.Marshal(bad4)
	if err := (&Kalooki{}).UnmarshalJSON(data4); err == nil {
		t.Error("bad phase should be rejected")
	}
}

func TestKalookiPlayer_JSONRoundTrip(t *testing.T) {
	p := NewKalookiPlayer(true)
	p.AddCard(klCard(CardDesignSpade, 5))
	p.SetHasOpened(true)
	p.AppendMeld([]*Card{klCard(CardDesignHeart, 7), klCard(CardDesignHeart, 8), klCard(CardDesignHeart, 9)})
	data, err := p.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var p2 KalookiPlayer
	if err := p2.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !p2.HasOpened() {
		t.Error("hasOpened not preserved")
	}
	if p2.GetMeldCount() != 1 {
		t.Error("melds not preserved")
	}
}

func TestKalooki_GetActionLog(t *testing.T) {
	g := newTestKalooki()
	g.SetCurrentPlayerIdx(0)
	_ = g.PlayerDrawFromStock()
	if len(g.GetActionLog()) == 0 {
		t.Error("action log should record the draw")
	}
}
