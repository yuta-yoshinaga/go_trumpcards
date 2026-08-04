//go:build test

package presenter_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func cegoCuiGame() *domain.Cego {
	g := domain.NewDefaultCego()
	g.Reset()
	return g
}

func TestCegoCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CegoCuiPresenter)

	t.Run("bid phase", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetBidPlayerIdx(0)
		result := p.Output(g, nil)
		assert.Contains(t, result, "チェゴ") // helpTitle
	})

	t.Run("contract phase", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhaseContract)
		result := p.Output(g, nil)
		assert.Contains(t, result, "コントラクト")
	})

	// **コマンド名だけでは選べない。**Web には契約ごとのリスク・リターンを
	// 説明する箱があるのに、CUI は構文リマインダーしか出していなかった (#4931)。
	t.Run("contract phase explains both contracts", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhaseContract)
		result := p.Output(g, nil)

		assert.Contains(t, result, "契約の選び方")
		// 盲札の枚数が説明に入る (受け入れ条件2)。
		assert.Contains(t, result, strconv.Itoa(g.GetBlindCount())+"枚の盲札")
		assert.Positive(t, g.GetBlindCount(), "盲札が 0 枚だと枚数の検査が意味を失う")
		assert.Contains(t, result, "大量得点")
		assert.Contains(t, result, "堅実")
		// 既存の構文リマインダーは残す。
		assert.Contains(t, result, "cego ... 場札と交換")
	})

	// 契約フェーズ以外では出さない。毎画面に出ると邪魔になる。
	t.Run("the contract explanation is confined to the contract phase", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetContractType(domain.CegoContractCego)
		g.SetPhase(domain.CegoPhaseExchange)
		assert.NotContains(t, p.Output(g, nil), "契約の選び方")
	})

	// ja / en 双方に説明文があり、かつ別文字列であること。
	t.Run("contract explanation is translated in both languages", func(t *testing.T) {
		defer i18n.SetLang("ja")
		for _, key := range []string{"cego.contractExplainTitle", "cego.contractCegoDesc", "cego.contractHandspielDesc"} {
			i18n.SetLang("ja")
			ja := i18n.T(key)
			assert.NotEqual(t, key, ja, "%s missing from ja", key)
			i18n.SetLang("en")
			en := i18n.T(key)
			assert.NotEqual(t, key, en, "%s missing from en", key)
			assert.NotEqual(t, ja, en, "%s is the same in both languages", key)
		}
	})

	t.Run("exchange phase", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetContractType(domain.CegoContractCego)
		g.SetPhase(domain.CegoPhaseExchange)
		result := p.Output(g, nil)
		assert.Contains(t, result, "場札交換")
	})

	t.Run("play phase renders trump and skus", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhasePlay)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CegoTrumpDesign, 7, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CegoSkusDesign, 0, false))
		g.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		result := p.Output(g, nil)
		assert.Contains(t, result, "T7")
		assert.Contains(t, result, "Sküs")
	})

	// declarerPrefix is the leading literal of the declarer line (before the
	// {{name}} placeholder), stable enough to assert on under the default locale.
	declarerPrefix := strings.SplitN(i18n.T("cego.declarerLine"), "{{", 2)[0]

	t.Run("declarer line shown once declarer is set", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhasePlay)
		result := p.Output(g, nil)
		assert.Contains(t, result, declarerPrefix)
	})

	t.Run("declarer line hidden before declarer is set", func(t *testing.T) {
		g := cegoCuiGame() // default declarer idx is -1 during bidding
		g.SetBidPlayerIdx(0)
		result := p.Output(g, nil)
		assert.NotContains(t, result, declarerPrefix)
	})

	t.Run("trick end", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetPhase(domain.CegoPhaseTrickEnd)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("round end", func(t *testing.T) {
		g := cegoCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.CegoBidPlay)
		g.SetPhase(domain.CegoPhaseRoundEnd)
		assert.NotEmpty(t, p.Output(g, nil))
	})
}

func TestCegoCuiPresenter_GameEnd(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CegoCuiPresenter)

	t.Run("solo winner", func(t *testing.T) {
		g := cegoDriveToGameEnd([domain.CegoPlayerCnt]int{0, 100, 0, 0}) // CPU 1 wins
		assert.True(t, g.GetGameEndFlag())
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("draw", func(t *testing.T) {
		g := cegoDriveToGameEnd([domain.CegoPlayerCnt]int{40, 0, 0, 0}) // tie -> no winner name
		assert.Equal(t, -1, g.GetWinnerPlayer())
		assert.NotEmpty(t, p.Output(g, nil))
	})
}

func TestCegoCuiPresenter_HintAndLog(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CegoCuiPresenter)

	g := cegoCuiGame()
	g.SetBidPlayerIdx(0)
	assert.NotEmpty(t, p.HintOutput(g))
	assert.NotEmpty(t, p.ActionLogOutput(g))

	// Contract-phase hint (recommends Cego/Handspiel).
	g2 := cegoCuiGame()
	g2.SetDeclarerIdx(0)
	g2.SetContract(domain.CegoBidPlay)
	g2.SetPhase(domain.CegoPhaseContract)
	assert.NotEmpty(t, p.HintOutput(g2))
}
