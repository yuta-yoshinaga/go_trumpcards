//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestCourtPiece(humanSeat0 bool) *domain.CourtPiece {
	players := []*domain.CourtPiecePlayer{
		domain.NewCourtPiecePlayer(humanSeat0, 0),
		domain.NewCourtPiecePlayer(false, 1),
		domain.NewCourtPiecePlayer(false, 0),
		domain.NewCourtPiecePlayer(false, 1),
	}
	return domain.NewCourtPiece(domain.NewTrumpCards(0), players, domain.DefaultCourtPieceConfig())
}

func cpCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func TestCourtPieceConfig_Validate(t *testing.T) {
	cfg := domain.DefaultCourtPieceConfig()
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, domain.CourtPieceDefaultPointLimit, cfg.PointLimit)

	bad := domain.CourtPieceConfig{CpuDifficulty: 99, PointLimit: 7}
	assert.Error(t, bad.Validate())
	bad2 := domain.CourtPieceConfig{CpuDifficulty: domain.CourtPieceCpuDifficultyNormal, PointLimit: 0}
	assert.Error(t, bad2.Validate())
	bad3 := domain.CourtPieceConfig{CpuDifficulty: domain.CourtPieceCpuDifficultyNormal, PointLimit: 999}
	assert.Error(t, bad3.Validate())
}

func TestCourtPiecePlayer_ResetRound(t *testing.T) {
	p := domain.NewCourtPiecePlayer(true, 0)
	p.AddCard(cpCard(domain.CardDesignSpade, 1))
	p.AddTrick([]*domain.Card{cpCard(domain.CardDesignHeart, 2)})
	p.ResetRound()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.Equal(t, 0, p.GetTeam())
}

func TestNewDefaultCourtPiece(t *testing.T) {
	c := domain.NewDefaultCourtPiece()
	require.NotNil(t, c)
	assert.Equal(t, domain.CourtPiecePlayerCnt, c.GetPlayerCnt())
	assert.True(t, c.GetPlayer(0).GetIsHuman())
	assert.Equal(t, 0, c.GetPlayer(0).GetTeam())
	assert.Equal(t, 1, c.GetPlayer(1).GetTeam())
	assert.Equal(t, -1, c.GetWinnerTeam())
	assert.Nil(t, c.GetPlayer(99))
}

func TestCourtPiece_ResetDealsPeekToCaller(t *testing.T) {
	c := newTestCourtPiece(true)
	c.Reset()
	assert.Equal(t, domain.CourtPiecePhaseTrumpDeclaration, c.GetPhase())
	assert.Equal(t, 1, c.GetRoundNumber())
	assert.Equal(t, 0, c.GetCallerIdx())
	// Only the caller has cards (the peek), others empty until declaration.
	assert.Equal(t, domain.CourtPiecePeekSize, c.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 0, c.GetPlayer(1).GetCardsSize())
}

func TestCourtPiece_DeclareTrumpDealsRest(t *testing.T) {
	c := newTestCourtPiece(true)
	c.Reset()
	require.NoError(t, c.PlayerDeclareTrump(domain.CardDesignSpade))
	assert.Equal(t, domain.CourtPiecePhasePlay, c.GetPhase())
	assert.Equal(t, domain.CardDesignSpade, c.GetTrumpSuit())
	for i := 0; i < domain.CourtPiecePlayerCnt; i++ {
		assert.Equal(t, domain.CourtPieceHandSize, c.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Equal(t, 1, c.GetTrickNumber())
	assert.Equal(t, 0, c.GetCurrentPlayerIdx())
}

func TestCourtPiece_DeclareTrumpErrors(t *testing.T) {
	c := newTestCourtPiece(true)
	c.Reset()
	assert.Error(t, c.PlayerDeclareTrump(99)) // invalid suit
	c.SetPhase(domain.CourtPiecePhasePlay)
	assert.ErrorIs(t, c.PlayerDeclareTrump(domain.CardDesignSpade), domain.ErrWrongPhase)

	// CPU caller → human declaration is rejected.
	cpu := newTestCourtPiece(false)
	cpu.Reset()
	assert.ErrorIs(t, cpu.PlayerDeclareTrump(domain.CardDesignSpade), domain.ErrNotHumanTurn)
}

func TestCourtPiece_CpuDeclareTrump(t *testing.T) {
	c := newTestCourtPiece(false)
	c.Reset()
	c.CpuDeclareTrump()
	assert.Equal(t, domain.CourtPiecePhasePlay, c.GetPhase())
	assert.True(t, c.GetTrumpSuit() >= domain.CardDesignSpade && c.GetTrumpSuit() <= domain.CardDesignDiamond)
}

func TestCourtPiece_MustFollowSuit(t *testing.T) {
	c := newTestCourtPiece(true)
	c.SetPhase(domain.CourtPiecePhasePlay)
	c.SetTrumpSuit(domain.CardDesignSpade)
	c.SetCurrentPlayerIdx(0)
	c.SetTrickNumber(1)
	// Lead a heart; seat 0 holds a heart and a club -> must follow heart.
	c.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: cpCard(domain.CardDesignHeart, 7)},
	})
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(cpCard(domain.CardDesignHeart, 9))  // idx 0 (legal)
	p.AddCard(cpCard(domain.CardDesignClover, 5)) // idx 1 (illegal: has heart)
	assert.Error(t, c.PlayerPlay(1))
	require.NoError(t, c.PlayerPlay(0))
}

func TestCourtPiece_TrickWinnerTrumpBeatsFailLead(t *testing.T) {
	c := newTestCourtPiece(true)
	c.SetTrumpSuit(domain.CardDesignSpade)
	c.SetPhase(domain.CourtPiecePhaseTrickEnd)
	c.SetTrickNumber(1)
	// Lead a high heart, but a low trump spade should win.
	c.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: cpCard(domain.CardDesignHeart, 1)},   // A♥ lead
		{PlayerIdx: 1, Card: cpCard(domain.CardDesignHeart, 13)},  // K♥
		{PlayerIdx: 2, Card: cpCard(domain.CardDesignSpade, 2)},   // 2♠ trump
		{PlayerIdx: 3, Card: cpCard(domain.CardDesignDiamond, 5)}, // off
	})
	c.ResolveTrick()
	assert.Equal(t, 1, c.GetPlayer(2).GetTrickCount())
}

func TestCourtPiece_TrickWinnerAceHigh(t *testing.T) {
	c := newTestCourtPiece(true)
	c.SetTrumpSuit(domain.CardDesignSpade) // no trump in trick
	c.SetPhase(domain.CourtPiecePhaseTrickEnd)
	c.SetTrickNumber(1)
	c.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: cpCard(domain.CardDesignHeart, 13)}, // K♥
		{PlayerIdx: 1, Card: cpCard(domain.CardDesignHeart, 1)},  // A♥ wins (ace-high)
		{PlayerIdx: 2, Card: cpCard(domain.CardDesignHeart, 10)},
		{PlayerIdx: 3, Card: cpCard(domain.CardDesignHeart, 2)},
	})
	c.ResolveTrick()
	assert.Equal(t, 1, c.GetPlayer(1).GetTrickCount())
}

func TestCourtPiece_ScoreRoundSarAndCallerRotation(t *testing.T) {
	c := newTestCourtPiece(true)
	c.SetPhase(domain.CourtPiecePhaseRoundEnd)
	c.SetCallerIdx(0) // team 0 caller
	// Team 1 (seats 1,3) takes 7 tricks -> team 1 wins, caller (team 0) loses -> rotates.
	for i := 0; i < 4; i++ {
		c.GetPlayer(1).AddTrick([]*domain.Card{cpCard(domain.CardDesignSpade, 2)})
	}
	for i := 0; i < 3; i++ {
		c.GetPlayer(3).AddTrick([]*domain.Card{cpCard(domain.CardDesignSpade, 3)})
	}
	for i := 0; i < 6; i++ {
		c.GetPlayer(0).AddTrick([]*domain.Card{cpCard(domain.CardDesignSpade, 4)})
	}
	c.ScoreRound()
	assert.Equal(t, 1, c.GetTeamScore(1))
	assert.Equal(t, 0, c.GetTeamScore(0))
	assert.Equal(t, 1, c.GetLastWinnerTeam())
	assert.Equal(t, 1, c.GetCallerIdx()) // rotated since caller team lost
}

func TestCourtPiece_ScoreRoundCourtConsecutive(t *testing.T) {
	c := newTestCourtPiece(true)
	c.SetCallerIdx(0)
	// Round 1: team 0 wins 7-6.
	dealTricks(c, 0, 7)
	dealTricks(c, 1, 6)
	c.SetPhase(domain.CourtPiecePhaseRoundEnd)
	c.ScoreRound()
	assert.Equal(t, 1, c.GetTeamScore(0))
	assert.Equal(t, 0, c.GetCallerIdx()) // team 0 kept the call

	// Round 2: team 0 wins again -> Court bonus (+2).
	resetTricks(c)
	dealTricks(c, 0, 8)
	dealTricks(c, 1, 5)
	c.SetPhase(domain.CourtPiecePhaseRoundEnd)
	c.ScoreRound()
	assert.Equal(t, 3, c.GetTeamScore(0)) // 1 + 2 (court)
	assert.True(t, c.IsLastRoundCourt())
	assert.GreaterOrEqual(t, c.GetConsecutiveWins(), 2)
}

func TestCourtPiece_ScoreRoundCleanSweepIsCourt(t *testing.T) {
	c := newTestCourtPiece(true)
	c.SetCallerIdx(0)
	dealTricks(c, 0, 13) // all 13 -> clean Court
	c.SetPhase(domain.CourtPiecePhaseRoundEnd)
	c.ScoreRound()
	assert.Equal(t, 2, c.GetTeamScore(0))
	assert.True(t, c.IsLastRoundCourt())
}

func TestCourtPiece_GameEnd(t *testing.T) {
	c := newTestCourtPiece(true)
	c.SetCallerIdx(0)
	c.SetTeamScore(0, domain.CourtPieceDefaultPointLimit-1)
	dealTricks(c, 0, 7)
	dealTricks(c, 1, 6)
	c.SetPhase(domain.CourtPiecePhaseRoundEnd)
	c.ScoreRound()
	assert.True(t, c.GetGameEndFlag())
	assert.Equal(t, domain.CourtPiecePhaseGameEnd, c.GetPhase())
	assert.Equal(t, 0, c.GetWinnerTeam())
}

func dealTricks(c *domain.CourtPiece, team, n int) {
	seat := team // seat 0 for team 0, seat 1 for team 1
	for i := 0; i < n; i++ {
		c.GetPlayer(seat).AddTrick([]*domain.Card{cpCard(domain.CardDesignSpade, 2)})
	}
}

func resetTricks(c *domain.CourtPiece) {
	for i := 0; i < c.GetPlayerCnt(); i++ {
		c.GetPlayer(i).ResetTricks()
	}
}

func TestCourtPiece_Hint(t *testing.T) {
	c := newTestCourtPiece(true)
	c.Reset()
	// Declaration phase hint.
	h := c.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.TrumpSuit)
	require.NoError(t, c.PlayerDeclareTrump(*h.TrumpSuit))
	// Play phase hint.
	h2 := c.GetHint()
	require.NotNil(t, h2)
	require.NotNil(t, h2.CardIndex)
}

func TestCourtPiece_FullCpuGame(t *testing.T) {
	c := newTestCourtPiece(false) // all CPU
	c.Reset()
	guard := 0
	for !c.GetGameEndFlag() && guard < 100000 {
		guard++
		switch c.GetPhase() {
		case domain.CourtPiecePhaseTrumpDeclaration:
			c.CpuDeclareTrump()
		case domain.CourtPiecePhasePlay:
			c.CpuPlay()
		case domain.CourtPiecePhaseTrickEnd:
			c.ResolveTrick()
			c.NextTrick()
		case domain.CourtPiecePhaseRoundEnd:
			c.ScoreRound()
			if !c.GetGameEndFlag() {
				c.NextRound()
			}
		}
	}
	assert.True(t, c.GetGameEndFlag(), "game should terminate")
	assert.True(t, c.GetWinnerTeam() == 0 || c.GetWinnerTeam() == 1)
	assert.GreaterOrEqual(t, c.GetTeamScore(c.GetWinnerTeam()), domain.CourtPieceDefaultPointLimit)
}

func TestCourtPiece_JSONRoundTrip(t *testing.T) {
	c := newTestCourtPiece(true)
	c.Reset()
	require.NoError(t, c.PlayerDeclareTrump(domain.CardDesignHeart))
	data, err := json.Marshal(c)
	require.NoError(t, err)

	var c2 domain.CourtPiece
	require.NoError(t, json.Unmarshal(data, &c2))
	assert.Equal(t, c.GetPhase(), c2.GetPhase())
	assert.Equal(t, c.GetTrumpSuit(), c2.GetTrumpSuit())
	assert.Equal(t, c.GetCallerIdx(), c2.GetCallerIdx())
	assert.Equal(t, c.GetPlayerCnt(), c2.GetPlayerCnt())
}

func TestCourtPiece_UnmarshalRejectsInvalidState(t *testing.T) {
	c := newTestCourtPiece(true)
	c.Reset()
	require.NoError(t, c.PlayerDeclareTrump(domain.CardDesignHeart))
	data, err := json.Marshal(c)
	require.NoError(t, err)

	// Tamper the caller index out of range.
	tampered := strings.Replace(string(data), `"ka":0`, `"ka":9`, 1)
	require.NotEqual(t, string(data), tampered)
	var bad domain.CourtPiece
	assert.Error(t, bad.UnmarshalJSON([]byte(tampered)))

	// Malformed JSON.
	var bad2 domain.CourtPiece
	assert.Error(t, bad2.UnmarshalJSON([]byte(`not json`)))

	// Out-of-range team on a player must be rejected.
	tamperedTeam := strings.Replace(string(data), `"tm":1`, `"tm":9`, 1)
	require.NotEqual(t, string(data), tamperedTeam)
	var bad3 domain.CourtPiece
	assert.Error(t, bad3.UnmarshalJSON([]byte(tamperedTeam)))

	// Wrong player count must be rejected.
	var bad4 domain.CourtPiece
	assert.Error(t, bad4.UnmarshalJSON([]byte(`{"ps":[null,null]}`)))
}

// **マストフォローの判定を1箇所に寄せた。**それまで Go の validatePlay と
// TypeScript の courtPieceLegalPlayIndices に別々の実装があり、ルールを変えたら
// 片方だけ直してずれる状態だった。
//
// ここで確かめるのは「GetPlayableIndices が validatePlay と同じ答えを返す」こと。
// 期待値を手で書くとその写経がもう1つの実装になってしまうので、全札を
// validatePlay に通した結果と突き合わせる。
func TestCourtPiece_GetPlayableIndicesAgreesWithPlayerPlay(t *testing.T) {
	c := domain.NewDefaultCourtPiece()
	c.Reset()
	c.SetPhase(domain.CourtPiecePhasePlay)
	c.SetCurrentPlayerIdx(0)

	human := c.GetPlayer(0)
	if human.GetCardsSize() == 0 {
		t.Fatal("前提: 人間に手札が配られていること")
	}

	// リード時は全札が合法。
	c.SetCurrentTrick(nil)
	lead := c.GetPlayableIndices(0)
	if len(lead) != human.GetCardsSize() {
		t.Errorf("リード時は全札が合法のはず: %d/%d", len(lead), human.GetCardsSize())
	}

	// 人間が持っているスートをリードさせると、そのスートだけが合法になる。
	leadSuit := human.GetCard(0).GetDesign()
	c.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(leadSuit, 5, false)},
	})

	got := c.GetPlayableIndices(0)
	if len(got) == 0 {
		t.Fatal("リードスートを1枚は持っているので空にはならない")
	}
	for _, idx := range got {
		if d := human.GetCard(idx).GetDesign(); d != leadSuit {
			t.Errorf("index %d はリードスート %d でないのに合法とされた (design=%d)", idx, leadSuit, d)
		}
	}
	// 逆側: リードスートの札がすべて含まれていること。
	for i := 0; i < human.GetCardsSize(); i++ {
		if human.GetCard(i).GetDesign() != leadSuit {
			continue
		}
		found := false
		for _, idx := range got {
			if idx == i {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("index %d はリードスートなのに合法に含まれていない", i)
		}
	}

	// プレイフェーズ以外では空。
	c.SetPhase(domain.CourtPiecePhaseTrumpDeclaration)
	if len(c.GetPlayableIndices(0)) != 0 {
		t.Error("プレイフェーズ以外では空を返す")
	}
	// 範囲外のプレイヤーでも落ちない。
	c.SetPhase(domain.CourtPiecePhasePlay)
	if len(c.GetPlayableIndices(99)) != 0 {
		t.Error("範囲外のプレイヤーでは空を返す")
	}
}
