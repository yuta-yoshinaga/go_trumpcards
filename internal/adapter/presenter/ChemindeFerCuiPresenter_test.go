//go:build test

package presenter

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newChemindeFerForPresenter は本物のドメインを返す。
//
// **モックではなく本物を使う。** アクセサを 30 個モックすると、返す値を自分で決めて
// しまうので「プレゼンタが盤面を正しく読めているか」を検査したことにならない。
func newChemindeFerForPresenter(t *testing.T) *domain.ChemindeFer {
	t.Helper()
	g := domain.NewDefaultChemindeFer()
	g.Reset()
	return g
}

// chemindeFerPresenterPosition は指定の合計・フェーズの局面を返す。
func chemindeFerPresenterPosition(t *testing.T, punterTotal, bankerTotal int,
	phase domain.ChemindeFerPhase,
) *domain.ChemindeFer {
	t.Helper()
	g := newChemindeFerForPresenter(t)
	g.SetupCoupForTest(punterTotal, bankerTotal, phase)
	return g
}

// **生のキーが画面に出ていないこと。**
//
// ロケールが 1 か所でもネストしていると Go 側は `map[string]string` に読めず、
// そのゲームの訳が丸ごと落ちて画面がキー名だらけになる。`Contains(out, i18n.T(k))` は
// 両辺が生キーになって素通りするので、**日本語の実文字列**で確かめる。
func TestChemindeFerCuiPresenter_RendersJapaneseNotRawKeys(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)
	g := newChemindeFerForPresenter(t)
	out := cp.Output(g, nil)

	assert.Contains(t, out, "フェーズ:")
	assert.Contains(t, out, "ラウンド:")
	assert.Contains(t, out, "親: 席")
	assert.Contains(t, out, "チップ:")
	assert.NotContains(t, out, "chemindefer.", "生の i18n キーが出力に混ざっている")
}

func TestChemindeFerCuiPresenter_ShowsBettingProgress(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)
	g := newChemindeFerForPresenter(t)
	require.NoError(t, g.StakeForTest(300)) // CPU を進めない (賭けの途中を保つ)

	out := cp.Output(g, nil)
	assert.Contains(t, out, "バンク額 300")
	assert.Contains(t, out, "賭け番")
	assert.NotContains(t, out, "chemindefer.")
}

// **選べない合計であることを画面に出す。**
//
// 黙って引かせると、なぜ手が飛んだのか読み取れない。
func TestChemindeFerCuiPresenter_MarksTheForcedTotals(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)

	forced := chemindeFerPresenterPosition(t, 3, 2, domain.ChemindeFerPhasePunterDraw)
	assert.Contains(t, cp.Output(forced, nil), "選択の余地はありません")

	free := chemindeFerPresenterPosition(t, domain.ChemindeFerPunterFreeTotal, 2,
		domain.ChemindeFerPhasePunterDraw)
	assert.NotContains(t, cp.Output(free, nil), "選択の余地はありません",
		"合計 5 は選べるのに強制と表示している")
}

func TestChemindeFerCuiPresenter_ShowsHandsAndResult(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)
	g := chemindeFerPresenterPosition(t, 7, 3, domain.ChemindeFerPhaseBankerDraw)
	require.NoError(t, g.BankerStand())

	out := cp.Output(g, nil)
	assert.Contains(t, out, "子:")
	assert.Contains(t, out, "親:")
	assert.Contains(t, out, "決着:")
	assert.Contains(t, out, "子側の勝ち")
}

// **卓の結果と自分の損益は別の情報** (#5774)。人間は席 0 で、この局面では
// 親。子側の勝ち = 自分の負け、という取り違えやすい向きをそのまま見る。
func TestChemindeFerCuiPresenter_ShowsYourOwnNet(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)

	lost := chemindeFerPresenterPosition(t, 7, 3, domain.ChemindeFerPhaseBankerDraw)
	require.Equal(t, 0, lost.GetBankerIdx(), "人間の席が親")
	// **子側が勝つとバンクは隣へ渡る。** 席番号は決着後に変わるので先に見る。
	require.NoError(t, lost.BankerStand())
	require.Negative(t, lost.GetLastNet(0))
	out := cp.Output(lost, nil)
	assert.Contains(t, out, "子側の勝ち")
	assert.Contains(t, out, i18n.Tf("chemindefer.netLossLine", "n", strconv.Itoa(-lost.GetLastNet(0))))
	assert.NotContains(t, out, fixedPart("chemindefer.netWinLine"))

	won := chemindeFerPresenterPosition(t, 3, 7, domain.ChemindeFerPhaseBankerDraw)
	require.NoError(t, won.BankerStand())
	assert.Contains(t, cp.Output(won, nil),
		i18n.Tf("chemindefer.netWinLine", "n", strconv.Itoa(won.GetLastNet(0))))

	// **賭けが動かない回も行は出す。** 消すと、勝ったのか賭けていないのかが読めない。
	flat := chemindeFerPresenterPosition(t, 5, 5, domain.ChemindeFerPhaseBankerDraw)
	require.NoError(t, flat.BankerStand())
	assert.Contains(t, cp.Output(flat, nil), i18n.T("chemindefer.netFlatLine"))
}

func TestChemindeFerCuiPresenter_ShowsErrors(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)
	g := newChemindeFerForPresenter(t)
	out := cp.Output(g, errors.New("張り額が範囲外です"))
	assert.Contains(t, out, "張り額が範囲外です")
}

func TestChemindeFerCuiPresenter_ShowsGameEnd(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)
	g := newChemindeFerForPresenter(t)
	g.GiveUp()

	out := cp.Output(g, nil)
	assert.Contains(t, out, "ゲーム終了")
}

func TestChemindeFerCuiPresenter_HintOutput(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)

	t.Run("判断どころでなければその旨", func(t *testing.T) {
		g := chemindeFerPresenterPosition(t, 3, 2, domain.ChemindeFerPhaseRoundEnd)
		assert.Equal(t, i18n.T("chemindefer.hintNone"), cp.HintOutput(g))
	})

	t.Run("親の判断には引くか立つかを薦める", func(t *testing.T) {
		g := chemindeFerPresenterPosition(t, 4, 2, domain.ChemindeFerPhaseBankerDraw)
		out := cp.HintOutput(g)
		assert.Contains(t, out, "引く")
		assert.NotContains(t, out, "chemindefer.")
	})
}

func TestChemindeFerCuiPresenter_ActionLogOutput(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)
	g := newChemindeFerForPresenter(t)
	assert.NotEmpty(t, cp.ActionLogOutput(g))
}

// **未知のフェーズでも落ちない。**
//
// 以前この検査は既定の卓 (= 必ず張り待ち) を渡していたので、phaseStr の default 分岐に
// 一度も到達せず、名前だけが「未知のフェーズ」を名乗っていた。**範囲外の値を実際に
// 入れて**初めて検査になる。
func TestChemindeFerCuiPresenter_UnknownPhase(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)
	g := chemindeFerPresenterPosition(t, 3, 2, domain.ChemindeFerPhase(99))

	out := cp.Output(g, nil)
	assert.Contains(t, out, "UNKNOWN", "範囲外のフェーズが UNKNOWN として出ていない")
	assert.NotContains(t, out, "chemindefer.")
}

// 既定の卓は張り待ちで始まる。
func TestChemindeFerCuiPresenter_StartsAtStake(t *testing.T) {
	cp := new(ChemindeFerCuiPresenter)
	out := cp.Output(newChemindeFerForPresenter(t), nil)
	assert.True(t, strings.Contains(out, "STAKE"), "既定の卓は張り待ちのはず: %s", out)
}
