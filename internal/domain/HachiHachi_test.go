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

// hachihachiCard は Hachi-Hachi テスト用のカード生成ショートカット (design=月, value=index)。
func hachihachiCard(month, index int) *domain.Card { return domain.NewCard(month, index, false) }

// hachihachiFullDeck は花札 48 枚を返す。
func hachihachiFullDeck() []*domain.Card {
	deck := make([]*domain.Card, 0, 48)
	for m := 1; m <= 12; m++ {
		for i := 1; i <= 4; i++ {
			deck = append(deck, hachihachiCard(m, i))
		}
	}
	return deck
}

// hachihachiYakuMap は成立出来役をキー→ボーナス点のマップにして返す。
func hachihachiYakuMap(cards []*domain.Card) (map[string]int, int, int) {
	raw, yaku, bonus := domain.HachiHachiEvaluateScore(cards)
	m := make(map[string]int)
	for _, y := range yaku {
		m[y.Key] = y.Points
	}
	return m, raw, bonus
}

// 代表札。
var (
	hhCrane     = hachihachiCard(1, 1)  // 松 光
	hhCurtain   = hachihachiCard(3, 1)  // 桜 光
	hhMoon      = hachihachiCard(8, 1)  // 芒 光
	hhRainman   = hachihachiCard(11, 1) // 柳 光 (雨)
	hhPhoenix   = hachihachiCard(12, 1) // 桐 光
	hhBoar      = hachihachiCard(7, 1)  // 萩 タネ
	hhDeer      = hachihachiCard(10, 1) // 紅葉 タネ
	hhButterfly = hachihachiCard(6, 1)  // 牡丹 タネ
)

// --- 素点 (card-point values) ---

func TestHachiHachiCardPoints(t *testing.T) {
	assert.Equal(t, 20, domain.HachiHachiCardPoints(hhCrane))             // 光
	assert.Equal(t, 10, domain.HachiHachiCardPoints(hhBoar))              // タネ
	assert.Equal(t, 5, domain.HachiHachiCardPoints(hachihachiCard(1, 2))) // 短冊
	assert.Equal(t, 1, domain.HachiHachiCardPoints(hachihachiCard(1, 3))) // カス
	assert.Equal(t, 1, domain.HachiHachiCardPoints(nil))                  // nil-safe
}

func TestHachiHachiFullDeckPoints(t *testing.T) {
	raw, _, _ := domain.HachiHachiEvaluateScore(hachihachiFullDeck())
	// 5×20 + 9×10 + 10×5 + 24×1 = 264 = 3×88。
	assert.Equal(t, 264, raw)
	assert.Equal(t, 3*domain.HachiHachiBaseline, raw)
}

// --- 出来役 ---

func TestHachiHachiYaku_Goko(t *testing.T) {
	m, raw, bonus := hachihachiYakuMap([]*domain.Card{hhCrane, hhCurtain, hhMoon, hhRainman, hhPhoenix})
	assert.Equal(t, 100, m["goko"])
	assert.Equal(t, 100, raw) // 光 5 枚 × 20
	assert.Equal(t, 100, bonus)
}

func TestHachiHachiYaku_Shiko(t *testing.T) {
	m, _, bonus := hachihachiYakuMap([]*domain.Card{hhCrane, hhCurtain, hhMoon, hhPhoenix})
	assert.Equal(t, 80, m["shiko"])
	assert.NotContains(t, m, "goko")
	assert.Equal(t, 80, bonus)
}

func TestHachiHachiYaku_AmeShiko(t *testing.T) {
	m, _, _ := hachihachiYakuMap([]*domain.Card{hhCrane, hhCurtain, hhRainman, hhPhoenix})
	assert.Equal(t, 60, m["ameshiko"])
	assert.NotContains(t, m, "shiko")
}

func TestHachiHachiYaku_Sanko(t *testing.T) {
	m, _, _ := hachihachiYakuMap([]*domain.Card{hhCrane, hhCurtain, hhPhoenix})
	assert.Equal(t, 40, m["sanko"])
}

func TestHachiHachiYaku_SankoBlockedByRain(t *testing.T) {
	m, _, _ := hachihachiYakuMap([]*domain.Card{hhCrane, hhRainman, hhPhoenix})
	assert.NotContains(t, m, "sanko")
}

func TestHachiHachiYaku_Inoshikacho(t *testing.T) {
	m, _, _ := hachihachiYakuMap([]*domain.Card{hhBoar, hhDeer, hhButterfly})
	assert.Equal(t, 50, m["inoshikacho"])
}

func TestHachiHachiYaku_Akatan(t *testing.T) {
	m, _, _ := hachihachiYakuMap([]*domain.Card{hachihachiCard(1, 2), hachihachiCard(2, 2), hachihachiCard(3, 2)})
	assert.Equal(t, 40, m["akatan"])
}

func TestHachiHachiYaku_Aotan(t *testing.T) {
	m, _, _ := hachihachiYakuMap([]*domain.Card{hachihachiCard(6, 2), hachihachiCard(9, 2), hachihachiCard(10, 2)})
	assert.Equal(t, 40, m["aotan"])
}

func TestHachiHachiYaku_Empty(t *testing.T) {
	m, raw, bonus := hachihachiYakuMap(nil)
	assert.Empty(t, m)
	assert.Equal(t, 0, raw)
	assert.Equal(t, 0, bonus)
}

func TestHachiHachiYaku_NilCardSkipped(t *testing.T) {
	m, raw, _ := hachihachiYakuMap([]*domain.Card{hhCrane, nil, hhCurtain})
	assert.Equal(t, 40, raw) // 2 brights, nil skipped
	assert.NotContains(t, m, "sanko")
}

// --- 88 精算 (ゼロ和) ---

// hachihachiDelta は素点 + 役ボーナス − 88 を返す。
func hachihachiDelta(cards []*domain.Card) int {
	raw, _, bonus := domain.HachiHachiEvaluateScore(cards)
	return raw + bonus - domain.HachiHachiBaseline
}

func TestHachiHachiSettlement_ZeroSumEvenSplit(t *testing.T) {
	deck := hachihachiFullDeck()
	// 16 枚ずつ 3 分割 (出来役が偶発しても素点部分だけをゼロ和検証)。
	var p [3][]*domain.Card
	for i, c := range deck {
		p[i%3] = append(p[i%3], c)
	}
	sum := 0
	for i := 0; i < 3; i++ {
		raw, _, _ := domain.HachiHachiEvaluateScore(p[i])
		sum += raw - domain.HachiHachiBaseline
	}
	assert.Equal(t, 0, sum, "素点 − 88 の 3 人合計は 0 (ゼロ和)")
}

func TestHachiHachiSettlement_ZeroSumLopsided(t *testing.T) {
	deck := hachihachiFullDeck()
	// 1 人が全 48 枚を捕獲した極端なケース。
	raw0, _, _ := domain.HachiHachiEvaluateScore(deck)
	d0 := raw0 - domain.HachiHachiBaseline
	d1 := 0 - domain.HachiHachiBaseline
	d2 := 0 - domain.HachiHachiBaseline
	assert.Equal(t, 176, d0)
	assert.Equal(t, 0, d0+d1+d2)
}

func TestHachiHachiDelta(t *testing.T) {
	// 五光 = 素点 100 + 役 100 − 88 = 112。
	assert.Equal(t, 112, hachihachiDelta([]*domain.Card{hhCrane, hhCurtain, hhMoon, hhRainman, hhPhoenix}))
}

// --- ゲーム進行 ---

func newTestHachiHachi(t *testing.T, diff domain.HachiHachiCpuDifficulty) *domain.HachiHachi {
	t.Helper()
	g := domain.NewDefaultHachiHachi()
	cfg := domain.DefaultHachiHachiConfig()
	cfg.CpuDifficulty = diff
	g.SetConfig(cfg)
	g.Reset()
	return g
}

func TestHachiHachiReset_Deal(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	assert.Equal(t, domain.HachiHachiPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, domain.HachiHachiFieldSize, len(g.GetFieldCards()))
	assert.Equal(t, 3, g.GetPlayerCnt())
	total := len(g.GetFieldCards()) + g.GetRemainingDeck()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.HachiHachiHandSize, g.GetPlayer(i).GetCardsSize())
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 48, total)
	assert.Equal(t, 21, g.GetRemainingDeck())
}

func TestHachiHachiPlayerPlay_Errors(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	assert.Error(t, g.PlayerPlay(99, -1)) // out of range
	g.SetPhase(domain.HachiHachiPhaseRoundEnd)
	assert.Error(t, g.PlayerPlay(0, -1)) // wrong phase
	g.SetPhase(domain.HachiHachiPhasePlay)
	g.SetCurrentTurn(1)
	assert.ErrorIs(t, g.PlayerPlay(0, -1), domain.ErrNotHumanTurn)
}

func TestHachiHachiPlayerPlay_InvalidFieldChoice(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(hachihachiCard(1, 1))
	g.SetFieldCards([]*domain.Card{hachihachiCard(5, 1)})
	assert.Error(t, g.PlayerPlay(0, 0)) // chosen field month 5 != played month 1
}

func TestHachiHachiPlayerPlay_GameEnded(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	cfg := domain.DefaultHachiHachiConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	hachihachiDrive(t, g)
	assert.ErrorIs(t, g.PlayerPlay(0, -1), domain.ErrGameEnded)
}

func TestHachiHachiCapture(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(hachihachiCard(3, 3)) // 桜 カス
	g.SetFieldCards([]*domain.Card{hachihachiCard(3, 4), hachihachiCard(7, 3)})
	require.NoError(t, g.PlayerPlay(0, -1))
	assert.GreaterOrEqual(t, human.CapturedCount(), 2)
}

func TestHachiHachiNextRound_GuardedToRoundEnd(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	before := g.GetRoundNumber()
	g.NextRound() // in Play phase → no-op
	assert.Equal(t, before, g.GetRoundNumber())
}

func TestHachiHachiHint(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.GreaterOrEqual(t, h.CardIndex, 0)
	// CPU の手番・別フェーズではヒント無し。
	g.SetCurrentTurn(1)
	assert.Nil(t, g.GetHint())
	g.SetCurrentTurn(0)
	g.SetPhase(domain.HachiHachiPhaseRoundEnd)
	assert.Nil(t, g.GetHint())
}

// hachihachiDrive はドメインだけで CPU/人間を駆動し終局まで進める。
func hachihachiDrive(t *testing.T, g *domain.HachiHachi) {
	t.Helper()
	for step := 0; step < 20000 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.HachiHachiPhasePlay:
			if g.IsHumanTurn() {
				require.NoError(t, g.PlayerPlay(0, -1))
			} else {
				g.CpuPlay()
			}
		case domain.HachiHachiPhaseRoundEnd:
			g.NextRound()
		case domain.HachiHachiPhaseGameEnd:
			return
		}
	}
}

func TestHachiHachiFullGame_Normal(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	hachihachiDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.HachiHachiPhaseGameEnd, g.GetPhase())
	// 累計得点の 3 人合計は素点部分がゼロ和 (出来役ボーナス分だけプラスになり得る)。
	sum := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		sum += g.GetPlayer(i).GetScore()
	}
	assert.GreaterOrEqual(t, sum, 0)
}

func TestHachiHachiFullGame_Easy(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyEasy)
	hachihachiDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
}

func TestHachiHachiFullGame_Hard(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyHard)
	hachihachiDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
	assert.NotEqual(t, domain.HachiHachiResultNone, g.GetResult())
}

// --- endRound 精算 ---

func TestHachiHachiEndRound_SweepAndScore(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	cfg := domain.DefaultHachiHachiConfig()
	cfg.TargetRounds = 3
	g.SetConfig(cfg)
	g.Reset()

	// 1 ラウンドを実際に完走させて精算を発生させる。
	for step := 0; step < 20000 && g.GetPhase() == domain.HachiHachiPhasePlay; step++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerPlay(0, -1))
		} else {
			g.CpuPlay()
		}
	}
	require.Equal(t, domain.HachiHachiPhaseRoundEnd, g.GetPhase())

	res := g.GetLastRoundResult()
	require.NotNil(t, res)
	require.Len(t, res.Scores, 3)

	// 場札の掃き寄せにより全 48 枚が分配され、素点合計は 264、素点精算はゼロ和。
	rawSum, deltaByRaw := 0, 0
	for i, s := range res.Scores {
		rawSum += s.RawScore
		deltaByRaw += s.RawScore - domain.HachiHachiBaseline
		// 差分 = 素点 + 役ボーナス − 88。
		assert.Equal(t, s.RawScore+s.Bonus-domain.HachiHachiBaseline, s.Delta)
		// roundDelta がプレイヤーへ反映されている。
		assert.Equal(t, s.Delta, g.GetPlayer(i).GetRoundDelta())
	}
	assert.Equal(t, 264, rawSum)
	assert.Equal(t, 0, deltaByRaw, "素点 − 88 の合計はゼロ和")
	assert.Equal(t, 0, len(g.GetFieldCards()), "掃き寄せ後は場札が空")
}

func TestHachiHachiEndRound_FinishesGameAtTargetRounds(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	cfg := domain.DefaultHachiHachiConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	hachihachiDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.HachiHachiPhaseGameEnd, g.GetPhase())
	assert.False(t, g.IsHumanTurn())
}

func TestHachiHachiNextRound_Advances(t *testing.T) {
	g := domain.NewDefaultHachiHachi()
	cfg := domain.DefaultHachiHachiConfig()
	cfg.TargetRounds = 3
	g.SetConfig(cfg)
	g.Reset()
	// 1 ラウンド完走 → RoundEnd。
	for step := 0; step < 20000 && g.GetPhase() == domain.HachiHachiPhasePlay; step++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerPlay(0, -1))
		} else {
			g.CpuPlay()
		}
	}
	require.Equal(t, domain.HachiHachiPhaseRoundEnd, g.GetPhase())
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.HachiHachiPhasePlay, g.GetPhase())
}

func TestHachiHachiCpuPlay_TwoMatchFieldChoice(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(1)
	cpu := g.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(hachihachiCard(3, 1)) // 桜 光 (month 3)
	g.SetFieldCards([]*domain.Card{hachihachiCard(3, 2), hachihachiCard(3, 3), hachihachiCard(7, 4)})
	g.CpuPlay()
	assert.GreaterOrEqual(t, cpu.CapturedCount(), 2)
}

func TestHachiHachiCpuPlay_ThreeMatchSweep(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyHard)
	g.SetCurrentTurn(1)
	cpu := g.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(hachihachiCard(5, 1))
	g.SetFieldCards([]*domain.Card{hachihachiCard(5, 2), hachihachiCard(5, 3), hachihachiCard(5, 4)})
	g.CpuPlay()
	assert.GreaterOrEqual(t, cpu.CapturedCount(), 4)
}

func TestHachiHachiCpuGuards(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0) // human turn
	g.CpuPlay()         // not CPU → no-op
	g.SetPhase(domain.HachiHachiPhaseRoundEnd)
	g.CpuPlay() // wrong phase → no-op
	assert.Equal(t, domain.HachiHachiPhaseRoundEnd, g.GetPhase())
}

func TestHachiHachiCaptureOptions(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(hachihachiCard(3, 1))  // matches field month 3
	human.AddCard(hachihachiCard(12, 1)) // no match
	g.SetFieldCards([]*domain.Card{hachihachiCard(3, 2), hachihachiCard(3, 3)})
	opts := g.GetCaptureOptions(0)
	assert.Contains(t, opts, 0)
	assert.NotContains(t, opts, 1)
	g.SetPhase(domain.HachiHachiPhaseRoundEnd)
	assert.Empty(t, g.GetCaptureOptions(0))
	g.SetPhase(domain.HachiHachiPhasePlay)
	assert.Empty(t, g.GetCaptureOptions(99))
}

func TestHachiHachiPlayableIndices(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	assert.Len(t, g.GetPlayableIndices(0), domain.HachiHachiHandSize)
	assert.Nil(t, g.GetPlayableIndices(99))
	g.SetPhase(domain.HachiHachiPhaseRoundEnd)
	assert.Nil(t, g.GetPlayableIndices(0))
}

func TestHachiHachiAccessors(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	y, raw := g.GetYaku(99)
	assert.Nil(t, y)
	assert.Equal(t, 0, raw)
	assert.Equal(t, domain.HachiHachiCpuDifficultyNormal, g.GetConfig().CpuDifficulty)
	assert.NotNil(t, g.GetActionLog())
	assert.Equal(t, -1, g.GetWinner())
	assert.Equal(t, domain.HachiHachiResultNone, g.GetResult())
}

func TestHachiHachiCardAccessors(t *testing.T) {
	assert.NotEmpty(t, domain.HachiHachiCardGlyph(hhCrane))
	assert.Equal(t, domain.HachiHachiBright, domain.HachiHachiCardCategory(hhCrane))
	assert.Equal(t, domain.HachiHachiRibbonBlue, domain.HachiHachiCardRibbonColor(hachihachiCard(6, 2)))
	assert.NotEmpty(t, domain.HachiHachiCardLabel(hhCrane))
	assert.Equal(t, domain.HachiHachiChaff, domain.HachiHachiCardCategory(nil))
	assert.Equal(t, "??", domain.HachiHachiCardLabel(nil))
}

func TestHachiHachiIsHumanTurn(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentTurn(1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentTurn(0)
	g.SetPhase(domain.HachiHachiPhaseRoundEnd)
	assert.False(t, g.IsHumanTurn())
}

// --- Config ---

func TestHachiHachiConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultHachiHachiConfig().Validate())
	assert.Error(t, domain.HachiHachiConfig{CpuDifficulty: 99, TargetRounds: 3}.Validate())
	assert.Error(t, domain.HachiHachiConfig{CpuDifficulty: 0, TargetRounds: 0}.Validate())
	assert.Error(t, domain.HachiHachiConfig{CpuDifficulty: 0, TargetRounds: 99999}.Validate())
}

// --- JSON ---

func TestHachiHachiJSON_RoundTrip(t *testing.T) {
	g := newTestHachiHachi(t, domain.HachiHachiCpuDifficultyNormal)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.HachiHachi
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, len(g.GetFieldCards()), len(restored.GetFieldCards()))
	assert.Equal(t, g.GetRemainingDeck(), restored.GetRemainingDeck())
}

func TestHachiHachiJSON_Invalid(t *testing.T) {
	var g domain.HachiHachi
	assert.Error(t, json.Unmarshal([]byte("not json"), &g))
	// invalid player count
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[],"cf":{"cd":0,"tr":3},"ph":0,"ct":0,"lc":-1,"wn":-1}`), &g))
	// invalid phase
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[{"gp":{}},{"gp":{}},{"gp":{}}],"cf":{"cd":0,"tr":3},"ph":9,"ct":0,"lc":-1,"wn":-1}`), &g))
	// out-of-range card in field
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[{"gp":{}},{"gp":{}},{"gp":{}}],"cf":{"cd":0,"tr":3},"ph":0,"ct":0,"lc":-1,"wn":-1,"fd":[{"d":99,"v":1}]}`), &g))
}

func TestHachiHachiJSON_MoreInvalid(t *testing.T) {
	valid3 := `"pl":[{"gp":{}},{"gp":{}},{"gp":{}}]`
	base := func(tail string) string {
		return `{` + valid3 + `,"cf":{"cd":0,"tr":3},"ph":0,"ct":0,"lc":-1,"wn":-1` + tail + `}`
	}
	cases := []string{
		// invalid config (bad difficulty)
		`{` + valid3 + `,"cf":{"cd":99,"tr":3},"ph":0,"ct":0,"lc":-1,"wn":-1}`,
		// nil player
		`{"pl":[null,null,null],"cf":{"cd":0,"tr":3},"ph":0,"ct":0,"lc":-1,"wn":-1}`,
		// current turn out of range
		`{` + valid3 + `,"cf":{"cd":0,"tr":3},"ph":0,"ct":9,"lc":-1,"wn":-1}`,
		// last capturer out of range
		`{` + valid3 + `,"cf":{"cd":0,"tr":3},"ph":0,"ct":0,"lc":9,"wn":-1}`,
		// winner out of range
		`{` + valid3 + `,"cf":{"cd":0,"tr":3},"ph":0,"ct":0,"lc":-1,"wn":9}`,
		// bad card in draw pile
		base(`,"dp":[{"d":0,"v":0}]`),
		// nil card in field
		base(`,"fd":[null]`),
	}
	for i, c := range cases {
		var g domain.HachiHachi
		assert.Errorf(t, json.Unmarshal([]byte(c), &g), "case %d should error", i)
	}

	// oversized slice (actionLog > hachihachiMaxSliceLen)
	var sb strings.Builder
	sb.WriteString(`{` + valid3 + `,"cf":{"cd":0,"tr":3},"ph":0,"ct":0,"lc":-1,"wn":-1,"al":[`)
	for i := 0; i < 1001; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(`{}`)
	}
	sb.WriteString(`]}`)
	var g domain.HachiHachi
	assert.Error(t, json.Unmarshal([]byte(sb.String()), &g))
}

func TestHachiHachiPlayerJSON(t *testing.T) {
	p := domain.NewHachiHachiPlayer(true)
	p.AddCaptured([]*domain.Card{hhCrane})
	p.AddScore(-7)
	p.SetRoundDelta(5)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored domain.HachiHachiPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, -7, restored.GetScore())
	assert.Equal(t, 5, restored.GetRoundDelta())
	assert.Equal(t, 1, restored.CapturedCount())
	restored.ResetCaptured()
	assert.Equal(t, 0, restored.CapturedCount())
}
