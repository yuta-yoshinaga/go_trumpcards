package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultScopa(t *testing.T) {
	s := NewDefaultScopa()
	if s.GetPlayerCnt() != ScopaPlayerCnt {
		t.Fatalf("expected %d players, got %d", ScopaPlayerCnt, s.GetPlayerCnt())
	}
	if !s.GetPlayer(0).GetIsHuman() {
		t.Error("player 0 should be human")
	}
	if s.GetPlayer(1).GetIsHuman() {
		t.Error("player 1 should be CPU")
	}
	if s.GetConfig().TargetScore != ScopaDefaultTargetScore {
		t.Errorf("expected target %d", ScopaDefaultTargetScore)
	}
	if s.GetPlayer(99) != nil {
		t.Error("out-of-range GetPlayer should be nil")
	}
}

func TestScopaReset_DealsCorrectly(t *testing.T) {
	s := NewDefaultScopa()
	s.Reset()
	if s.GetPhase() != ScopaPhasePlayerTurn {
		t.Errorf("expected playerTurn phase, got %s", s.GetPhase())
	}
	for i := 0; i < ScopaPlayerCnt; i++ {
		if got := s.GetPlayer(i).GetCardsSize(); got != ScopaHandSize {
			t.Errorf("player %d hand = %d, want %d", i, got, ScopaHandSize)
		}
	}
	if got := len(s.GetTableCards()); got != ScopaInitialTableSize {
		t.Errorf("table = %d, want %d", got, ScopaInitialTableSize)
	}
	// 40 - (3*2) - 4 = 30
	if got := s.GetRemainingDeck(); got != 30 {
		t.Errorf("remaining deck = %d, want 30", got)
	}
	if !s.IsHumanTurn() {
		t.Error("expected human turn after reset")
	}
}

// scopaPlayReady returns a drained-deck game in playerTurn phase for
// deterministic single-play tests (no auto re-deal during the play).
func scopaPlayReady(t *testing.T) *Scopa {
	t.Helper()
	s := ScopaTestNew(DefaultScopaConfig())
	s.ScopaTestDrainDeck()
	s.ScopaTestSetPhase(ScopaPhasePlayerTurn)
	s.ScopaTestSetCurrentTurn(0)
	return s
}

func TestScopaPlayerPlay_Capture(t *testing.T) {
	s := scopaPlayReady(t)
	s.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	s.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 2, false)) // keep hands non-empty
	s.ScopaTestSetTable([]*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 3, false),
	})
	if err := s.PlayerPlay(0, []int{0}); err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if got := s.GetPlayer(0).CapturedCount(); got != 2 {
		t.Errorf("captured = %d, want 2 (played + table)", got)
	}
	if got := len(s.GetTableCards()); got != 1 {
		t.Errorf("table remaining = %d, want 1", got)
	}
	if s.GetLastCaptureIdx() != 0 {
		t.Errorf("lastCapture = %d, want 0", s.GetLastCaptureIdx())
	}
	if s.GetPlayer(0).GetScopaCount() != 0 {
		t.Error("table not cleared, scopa must not trigger")
	}
}

func TestScopaPlayerPlay_ForcedCaptureRejected(t *testing.T) {
	s := scopaPlayReady(t)
	s.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	s.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 2, false))
	s.ScopaTestSetTable([]*Card{NewCard(CardDesignSpade, 5, false)})
	err := s.PlayerPlay(0, nil)
	if err == nil || !errors.Is(err, ErrInvalidPlay) {
		t.Fatalf("expected ErrInvalidPlay when a capture is available, got %v", err)
	}
}

func TestScopaPlayerPlay_Lay(t *testing.T) {
	s := scopaPlayReady(t)
	s.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	s.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 2, false))
	s.ScopaTestSetTable([]*Card{NewCard(CardDesignSpade, 2, false)}) // no capture for a 5
	if err := s.PlayerPlay(0, nil); err != nil {
		t.Fatalf("lay failed: %v", err)
	}
	if got := len(s.GetTableCards()); got != 2 {
		t.Errorf("table = %d, want 2 after lay", got)
	}
	if s.GetPlayer(0).CapturedCount() != 0 {
		t.Error("laying must not capture")
	}
}

func TestScopaPlayerPlay_InvalidSelection(t *testing.T) {
	s := scopaPlayReady(t)
	s.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	s.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 2, false))
	s.ScopaTestSetTable([]*Card{NewCard(CardDesignSpade, 2, false)})
	if err := s.PlayerPlay(0, []int{0}); !errors.Is(err, ErrInvalidPlay) {
		t.Fatalf("expected ErrInvalidPlay for impossible capture, got %v", err)
	}
}

func TestScopaScopaDetection(t *testing.T) {
	s := scopaPlayReady(t)
	s.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	s.GetPlayer(1).AddCard(NewCard(CardDesignClover, 9, false)) // value 9, keeps hands non-empty
	s.ScopaTestSetTable([]*Card{NewCard(CardDesignSpade, 5, false)})
	if err := s.PlayerPlay(0, []int{0}); err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if s.GetPlayer(0).GetScopaCount() != 1 {
		t.Errorf("scopa count = %d, want 1 (table cleared, not last play)", s.GetPlayer(0).GetScopaCount())
	}
	act := s.GetHumanAction()
	if act == nil || !act.IsScopa {
		t.Error("human action should record a scopa")
	}
}

func TestScopaLastPlayNoScopa_TriggersFinish(t *testing.T) {
	s := scopaPlayReady(t)
	s.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	// player 1 has no cards → after this play, round is over (last play).
	s.ScopaTestSetTable([]*Card{NewCard(CardDesignSpade, 5, false)})
	if err := s.PlayerPlay(0, []int{0}); err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if s.GetPlayer(0).GetScopaCount() != 0 {
		t.Error("clearing the table on the final play must not score a scopa")
	}
	if s.GetPhase() != ScopaPhaseRoundEnd && s.GetPhase() != ScopaPhaseGameEnd {
		t.Errorf("expected round to finish, phase = %s", s.GetPhase())
	}
}

func TestScopaGuards(t *testing.T) {
	// game ended
	s := scopaPlayReady(t)
	s.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	s.ScopaTestSetGameEnd(true)
	if err := s.PlayerPlay(0, nil); !errors.Is(err, ErrGameEnded) {
		t.Errorf("expected ErrGameEnded, got %v", err)
	}

	// wrong phase
	s2 := scopaPlayReady(t)
	s2.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	s2.ScopaTestSetPhase(ScopaPhaseDealing)
	if err := s2.PlayerPlay(0, nil); !errors.Is(err, ErrWrongPhase) {
		t.Errorf("expected ErrWrongPhase, got %v", err)
	}

	// not human turn
	s3 := scopaPlayReady(t)
	s3.ScopaTestSetCurrentTurn(1)
	if err := s3.PlayerPlay(0, nil); !errors.Is(err, ErrNotHumanTurn) {
		t.Errorf("expected ErrNotHumanTurn, got %v", err)
	}

	// invalid card index
	s4 := scopaPlayReady(t)
	if err := s4.PlayerPlay(5, nil); !errors.Is(err, ErrInvalidCard) {
		t.Errorf("expected ErrInvalidCard, got %v", err)
	}
}

func TestScopaScoreRound(t *testing.T) {
	s := ScopaTestNew(DefaultScopaConfig())
	p0 := s.GetPlayer(0)
	p1 := s.GetPlayer(1)
	p0.AddCaptured([]*Card{
		NewCard(CardDesignDiamond, 7, false), // settebello + seven + diamond
		NewCard(CardDesignSpade, 7, false),   // seven
		NewCard(CardDesignDiamond, 2, false), // diamond
		NewCard(CardDesignDiamond, 3, false), // diamond
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignClover, 5, false),
	})
	p0.IncrementScopa()
	p1.AddCaptured([]*Card{
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignClover, 11, false),
	})
	det := s.ScopaTestScoreRound()
	if det.HasSetteBello != 0 {
		t.Errorf("settebello holder = %d, want 0", det.HasSetteBello)
	}
	// p0: carte + denari + sevens + settebello + 1 scopa = 5
	if det.Gained[0] != 5 {
		t.Errorf("p0 gained = %d, want 5", det.Gained[0])
	}
	if det.Gained[1] != 0 {
		t.Errorf("p1 gained = %d, want 0", det.Gained[1])
	}
}

func TestScopaScoreRound_TieAwardsNobody(t *testing.T) {
	s := ScopaTestNew(DefaultScopaConfig())
	// Equal card counts and no diamonds/sevens/settebello → carte tie, nobody scores.
	s.GetPlayer(0).AddCaptured([]*Card{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignSpade, 3, false)})
	s.GetPlayer(1).AddCaptured([]*Card{NewCard(CardDesignClover, 2, false), NewCard(CardDesignClover, 3, false)})
	det := s.ScopaTestScoreRound()
	if det.Gained[0] != 0 || det.Gained[1] != 0 {
		t.Errorf("tie should award nobody, got %v", det.Gained)
	}
	if det.HasSetteBello != -1 {
		t.Errorf("no settebello expected, got %d", det.HasSetteBello)
	}
}

func TestScopaFinishRound_LastTakeAndGameEnd(t *testing.T) {
	cfg := DefaultScopaConfig()
	cfg.TargetScore = 3
	s := ScopaTestNew(cfg)
	s.ScopaTestDrainDeck()
	// Leftover table cards go to the last capturer (player 0).
	s.ScopaTestSetTable([]*Card{
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignDiamond, 2, false),
		NewCard(CardDesignSpade, 7, false),
	})
	s.ScopaTestSetLastCapture(0)
	s.ScopaTestFinishRound()
	if s.GetPlayer(0).CapturedCount() != 3 {
		t.Errorf("leftover not awarded to last capturer: %d", s.GetPlayer(0).CapturedCount())
	}
	if !s.GetGameEndFlag() {
		t.Error("expected game to end at low target score")
	}
	if s.GetPhase() != ScopaPhaseGameEnd {
		t.Errorf("expected gameEnd phase, got %s", s.GetPhase())
	}
	winners := s.GetRoundWinners()
	if len(winners) != 1 || winners[0] != 0 {
		t.Errorf("winners = %v, want [0]", winners)
	}
	if d := s.GetLastRoundDetail(); d == nil {
		t.Error("last round detail should be set")
	}
}

func TestScopaNextRound(t *testing.T) {
	cfg := DefaultScopaConfig()
	cfg.TargetScore = 999 // prevent game end
	s := ScopaTestNew(cfg)
	s.ScopaTestDrainDeck()
	s.ScopaTestSetTable([]*Card{NewCard(CardDesignSpade, 4, false)})
	s.ScopaTestSetLastCapture(0)
	s.ScopaTestFinishRound()
	if s.GetGameEndFlag() {
		t.Fatal("game should not end with high target")
	}
	s.NextRound()
	if s.GetPhase() != ScopaPhasePlayerTurn {
		t.Errorf("expected playerTurn after NextRound, got %s", s.GetPhase())
	}
	for i := 0; i < ScopaPlayerCnt; i++ {
		if s.GetPlayer(i).GetCardsSize() != ScopaHandSize {
			t.Errorf("player %d not dealt a fresh hand", i)
		}
		if s.GetPlayer(i).CapturedCount() != 0 {
			t.Errorf("player %d captures not reset", i)
		}
	}
	if len(s.GetTableCards()) != ScopaInitialTableSize {
		t.Errorf("new round should place %d table cards, got %d", ScopaInitialTableSize, len(s.GetTableCards()))
	}
}

func TestScopaNextRound_NoOpWhenEnded(t *testing.T) {
	s := NewDefaultScopa()
	s.Reset()
	s.ScopaTestSetGameEnd(true)
	before := s.GetPhase()
	s.ScopaTestSetGameEnd(true)
	s.NextRound()
	_ = before
	if !s.GetGameEndFlag() {
		t.Error("NextRound must be a no-op once the game has ended")
	}
}

func TestScopaJSONRoundTrip(t *testing.T) {
	s := NewDefaultScopa()
	s.Reset()
	// advance a CPU turn to populate some state
	s.ScopaTestSetCurrentTurn(0)
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Scopa
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetPlayerCnt() != s.GetPlayerCnt() {
		t.Errorf("player count mismatch after round-trip")
	}
	if got.GetRemainingDeck() != s.GetRemainingDeck() {
		t.Errorf("deck mismatch: %d vs %d", got.GetRemainingDeck(), s.GetRemainingDeck())
	}
	if got.GetPhase() != s.GetPhase() {
		t.Errorf("phase mismatch")
	}
}

func TestScopaUnmarshalRejectsMissingDeck(t *testing.T) {
	var s Scopa
	if err := json.Unmarshal([]byte(`{"pl":[]}`), &s); err == nil {
		t.Error("expected error for missing trump cards")
	}
}

// **なぜその点数になったのかを CUI は一切出していなかった (#4756)。**内訳は
// 得点計算そのものと同じ判定 (uniqueMaxIndex) から出す。**別実装にすると、
// 内訳の合計が実際の得点と合わなくなる。**
func TestScopaCategoryWinners(t *testing.T) {
	t.Run("nothing to report without a detail", func(t *testing.T) {
		assert.Nil(t, ScopaCategoryWinners(nil))
	})

	t.Run("names the unique leader of each category", func(t *testing.T) {
		rows := ScopaCategoryWinners(&ScopaScoreDetail{
			Cards:         map[int]int{0: 21, 1: 19},
			Diamonds:      map[int]int{0: 4, 1: 6},
			Sevens:        map[int]int{0: 3, 1: 1},
			HasSetteBello: 1,
		})
		byKey := map[string]ScopaCategoryWinner{}
		for _, r := range rows {
			byKey[r.Key] = r
		}
		assert.Equal(t, 0, byKey["cards"].Winner)
		assert.Equal(t, 1, byKey["denari"].Winner)
		assert.Equal(t, 0, byKey["primiera"].Winner)
		assert.Equal(t, 1, byKey["settebello"].Winner)
	})

	// **同点は誰も取らない。**勝者を1人でっち上げると、内訳の合計が実際の
	// 得点より大きくなる。
	t.Run("awards nothing on a tie", func(t *testing.T) {
		rows := ScopaCategoryWinners(&ScopaScoreDetail{
			Cards:         map[int]int{0: 20, 1: 20},
			Diamonds:      map[int]int{0: 5, 1: 5},
			Sevens:        map[int]int{0: 2, 1: 2},
			HasSetteBello: -1,
		})
		for _, r := range rows {
			assert.Equal(t, -1, r.Winner, "%s に勝者が付いている", r.Key)
			assert.Equal(t, 0, r.Points, "%s に点が付いている", r.Key)
		}
	})

	// **内訳の合計は実際に加算された点数と一致しなければならない。**
	t.Run("the category points add up to what the round awarded", func(t *testing.T) {
		det := &ScopaScoreDetail{
			Cards:         map[int]int{0: 21, 1: 19},
			Diamonds:      map[int]int{0: 6, 1: 4},
			Sevens:        map[int]int{0: 1, 1: 3},
			HasSetteBello: 1,
			Scopas:        map[int]int{0: 2},
		}
		perPlayer := map[int]int{}
		for _, r := range ScopaCategoryWinners(det) {
			if r.Winner >= 0 {
				perPlayer[r.Winner] += r.Points
			}
		}
		for i, n := range det.Scopas {
			perPlayer[i] += n * ScopaScoreScopa
		}
		assert.Equal(t, ScopaScoreMostCards+ScopaScoreMostDiamonds+2*ScopaScoreScopa, perPlayer[0])
		assert.Equal(t, ScopaScoreMostSevens+ScopaScoreSetteBello, perPlayer[1])
	})
}
