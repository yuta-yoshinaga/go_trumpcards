//go:build test

package presenter_test

import (
	"errors"
	"strconv"
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

func setupSambaCuiMock() *interfaces.MockSambaGame {
	m := new(interfaces.MockSambaGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(80)
	m.On("GetDiscardPileCount").Return(0)
	m.On("GetIsFrozen").Return(false)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SambaPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetMinimumMeldValue", mock.Anything).Return(50)
	return m
}

func setupSambaCuiMockWithPlayers() (*interfaces.MockSambaGame, []*domain.SambaPlayer) {
	m := setupSambaCuiMock()
	players := makeSambaPlayers()
	m.On("GetPlayerCnt").Return(4)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestSambaCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SambaCuiPresenter)

	t.Run("initial state header and players", func(t *testing.T) {
		m, players := setupSambaCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Samba (サンバ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 80枚")
		assert.Contains(t, result, "チーム0:")
		assert.Contains(t, result, "あなた (チーム0)")
		assert.Contains(t, result, "[0]SPADE 5")
		assert.Contains(t, result, "手番: あなた")
		// Not frozen → no frozen draw-help note.
		assert.NotContains(t, result, i18n.T("samba.promptDrawHelpFrozen"))
	})

	t.Run("frozen tag and frozen draw help", func(t *testing.T) {
		m, _ := setupSambaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsFrozen")
		m.On("GetIsFrozen").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "[フリーズ]")
		// Frozen draw phase spells out the take-condition.
		assert.Contains(t, result, i18n.T("samba.promptDrawHelpFrozen"))
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupSambaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札: HEART 7")
	})

	t.Run("sequence meld shows samba label", func(t *testing.T) {
		m, players := setupSambaCuiMockWithPlayers()
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
		}
		players[0].AddMeld(&domain.SambaMeld{Cards: cards, Kind: domain.SambaMeldSequence, IsNatural: true})
		result := p.Output(m, nil)
		assert.Contains(t, result, "シーケンス")
		assert.Contains(t, result, "サンバ")
		assert.Contains(t, result, "★サンバ")
	})

	t.Run("set meld shows canasta label", func(t *testing.T) {
		m, players := setupSambaCuiMockWithPlayers()
		cards := make([]*domain.Card, 7)
		for i := range cards {
			cards[i] = domain.NewCard(domain.CardDesignSpade, 5, false)
		}
		players[0].AddMeld(&domain.SambaMeld{Cards: cards, Kind: domain.SambaMeldSet, IsNatural: true})
		result := p.Output(m, nil)
		assert.Contains(t, result, "ナチュラル")
		assert.Contains(t, result, "カナスタ")
		assert.Contains(t, result, "★カナスタ")
	})

	t.Run("red 3 tag", func(t *testing.T) {
		m, players := setupSambaCuiMockWithPlayers()
		players[0].AddRed3(domain.NewCard(domain.CardDesignHeart, 3, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "赤3: 1枚")
	})

	t.Run("error message shown", func(t *testing.T) {
		m, _ := setupSambaCuiMockWithPlayers()
		result := p.Output(m, errors.New("invalid card index"))
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game ended human team wins", func(t *testing.T) {
		m, _ := setupSambaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "あなたのチームの勝利です！")
	})

	t.Run("meld phase commands", func(t *testing.T) {
		m, _ := setupSambaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SambaPhaseMeld)
		result := p.Output(m, nil)
		assert.Contains(t, result, "メルドフェーズ")
		assert.Contains(t, result, "sm")
	})

	t.Run("discard phase commands", func(t *testing.T) {
		m, _ := setupSambaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SambaPhaseDiscard)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "go")
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupSambaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SambaPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "nr / nextround")
	})
}

func TestSambaCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SambaCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockSambaGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "drew"},
		})
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewSambaPlayer(true, 0)).Maybe()
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "draw_stock")
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockSambaGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
		m.AssertExpectations(t)
	})
}

// #5702: 初回メルドの最低点はチーム累積点で 15/50/90/120 と変わる。Web は
// sa-meld-points に「必要点数」を常時出しているのに、CUI は一般的な案内文だけで、
// 足りているかはメルドを試すまで分からなかった。
func TestSambaCuiPresenter_ShowsTheInitialMeldRequirement(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SambaCuiPresenter)

	meldTurn := func(minValue int, melded bool) *interfaces.MockSambaGame {
		m, players := setupSambaCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SambaPhaseMeld)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMinimumMeldValue")
		m.On("GetMinimumMeldValue", mock.Anything).Return(minValue)
		players[0].SetHasInitMeld(melded)
		return m
	}

	// スコア帯ごとの値はドメインが決める。表示側はそれをそのまま出す。
	for _, minValue := range []int{15, 50, 90, 120} {
		out := p.Output(meldTurn(minValue, false), nil)

		assert.Contains(t, out, i18n.Tf("samba.promptMeldMinimum",
			"points", strconv.Itoa(minValue)))
	}

	t.Run("says nothing once the team has melded", func(t *testing.T) {
		out := p.Output(meldTurn(90, true), nil)

		assert.NotContains(t, out, strings.Split(i18n.T("samba.promptMeldMinimum"), "{{")[0])
	})
}
