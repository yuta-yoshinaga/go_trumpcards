//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// koikoiCard は Koi-Koi テスト用のカード生成ショートカット (design=月, value=index)。
func koikoiCard(month, index int) *domain.Card { return domain.NewCard(month, index, false) }

// koikoiYakuOf は取り札集合に対する成立役をキー→点数のマップと合計点で返す。
// 非公開の koikoiEvaluateYaku を GetYaku 経由で駆動する。
func koikoiYakuOf(cards []*domain.Card) (map[string]int, int) {
	human := domain.NewKoiKoiPlayer(true)
	human.AddCaptured(cards)
	cpu := domain.NewKoiKoiPlayer(false)
	g := domain.NewKoiKoi([]*domain.KoiKoiPlayer{human, cpu}, domain.DefaultKoiKoiConfig())
	yakus, total := g.GetYaku(0)
	m := make(map[string]int)
	for _, y := range yakus {
		m[y.Key] = y.Points
	}
	return m, total
}

// 各カードの (月, index) 定義 (KoiKoi.go の対応表と一致)。
var (
	koikoiCrane     = koikoiCard(1, 1)  // 松 光
	koikoiCurtain   = koikoiCard(3, 1)  // 桜 光
	koikoiMoon      = koikoiCard(8, 1)  // 芒 光
	koikoiRainman   = koikoiCard(11, 1) // 柳 光 (雨)
	koikoiPhoenix   = koikoiCard(12, 1) // 桐 光
	koikoiBoar      = koikoiCard(7, 1)  // 萩 タネ
	koikoiDeer      = koikoiCard(10, 1) // 紅葉 タネ
	koikoiButterfly = koikoiCard(6, 1)  // 牡丹 タネ
	koikoiSakeCup   = koikoiCard(9, 1)  // 菊 タネ (盃)
)

func TestKoiKoiYaku_Goko(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiCrane, koikoiCurtain, koikoiMoon, koikoiRainman, koikoiPhoenix})
	assert.Equal(t, 10, m["goko"])
	// 五光には桜(幕)と芒(月)が含まれるが盃が無いので花見/月見は成立しない。
	assert.Equal(t, 10, total)
}

func TestKoiKoiYaku_Shiko(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiCrane, koikoiCurtain, koikoiMoon, koikoiPhoenix})
	assert.Equal(t, 8, m["shiko"])
	assert.NotContains(t, m, "goko")
	assert.Equal(t, 8, total)
}

func TestKoiKoiYaku_AmeShiko(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiCrane, koikoiCurtain, koikoiRainman, koikoiPhoenix})
	assert.Equal(t, 7, m["ameshiko"])
	assert.NotContains(t, m, "shiko")
	assert.Equal(t, 7, total)
}

func TestKoiKoiYaku_Sanko(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiCrane, koikoiCurtain, koikoiPhoenix})
	assert.Equal(t, 5, m["sanko"])
	assert.Equal(t, 5, total)
}

func TestKoiKoiYaku_SankoBlockedByRain(t *testing.T) {
	// 雨を含む 3 光は三光にならない。
	m, total := koikoiYakuOf([]*domain.Card{koikoiCrane, koikoiRainman, koikoiPhoenix})
	assert.NotContains(t, m, "sanko")
	assert.Equal(t, 0, total)
}

func TestKoiKoiYaku_Inoshikacho(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiBoar, koikoiDeer, koikoiButterfly})
	assert.Equal(t, 5, m["inoshikacho"])
	assert.NotContains(t, m, "tane") // 3 枚では タネ 未成立
	assert.Equal(t, 5, total)
}

func TestKoiKoiYaku_Akatan(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiCard(1, 2), koikoiCard(2, 2), koikoiCard(3, 2)})
	assert.Equal(t, 5, m["akatan"])
	assert.NotContains(t, m, "tanzaku") // 3 枚では 短冊 未成立
	assert.Equal(t, 5, total)
}

func TestKoiKoiYaku_Aotan(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiCard(6, 2), koikoiCard(9, 2), koikoiCard(10, 2)})
	assert.Equal(t, 5, m["aotan"])
	assert.Equal(t, 5, total)
}

func TestKoiKoiYaku_Tanzaku(t *testing.T) {
	// 5 短冊 (赤/青の 3 枚役は避ける)。
	m, total := koikoiYakuOf([]*domain.Card{
		koikoiCard(1, 2), koikoiCard(4, 2), koikoiCard(5, 2), koikoiCard(7, 2), koikoiCard(11, 3),
	})
	assert.Equal(t, 1, m["tanzaku"])
	assert.NotContains(t, m, "akatan")
	assert.NotContains(t, m, "aotan")
	assert.Equal(t, 1, total)
}

func TestKoiKoiYaku_TanzakuExtra(t *testing.T) {
	// 6 短冊 → 2 点 (赤短は 2 枚に留めて役化を避ける)。
	m, _ := koikoiYakuOf([]*domain.Card{
		koikoiCard(1, 2), koikoiCard(2, 2), koikoiCard(4, 2), koikoiCard(5, 2), koikoiCard(7, 2), koikoiCard(11, 3),
	})
	assert.Equal(t, 2, m["tanzaku"])
	assert.NotContains(t, m, "akatan")
}

func TestKoiKoiYaku_Tane(t *testing.T) {
	// 5 タネ (猪鹿蝶を避ける)。
	m, total := koikoiYakuOf([]*domain.Card{
		koikoiCard(2, 1), koikoiCard(4, 1), koikoiCard(5, 1), koikoiCard(8, 2), koikoiCard(11, 2),
	})
	assert.Equal(t, 1, m["tane"])
	assert.NotContains(t, m, "inoshikacho")
	assert.Equal(t, 1, total)
}

func TestKoiKoiYaku_TaneExtra(t *testing.T) {
	m, _ := koikoiYakuOf([]*domain.Card{
		koikoiCard(2, 1), koikoiCard(4, 1), koikoiCard(5, 1), koikoiCard(8, 2), koikoiCard(11, 2), koikoiSakeCup,
	})
	assert.Equal(t, 2, m["tane"])
}

func TestKoiKoiYaku_Kasu(t *testing.T) {
	cards := []*domain.Card{
		koikoiCard(1, 3), koikoiCard(1, 4), koikoiCard(2, 3), koikoiCard(2, 4), koikoiCard(3, 3),
		koikoiCard(3, 4), koikoiCard(4, 3), koikoiCard(4, 4), koikoiCard(5, 3), koikoiCard(5, 4),
	}
	m, total := koikoiYakuOf(cards)
	assert.Equal(t, 1, m["kasu"])
	assert.Equal(t, 1, total)

	// +1 枚で 2 点。
	m2, _ := koikoiYakuOf(append(cards, koikoiCard(6, 3)))
	assert.Equal(t, 2, m2["kasu"])
}

func TestKoiKoiYaku_Hanami(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiCurtain, koikoiSakeCup})
	assert.Equal(t, 5, m["hanami"])
	assert.Equal(t, 5, total)
}

func TestKoiKoiYaku_Tsukimi(t *testing.T) {
	m, total := koikoiYakuOf([]*domain.Card{koikoiMoon, koikoiSakeCup})
	assert.Equal(t, 5, m["tsukimi"])
	assert.Equal(t, 5, total)
}

func TestKoiKoiYaku_OverlapGokoWithSake(t *testing.T) {
	// 五光 + 盃 → 花見酒 + 月見酒 も成立 (桜/芒が五光に含まれるため)。
	m, total := koikoiYakuOf([]*domain.Card{
		koikoiCrane, koikoiCurtain, koikoiMoon, koikoiRainman, koikoiPhoenix, koikoiSakeCup,
	})
	assert.Equal(t, 10, m["goko"])
	assert.Equal(t, 5, m["hanami"])
	assert.Equal(t, 5, m["tsukimi"])
	assert.Equal(t, 20, total)
}

func TestKoiKoiYaku_Empty(t *testing.T) {
	m, total := koikoiYakuOf(nil)
	assert.Empty(t, m)
	assert.Equal(t, 0, total)
}

// --- ゲーム進行 ---

func newTestKoiKoi(t *testing.T, diff domain.KoiKoiCpuDifficulty) *domain.KoiKoi {
	t.Helper()
	g := domain.NewDefaultKoiKoi()
	cfg := domain.DefaultKoiKoiConfig()
	cfg.CpuDifficulty = diff
	g.SetConfig(cfg)
	g.Reset()
	return g
}

func TestKoiKoiReset_Deal(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	assert.Equal(t, domain.KoiKoiPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, domain.KoiKoiFieldSize, len(g.GetFieldCards()))
	total := len(g.GetFieldCards()) + g.GetRemainingDeck()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.KoiKoiHandSize, g.GetPlayer(i).GetCardsSize())
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 48, total)
	assert.Equal(t, 24, g.GetRemainingDeck())
}

func TestKoiKoiPlayerPlay_Errors(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	// out of range
	assert.Error(t, g.PlayerPlay(99, -1))
	// wrong phase
	g.SetPhase(domain.KoiKoiPhaseRoundEnd)
	assert.Error(t, g.PlayerPlay(0, -1))
	g.SetPhase(domain.KoiKoiPhasePlay)
	// not human turn
	g.SetCurrentTurn(1)
	assert.ErrorIs(t, g.PlayerPlay(0, -1), domain.ErrNotHumanTurn)
}

func TestKoiKoiPlayerPlay_InvalidFieldChoice(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	// human hand: single card month 1; field has no month-1 card → fieldIdx 0 mismatch.
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(koikoiCard(1, 1))
	g.SetFieldCards([]*domain.Card{koikoiCard(5, 1)})
	assert.Error(t, g.PlayerPlay(0, 0)) // chosen field card month 5 != played month 1
}

func TestKoiKoiCapture(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(koikoiCard(3, 3)) // 桜 カス
	g.SetFieldCards([]*domain.Card{koikoiCard(3, 4), koikoiCard(7, 3)})
	require.NoError(t, g.PlayerPlay(0, -1))
	// 出した札 + 同月の (3,4) を捕獲 (めくり札で更に増える可能性あり)。
	assert.GreaterOrEqual(t, human.CapturedCount(), 2)
}

func TestKoiKoiDecide_Errors(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	// wrong phase (Play, not Decision)
	assert.Error(t, g.PlayerDecide(true))
}

func TestKoiKoiNextRound_GuardedToRoundEnd(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	before := g.GetRoundNumber()
	g.NextRound() // in Play phase → no-op
	assert.Equal(t, before, g.GetRoundNumber())
}

func TestKoiKoiHint(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.GreaterOrEqual(t, h.CardIndex, 0)
	assert.Equal(t, -1, h.KoiKoi) // play phase → not a decision hint

	// decision phase hint
	g.SetPhase(domain.KoiKoiPhaseKoiKoiDecision)
	h2 := g.GetHint()
	require.NotNil(t, h2)
	assert.NotEqual(t, -1, h2.KoiKoi)
}

// koikoiDrive はドメインだけで CPU/人間を駆動し終局まで進める。
func koikoiDrive(t *testing.T, g *domain.KoiKoi) {
	t.Helper()
	for step := 0; step < 20000 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.KoiKoiPhasePlay:
			if g.IsHumanTurn() {
				require.NoError(t, g.PlayerPlay(0, -1))
			} else {
				g.CpuPlay()
			}
		case domain.KoiKoiPhaseKoiKoiDecision:
			if g.IsHumanTurn() {
				require.NoError(t, g.PlayerDecide(false)) // 人間は常に勝負
			} else {
				g.CpuDecide()
			}
		case domain.KoiKoiPhaseRoundEnd:
			g.NextRound()
		case domain.KoiKoiPhaseGameEnd:
			return
		}
	}
}

func TestKoiKoiFullGame_Normal(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	koikoiDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.KoiKoiPhaseGameEnd, g.GetPhase())
}

func TestKoiKoiFullGame_Hard(t *testing.T) {
	// Hard は積極的にこいこいを宣言する → こいこい経路と倍率を通す。
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyHard)
	koikoiDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetPlayer(0).GetScore(), 0)
	assert.GreaterOrEqual(t, g.GetPlayer(1).GetScore(), 0)
}

func TestKoiKoiFullGame_Easy(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyEasy)
	koikoiDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
}

// --- Config ---

func TestKoiKoiConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultKoiKoiConfig().Validate())
	assert.Error(t, domain.KoiKoiConfig{CpuDifficulty: 99, TargetScore: 15}.Validate())
	assert.Error(t, domain.KoiKoiConfig{CpuDifficulty: 0, TargetScore: 0}.Validate())
	assert.Error(t, domain.KoiKoiConfig{CpuDifficulty: 0, TargetScore: 99999}.Validate())
}

// --- JSON ---

func TestKoiKoiJSON_RoundTrip(t *testing.T) {
	g := newTestKoiKoi(t, domain.KoiKoiCpuDifficultyNormal)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.KoiKoi
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, len(g.GetFieldCards()), len(restored.GetFieldCards()))
	assert.Equal(t, g.GetRemainingDeck(), restored.GetRemainingDeck())
}

func TestKoiKoiJSON_Invalid(t *testing.T) {
	var g domain.KoiKoi
	assert.Error(t, json.Unmarshal([]byte("not json"), &g))
	// invalid player count
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[],"cf":{"cd":0,"ts":15},"ph":0,"ct":0,"rw":-1,"wn":-1}`), &g))
	// invalid phase
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[{"gp":{}},{"gp":{}}],"cf":{"cd":0,"ts":15},"ph":9,"ct":0,"rw":-1,"wn":-1}`), &g))
	// out-of-range card in field
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[{"gp":{}},{"gp":{}}],"cf":{"cd":0,"ts":15},"ph":0,"ct":0,"rw":-1,"wn":-1,"fd":[{"d":99,"v":1}]}`), &g))
}

func TestKoiKoiPlayerJSON(t *testing.T) {
	p := domain.NewKoiKoiPlayer(true)
	p.AddCaptured([]*domain.Card{koikoiCrane})
	p.AddScore(7)
	p.SetCalledKoiKoi(true)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored domain.KoiKoiPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 7, restored.GetScore())
	assert.True(t, restored.GetCalledKoiKoi())
	assert.Equal(t, 1, restored.CapturedCount())
}
