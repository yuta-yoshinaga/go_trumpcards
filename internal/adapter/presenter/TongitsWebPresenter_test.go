//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupTongitsWebMock() *interfaces.MockTongitsGame {
	m := new(interfaces.MockTongitsGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(41)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TongitsPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultTongitsConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetKnockerIdx").Return(-1)
	m.On("GetKnockerMelds").Return(([][]*domain.Card)(nil))
	m.On("GetKnockerDeadwood").Return(([]*domain.Card)(nil))
	m.On("GetOpponentMelds").Return(([][]*domain.Card)(nil))
	m.On("GetOpponentDeadwood").Return(([]*domain.Card)(nil))
	m.On("GetIsTongits").Return(false)
	m.On("GetIsUndercut").Return(false)
	// #4750: ディスカードフェーズで最小デッドウッドを引く。
	m.On("IsHumanTurn").Return(false).Maybe()
	m.On("GetBestDeadwood", 0).Return(3, 0).Maybe()

	return m
}

func setupTongitsWebMockWithPlayers() (*interfaces.MockTongitsGame, []*domain.TongitsPlayer) {
	m := setupTongitsWebMock()
	players := makeTongitsPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestTongitsWebPresenter_Output(t *testing.T) {
	p := new(presenter.TongitsWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.TongitsWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 41, resObj.DrawPileCount)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, -1, resObj.KnockerIdx)
		assert.False(t, resObj.IsTongits)
		assert.Nil(t, resObj.DiscardTop)
	})

	t.Run("human cards shown, CPU cards hidden in draw phase", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("CPU cards shown in round end", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
	})

	t.Run("CPU cards shown in game end", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.TongitsPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
		assert.Equal(t, 0, resObj.WinnerIdx)
		assert.NotEmpty(t, resObj.MessageCode)
	})

	t.Run("error in lastErr", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		testErr := errors.New("oops")
		result := p.Output(m, testErr)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "oops", resObj.Message)
	})

	t.Run("knocker melds and deadwood serialized", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerMelds")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerDeadwood")
		m.On("GetKnockerIdx").Return(0)
		melds := [][]*domain.Card{
			{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
				domain.NewCard(domain.CardDesignDiamond, 5, false),
			},
		}
		m.On("GetKnockerMelds").Return(melds)
		m.On("GetKnockerDeadwood").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignClover, 7, false),
		})

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, 0, resObj.KnockerIdx)
		assert.Len(t, resObj.KnockerMelds, 1)
		assert.Len(t, resObj.KnockerMelds[0].Cards, 3)
		assert.Len(t, resObj.KnockerDeadwood, 1)
	})

	t.Run("discard top serialized", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 8, false))

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, 8, resObj.DiscardTop.Value)
	})

	t.Run("draw phase message code", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tongits.drawPhase", resObj.MessageCode)
	})

	t.Run("discard phase message code", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseDiscard)
		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tongits.discardPhase", resObj.MessageCode)
	})

	t.Run("round end with tongits on deal message code", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsTongits")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		m.On("GetIsTongits").Return(true)
		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tongits.tongitsOnDeal", resObj.MessageCode)
	})

	t.Run("round end normal message code", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tongits.roundEnd", resObj.MessageCode)
	})
}

// **CUI は毎ターン「ノック可能/不可」を出しているのに、Web はプレイヤーの
// 手計算に任せていた (#4750)。**判断の基準 (閾値) ごと送るので、フロントは
// 数値を写さずに済む。
func TestTongitsWebPresenter_BestDeadwood(t *testing.T) {
	p := new(presenter.TongitsWebPresenter)

	decode := func(t *testing.T, m *interfaces.MockTongitsGame) controller.TongitsWebOutput {
		t.Helper()
		var out controller.TongitsWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &out))
		return out
	}

	t.Run("human discard turn carries the domain's answer and the threshold", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBestDeadwood")
		m.On("GetPhase").Return(domain.TongitsPhaseDiscard)
		m.On("IsHumanTurn").Return(true)
		m.On("GetBestDeadwood", 0).Return(4, 1)

		out := decode(t, m)
		assert.Equal(t, 4, out.BestDeadwood)
		assert.Equal(t, domain.TongitsKnockThreshold, out.KnockThreshold)
	})

	// **-1 は「まだ聞くべき場面でない」印。**0 にすると「デッドウッド0 =
	// 必ずノック可能」と読めてしまい、ドローフェーズで誤った案内が出る。
	t.Run("outside the human discard turn it is -1, not 0", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("IsHumanTurn").Return(true) // フェーズは Draw のまま

		assert.Equal(t, -1, decode(t, m).BestDeadwood)
	})

	t.Run("cpu discard turn is -1 too", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseDiscard) // IsHumanTurn は既定 false

		assert.Equal(t, -1, decode(t, m).BestDeadwood)
	})
}

func TestTongitsWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TongitsWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockTongitsGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "knock", Detail: "knocks"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "knock")
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockTongitsGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
	})
}

// #5582: 閾値はドメインから渡すこと。画面に 2 を書くと、変えたとき Web と CUI で
// 警告の出る局面がずれる。
func TestTongitsWebPresenter_ShipsTheUndercutThreshold(t *testing.T) {
	g := domain.NewDefaultTongits()
	g.Reset()

	var out controller.TongitsWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(presenter.TongitsWebPresenter).Output(g, nil)), &out))
	assert.Equal(t, domain.TongitsUndercutRiskMax, out.UndercutRiskMax)
	assert.NotZero(t, out.UndercutRiskMax)
}
