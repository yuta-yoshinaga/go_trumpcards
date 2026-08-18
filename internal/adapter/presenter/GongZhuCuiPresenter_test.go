package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupGongZhuCuiMock() *interfaces.MockGongZhuGame {
	m := new(interfaces.MockGongZhuGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetExposure").Return(domain.GongZhuExposure{})
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GongZhuPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// 得点内訳の既定値 (#5630)。中身を見るテストは自分で上書きする。
	m.On("ScoreBreakdownFor", mock.Anything).Return(domain.GongZhuScoreBreakdown{}).Maybe()
	return m
}

func setupGongZhuCuiMockWithPlayers() (*interfaces.MockGongZhuGame, []*domain.GongZhuPlayer) {
	m := setupGongZhuCuiMock()
	players := makeGongZhuPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestGongZhuCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GongZhuCuiPresenter)

	t.Run("play phase shows player info and prompt", func(t *testing.T) {
		m, players := setupGongZhuCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Gong Zhu")
		assert.NotEmpty(t, result)
	})

	t.Run("captured point cards are listed per player", func(t *testing.T) {
		m, players := setupGongZhuCuiMockWithPlayers()
		// CPU 1 took a trick containing the pig (♠Q), a heart, and a plain card.
		players[1].AddTrick([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 12, false),  // pig -> shown
			domain.NewCard(domain.CardDesignHeart, 5, false),   // heart -> shown
			domain.NewCard(domain.CardDesignClover, 3, false),  // plain -> not shown
			domain.NewCard(domain.CardDesignDiamond, 8, false), // plain -> not shown
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "獲得: SPADE 12 HEART 5")
		// Only the one capturing player gets a line; the others took nothing.
		assert.Equal(t, 1, strings.Count(result, "獲得:"))
	})

	t.Run("expose phase prompt", func(t *testing.T) {
		m, _ := setupGongZhuCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetExposure")
		m.On("GetPhase").Return(domain.GongZhuPhaseExpose)
		m.On("GetExposure").Return(domain.GongZhuExposure{Pig: true, Sheep: true, Ace: true, Doubler: true})
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// Exposure summary uses the localized point-card symbol keys for every exposed card.
		assert.Contains(t, result, i18n.T("gongzhu.card.spadeQueen"))
		assert.Contains(t, result, i18n.T("gongzhu.card.diamondJack"))
		assert.Contains(t, result, i18n.T("gongzhu.card.heartAce"))
		assert.Contains(t, result, i18n.T("gongzhu.card.clubTen"))
	})

	t.Run("trick end and round end prompts", func(t *testing.T) {
		for _, phase := range []domain.GongZhuPhase{domain.GongZhuPhaseTrickEnd, domain.GongZhuPhaseRoundEnd} {
			m, _ := setupGongZhuCuiMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(phase)
			assert.NotEmpty(t, p.Output(m, nil))
		}
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupGongZhuCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})
}

func TestGongZhuCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.GongZhuCuiPresenter)

	t.Run("expose hint", func(t *testing.T) {
		m, players := setupGongZhuCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))
		m.On("GetHint").Return(&domain.GongZhuHint{CardIndices: []int{0}, Reason: "expose_sheep"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("expose none hint (empty indices)", func(t *testing.T) {
		m, _ := setupGongZhuCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.GongZhuHint{CardIndices: []int{}, Reason: "expose_none"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupGongZhuCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.GongZhuHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestGongZhuCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GongZhuCuiPresenter)
	m := new(interfaces.MockGongZhuGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠5"},
	})
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

// #5630: 得点は「ハート合計 → 全ハートボーナス → ♥A で倍 → 猪/羊 → 猪抜きの倍率」と
// 何段も重なるのに、画面には最終値しか出ず、なぜその点なのかを確かめる手段が無かった。
func TestGongZhuCuiPresenterShowsTheRoundBreakdown(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	m, _ := setupGongZhuCuiMockWithPlayers()
	m.ExpectedCalls = filterGongZhuCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.GongZhuPhaseRoundEnd)
	// 既定 (共有ヘルパの .Maybe()) を先に消す。testify は最初に一致した期待値を
	// 返すので、消さずに足すと上書きしたつもりのケースが何も確かめない。
	m.ExpectedCalls = filterGongZhuCall(m.ExpectedCalls, "ScoreBreakdownFor")
	m.On("ScoreBreakdownFor", mock.Anything).Return(domain.GongZhuScoreBreakdown{
		HeartCount: 3, HeartsSum: -120, AceExposed: true,
		HasPig: true, PigExposed: true,
		HasSheep:   true,
		HasDoubler: true, DoublerMultiplier: 2,
		Subtotal: -220, Total: -440,
	})

	out := new(presenter.GongZhuCuiPresenter).Output(m, nil)

	// 各段が個別に読める。合計だけでは検算できない。
	assert.Contains(t, out, i18n.Tf("gongzhu.breakdownHearts", "count", "3", "sum", "-120"))
	assert.Contains(t, out, i18n.T("gongzhu.breakdownPigExposed"))
	assert.Contains(t, out, i18n.T("gongzhu.breakdownSheep"))
	assert.Contains(t, out, i18n.Tf("gongzhu.breakdownDoubler", "mult", "2"))
	assert.Contains(t, out, i18n.Tf("gongzhu.breakdownTotal", "total", "-440"))
}

// 起きていないことは書かない。猪を取っていないのに猪の行が出ると、
// 何が起きたのか読めなくなる。
func TestGongZhuCuiPresenterOmitsTheItemsThatDidNotHappen(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	m, _ := setupGongZhuCuiMockWithPlayers()
	m.ExpectedCalls = filterGongZhuCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.GongZhuPhaseRoundEnd)
	// 既定 (共有ヘルパの .Maybe()) を先に消す。testify は最初に一致した期待値を
	// 返すので、消さずに足すと上書きしたつもりのケースが何も確かめない。
	m.ExpectedCalls = filterGongZhuCall(m.ExpectedCalls, "ScoreBreakdownFor")
	m.On("ScoreBreakdownFor", mock.Anything).Return(domain.GongZhuScoreBreakdown{
		HeartCount: 2, HeartsSum: -90, Subtotal: -90, Total: -90,
	})

	out := new(presenter.GongZhuCuiPresenter).Output(m, nil)
	assert.NotContains(t, out, i18n.T("gongzhu.breakdownPig"))
	assert.NotContains(t, out, i18n.T("gongzhu.breakdownSheep"))
	assert.NotContains(t, out, i18n.Tf("gongzhu.breakdownDoubler", "mult", "2"))
}

// ラウンド終了以外では内訳を出さない (途中経過の点は確定していない)。
func TestGongZhuCuiPresenterOmitsTheBreakdownDuringPlay(t *testing.T) {
	m, _ := setupGongZhuCuiMockWithPlayers()
	m.On("ScoreBreakdownFor", mock.Anything).Return(domain.GongZhuScoreBreakdown{Total: -90}).Maybe()

	out := new(presenter.GongZhuCuiPresenter).Output(m, nil)
	// 見出しごと出ない。特定の数字だけ見ると、別の理由で消えていても気づけない。
	assert.NotContains(t, out, i18n.T("gongzhu.breakdownTitle"))
}

// filterGongZhuCall removes an expectation so a test can override it.
func filterGongZhuCall(calls []*mock.Call, method string) []*mock.Call {
	out := calls[:0]
	for _, c := range calls {
		if c.Method != method {
			out = append(out, c)
		}
	}
	return out
}
