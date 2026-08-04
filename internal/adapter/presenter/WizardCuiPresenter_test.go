//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupWizardCuiMock() *interfaces.MockWizardGame {
	m := new(interfaces.MockWizardGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTotalRounds").Return(15)
	m.On("GetHandSize").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.WizardPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpCard").Return(domain.NewCard(domain.CardDesignHeart, 5, false))
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetRestrictedBid").Return(-1)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultWizardConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupWizardCuiMockWithPlayers() (*interfaces.MockWizardGame, []*domain.WizardPlayer) {
	m := setupWizardCuiMock()
	players := makeWizardPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestWizardCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.WizardCuiPresenter)

	t.Run("play phase", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Wizard")
		assert.Contains(t, result, "ラウンド: 1/15")
		assert.Contains(t, result, "手札枚数: 1")
		assert.Contains(t, result, "切り札:")
		assert.Contains(t, result, "手番:")
		// Empty trick → no established lead suit yet.
		assert.Contains(t, result, i18n.T("wizard.leadNone"))
	})

	t.Run("play phase names the lead suit", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		// A Jester (skipped) then a heart establishes hearts as the lead suit.
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.WizardDesignJester, 1, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "HEART")
		assert.NotContains(t, result, i18n.T("wizard.leadNone"))
	})

	t.Run("renders wizard and jester cards in hand", func(t *testing.T) {
		m, players := setupWizardCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.WizardDesignWizard, 1, false))
		players[0].AddCard(domain.NewCard(domain.WizardDesignJester, 1, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Wizard")
		assert.Contains(t, result, "Jester")
	})

	// **出せる札に印を付ける。**リードスート名だけ出しても、CUI プレイヤーは
	// 毎トリック自分の手札と暗算で照合させられる (#4927)。
	t.Run("play phase marks the cards that may legally be played", func(t *testing.T) {
		m, players := setupWizardCuiMockWithPlayers()
		players[0].Reset()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))    // follows
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))    // must not
		players[0].AddCard(domain.NewCard(domain.WizardDesignWizard, 1, false)) // always legal
		players[0].AddCard(domain.NewCard(domain.WizardDesignJester, 1, false)) // always legal
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]HEART 9"+presenter.WizardLegalMark)
		assert.NotContains(t, result, "[1]SPADE 3"+presenter.WizardLegalMark)
		assert.Contains(t, result, "[2]Wizard"+presenter.WizardLegalMark)
		assert.Contains(t, result, "[3]Jester"+presenter.WizardLegalMark)
		// 凡例が無いと印の意味が分からない。
		assert.Contains(t, result, i18n.T("wizard.legalMark"))
	})

	// 他人の手番なら印は出ない。出せる札は手番が来てから分かればよい。
	t.Run("no marks while it is not the human's turn", func(t *testing.T) {
		m, players := setupWizardCuiMockWithPlayers()
		players[0].Reset()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(2)
		result := p.Output(m, nil)
		assert.NotContains(t, result, presenter.WizardLegalMark)
		assert.NotContains(t, result, i18n.T("wizard.legalMark"))
	})

	// 全部出せるときは印だらけになるだけなので、印も凡例も出さない。
	t.Run("no marks while leading, where every card is legal", func(t *testing.T) {
		m, players := setupWizardCuiMockWithPlayers()
		players[0].Reset()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		result := p.Output(m, nil) // 既定のトリックは空 = リード
		assert.NotContains(t, result, presenter.WizardLegalMark)
		assert.NotContains(t, result, i18n.T("wizard.legalMark"))
	})

	// ja / en 双方に凡例がある。片方だけだと `--lang en` でキーが出る。
	t.Run("legal-mark legend is translated in both languages", func(t *testing.T) {
		defer i18n.SetLang("ja")
		i18n.SetLang("ja")
		ja := i18n.T("wizard.legalMark")
		assert.NotEqual(t, "wizard.legalMark", ja)
		i18n.SetLang("en")
		en := i18n.T("wizard.legalMark")
		assert.NotEqual(t, "wizard.legalMark", en)
		assert.NotEqual(t, ja, en)
	})

	t.Run("bid phase", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseBid)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッドフェーズ")
	})

	t.Run("bid phase with restriction", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRestrictedBid")
		m.On("GetPhase").Return(domain.WizardPhaseBid)
		m.On("GetRestrictedBid").Return(3)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッド3は不可")
	})

	t.Run("trick end", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.WizardPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("no trump", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpCard")
		m.On("GetTrumpCard").Return((*domain.Card)(nil))
		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: なし")
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupWizardCuiMockWithPlayers()
		result := p.Output(m, domain.ErrInvalidPlay)
		assert.Contains(t, result, "invalid play")
	})
}

func TestWizardCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	p := new(presenter.WizardCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		m := setupWizardCuiMock()
		bid := 3
		m.On("GetHint").Return(&domain.WizardHint{Bid: &bid, Reason: "strategic_bid"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ビッド 3")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupWizardCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		cardIdx := 0
		m.On("GetHint").Return(&domain.WizardHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "[0]")
	})

	t.Run("nil hint", func(t *testing.T) {
		m := setupWizardCuiMock()
		m.On("GetHint").Return((*domain.WizardHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint with nil bid and nil cardIndex", func(t *testing.T) {
		m := setupWizardCuiMock()
		m.On("GetHint").Return(&domain.WizardHint{Reason: "unknown"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("card hint with no human player", func(t *testing.T) {
		m := setupWizardCuiMock()
		cpuPlayers := []*domain.WizardPlayer{
			domain.NewWizardPlayer(false),
			domain.NewWizardPlayer(false),
			domain.NewWizardPlayer(false),
			domain.NewWizardPlayer(false),
		}
		m.On("GetPlayerCnt").Return(4)
		m.On("GetPlayer", 0).Return(cpuPlayers[0])
		m.On("GetPlayer", 1).Return(cpuPlayers[1])
		m.On("GetPlayer", 2).Return(cpuPlayers[2])
		m.On("GetPlayer", 3).Return(cpuPlayers[3])
		cardIdx := 0
		m.On("GetHint").Return(&domain.WizardHint{CardIndex: &cardIdx, Reason: "follow_suit"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestWizardCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WizardCuiPresenter)
	m := setupWizardCuiMock()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "You bids 3"},
	})
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
