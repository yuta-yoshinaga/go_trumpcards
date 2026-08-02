//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestGongZhu() *domain.GongZhu {
	players := []*domain.GongZhuPlayer{
		domain.NewGongZhuPlayer(true),  // player 0 = human
		domain.NewGongZhuPlayer(false), // player 1 = CPU
		domain.NewGongZhuPlayer(false), // player 2 = CPU
		domain.NewGongZhuPlayer(false), // player 3 = CPU
	}
	return domain.NewGongZhu(domain.NewTrumpCards(0), players, domain.DefaultGongZhuConfig())
}

func gzCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

// gzWithHighLimit sets a large point limit so single-round scoring tests never
// trigger the game-end condition.
func gzWithHighLimit(g *domain.GongZhu) {
	cfg := g.GetConfig()
	cfg.PointLimit = 1000000
	g.SetConfig(cfg)
}

// gzScoreOneRound puts the game in round-end, runs ScoreRound, and returns the
// human player's resulting round score.
func gzScoreOneRound(t *testing.T, tricks []*domain.Card, e domain.GongZhuExposure) int {
	t.Helper()
	g := newTestGongZhu()
	gzWithHighLimit(g)
	g.SetExposure(e)
	g.GetPlayer(0).AddTrick(tricks)
	g.SetPhase(domain.GongZhuPhaseRoundEnd)
	g.ScoreRound()
	return g.GetPlayer(0).GetRoundScore()
}

// --- construction ---

func TestNewGongZhu(t *testing.T) {
	g := newTestGongZhu()
	assert.Equal(t, domain.GongZhuPhase(0), g.GetPhase())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetActionLog())
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestNewDefaultGongZhu(t *testing.T) {
	g := domain.NewDefaultGongZhu()
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())
}

func TestGongZhuReset(t *testing.T) {
	g := newTestGongZhu()
	g.Reset()
	assert.Equal(t, domain.GongZhuPhaseExpose, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
		assert.Equal(t, 13, g.GetPlayer(i).GetCardsSize())
	}
	assert.Equal(t, 52, total)
}

// --- config ---

func TestGongZhuConfigValidate(t *testing.T) {
	assert.NoError(t, domain.DefaultGongZhuConfig().Validate())
	bad := domain.GongZhuConfig{CpuDifficulty: 99, PointLimit: 1000}
	assert.Error(t, bad.Validate())
	bad2 := domain.GongZhuConfig{CpuDifficulty: domain.GongZhuCpuDifficultyEasy, PointLimit: 0}
	assert.Error(t, bad2.Validate())
}

// --- scoring ---

func TestGongZhuScoreHeartsGraded(t *testing.T) {
	// A(-50) + K(-40) + Q(-30) + J(-20) + 10(-10) + 4(0) = -150
	tricks := []*domain.Card{
		gzCard(domain.CardDesignHeart, 1),
		gzCard(domain.CardDesignHeart, 13),
		gzCard(domain.CardDesignHeart, 12),
		gzCard(domain.CardDesignHeart, 11),
		gzCard(domain.CardDesignHeart, 10),
		gzCard(domain.CardDesignHeart, 4),
	}
	assert.Equal(t, -150, gzScoreOneRound(t, tricks, domain.GongZhuExposure{}))
}

func TestGongZhuScorePig(t *testing.T) {
	tricks := []*domain.Card{gzCard(domain.CardDesignSpade, 12)}
	assert.Equal(t, -100, gzScoreOneRound(t, tricks, domain.GongZhuExposure{}))
	assert.Equal(t, -200, gzScoreOneRound(t, tricks, domain.GongZhuExposure{Pig: true}))
}

func TestGongZhuScoreSheep(t *testing.T) {
	tricks := []*domain.Card{gzCard(domain.CardDesignDiamond, 11)}
	assert.Equal(t, 100, gzScoreOneRound(t, tricks, domain.GongZhuExposure{}))
	assert.Equal(t, 200, gzScoreOneRound(t, tricks, domain.GongZhuExposure{Sheep: true}))
}

func TestGongZhuScoreDoublerStandalone(t *testing.T) {
	tricks := []*domain.Card{gzCard(domain.CardDesignClover, 10)}
	assert.Equal(t, 50, gzScoreOneRound(t, tricks, domain.GongZhuExposure{}))
	assert.Equal(t, 100, gzScoreOneRound(t, tricks, domain.GongZhuExposure{Doubler: true}))
}

func TestGongZhuScoreDoublerMultiplies(t *testing.T) {
	// pig (-100) with doubler => -200; exposed doubler => -400
	tricks := []*domain.Card{gzCard(domain.CardDesignSpade, 12), gzCard(domain.CardDesignClover, 10)}
	assert.Equal(t, -200, gzScoreOneRound(t, tricks, domain.GongZhuExposure{}))
	assert.Equal(t, -400, gzScoreOneRound(t, tricks, domain.GongZhuExposure{Doubler: true}))
}

func TestGongZhuScoreSheepWithDoubler(t *testing.T) {
	tricks := []*domain.Card{gzCard(domain.CardDesignDiamond, 11), gzCard(domain.CardDesignClover, 10)}
	assert.Equal(t, 200, gzScoreOneRound(t, tricks, domain.GongZhuExposure{}))
}

func TestGongZhuScoreAllHearts(t *testing.T) {
	var tricks []*domain.Card
	for v := 1; v <= 13; v++ {
		tricks = append(tricks, gzCard(domain.CardDesignHeart, v))
	}
	assert.Equal(t, 200, gzScoreOneRound(t, tricks, domain.GongZhuExposure{}))
	// exposed ace doubles the all-hearts bonus
	assert.Equal(t, 400, gzScoreOneRound(t, tricks, domain.GongZhuExposure{Ace: true}))
}

func TestGongZhuScoreAceExposureDoublesHearts(t *testing.T) {
	// A(-50) + K(-40) = -90; exposed ace => -180
	tricks := []*domain.Card{gzCard(domain.CardDesignHeart, 1), gzCard(domain.CardDesignHeart, 13)}
	assert.Equal(t, -90, gzScoreOneRound(t, tricks, domain.GongZhuExposure{}))
	assert.Equal(t, -180, gzScoreOneRound(t, tricks, domain.GongZhuExposure{Ace: true}))
}

func TestGongZhuScoreEmpty(t *testing.T) {
	assert.Equal(t, 0, gzScoreOneRound(t, []*domain.Card{gzCard(domain.CardDesignSpade, 3)}, domain.GongZhuExposure{}))
}

// --- exposure flow ---

func TestGongZhuPlayerExpose(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseExpose)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignSpade, 12))   // pig idx 0
	p.AddCard(gzCard(domain.CardDesignDiamond, 11)) // sheep idx 1
	p.AddCard(gzCard(domain.CardDesignSpade, 3))    // non-special idx 2

	assert.Equal(t, []int{0, 1}, g.GetExposableIndices(0))

	err := g.PlayerExpose([]int{0})
	assert.NoError(t, err)
	assert.True(t, g.GetExposure().Pig)
	assert.False(t, g.GetExposure().Sheep)
	assert.True(t, g.GetExposeReady()[0])

	// second expose rejected (already ready)
	assert.Error(t, g.PlayerExpose([]int{1}))
}

func TestGongZhuPlayerExposeErrors(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseExpose)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignSpade, 12)) // pig
	p.AddCard(gzCard(domain.CardDesignSpade, 3))  // non-special

	assert.Error(t, g.PlayerExpose([]int{5}))    // out of range
	assert.Error(t, g.PlayerExpose([]int{1}))    // non-special
	assert.Error(t, g.PlayerExpose([]int{0, 0})) // duplicate

	// wrong phase
	g.SetPhase(domain.GongZhuPhasePlay)
	assert.ErrorIs(t, g.PlayerExpose([]int{0}), domain.ErrWrongPhase)
}

func TestGongZhuExposeEmptyAllowed(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseExpose)
	g.GetPlayer(0).Reset()
	assert.NoError(t, g.PlayerExpose([]int{}))
	assert.True(t, g.GetExposeReady()[0])
}

func TestGongZhuExecuteExposeStartsPlay(t *testing.T) {
	g := newTestGongZhu()
	g.Reset() // deals + expose phase
	_ = g.PlayerExpose([]int{})
	// not all ready yet
	g.ExecuteExpose()
	assert.Equal(t, domain.GongZhuPhaseExpose, g.GetPhase())
	g.CpuExpose()
	g.ExecuteExpose()
	assert.Equal(t, domain.GongZhuPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetTrickNumber())
	assert.GreaterOrEqual(t, g.GetLeadPlayerIdx(), 0)
}

// --- validate play ---

func TestGongZhuValidatePlayFollowSuit(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetTrickNumber(2)
	g.SetCurrentPlayerIdx(0)
	g.SetHeartsBroken(true)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: gzCard(domain.CardDesignSpade, 5)},
	})
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignSpade, 8)) // follows
	p.AddCard(gzCard(domain.CardDesignHeart, 4)) // off-suit

	// must follow spade -> playing heart (idx 1) fails
	assert.Error(t, g.PlayerPlay(1))
	// playing spade (idx 0) ok
	assert.NoError(t, g.PlayerPlay(0))
}

func TestGongZhuValidateLeadHeartsRestriction(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetTrickNumber(3)
	g.SetCurrentPlayerIdx(0)
	g.SetHeartsBroken(false)
	g.SetCurrentTrick(nil)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignHeart, 5))
	p.AddCard(gzCard(domain.CardDesignSpade, 9))
	// leading hearts while not broken & holding non-heart -> error
	assert.Error(t, g.PlayerPlay(0))
	// leading spade ok
	assert.NoError(t, g.PlayerPlay(1))
}

func TestGongZhuLeadHeartsWhenOnlyHearts(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetTrickNumber(3)
	g.SetCurrentPlayerIdx(0)
	g.SetHeartsBroken(false)
	g.SetCurrentTrick(nil)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignHeart, 5))
	assert.NoError(t, g.PlayerPlay(0))
	assert.True(t, g.GetHeartsBroken())
}

// --- trick resolution ---

func TestGongZhuTrickWinnerAceHigh(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseTrickEnd)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: gzCard(domain.CardDesignSpade, 13)}, // K
		{PlayerIdx: 1, Card: gzCard(domain.CardDesignSpade, 1)},  // A (wins)
		{PlayerIdx: 2, Card: gzCard(domain.CardDesignSpade, 9)},
		{PlayerIdx: 3, Card: gzCard(domain.CardDesignHeart, 5)}, // off-suit
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
	assert.Equal(t, 1, g.GetPlayer(1).GetTrickCount())
}

func TestGongZhuNextTrick(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(1)
	g.NextTrick()
	assert.Equal(t, domain.GongZhuPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 2, g.GetTrickNumber())
}

func TestGongZhuResolveTrickLastSetsRoundEnd(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseTrickEnd)
	g.SetTrickNumber(domain.GongZhuHandSize)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: gzCard(domain.CardDesignSpade, 13)},
		{PlayerIdx: 1, Card: gzCard(domain.CardDesignSpade, 2)},
		{PlayerIdx: 2, Card: gzCard(domain.CardDesignSpade, 9)},
		{PlayerIdx: 3, Card: gzCard(domain.CardDesignSpade, 5)},
	})
	g.ResolveTrick()
	assert.Equal(t, domain.GongZhuPhaseRoundEnd, g.GetPhase())
}

// --- score round / game end ---

func TestGongZhuScoreRoundGameEnd(t *testing.T) {
	g := newTestGongZhu()
	cfg := g.GetConfig()
	cfg.PointLimit = 100
	g.SetConfig(cfg)
	g.SetPhase(domain.GongZhuPhaseRoundEnd)
	// player 2 takes the pig (-100) => reaches -limit
	g.GetPlayer(2).AddTrick([]*domain.Card{gzCard(domain.CardDesignSpade, 12)})
	// player 1 takes the sheep (+100) => highest
	g.GetPlayer(1).AddTrick([]*domain.Card{gzCard(domain.CardDesignDiamond, 11)})
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.GongZhuPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestGongZhuScoreRoundWrongPhaseNoop(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhasePlay)
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag())
}

func TestGongZhuNextRound(t *testing.T) {
	g := newTestGongZhu()
	g.Reset()
	g.SetPhase(domain.GongZhuPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.GongZhuPhaseExpose, g.GetPhase())
	assert.Equal(t, 13, g.GetPlayer(0).GetCardsSize())
}

// --- CPU play ---

func TestGongZhuCpuPlayProducesValidCard(t *testing.T) {
	for _, diff := range []domain.GongZhuCpuDifficulty{
		domain.GongZhuCpuDifficultyEasy,
		domain.GongZhuCpuDifficultyNormal,
		domain.GongZhuCpuDifficultyHard,
	} {
		g := newTestGongZhu()
		cfg := g.GetConfig()
		cfg.CpuDifficulty = diff
		g.SetConfig(cfg)
		g.Reset()
		_ = g.PlayerExpose([]int{})
		g.CpuExpose()
		g.ExecuteExpose()
		// run a full round of CPU/auto play to exercise AI branches
		guard := 0
		for g.GetPhase() != domain.GongZhuPhaseGameEnd && g.GetPhase() != domain.GongZhuPhaseRoundEnd && guard < 200 {
			guard++
			switch g.GetPhase() {
			case domain.GongZhuPhasePlay:
				if g.IsHumanTurn() {
					idx := 0 // human plays first legal card
					// find a legal card
					hp := g.GetPlayer(g.GetCurrentPlayerIdx())
					for i := 0; i < hp.GetCardsSize(); i++ {
						if g.PlayerPlay(i) == nil {
							idx = -1
							break
						}
					}
					_ = idx
				} else {
					g.CpuPlay()
				}
			case domain.GongZhuPhaseTrickEnd:
				g.ResolveTrick()
				if g.GetPhase() == domain.GongZhuPhaseTrickEnd {
					g.NextTrick()
				}
			}
		}
		assert.Less(t, guard, 200, "round should finish")
	}
}

func TestGongZhuCpuFollowCapturesSheep(t *testing.T) {
	g := newTestGongZhu()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.GongZhuCpuDifficultyHard
	g.SetConfig(cfg)
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetTrickNumber(2)
	g.SetCurrentPlayerIdx(1)
	g.SetHeartsBroken(true)
	// lead diamond 5, sheep (J) already in trick -> CPU should try to win with A
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: gzCard(domain.CardDesignDiamond, 5)},
		{PlayerIdx: 3, Card: gzCard(domain.CardDesignDiamond, 11)}, // sheep
	})
	p := g.GetPlayer(1)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignDiamond, 2))
	p.AddCard(gzCard(domain.CardDesignDiamond, 13)) // K beats 11 -> capture sheep
	g.CpuPlay()
	last := g.GetCurrentTrick()[len(g.GetCurrentTrick())-1]
	assert.Equal(t, 13, last.Card.GetValue())
}

func TestGongZhuCpuDiscardDumpsPig(t *testing.T) {
	g := newTestGongZhu()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.GongZhuCpuDifficultyHard
	g.SetConfig(cfg)
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetTrickNumber(2)
	g.SetCurrentPlayerIdx(1)
	g.SetHeartsBroken(true)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: gzCard(domain.CardDesignClover, 5)},
	})
	p := g.GetPlayer(1)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignSpade, 12)) // pig - should be dumped
	p.AddCard(gzCard(domain.CardDesignSpade, 3))
	g.CpuPlay()
	last := g.GetCurrentTrick()[len(g.GetCurrentTrick())-1]
	assert.True(t, last.Card.GetDesign() == domain.CardDesignSpade && last.Card.GetValue() == 12)
}

// --- hint ---

func TestGongZhuHintExpose(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseExpose)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignDiamond, 11)) // sheep
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "expose_sheep", hint.Reason)

	p.Reset()
	p.AddCard(gzCard(domain.CardDesignSpade, 3))
	hint = g.GetHint()
	assert.Equal(t, "expose_none", hint.Reason)
}

func TestGongZhuHintPlay(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetTrickNumber(2)
	g.SetHeartsBroken(true)
	g.SetCurrentTrick(nil)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(gzCard(domain.CardDesignSpade, 4))
	p.AddCard(gzCard(domain.CardDesignClover, 9))
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.Len(t, hint.CardIndices, 1)
}

func TestGongZhuHintNilOtherPhase(t *testing.T) {
	g := newTestGongZhu()
	g.SetPhase(domain.GongZhuPhaseRoundEnd)
	assert.Nil(t, g.GetHint())
}

// --- JSON ---

func TestGongZhuJSONRoundTrip(t *testing.T) {
	g := newTestGongZhu()
	g.Reset()
	g.SetExposure(domain.GongZhuExposure{Pig: true, Sheep: true})
	data, err := json.Marshal(g)
	assert.NoError(t, err)

	var g2 domain.GongZhu
	assert.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), g2.GetRoundNumber())
	assert.Equal(t, 4, g2.GetPlayerCnt())
	assert.True(t, g2.GetExposure().Pig)
	assert.True(t, g2.GetExposure().Sheep)
}

func TestGongZhuUnmarshalEmpty(t *testing.T) {
	var g domain.GongZhu
	assert.NoError(t, json.Unmarshal([]byte(`{}`), &g))
	assert.Equal(t, 0, g.GetPlayerCnt())
	assert.NotNil(t, g.GetActionLog())
}

func TestGongZhuUnmarshalInvalid(t *testing.T) {
	var g domain.GongZhu
	assert.Error(t, json.Unmarshal([]byte(`{`), &g))
}

func TestGongZhuPlayerJSONRoundTrip(t *testing.T) {
	p := domain.NewGongZhuPlayer(true)
	p.AddCard(gzCard(domain.CardDesignHeart, 5))
	p.SetRoundScore(-30)
	p.SetCumulativeScore(-90)
	data, err := json.Marshal(p)
	assert.NoError(t, err)
	var p2 domain.GongZhuPlayer
	assert.NoError(t, json.Unmarshal(data, &p2))
	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, -30, p2.GetRoundScore())
	assert.Equal(t, -90, p2.GetCumulativeScore())
}

// **手札が尽きた CPU に手番が回ってもパニックしない。**
// cpuSelectPlayCard は候補が無いとき 0 を返し、RemoveCard(0) は手札が空なら
// nil を返す。それをそのまま playCard に渡していたため、GetDesign() で
// nil デリファレンスになりサーバごと落ちていた (E2E で観測、GongZhu.go:592)。
func TestGongZhuCpuPlayWithNoCards(t *testing.T) {
	g := domain.NewDefaultGongZhu()
	g.Reset()
	g.SetPhase(domain.GongZhuPhasePlay)

	// 席 1 は CPU。手札を空にする。
	cpu := g.GetPlayer(1)
	for cpu.GetCardsSize() > 0 {
		cpu.RemoveCard(0)
	}
	g.SetCurrentPlayerIdx(1)

	g.CpuPlay() // パニックしないこと

	if len(g.GetCurrentTrick()) != 0 {
		t.Errorf("手札の無い席が札を出した: trick=%d", len(g.GetCurrentTrick()))
	}
}
