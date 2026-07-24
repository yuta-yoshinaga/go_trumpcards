//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSixCardGolf() *SixCardGolf {
	g := NewDefaultSixCardGolf()
	g.SetRand(rand.New(rand.NewSource(42)))
	return g
}

func setupGridAllFaceUp(g *SixCardGolf, idx int) {
	p := g.GetPlayer(idx)
	for i := range p.Grid {
		p.Grid[i].FaceUp = true
	}
}

func makeSlot(value int, suit int, faceUp bool) SixCardGolfSlot {
	return SixCardGolfSlot{Card: NewCard(suit, value, false), FaceUp: faceUp}
}

// --- Config ---

func TestSixCardGolfConfig_Default(t *testing.T) {
	cfg := DefaultSixCardGolfConfig()
	assert.Equal(t, 2, cfg.PlayerCount)
	assert.Equal(t, SixCardGolfCpuNormal, cfg.CpuDifficulty)
	assert.Equal(t, 9, cfg.Rounds)
	assert.NoError(t, cfg.Validate())
}

func TestSixCardGolfConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SixCardGolfConfig
		wantErr bool
	}{
		{"valid", DefaultSixCardGolfConfig(), false},
		{"too few players", SixCardGolfConfig{PlayerCount: 1, CpuDifficulty: SixCardGolfCpuNormal, Rounds: 9}, true},
		{"too many players", SixCardGolfConfig{PlayerCount: 5, CpuDifficulty: SixCardGolfCpuNormal, Rounds: 9}, true},
		{"invalid difficulty", SixCardGolfConfig{PlayerCount: 2, CpuDifficulty: SixCardGolfCpuDifficulty(99), Rounds: 9}, true},
		{"zero rounds", SixCardGolfConfig{PlayerCount: 2, CpuDifficulty: SixCardGolfCpuNormal, Rounds: 0}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- Player helpers ---

func TestSixCardGolfPlayer_AllFaceUp(t *testing.T) {
	p := &SixCardGolfPlayer{}
	for i := range p.Grid {
		p.Grid[i] = makeSlot(i+1, CardDesignSpade, true)
	}
	assert.True(t, p.AllFaceUp())

	p.Grid[3].FaceUp = false
	assert.False(t, p.AllFaceUp())
}

func TestSixCardGolfPlayer_FaceUpCount(t *testing.T) {
	p := &SixCardGolfPlayer{}
	for i := range p.Grid {
		p.Grid[i] = makeSlot(i+1, CardDesignSpade, false)
	}
	assert.Equal(t, 0, p.FaceUpCount())
	p.Grid[0].FaceUp = true
	p.Grid[2].FaceUp = true
	assert.Equal(t, 2, p.FaceUpCount())
}

// --- Reset ---

func TestSixCardGolf_Reset(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()

	assert.Equal(t, SixCardGolfPhaseSetup, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 2, g.GetPlayerCnt())

	totalCards := g.GetDrawPileCount() + len(g.discardPile)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalCards += SixCardGolfGridSize
	}
	assert.Equal(t, 52, totalCards)

	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		assert.Equal(t, 0, p.FaceUpCount())
	}
}

// --- Setup phase ---

func TestSixCardGolf_FlipInitial(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()

	assert.NoError(t, g.FlipInitial(0))
	assert.True(t, g.GetPlayer(0).Grid[0].FaceUp)
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())

	assert.NoError(t, g.FlipInitial(3))
	assert.True(t, g.GetPlayer(0).Grid[3].FaceUp)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())

	assert.NoError(t, g.FlipInitial(1))
	assert.NoError(t, g.FlipInitial(4))
	assert.Equal(t, SixCardGolfPhasePlayerTurn, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
}

func TestSixCardGolf_FlipInitial_Errors(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()

	assert.Error(t, g.FlipInitial(-1))
	assert.Error(t, g.FlipInitial(6))

	assert.NoError(t, g.FlipInitial(0))
	assert.Error(t, g.FlipInitial(0))
}

func TestSixCardGolf_FlipInitial_WrongPhase(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	g.SetPhase(SixCardGolfPhasePlayerTurn)
	assert.ErrorIs(t, g.FlipInitial(0), ErrWrongPhase)
}

// --- Draw ---

func TestSixCardGolf_DrawStock(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	before := g.GetDrawPileCount()
	assert.NoError(t, g.DrawStock())
	assert.Equal(t, SixCardGolfPhaseDrawPending, g.GetPhase())
	assert.NotNil(t, g.GetDrawnCard())
	assert.False(t, g.GetDrawnFromDiscard())
	assert.Equal(t, before-1, g.GetDrawPileCount())
}

func TestSixCardGolf_DrawDiscard(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	assert.True(t, len(g.discardPile) > 0)
	assert.NoError(t, g.DrawDiscard())
	assert.Equal(t, SixCardGolfPhaseDrawPending, g.GetPhase())
	assert.NotNil(t, g.GetDrawnCard())
	assert.True(t, g.GetDrawnFromDiscard())
}

func TestSixCardGolf_DrawStock_WrongPhase(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	assert.ErrorIs(t, g.DrawStock(), ErrWrongPhase)
}

func TestSixCardGolf_DrawDiscard_WrongPhase(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	assert.ErrorIs(t, g.DrawDiscard(), ErrWrongPhase)
}

func TestSixCardGolf_DrawStock_GameEnded(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.DrawStock(), ErrGameEnded)
}

func TestSixCardGolf_DrawDiscard_GameEnded(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.gameEndFlag = true
	assert.ErrorIs(t, g.DrawDiscard(), ErrGameEnded)
}

func TestSixCardGolf_DrawStock_CpuTurn(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.DrawStock(), ErrNotHumanTurn)
}

func TestSixCardGolf_DrawDiscard_EmptyDiscard(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.SetDiscardPile(nil)
	assert.Error(t, g.DrawDiscard())
}

// --- Swap ---

func TestSixCardGolf_SwapCard(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	require.NoError(t, g.DrawStock())
	drawn := g.GetDrawnCard()
	require.NotNil(t, drawn)

	oldCard := g.GetPlayer(0).Grid[2].Card
	require.NoError(t, g.SwapCard(2))

	assert.True(t, g.GetPlayer(0).Grid[2].FaceUp)
	assert.Equal(t, drawn, g.GetPlayer(0).Grid[2].Card)
	assert.Nil(t, g.GetDrawnCard())

	discardTop := g.GetDiscardTop()
	assert.Equal(t, oldCard, discardTop)
}

func TestSixCardGolf_SwapCard_Errors(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	assert.Error(t, g.SwapCard(0))

	require.NoError(t, g.DrawStock())
	assert.Error(t, g.SwapCard(-1))
	assert.Error(t, g.SwapCard(6))
}

func TestSixCardGolf_SwapCard_GameEnded(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	require.NoError(t, g.DrawStock())
	g.gameEndFlag = true
	assert.ErrorIs(t, g.SwapCard(0), ErrGameEnded)
}

func TestSixCardGolf_SwapCard_CpuTurn(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.SetPhase(SixCardGolfPhaseDrawPending)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.SwapCard(0), ErrNotHumanTurn)
}

// --- Discard ---

func TestSixCardGolf_DiscardDrawn_FromStock(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	require.NoError(t, g.DrawStock())
	assert.False(t, g.GetDrawnFromDiscard())

	require.NoError(t, g.DiscardDrawn())
	assert.Nil(t, g.GetDrawnCard())
	assert.True(t, g.GetCanFlip())
	assert.Equal(t, SixCardGolfPhasePlayerTurn, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
}

func TestSixCardGolf_DiscardDrawn_FromDiscard(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	require.NoError(t, g.DrawDiscard())
	assert.True(t, g.GetDrawnFromDiscard())

	require.NoError(t, g.DiscardDrawn())
	assert.False(t, g.GetCanFlip())
}

func TestSixCardGolf_DiscardDrawn_Errors(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	assert.Error(t, g.DiscardDrawn())

	require.NoError(t, g.DrawStock())
	g.gameEndFlag = true
	assert.ErrorIs(t, g.DiscardDrawn(), ErrGameEnded)
}

func TestSixCardGolf_DiscardDrawn_CpuTurn(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.SetPhase(SixCardGolfPhaseDrawPending)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.DiscardDrawn(), ErrNotHumanTurn)
}

// --- FlipCard ---

func TestSixCardGolf_FlipCard(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	require.NoError(t, g.DrawStock())
	require.NoError(t, g.DiscardDrawn())
	assert.True(t, g.GetCanFlip())

	require.NoError(t, g.FlipCard(2))
	assert.True(t, g.GetPlayer(0).Grid[2].FaceUp)
	assert.False(t, g.GetCanFlip())
}

func TestSixCardGolf_FlipCard_Errors(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	assert.Error(t, g.FlipCard(0))

	require.NoError(t, g.DrawStock())
	require.NoError(t, g.DiscardDrawn())
	assert.True(t, g.GetCanFlip())

	assert.Error(t, g.FlipCard(-1))
	assert.Error(t, g.FlipCard(6))

	p := g.GetPlayer(0)
	faceUpIdx := -1
	for i := range p.Grid {
		if p.Grid[i].FaceUp {
			faceUpIdx = i
			break
		}
	}
	if faceUpIdx >= 0 {
		assert.Error(t, g.FlipCard(faceUpIdx))
	}
}

func TestSixCardGolf_FlipCard_GameEnded(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	require.NoError(t, g.DrawStock())
	require.NoError(t, g.DiscardDrawn())
	g.gameEndFlag = true
	assert.ErrorIs(t, g.FlipCard(1), ErrGameEnded)
}

func TestSixCardGolf_FlipCard_CpuTurn(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	require.NoError(t, g.DrawStock())
	require.NoError(t, g.DiscardDrawn())
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.FlipCard(0), ErrNotHumanTurn)
}

// --- SkipFlip ---

func TestSixCardGolf_SkipFlip(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	require.NoError(t, g.DrawStock())
	require.NoError(t, g.DiscardDrawn())
	assert.True(t, g.GetCanFlip())

	require.NoError(t, g.SkipFlip())
	assert.False(t, g.GetCanFlip())
}

func TestSixCardGolf_SkipFlip_Error(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	assert.Error(t, g.SkipFlip())
}

func TestSixCardGolf_SkipFlip_GameEnded(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.canFlip = true
	g.gameEndFlag = true
	assert.ErrorIs(t, g.SkipFlip(), ErrGameEnded)
}

// --- Scoring ---

func TestSixCardGolf_ScorePlayer_Basic(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	p := g.GetPlayer(0)
	p.Grid[0] = makeSlot(1, CardDesignSpade, true)
	p.Grid[1] = makeSlot(5, CardDesignHeart, true)
	p.Grid[2] = makeSlot(13, CardDesignClover, true)
	p.Grid[3] = makeSlot(10, CardDesignDiamond, true)
	p.Grid[4] = makeSlot(11, CardDesignSpade, true)
	p.Grid[5] = makeSlot(12, CardDesignHeart, true)

	score := g.ScorePlayer(0)
	assert.Equal(t, 1+5+0+10+10+10, score)
}

func TestSixCardGolf_ScorePlayer_ColumnMatch(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	p := g.GetPlayer(0)
	p.Grid[0] = makeSlot(7, CardDesignSpade, true)
	p.Grid[1] = makeSlot(3, CardDesignHeart, true)
	p.Grid[2] = makeSlot(9, CardDesignClover, true)
	p.Grid[3] = makeSlot(7, CardDesignDiamond, true)
	p.Grid[4] = makeSlot(3, CardDesignSpade, true)
	p.Grid[5] = makeSlot(2, CardDesignHeart, true)

	score := g.ScorePlayer(0)
	assert.Equal(t, 0+0+9+0+0+2, score)
}

func TestSixCardGolf_ScorePlayer_AllColumnMatch(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	p := g.GetPlayer(0)
	p.Grid[0] = makeSlot(4, CardDesignSpade, true)
	p.Grid[1] = makeSlot(8, CardDesignHeart, true)
	p.Grid[2] = makeSlot(2, CardDesignClover, true)
	p.Grid[3] = makeSlot(4, CardDesignDiamond, true)
	p.Grid[4] = makeSlot(8, CardDesignSpade, true)
	p.Grid[5] = makeSlot(2, CardDesignHeart, true)

	assert.Equal(t, 0, g.ScorePlayer(0))
}

func TestSixCardGolf_ScorePlayer_FaceDownIgnored(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	p := g.GetPlayer(0)
	p.Grid[0] = makeSlot(5, CardDesignSpade, true)
	p.Grid[1] = makeSlot(5, CardDesignHeart, false)
	p.Grid[2] = makeSlot(5, CardDesignClover, true)
	p.Grid[3] = makeSlot(5, CardDesignDiamond, true)
	p.Grid[4] = makeSlot(5, CardDesignSpade, true)
	p.Grid[5] = makeSlot(5, CardDesignHeart, true)

	score := g.ScorePlayer(0)
	assert.Equal(t, 5+0+0, score)
}

func TestSixCardGolf_ScorePlayer_OutOfRange(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	assert.Equal(t, 0, g.ScorePlayer(-1))
	assert.Equal(t, 0, g.ScorePlayer(99))
}

func TestSixCardGolf_CardScore(t *testing.T) {
	g := newTestSixCardGolf()
	tests := []struct {
		value int
		want  int
	}{
		{13, 0},
		{1, 1},
		{2, 2},
		{10, 10},
		{11, 10},
		{12, 10},
		{5, 5},
	}
	for _, tc := range tests {
		c := NewCard(CardDesignSpade, tc.value, false)
		assert.Equal(t, tc.want, g.sixCardGolfCardScore(c))
	}
	assert.Equal(t, 0, g.sixCardGolfCardScore(nil))
}

// --- End Trigger & Final Turns ---

func TestSixCardGolf_EndTrigger(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	setupGridAllFaceUp(g, 0)
	p0 := g.GetPlayer(0)
	p0.Grid[5].FaceUp = false

	require.NoError(t, g.DrawStock())
	require.NoError(t, g.SwapCard(5))

	assert.Equal(t, 0, g.GetFinalTurnTrigger())
}

func TestSixCardGolf_FinalTurnFlow(t *testing.T) {
	g := NewSixCardGolf(NewTrumpCards(0), SixCardGolfConfig{PlayerCount: 2, CpuDifficulty: SixCardGolfCpuNormal, Rounds: 9})
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()
	skipSetup(g)

	for i := range g.GetPlayer(0).Grid {
		g.GetPlayer(0).Grid[i].FaceUp = true
	}
	for i := range g.GetPlayer(1).Grid {
		g.GetPlayer(1).Grid[i].FaceUp = true
	}
	g.GetPlayer(0).Grid[5].FaceUp = false

	g.SetDrawnCard(NewCard(CardDesignSpade, 1, false))
	g.SetPhase(SixCardGolfPhaseDrawPending)

	require.NoError(t, g.SwapCard(5))

	assert.Equal(t, 0, g.GetFinalTurnTrigger())
	assert.Equal(t, SixCardGolfPhasePlayerTurn, g.GetPhase())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())

	g.CpuPlay()
	assert.Equal(t, SixCardGolfPhaseRoundOver, g.GetPhase())
}

// --- Multi-Round ---

func TestSixCardGolf_NextRound(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	g.SetPhase(SixCardGolfPhaseRoundOver)
	g.GetPlayer(0).RoundScore = 10
	g.GetPlayer(0).CumulativeScore = 10

	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, SixCardGolfPhaseSetup, g.GetPhase())
}

func TestSixCardGolf_NextRound_WrongPhase(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	g.NextRound()
	assert.Equal(t, 1, g.GetRoundNumber())
}

func TestSixCardGolf_GameOver_AfterLastRound(t *testing.T) {
	g := NewSixCardGolf(NewTrumpCards(0), SixCardGolfConfig{PlayerCount: 2, CpuDifficulty: SixCardGolfCpuNormal, Rounds: 1})
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()
	skipSetup(g)

	for i := range g.GetPlayer(0).Grid {
		g.GetPlayer(0).Grid[i].FaceUp = true
	}
	for i := range g.GetPlayer(1).Grid {
		g.GetPlayer(1).Grid[i].FaceUp = true
	}
	g.GetPlayer(0).Grid[5].FaceUp = false

	g.SetDrawnCard(NewCard(CardDesignSpade, 13, false))
	g.SetPhase(SixCardGolfPhaseDrawPending)

	require.NoError(t, g.SwapCard(5))
	assert.Equal(t, 0, g.GetFinalTurnTrigger())

	g.CpuPlay()

	assert.Equal(t, SixCardGolfPhaseGameOver, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
}

// --- CPU ---

func TestSixCardGolf_CpuPlay_Setup(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()

	require.NoError(t, g.FlipInitial(0))
	require.NoError(t, g.FlipInitial(3))

	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	g.CpuPlay()
	assert.Equal(t, SixCardGolfPhasePlayerTurn, g.GetPhase())
	assert.Equal(t, 2, g.GetPlayer(1).FaceUpCount())
}

func TestSixCardGolf_CpuPlay_DrawAndSwap(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.SetCurrentPlayerIdx(1)

	g.CpuPlay()
	assert.NotEqual(t, SixCardGolfPhaseDrawPending, g.GetPhase())
}

func TestSixCardGolf_CpuPlay_NotCpuTurn(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	beforePhase := g.GetPhase()
	g.CpuPlay()
	assert.Equal(t, beforePhase, g.GetPhase())
}

func TestSixCardGolf_CpuPlay_GameEnded(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.SetCurrentPlayerIdx(1)
	g.gameEndFlag = true
	g.CpuPlay()
}

// --- IsHumanTurn ---

func TestSixCardGolf_IsHumanTurn(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
}

// --- Getters ---

func TestSixCardGolf_GetPlayer_OutOfRange(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestSixCardGolf_GetDiscardTop_Empty(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	g.SetDiscardPile(nil)
	assert.Nil(t, g.GetDiscardTop())
}

// --- Draw pile refill ---

func TestSixCardGolf_RefillDrawPile(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	c1 := NewCard(CardDesignSpade, 1, false)
	c2 := NewCard(CardDesignHeart, 2, false)
	c3 := NewCard(CardDesignClover, 3, false)

	g.SetDrawPile(nil)
	g.SetDiscardPile([]*Card{c1, c2, c3})

	require.NoError(t, g.DrawStock())
	assert.True(t, g.GetDrawPileCount() > 0 || g.GetDrawnCard() != nil)
}

// --- JSON ---

func TestSixCardGolf_JSONRoundtrip(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	g2 := &SixCardGolf{}
	require.NoError(t, json.Unmarshal(data, g2))

	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), g2.GetRoundNumber())
	assert.Equal(t, g.GetCurrentPlayerIdx(), g2.GetCurrentPlayerIdx())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
	assert.Equal(t, g.GetDrawPileCount(), g2.GetDrawPileCount())
}

func TestSixCardGolf_UnmarshalRejectsWinnerWhileNotGameOver(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)
	g.winnerIdx = 0

	data, err := json.Marshal(g)
	require.NoError(t, err)

	g2 := &SixCardGolf{}
	require.NoError(t, json.Unmarshal(data, g2))
	assert.Equal(t, -1, g2.GetWinnerIdx())
}

func TestSixCardGolf_UnmarshalRejectsInvalidPlayers(t *testing.T) {
	raw := `{"pl":[{"gr":[{},{},{},{},{},{}],"ic":false}],"cf":{"pc":1}}`
	g := &SixCardGolf{}
	assert.Error(t, json.Unmarshal([]byte(raw), g))
}

func TestSixCardGolf_UnmarshalRejectsOversizeArrays(t *testing.T) {
	big := make([]*Card, 201)
	j := sixCardGolfJSON{
		DrawPile: big,
		Players:  []*sixCardGolfPlayerJS{{}, {}},
	}
	data, _ := json.Marshal(j)
	g := &SixCardGolf{}
	assert.Error(t, json.Unmarshal(data, g))
}

// --- Column match helper ---

func TestSixCardGolf_WouldColumnMatch(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	p := g.GetPlayer(0)
	p.Grid[3] = makeSlot(7, CardDesignSpade, true)

	assert.True(t, g.wouldColumnMatch(0, 0, NewCard(CardDesignHeart, 7, false)))
	assert.False(t, g.wouldColumnMatch(0, 0, NewCard(CardDesignHeart, 5, false)))
}

func TestSixCardGolf_WouldColumnMatch_FaceDown(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	p := g.GetPlayer(0)
	p.Grid[3] = makeSlot(7, CardDesignSpade, false)

	assert.False(t, g.wouldColumnMatch(0, 0, NewCard(CardDesignHeart, 7, false)))
}

// --- DiscardDrawn when all face up after discard from stock ---

func TestSixCardGolf_DiscardDrawn_AllFaceUpNoFlip(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	skipSetup(g)

	setupGridAllFaceUp(g, 0)

	g.SetPhase(SixCardGolfPhasePlayerTurn)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.DrawStock())
	require.NoError(t, g.DiscardDrawn())
	assert.False(t, g.GetCanFlip())
}

// --- SetPlayerGrid ---

func TestSixCardGolf_SetPlayerGrid(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()

	var grid [SixCardGolfGridSize]SixCardGolfSlot
	for i := range grid {
		grid[i] = makeSlot(i+1, CardDesignSpade, true)
	}
	g.SetPlayerGrid(0, grid)
	assert.True(t, g.GetPlayer(0).AllFaceUp())

	g.SetPlayerGrid(-1, grid)
	g.SetPlayerGrid(99, grid)
}

// --- ActionLog ---

func TestSixCardGolf_ActionLog(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	require.NoError(t, g.FlipInitial(0))
	assert.True(t, len(g.GetActionLog()) > 0)
}

// --- 4 players ---

func TestSixCardGolf_FourPlayers(t *testing.T) {
	cfg := SixCardGolfConfig{PlayerCount: 4, CpuDifficulty: SixCardGolfCpuNormal, Rounds: 1}
	g := NewSixCardGolf(NewTrumpCards(0), cfg)
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()

	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.Equal(t, SixCardGolfPhaseSetup, g.GetPhase())

	totalCards := g.GetDrawPileCount() + len(g.discardPile)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalCards += SixCardGolfGridSize
	}
	assert.Equal(t, 52, totalCards)
}

// --- FinalTurnDone getters/setters ---

func TestSixCardGolf_FinalTurnDone(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()

	done := []bool{true, false}
	g.SetFinalTurnDone(done)
	assert.Equal(t, done, g.GetFinalTurnDone())

	g.SetFinalTurnTrigger(0)
	assert.Equal(t, 0, g.GetFinalTurnTrigger())
}

// --- Helpers ---

func skipSetup(g *SixCardGolf) {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		flipped := 0
		for j := range p.Grid {
			if !p.Grid[j].FaceUp && flipped < SixCardGolfInitialFlips {
				p.Grid[j].FaceUp = true
				flipped++
			}
		}
	}
	g.SetPhase(SixCardGolfPhasePlayerTurn)
	g.SetCurrentPlayerIdx(0)
}

func TestSixCardGolf_ShouldDrawFromDiscard(t *testing.T) {
	g := newTestSixCardGolf()
	g.Reset()
	g.SetDiscardPile([]*Card{NewCard(CardDesignSpade, 13, false)}) // King = score 0
	assert.True(t, g.ShouldDrawFromDiscard())
	g.SetDiscardPile([]*Card{NewCard(CardDesignSpade, 9, false)}) // 9 > 3
	assert.False(t, g.ShouldDrawFromDiscard())
	g.SetDiscardPile(nil)
	assert.False(t, g.ShouldDrawFromDiscard())
}

func TestSixCardGolf_RecommendedSwap(t *testing.T) {
	fill := func(g *SixCardGolf, value int) {
		p := g.GetPlayer(0)
		for i := range p.Grid {
			p.Grid[i] = makeSlot(value, CardDesignSpade, true)
		}
	}

	t.Run("swaps a low drawn card into a high grid", func(t *testing.T) {
		g := newTestSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(SixCardGolfPhaseDrawPending)
		fill(g, 11)                                        // Jacks = 10 each
		g.SetDrawnCard(NewCard(CardDesignSpade, 1, false)) // Ace = 1
		pos, pair := g.RecommendedSwap()
		assert.GreaterOrEqual(t, pos, 0)
		assert.False(t, pair)
	})

	t.Run("recommends discard for a high drawn card over a low grid", func(t *testing.T) {
		g := newTestSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(SixCardGolfPhaseDrawPending)
		fill(g, 1)                                          // Aces = 1 each
		g.SetDrawnCard(NewCard(CardDesignSpade, 11, false)) // Jack = 10
		pos, _ := g.RecommendedSwap()
		assert.Equal(t, -1, pos)
	})

	t.Run("flags a column pair", func(t *testing.T) {
		g := newTestSixCardGolf()
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(SixCardGolfPhaseDrawPending)
		fill(g, 9) // 9s so no accidental pairs
		p := g.GetPlayer(0)
		p.Grid[3] = makeSlot(5, CardDesignSpade, true) // column 0 partner is a 5
		g.SetDrawnCard(NewCard(CardDesignHeart, 5, false))
		pos, pair := g.RecommendedSwap()
		assert.Equal(t, 0, pos)
		assert.True(t, pair)
	})

	t.Run("no recommendation outside DrawPending", func(t *testing.T) {
		g := newTestSixCardGolf()
		g.Reset()
		g.SetPhase(SixCardGolfPhasePlayerTurn)
		pos, _ := g.RecommendedSwap()
		assert.Equal(t, -1, pos)
	})
}
