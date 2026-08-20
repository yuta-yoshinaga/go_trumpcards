//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func ofcCardP(d, v int) *domain.Card { return domain.NewCard(d, v, false) }

func setupOpenFaceChineseCuiMock() *interfaces.MockOpenFaceChineseGame {
	m := new(interfaces.MockOpenFaceChineseGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetConfig").Return(domain.DefaultOpenFaceChineseConfig())
	m.On("GetPlayerCnt").Return(2)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	human := domain.NewOpenFaceChinesePlayer(true)
	human.SetPending([]*domain.Card{ofcCardP(domain.CardDesignSpade, 13)})
	human.SetFront([]*domain.Card{ofcCardP(domain.CardDesignHeart, 5)})
	cpu := domain.NewOpenFaceChinesePlayer(false)
	m.On("GetPlayer", 0).Return(human)
	m.On("GetPlayer", 1).Return(cpu)
	m.On("GetCurrentCard").Return(ofcCardP(domain.CardDesignSpade, 13))
	return m
}

func TestOpenFaceChineseCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.OpenFaceChineseCuiPresenter)

	t.Run("placing phase shows rows and prompt", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		result := p.Output(m, nil)
		// i18n is loaded (ja) in this test build → assert on rendered Japanese text.
		assert.Contains(t, result, "ラウンド")
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.OpenFaceChinesePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner with winner", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end draw", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestOpenFaceChineseCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.OpenFaceChineseCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.On("GetHint").Return((*domain.OpenFaceChineseHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("hint back row", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.On("GetHint").Return(&domain.OpenFaceChineseHint{Row: domain.OpenFaceChineseRowBack, Reason: "strong_back"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint front row", func(t *testing.T) {
		m := setupOpenFaceChineseCuiMock()
		m.On("GetHint").Return(&domain.OpenFaceChineseHint{Row: domain.OpenFaceChineseRowFront, Reason: "weak_front"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestOpenFaceChineseCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.OpenFaceChineseCuiPresenter)
	m := setupOpenFaceChineseCuiMock()
	result := p.ActionLogOutput(m)
	assert.NotNil(t, result)
}

// #5676: 反則 (front > middle または middle > back) は全段負け扱いという重い結果
// なのに、CUI は置いてラウンドが終わるまで気づけなかった。Web は各段のボタンに
// その場で警告を出している。
func TestOpenFaceChineseCuiPresenter_WarnsAboutFoulingPlacements(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.OpenFaceChineseCuiPresenter)
	card := func(d, v int) *domain.Card { return ofcCardP(d, v) }

	// middle が A ハイで埋まり、back はあと 1 枚。K ハイで埋めると middle > back。
	build := func(pending *domain.Card) *interfaces.MockOpenFaceChineseGame {
		m := new(interfaces.MockOpenFaceChineseGame)
		m.On("GetRoundNumber").Return(1)
		m.On("GetConfig").Return(domain.DefaultOpenFaceChineseConfig())
		m.On("GetPlayerCnt").Return(2)
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.OpenFaceChinesePhasePlacing)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		human := domain.NewOpenFaceChinesePlayer(true)
		human.SetMiddle([]*domain.Card{
			card(domain.CardDesignSpade, 1), card(domain.CardDesignHeart, 3),
			card(domain.CardDesignClover, 5), card(domain.CardDesignDiamond, 7),
			card(domain.CardDesignSpade, 9),
		})
		human.SetBack([]*domain.Card{
			card(domain.CardDesignHeart, 13), card(domain.CardDesignClover, 2),
			card(domain.CardDesignDiamond, 4), card(domain.CardDesignSpade, 6),
		})
		human.SetPending([]*domain.Card{pending})
		m.On("GetPlayer", 0).Return(human)
		m.On("GetPlayer", 1).Return(domain.NewOpenFaceChinesePlayer(false))
		m.On("GetCurrentCard").Return(pending)
		return m
	}

	t.Run("names the row that would foul", func(t *testing.T) {
		// ♥8 を下段に置くと K ハイで埋まり、A ハイの中段を下回る = 確定で反則。
		out := p.Output(build(card(domain.CardDesignHeart, 8)), nil)

		assert.Contains(t, out, i18n.Tf("openfacechinese.foulRiskWarning",
			"rows", i18n.T("openfacechinese.rowBack")))
	})

	// **確定しない配置では出さない。**未確定を反則と呼ぶと、まだ挽回できる手まで
	// 避けさせてしまう。
	t.Run("stays quiet when the placement keeps the order", func(t *testing.T) {
		// ♠A を下段に置けば A ペア相当で中段を上回る。
		out := p.Output(build(card(domain.CardDesignClover, 1)), nil)

		prefix, _, ok := strings.Cut(i18n.Tf("openfacechinese.foulRiskWarning", "rows", "\x00"), "\x00")
		require.True(t, ok)
		require.NotEmpty(t, strings.TrimSpace(prefix))
		assert.NotContains(t, out, prefix)
	})

	// 既存の配置プロンプトは残す。
	t.Run("keeps the placement prompt", func(t *testing.T) {
		out := p.Output(build(card(domain.CardDesignHeart, 8)), nil)

		assert.Contains(t, out, i18n.T("openfacechinese.promptPlaceHelp"))
	})
}
