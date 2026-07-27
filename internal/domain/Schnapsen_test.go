package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestSchnapsen() *domain.Schnapsen {
	s := domain.NewDefaultSchnapsen()
	s.Reset()
	return s
}

// setHand clears player p's hand and adds the given cards.
func schnSetHand(p *domain.SchnapsenPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func schnCard(suit, val int) *domain.Card { return domain.NewCard(suit, val, false) }

// drainStock repeatedly draws from the stock (via NextTrick replenishment)
// until the second phase (endgame) is reached. Player hands grow but are reset
// by callers via setHand before the assertion of interest.
func drainStock(s *domain.Schnapsen) {
	guard := 0
	for !s.IsEndgame() && guard < 50 {
		guard++
		s.SetPhase(domain.SchnapsenPhaseTrickEnd)
		s.SetLeadPlayerIdx(0)
		s.NextTrick()
	}
}

func TestNewDefaultSchnapsen(t *testing.T) {
	s := domain.NewDefaultSchnapsen()
	if got := s.GetPlayerCnt(); got != 2 {
		t.Errorf("GetPlayerCnt() = %d, want 2", got)
	}
	if !s.GetPlayer(0).GetIsHuman() {
		t.Error("player 0 should be human")
	}
	if s.GetPlayer(1).GetIsHuman() {
		t.Error("player 1 should be CPU")
	}
	if s.GetPlayer(5) != nil {
		t.Error("out-of-range player should be nil")
	}
}

func TestSchnapsen_Reset(t *testing.T) {
	s := newTestSchnapsen()
	for i := 0; i < 2; i++ {
		if got := s.GetPlayer(i).GetCardsSize(); got != domain.SchnapsenHandSize {
			t.Errorf("player %d hand size = %d, want %d", i, got, domain.SchnapsenHandSize)
		}
	}
	if s.GetTrumpCard() == nil {
		t.Error("trump card should be set after Reset")
	}
	if s.GetTrumpSuit() == 0 {
		t.Error("trump suit should be set after Reset")
	}
	// 20 - 5*2 - 1 (trump upcard) = 9 cards remain in the stock.
	if got := s.GetStockRemaining(); got != 9 {
		t.Errorf("stock remaining = %d, want 9", got)
	}
	if s.GetPhase() != domain.SchnapsenPhasePlay {
		t.Errorf("phase = %v, want Play", s.GetPhase())
	}
	if s.GetGameEndFlag() {
		t.Error("game should not be ended after Reset")
	}
}

func TestSchnapsen_CardPointsAndRank(t *testing.T) {
	cases := []struct {
		val  int
		pts  int
		rank int
	}{
		{1, 11, 5},
		{10, 10, 4},
		{13, 4, 3},
		{12, 3, 2},
		{11, 2, 1},
	}
	for _, c := range cases {
		cd := schnCard(domain.CardDesignSpade, c.val)
		if got := domain.SchnapsenCardPoints(cd); got != c.pts {
			t.Errorf("SchnapsenCardPoints(val=%d) = %d, want %d", c.val, got, c.pts)
		}
		if got := domain.SchnapsenRankOrder(cd); got != c.rank {
			t.Errorf("SchnapsenRankOrder(val=%d) = %d, want %d", c.val, got, c.rank)
		}
	}
	if domain.SchnapsenCardPoints(nil) != 0 || domain.SchnapsenRankOrder(nil) != 0 {
		t.Error("nil card should return 0 for points and rank")
	}
}

func TestSchnapsen_TrickWinner(t *testing.T) {
	tests := []struct {
		name      string
		trumpSuit int
		lead      *domain.Card
		challenge *domain.Card
		want      int
	}{
		{"higher same suit wins", domain.CardDesignSpade,
			schnCard(domain.CardDesignClover, 13), schnCard(domain.CardDesignClover, 1), 1},
		{"lower same suit loses", domain.CardDesignSpade,
			schnCard(domain.CardDesignClover, 1), schnCard(domain.CardDesignClover, 13), 0},
		{"off-suit cannot win", domain.CardDesignSpade,
			schnCard(domain.CardDesignClover, 11), schnCard(domain.CardDesignHeart, 1), 0},
		{"trump beats non-trump", domain.CardDesignSpade,
			schnCard(domain.CardDesignClover, 1), schnCard(domain.CardDesignSpade, 11), 1},
		{"non-trump cannot beat trump lead", domain.CardDesignSpade,
			schnCard(domain.CardDesignSpade, 11), schnCard(domain.CardDesignHeart, 1), 0},
		{"higher trump beats lower trump", domain.CardDesignSpade,
			schnCard(domain.CardDesignSpade, 12), schnCard(domain.CardDesignSpade, 1), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSchnapsen()
			s.SetTrumpSuit(tt.trumpSuit)
			s.SetCurrentTrick([]*domain.TrickCard{
				{PlayerIdx: 0, Card: tt.lead},
				{PlayerIdx: 1, Card: tt.challenge},
			})
			s.SetPhase(domain.SchnapsenPhaseTrickEnd)
			s.ResolveTrick()
			if got := s.GetLeadPlayerIdx(); got != tt.want {
				t.Errorf("winner = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSchnapsen_ResolveTrick_AddsPointsAndWinsAt66(t *testing.T) {
	s := newTestSchnapsen()
	s.SetTrumpSuit(domain.CardDesignSpade)
	s.SetPlayerPoints(0, 60)
	s.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: schnCard(domain.CardDesignClover, 1)},  // A=11 (lead, wins)
		{PlayerIdx: 1, Card: schnCard(domain.CardDesignClover, 11)}, // J=2
	})
	s.SetPhase(domain.SchnapsenPhaseTrickEnd)
	s.ResolveTrick()
	if got := s.GetPlayerPoints(0); got != 73 {
		t.Errorf("player 0 points = %d, want 73", got)
	}
	if !s.GetGameEndFlag() {
		t.Error("game should end when reaching 66")
	}
	if s.GetWinnerIdx() != 0 {
		t.Errorf("winner = %d, want 0", s.GetWinnerIdx())
	}
}

func TestSchnapsen_ResolveTrick_GuardWrongPhase(t *testing.T) {
	s := newTestSchnapsen()
	s.SetPhase(domain.SchnapsenPhasePlay)
	s.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: schnCard(domain.CardDesignClover, 1)},
	})
	s.ResolveTrick() // not enough cards / wrong phase -> no-op
	if s.GetPlayerPoints(0) != 0 {
		t.Error("ResolveTrick should be a no-op in wrong phase")
	}
}

func TestSchnapsen_Marriage_Normal(t *testing.T) {
	s := newTestSchnapsen()
	s.SetTrumpSuit(domain.CardDesignSpade) // marriage in clover -> normal (20)
	s.SetCurrentPlayerIdx(0)
	s.SetLeadPlayerIdx(0)
	s.SetCurrentTrick(nil)
	schnSetHand(s.GetPlayer(0),
		schnCard(domain.CardDesignClover, 13), // K
		schnCard(domain.CardDesignClover, 12), // Q
		schnCard(domain.CardDesignHeart, 1),
	)
	idxs := s.GetMarriageIndices(0)
	if len(idxs) != 2 {
		t.Fatalf("marriage indices = %v, want 2 entries", idxs)
	}
	if err := s.PlayerDeclareMarriage(0); err != nil { // declare on K
		t.Fatalf("declare marriage err: %v", err)
	}
	if got := s.GetPlayerPoints(0); got != 20 {
		t.Errorf("points after normal marriage = %d, want 20", got)
	}
	// The K was led, so the current trick should contain one card.
	if len(s.GetCurrentTrick()) != 1 {
		t.Errorf("current trick len = %d, want 1 (marriage card led)", len(s.GetCurrentTrick()))
	}
}

func TestSchnapsen_Marriage_RoyalAndWin(t *testing.T) {
	s := newTestSchnapsen()
	s.SetTrumpSuit(domain.CardDesignSpade) // marriage in spade -> royal (40)
	s.SetCurrentPlayerIdx(0)
	s.SetLeadPlayerIdx(0)
	s.SetCurrentTrick(nil)
	s.SetPlayerPoints(0, 30)
	schnSetHand(s.GetPlayer(0),
		schnCard(domain.CardDesignSpade, 12), // Q trump
		schnCard(domain.CardDesignSpade, 13), // K trump
	)
	if err := s.PlayerDeclareMarriage(0); err != nil {
		t.Fatalf("declare royal marriage err: %v", err)
	}
	if got := s.GetPlayerPoints(0); got != 70 {
		t.Errorf("points = %d, want 70 (30+40)", got)
	}
	if !s.GetGameEndFlag() {
		t.Error("declaring royal marriage to reach 66 should end the game")
	}
	if s.GetWinnerIdx() != 0 {
		t.Errorf("winner = %d, want 0", s.GetWinnerIdx())
	}
}

func TestSchnapsen_Marriage_Errors(t *testing.T) {
	s := newTestSchnapsen()
	s.SetTrumpSuit(domain.CardDesignSpade)
	s.SetCurrentPlayerIdx(0)
	s.SetLeadPlayerIdx(0)
	s.SetCurrentTrick(nil)
	schnSetHand(s.GetPlayer(0),
		schnCard(domain.CardDesignClover, 13), // K, no partner Q
		schnCard(domain.CardDesignHeart, 1),
	)
	if err := s.PlayerDeclareMarriage(0); err == nil {
		t.Error("declaring without partner should error")
	}
	if err := s.PlayerDeclareMarriage(99); err == nil {
		t.Error("out-of-range index should error")
	}
	// Not on lead (trick has a card already)
	s.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: schnCard(domain.CardDesignHeart, 1)}})
	schnSetHand(s.GetPlayer(0), schnCard(domain.CardDesignClover, 13), schnCard(domain.CardDesignClover, 12))
	if err := s.PlayerDeclareMarriage(0); err == nil {
		t.Error("declaring when not on lead should error")
	}
}

func TestSchnapsen_Marriage_NotHumanTurn(t *testing.T) {
	s := newTestSchnapsen()
	s.SetCurrentPlayerIdx(1) // CPU
	if err := s.PlayerDeclareMarriage(0); !errors.Is(err, domain.ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
}

func TestSchnapsen_PlayerPlay_Validations(t *testing.T) {
	s := newTestSchnapsen()
	s.SetGameEndFlag(true)
	if err := s.PlayerPlay(0); !errors.Is(err, domain.ErrGameEnded) {
		t.Errorf("err = %v, want ErrGameEnded", err)
	}
	s.SetGameEndFlag(false)
	s.SetPhase(domain.SchnapsenPhaseTrickEnd)
	if err := s.PlayerPlay(0); !errors.Is(err, domain.ErrWrongPhase) {
		t.Errorf("err = %v, want ErrWrongPhase", err)
	}
	s.SetPhase(domain.SchnapsenPhasePlay)
	s.SetCurrentPlayerIdx(1) // CPU
	if err := s.PlayerPlay(0); !errors.Is(err, domain.ErrNotHumanTurn) {
		t.Errorf("err = %v, want ErrNotHumanTurn", err)
	}
	s.SetCurrentPlayerIdx(0)
	if err := s.PlayerPlay(99); !errors.Is(err, domain.ErrInvalidCard) {
		t.Errorf("err = %v, want ErrInvalidCard", err)
	}
}

func TestSchnapsen_PlayerPlay_LeadsCard(t *testing.T) {
	s := newTestSchnapsen()
	s.SetPhase(domain.SchnapsenPhasePlay)
	s.SetCurrentPlayerIdx(0)
	s.SetCurrentTrick(nil)
	schnSetHand(s.GetPlayer(0), schnCard(domain.CardDesignHeart, 1), schnCard(domain.CardDesignClover, 11))
	if err := s.PlayerPlay(0); err != nil {
		t.Fatalf("PlayerPlay err: %v", err)
	}
	if len(s.GetCurrentTrick()) != 1 {
		t.Errorf("trick len = %d, want 1", len(s.GetCurrentTrick()))
	}
	if s.GetCurrentPlayerIdx() != 1 {
		t.Errorf("turn should advance to 1, got %d", s.GetCurrentPlayerIdx())
	}
}

func TestSchnapsen_EndgameFollowRules(t *testing.T) {
	s := newTestSchnapsen()
	s.SetTrumpSuit(domain.CardDesignSpade)
	drainStock(s)
	if !s.IsEndgame() {
		t.Fatal("expected endgame after draining stock")
	}
	s.SetPhase(domain.SchnapsenPhasePlay)
	s.SetCurrentPlayerIdx(1)
	s.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: schnCard(domain.CardDesignClover, 13)}, // K clover lead
	})
	// Follower has clover A (wins) and clover J (loses) and a heart.
	schnSetHand(s.GetPlayer(1),
		schnCard(domain.CardDesignClover, 1),  // A: must play (must win)
		schnCard(domain.CardDesignClover, 11), // J: same suit but loses -> illegal
		schnCard(domain.CardDesignHeart, 10),  // off-suit -> illegal (has lead suit)
	)
	valid := s.GetValidPlayIndices(1)
	if len(valid) != 1 || valid[0] != 0 {
		t.Errorf("valid indices = %v, want [0] (must win with A)", valid)
	}
}

func TestSchnapsen_EndgameMustTrumpWhenVoid(t *testing.T) {
	s := newTestSchnapsen()
	s.SetTrumpSuit(domain.CardDesignSpade)
	drainStock(s)
	s.SetCurrentPlayerIdx(1)
	s.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: schnCard(domain.CardDesignClover, 13)},
	})
	schnSetHand(s.GetPlayer(1),
		schnCard(domain.CardDesignSpade, 11), // trump -> must play
		schnCard(domain.CardDesignHeart, 10), // off-suit non-trump -> illegal
	)
	valid := s.GetValidPlayIndices(1)
	if len(valid) != 1 || valid[0] != 0 {
		t.Errorf("valid indices = %v, want [0] (must trump)", valid)
	}
}

func TestSchnapsen_NextTrick_DrawAndAdvance(t *testing.T) {
	s := newTestSchnapsen()
	s.SetPhase(domain.SchnapsenPhaseTrickEnd)
	s.SetLeadPlayerIdx(0)
	before0 := s.GetPlayer(0).GetCardsSize()
	stockBefore := s.GetStockRemaining()
	s.NextTrick()
	if s.GetStockRemaining() != stockBefore-2 {
		t.Errorf("stock = %d, want %d", s.GetStockRemaining(), stockBefore-2)
	}
	if s.GetPlayer(0).GetCardsSize() != before0+1 {
		t.Error("winner should draw one card")
	}
	if s.GetPhase() != domain.SchnapsenPhasePlay {
		t.Error("phase should return to Play")
	}
	if s.GetCurrentPlayerIdx() != 0 {
		t.Error("lead player should be on turn")
	}
}

func TestSchnapsen_NextTrick_GuardWrongPhase(t *testing.T) {
	s := newTestSchnapsen()
	s.SetPhase(domain.SchnapsenPhasePlay)
	stock := s.GetStockRemaining()
	s.NextTrick()
	if s.GetStockRemaining() != stock {
		t.Error("NextTrick in wrong phase should be a no-op")
	}
}

func TestSchnapsen_NextTrick_EndsWhenHandsEmpty(t *testing.T) {
	s := newTestSchnapsen()
	drainStock(s)
	schnSetHand(s.GetPlayer(0)) // empty
	schnSetHand(s.GetPlayer(1)) // empty
	s.SetPhase(domain.SchnapsenPhaseTrickEnd)
	s.SetLeadPlayerIdx(1)
	s.SetPlayerPoints(0, 40)
	s.SetPlayerPoints(1, 30)
	s.NextTrick()
	if !s.GetGameEndFlag() {
		t.Error("game should end when all hands empty")
	}
	// Neither reached 66 -> last trick winner (lead idx 1) wins.
	if s.GetWinnerIdx() != 1 {
		t.Errorf("winner = %d, want 1 (last trick fallback)", s.GetWinnerIdx())
	}
}

func TestSchnapsen_DetermineWinner(t *testing.T) {
	cases := []struct {
		p0, p1, last, want int
	}{
		{70, 30, 1, 0},
		{30, 70, 0, 1},
		{40, 30, 1, 1},
		{40, 30, 0, 0},
	}
	for _, c := range cases {
		if got := domain.SchnapsenDetermineWinner(c.p0, c.p1, c.last); got != c.want {
			t.Errorf("DetermineWinner(%d,%d,%d) = %d, want %d", c.p0, c.p1, c.last, got, c.want)
		}
	}
}

func TestSchnapsen_CpuPlay(t *testing.T) {
	s := newTestSchnapsen()
	s.SetPhase(domain.SchnapsenPhasePlay)
	s.SetCurrentPlayerIdx(1)
	s.SetCurrentTrick(nil)
	before := s.GetPlayer(1).GetCardsSize()
	s.CpuPlay()
	if s.GetPlayer(1).GetCardsSize() != before-1 {
		t.Error("CPU should play one card")
	}
	// Human turn: CpuPlay is a no-op.
	s.SetCurrentPlayerIdx(0)
	s.SetCurrentTrick(nil)
	hb := s.GetPlayer(0).GetCardsSize()
	s.CpuPlay()
	if s.GetPlayer(0).GetCardsSize() != hb {
		t.Error("CpuPlay should be a no-op on human turn")
	}
}

func TestSchnapsen_CpuPlay_DeclaresMarriage(t *testing.T) {
	s := newTestSchnapsen()
	s.SetTrumpSuit(domain.CardDesignSpade)
	s.SetPhase(domain.SchnapsenPhasePlay)
	s.SetCurrentPlayerIdx(1)
	s.SetLeadPlayerIdx(1)
	s.SetCurrentTrick(nil)
	schnSetHand(s.GetPlayer(1),
		schnCard(domain.CardDesignSpade, 12), // Q trump
		schnCard(domain.CardDesignSpade, 13), // K trump
	)
	s.CpuPlay()
	if s.GetPlayerPoints(1) != 40 {
		t.Errorf("CPU royal marriage points = %d, want 40", s.GetPlayerPoints(1))
	}
}

func TestSchnapsen_GetHint(t *testing.T) {
	s := newTestSchnapsen()
	s.SetPhase(domain.SchnapsenPhasePlay)
	s.SetCurrentPlayerIdx(0)
	s.SetCurrentTrick(nil)
	if h := s.GetHint(); h == nil || h.CardIndex == nil {
		t.Error("expected a play hint on human lead")
	}
	// Marriage hint
	s.SetTrumpSuit(domain.CardDesignSpade)
	s.SetLeadPlayerIdx(0)
	schnSetHand(s.GetPlayer(0), schnCard(domain.CardDesignClover, 12), schnCard(domain.CardDesignClover, 13))
	h := s.GetHint()
	if h == nil || !h.IsMarriage {
		t.Error("expected a marriage hint")
	}
	// Not human turn -> nil
	s.SetCurrentPlayerIdx(1)
	if s.GetHint() != nil {
		t.Error("hint should be nil when not human turn")
	}
}

func TestSchnapsen_JSONRoundTrip(t *testing.T) {
	s := newTestSchnapsen()
	s.SetPlayerPoints(0, 12)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	var s2 domain.Schnapsen
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if s2.GetPlayerPoints(0) != 12 {
		t.Errorf("round-trip points = %d, want 12", s2.GetPlayerPoints(0))
	}
	if s2.GetTrumpSuit() != s.GetTrumpSuit() {
		t.Error("trump suit not preserved")
	}
}

func TestSchnapsen_UnmarshalValidation(t *testing.T) {
	bad := []string{
		`{"ps":[]}`,                             // wrong player count
		`{"ps":[null,{}]}`,                      // nil player pointer
		`{"ps":[{},{}],"ct":[{},{},{}]}`,        // trick too long
		`{"ps":[{},{}],"ct":[{"pi":0}]}`,        // nil card in trick
		`{"ps":[{},{}],"ct":[{"pi":5,"c":{}}]}`, // out-of-range trick player index
		`{"ps":[{},{}],"pp":[1,2,3]}`,           // wrong points length
		`{"ps":[{},{}],"al":[null]}`,            // nil action-log entry
		`{"ps":[{},{}],"ci":5}`,                 // out-of-range current player index
		`{"ps":[{},{}],"li":-1}`,                // out-of-range lead player index
		`{"ps":[{},{}],"di":9}`,                 // out-of-range dealer index
		`{"ps":[{},{}],"ph":9}`,                 // invalid phase
	}
	for _, b := range bad {
		var s domain.Schnapsen
		if err := json.Unmarshal([]byte(b), &s); err == nil {
			t.Errorf("expected error unmarshalling %s", b)
		}
	}
	// Valid minimal with nil slices filled in.
	var s domain.Schnapsen
	if err := json.Unmarshal([]byte(`{"ps":[{},{}]}`), &s); err != nil {
		t.Fatalf("valid minimal unmarshal err: %v", err)
	}
	if s.GetStockRemaining() < 0 {
		t.Error("stock should be initialised")
	}
}

func TestSchnapsenConfig_Validate(t *testing.T) {
	if err := domain.DefaultSchnapsenConfig().Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
	bad := domain.SchnapsenConfig{CpuDifficulty: domain.SchnapsenCpuDifficulty(9)}
	if err := bad.Validate(); err == nil {
		t.Error("out-of-range difficulty should be invalid")
	}
}

func TestSchnapsen_FullGameToCompletion(t *testing.T) {
	// Play full games with random shuffles; ensure a winner is always produced
	// and points never exceed sane bounds.
	for iter := 0; iter < 200; iter++ {
		s := domain.NewDefaultSchnapsen()
		s.Reset()
		guard := 0
		for !s.GetGameEndFlag() && guard < 1000 {
			guard++
			switch s.GetPhase() {
			case domain.SchnapsenPhasePlay:
				if s.IsHumanTurn() {
					idxs := s.GetValidPlayIndices(0)
					if len(idxs) == 0 {
						t.Fatal("no valid plays for human")
					}
					if err := s.PlayerPlay(idxs[0]); err != nil {
						t.Fatalf("human play err: %v", err)
					}
				} else {
					s.CpuPlay()
				}
			case domain.SchnapsenPhaseTrickEnd:
				s.ResolveTrick()
				if !s.GetGameEndFlag() {
					s.NextTrick()
				}
			}
		}
		if !s.GetGameEndFlag() {
			t.Fatal("game did not finish within guard limit")
		}
		if w := s.GetWinnerIdx(); w != 0 && w != 1 {
			t.Fatalf("invalid winner %d", w)
		}
	}
}
