//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func indianRummyCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func indianRummyJoker(value int) *domain.Card {
	return domain.NewCard(domain.CardDesignJoker, value, false)
}

func newTestIndianRummy(n int) *domain.IndianRummy {
	players := make([]*domain.IndianRummyPlayer, n)
	players[0] = domain.NewIndianRummyPlayer(true)
	for i := 1; i < n; i++ {
		players[i] = domain.NewIndianRummyPlayer(false)
	}
	cfg := domain.DefaultIndianRummyConfig()
	cfg.PlayerCount = n
	return domain.NewIndianRummy(domain.NewTrumpCardsWithDecks(2, 4), players, cfg)
}

func setIndianRummyHand(p *domain.IndianRummyPlayer, cards []*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// validIndianRummyHand は 2 つ以上のシーケンス（うち 1 つ以上ピュア）を含む有効宣言 13 枚。
func validIndianRummyHand() []*domain.Card {
	return []*domain.Card{
		// pure run ♠3-4-5
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignSpade, 4),
		indianRummyCard(domain.CardDesignSpade, 5),
		// pure run ♥6-7-8
		indianRummyCard(domain.CardDesignHeart, 6),
		indianRummyCard(domain.CardDesignHeart, 7),
		indianRummyCard(domain.CardDesignHeart, 8),
		// set of 9s
		indianRummyCard(domain.CardDesignClover, 9),
		indianRummyCard(domain.CardDesignDiamond, 9),
		indianRummyCard(domain.CardDesignSpade, 9),
		// run ♦10-11-12-13
		indianRummyCard(domain.CardDesignDiamond, 10),
		indianRummyCard(domain.CardDesignDiamond, 11),
		indianRummyCard(domain.CardDesignDiamond, 12),
		indianRummyCard(domain.CardDesignDiamond, 13),
	}
}

func TestNewIndianRummy(t *testing.T) {
	g := newTestIndianRummy(4)
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetDeclarerIdx())
}

func TestIndianRummy_Reset(t *testing.T) {
	g := newTestIndianRummy(4)
	g.Reset()

	assert.Equal(t, domain.IndianRummyPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 4, g.GetPlayerCnt())
	// 各プレイヤーに 13 枚配られている。
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.IndianRummyHandSize, g.GetPlayer(i).GetCardsSize())
	}
	assert.NotNil(t, g.GetWildJoker())
	assert.True(t, g.GetWildRank() >= 0 && g.GetWildRank() <= domain.CardValueMax)
	assert.NotNil(t, g.GetDiscardTop())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // 席 0 のディーラーの左隣
}

func TestIndianRummy_DefaultConstructor(t *testing.T) {
	g := domain.NewDefaultIndianRummy()
	g.Reset()
	assert.Equal(t, domain.IndianRummyDefaultPlayerCount, g.GetPlayerCnt())
}

func TestIndianRummyConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultIndianRummyConfig().Validate())

	bad := domain.DefaultIndianRummyConfig()
	bad.PlayerCount = 1
	assert.Error(t, bad.Validate())

	bad2 := domain.DefaultIndianRummyConfig()
	bad2.PlayerCount = 5
	assert.Error(t, bad2.Validate())

	bad3 := domain.DefaultIndianRummyConfig()
	bad3.CpuDifficulty = 9
	assert.Error(t, bad3.Validate())

	bad4 := domain.DefaultIndianRummyConfig()
	bad4.TargetRounds = 0
	assert.Error(t, bad4.Validate())
}

func TestIndianRummy_CardPoints(t *testing.T) {
	assert.Equal(t, 10, domain.IndianRummyCardPoints(indianRummyCard(domain.CardDesignSpade, 1), 0))  // Ace
	assert.Equal(t, 10, domain.IndianRummyCardPoints(indianRummyCard(domain.CardDesignSpade, 13), 0)) // King
	assert.Equal(t, 10, domain.IndianRummyCardPoints(indianRummyCard(domain.CardDesignSpade, 10), 0))
	assert.Equal(t, 5, domain.IndianRummyCardPoints(indianRummyCard(domain.CardDesignSpade, 5), 0))
	assert.Equal(t, 0, domain.IndianRummyCardPoints(indianRummyJoker(1), 0))                        // printed joker
	assert.Equal(t, 0, domain.IndianRummyCardPoints(indianRummyCard(domain.CardDesignSpade, 7), 7)) // wild rank
}

func TestIndianRummy_HasPureSequence(t *testing.T) {
	pure := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignSpade, 4),
		indianRummyCard(domain.CardDesignSpade, 5),
	}
	assert.True(t, domain.IndianRummyHasPureSequence(pure, 0))

	noPure := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignHeart, 7),
		indianRummyCard(domain.CardDesignDiamond, 11),
	}
	assert.False(t, domain.IndianRummyHasPureSequence(noPure, 0))

	// wild-rank card cannot make a pure sequence.
	withWild := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignSpade, 4),
		indianRummyCard(domain.CardDesignSpade, 5), // 5 is wild
	}
	assert.False(t, domain.IndianRummyHasPureSequence(withWild, 5))
}

func TestIndianRummy_ValidateDeclaration_Valid(t *testing.T) {
	assert.True(t, domain.IndianRummyValidateDeclaration(validIndianRummyHand(), 0))
}

func TestIndianRummy_ValidateDeclaration_WrongLength(t *testing.T) {
	short := validIndianRummyHand()[:12]
	assert.False(t, domain.IndianRummyValidateDeclaration(short, 0))
}

func TestIndianRummy_ValidateDeclaration_OnlyOneSequence(t *testing.T) {
	// 1 run + 3 sets: only a single sequence → invalid (needs 2 sequences).
	hand := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignSpade, 4),
		indianRummyCard(domain.CardDesignSpade, 5),
		indianRummyCard(domain.CardDesignSpade, 7),
		indianRummyCard(domain.CardDesignHeart, 7),
		indianRummyCard(domain.CardDesignDiamond, 7),
		indianRummyCard(domain.CardDesignSpade, 9),
		indianRummyCard(domain.CardDesignHeart, 9),
		indianRummyCard(domain.CardDesignDiamond, 9),
		indianRummyCard(domain.CardDesignSpade, 11),
		indianRummyCard(domain.CardDesignHeart, 11),
		indianRummyCard(domain.CardDesignDiamond, 11),
		indianRummyCard(domain.CardDesignClover, 11),
	}
	assert.False(t, domain.IndianRummyValidateDeclaration(hand, 0))
}

func TestIndianRummy_ValidateDeclaration_NoPureSequence(t *testing.T) {
	// 2 impure sequences (each uses a joker), no pure → invalid.
	hand := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignSpade, 4),
		indianRummyJoker(1), // fills ♠5
		indianRummyCard(domain.CardDesignHeart, 7),
		indianRummyCard(domain.CardDesignHeart, 8),
		indianRummyJoker(2), // fills ♥9
		indianRummyCard(domain.CardDesignSpade, 10),
		indianRummyCard(domain.CardDesignHeart, 10),
		indianRummyCard(domain.CardDesignDiamond, 10),
		indianRummyCard(domain.CardDesignSpade, 13),
		indianRummyCard(domain.CardDesignHeart, 13),
		indianRummyCard(domain.CardDesignDiamond, 13),
		indianRummyCard(domain.CardDesignClover, 13),
	}
	assert.False(t, domain.IndianRummyValidateDeclaration(hand, 0))
}

func TestIndianRummy_ValidateDeclaration_ImpurePlusPure(t *testing.T) {
	// 1 pure run + 1 impure run + set + set → valid.
	hand := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignSpade, 4),
		indianRummyCard(domain.CardDesignSpade, 5), // pure run
		indianRummyCard(domain.CardDesignHeart, 7),
		indianRummyCard(domain.CardDesignHeart, 8),
		indianRummyJoker(1), // impure run fills ♥9
		indianRummyCard(domain.CardDesignSpade, 10),
		indianRummyCard(domain.CardDesignHeart, 10),
		indianRummyCard(domain.CardDesignDiamond, 10),
		indianRummyCard(domain.CardDesignSpade, 13),
		indianRummyCard(domain.CardDesignHeart, 13),
		indianRummyCard(domain.CardDesignDiamond, 13),
		indianRummyCard(domain.CardDesignClover, 13),
	}
	assert.True(t, domain.IndianRummyValidateDeclaration(hand, 0))
}

func TestIndianRummy_DeadwoodScore(t *testing.T) {
	// No pure sequence → full cap 80.
	noPure := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignHeart, 7),
		indianRummyCard(domain.CardDesignDiamond, 11),
	}
	assert.Equal(t, domain.IndianRummyDeadwoodCap, domain.IndianRummyDeadwoodScore(noPure, 0))

	// Pure run + small deadwood.
	smallDW := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignSpade, 4),
		indianRummyCard(domain.CardDesignSpade, 5),
		indianRummyCard(domain.CardDesignHeart, 2),
	}
	assert.Equal(t, 2, domain.IndianRummyDeadwoodScore(smallDW, 0))

	// Pure run + 100 points of un-meldable high cards → capped at 80.
	capped := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 1),
		indianRummyCard(domain.CardDesignSpade, 2),
		indianRummyCard(domain.CardDesignSpade, 3), // pure run
		indianRummyCard(domain.CardDesignSpade, 10),
		indianRummyCard(domain.CardDesignSpade, 12),
		indianRummyCard(domain.CardDesignHeart, 11),
		indianRummyCard(domain.CardDesignHeart, 13),
		indianRummyCard(domain.CardDesignClover, 10),
		indianRummyCard(domain.CardDesignClover, 12),
		indianRummyCard(domain.CardDesignDiamond, 11),
		indianRummyCard(domain.CardDesignDiamond, 13),
		indianRummyCard(domain.CardDesignHeart, 1),
		indianRummyCard(domain.CardDesignDiamond, 1),
	}
	assert.Equal(t, domain.IndianRummyDeadwoodCap, domain.IndianRummyDeadwoodScore(capped, 0))
}

func TestIndianRummy_DrawAndDiscard(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDraw)

	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.IndianRummyPhaseDiscard, g.GetPhase())

	require.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.IndianRummyPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestIndianRummy_DrawFromDiscard(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDraw)
	g.SetDiscardPile([]*domain.Card{indianRummyCard(domain.CardDesignSpade, 7)})

	require.NoError(t, g.PlayerDrawFromDiscard())
	assert.Equal(t, domain.IndianRummyPhaseDiscard, g.GetPhase())
}

func TestIndianRummy_DrawGuards(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)

	g.SetPhase(domain.IndianRummyPhaseDiscard)
	assert.Error(t, g.PlayerDrawFromStock()) // wrong phase

	g.SetPhase(domain.IndianRummyPhaseDraw)
	g.SetCurrentPlayerIdx(1) // CPU turn
	assert.Error(t, g.PlayerDrawFromStock())

	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDraw)
	g.SetDiscardPile(nil)
	assert.Error(t, g.PlayerDrawFromDiscard()) // empty discard
}

func TestIndianRummy_DiscardOutOfRange(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDiscard)
	assert.Error(t, g.PlayerDiscard(-1))
	assert.Error(t, g.PlayerDiscard(999))
}

func TestIndianRummy_Declare_Valid(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDiscard)
	g.SetWildRank(0)

	hand := append(validIndianRummyHand(), indianRummyCard(domain.CardDesignClover, 2)) // 14th = finish
	setIndianRummyHand(g.GetPlayer(0), hand)
	// give opponent a no-pure hand.
	setIndianRummyHand(g.GetPlayer(1), []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 13),
		indianRummyCard(domain.CardDesignHeart, 11),
	})

	require.NoError(t, g.PlayerDeclare(13))
	assert.True(t, g.GetDeclarationValid())
	assert.Equal(t, 0, g.GetDeclarerIdx())
	assert.Equal(t, domain.IndianRummyPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetPlayer(0).GetRoundScore())                             // winner scores 0
	assert.Equal(t, domain.IndianRummyDeadwoodCap, g.GetPlayer(1).GetRoundScore()) // no pure → 80
}

func TestIndianRummy_Declare_Invalid(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDiscard)
	g.SetWildRank(0)

	// 13 non-melding cards + 1 finish.
	junk := []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 2),
		indianRummyCard(domain.CardDesignHeart, 5),
		indianRummyCard(domain.CardDesignDiamond, 8),
		indianRummyCard(domain.CardDesignClover, 11),
		indianRummyCard(domain.CardDesignSpade, 13),
		indianRummyCard(domain.CardDesignHeart, 3),
		indianRummyCard(domain.CardDesignDiamond, 6),
		indianRummyCard(domain.CardDesignClover, 9),
		indianRummyCard(domain.CardDesignSpade, 12),
		indianRummyCard(domain.CardDesignHeart, 4),
		indianRummyCard(domain.CardDesignDiamond, 7),
		indianRummyCard(domain.CardDesignClover, 10),
		indianRummyCard(domain.CardDesignSpade, 1),
		indianRummyCard(domain.CardDesignHeart, 13), // finish
	}
	setIndianRummyHand(g.GetPlayer(0), junk)

	require.NoError(t, g.PlayerDeclare(13))
	assert.False(t, g.GetDeclarationValid())
	assert.Equal(t, domain.IndianRummyDeadwoodCap, g.GetPlayer(0).GetRoundScore()) // invalid → 80
}

func TestIndianRummy_DeclareGuards(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDraw)
	assert.Error(t, g.PlayerDeclare(0)) // wrong phase

	g.SetPhase(domain.IndianRummyPhaseDiscard)
	assert.Error(t, g.PlayerDeclare(999)) // out of range
}

func TestIndianRummy_Recycle(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDraw)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 2),
		indianRummyCard(domain.CardDesignHeart, 3),
		indianRummyCard(domain.CardDesignDiamond, 4),
	})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, domain.IndianRummyPhaseDiscard, g.GetPhase())
}

func TestIndianRummy_StockOut(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDraw)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*domain.Card{indianRummyCard(domain.CardDesignSpade, 2)})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, domain.IndianRummyPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, -1, g.GetDeclarerIdx())
}

func TestIndianRummy_NextRound(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetRoundNumber(1)
	g.SetPhase(domain.IndianRummyPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.IndianRummyPhaseDraw, g.GetPhase())

	// wrong phase → no-op
	g.SetPhase(domain.IndianRummyPhaseDraw)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
}

func TestIndianRummy_NextRound_GameEnd(t *testing.T) {
	g := newTestIndianRummy(2)
	cfg := domain.DefaultIndianRummyConfig()
	cfg.PlayerCount = 2
	cfg.TargetRounds = 2
	g.SetConfig(cfg)
	g.Reset()
	g.SetRoundNumber(2)
	g.SetPhase(domain.IndianRummyPhaseRoundEnd)
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.IndianRummyPhaseGameEnd, g.GetPhase())
	assert.True(t, g.GetWinnerIdx() >= 0)
}

func TestIndianRummy_CpuBoundedLoop(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	for iter := 0; iter < 400 && !g.GetGameEndFlag(); iter++ {
		phase := g.GetPhase()
		if phase == domain.IndianRummyPhaseRoundEnd {
			g.NextRound()
			continue
		}
		if g.IsHumanTurn() {
			if g.GetPhase() == domain.IndianRummyPhaseDraw {
				_ = g.PlayerDrawFromStock()
			} else {
				_ = g.PlayerDiscard(0)
			}
			continue
		}
		g.CpuPlay()
	}
	// 一貫した状態であることのみ確認（ゲーム終了は保証しない）。
	p := g.GetPhase()
	assert.True(t, p >= domain.IndianRummyPhaseDraw && p <= domain.IndianRummyPhaseGameEnd)
}

func TestIndianRummy_CpuDifficulties(t *testing.T) {
	for _, d := range []domain.IndianRummyCpuDifficulty{
		domain.IndianRummyCpuDifficultyEasy,
		domain.IndianRummyCpuDifficultyNormal,
		domain.IndianRummyCpuDifficultyHard,
	} {
		g := newTestIndianRummy(2)
		cfg := domain.DefaultIndianRummyConfig()
		cfg.PlayerCount = 2
		cfg.CpuDifficulty = d
		g.SetConfig(cfg)
		g.Reset()
		g.SetCurrentPlayerIdx(1) // CPU
		g.SetPhase(domain.IndianRummyPhaseDraw)
		g.CpuPlay()
		g.CpuPlay()
		assert.NotNil(t, g.GetPlayer(1))
	}
}

func TestIndianRummy_PlayerHelpers(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetWildRank(0)
	setIndianRummyHand(g.GetPlayer(0), []*domain.Card{
		indianRummyCard(domain.CardDesignSpade, 3),
		indianRummyCard(domain.CardDesignSpade, 4),
		indianRummyCard(domain.CardDesignSpade, 5),
		indianRummyCard(domain.CardDesignHeart, 9),
	})
	assert.True(t, g.PlayerHasPureSequence(0))
	assert.Equal(t, 9, g.PlayerDeadwoodValue(0))
	assert.Nil(t, g.GetPlayer(99))
	assert.False(t, g.PlayerHasPureSequence(99))
	assert.Equal(t, 0, g.PlayerDeadwoodValue(99))
}

func TestIndianRummy_JSONRoundTrip(t *testing.T) {
	g := newTestIndianRummy(3)
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	var restored domain.IndianRummy
	require.NoError(t, restored.UnmarshalJSON(data))
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetWildRank(), restored.GetWildRank())
}

func TestIndianRummy_UnmarshalJSON_BadJSON(t *testing.T) {
	var g domain.IndianRummy
	assert.Error(t, g.UnmarshalJSON([]byte("not json")))
}

func TestIndianRummy_UnmarshalJSON_Validation(t *testing.T) {
	base := newTestIndianRummy(2)
	base.Reset()
	data, err := base.MarshalJSON()
	require.NoError(t, err)

	tamper := func(mut func(m map[string]json.RawMessage)) []byte {
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &m))
		mut(m)
		out, err := json.Marshal(m)
		require.NoError(t, err)
		return out
	}

	cases := map[string]func(m map[string]json.RawMessage){
		"currentPlayerIdx": func(m map[string]json.RawMessage) { m["ci"] = json.RawMessage("99") },
		"dealerIdx":        func(m map[string]json.RawMessage) { m["di"] = json.RawMessage("99") },
		"phase":            func(m map[string]json.RawMessage) { m["ps"] = json.RawMessage("99") },
		"roundNumber":      func(m map[string]json.RawMessage) { m["rn"] = json.RawMessage("-1") },
		"winnerIdx":        func(m map[string]json.RawMessage) { m["wi"] = json.RawMessage("99") },
		"declarerIdx":      func(m map[string]json.RawMessage) { m["de"] = json.RawMessage("99") },
		"wildRank":         func(m map[string]json.RawMessage) { m["wr"] = json.RawMessage("99") },
		"playerCount":      func(m map[string]json.RawMessage) { m["pl"] = json.RawMessage("[]") },
		"nilPlayer":        func(m map[string]json.RawMessage) { m["pl"] = json.RawMessage("[null,null]") },
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.IndianRummy
			assert.Error(t, g.UnmarshalJSON(tamper(mut)))
		})
	}
}

func TestIndianRummy_UnmarshalJSON_FiltersNilCards(t *testing.T) {
	base := newTestIndianRummy(2)
	base.Reset()
	data, err := base.MarshalJSON()
	require.NoError(t, err)

	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &m))
	m["dp"] = json.RawMessage(`[null,{"d":1,"v":5,"w":false}]`)
	out, err := json.Marshal(m)
	require.NoError(t, err)

	var g domain.IndianRummy
	require.NoError(t, g.UnmarshalJSON(out))
	// nil は除去され、有効カード 1 枚だけがトップに残る。
	assert.NotNil(t, g.GetDiscardTop())
}

func TestIndianRummy_ActionLogAccumulates(t *testing.T) {
	g := newTestIndianRummy(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDraw)
	require.NoError(t, g.PlayerDrawFromStock())
	assert.NotEmpty(t, g.GetActionLog())
}
