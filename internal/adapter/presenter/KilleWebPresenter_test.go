//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// makeKillePlayers builds four seats, the first human, each holding one card.
func makeKillePlayers(ranks ...domain.KilleRank) []*domain.KillePlayer {
	out := make([]*domain.KillePlayer, 0, len(ranks))
	for i, r := range ranks {
		p := domain.NewKillePlayer(i == 0)
		p.AddCard(domain.NewKilleCard(r))
		out = append(out, p)
	}
	return out
}

func setupKilleWebMock(phase domain.KillePhase) (*interfaces.MockKilleGame, []*domain.KillePlayer) {
	m := new(interfaces.MockKilleGame)
	players := makeKillePlayers(domain.KilleNum5, domain.KillePig, domain.KilleHarlequin, domain.KilleNum9)
	m.On("GetPhase").Return(phase)
	m.On("GetRoundNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(3)
	m.On("GetStockCount").Return(38)
	m.On("GetPot").Return(4)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLoserIdxs").Return([]int{})
	m.On("GetEvents").Return([]*domain.KilleEvent{
		{Kind: "swap", Actor: 0, Target: 1},
		nil, // nil 混入でも落ちないこと
	})
	m.On("GetConfig").Return(domain.DefaultKilleConfig())
	m.On("GetPlayers").Return(players)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("KilleStrength", i).Return(int(domain.KilleRankOf(players[i].GetCard(0))))
		m.On("KilleReentryCost", i).Return(1)
	}
	return m, players
}

func parseKilleOutput(t *testing.T, s string) *controller.KilleWebOutput {
	t.Helper()
	var out controller.KilleWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **他家の札は伏せる。**1 枚しか持たないゲームなので、漏れたら勝負が終わる。
func TestKilleWebPresenter_HidesEveryoneElsesCard(t *testing.T) {
	m, _ := setupKilleWebMock(domain.KillePhaseExchange)
	out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, nil))

	assert.Len(t, out.Players, 4)
	assert.NotNil(t, out.Players[0].Card, "the human sees its own card")
	for i := 1; i < 4; i++ {
		assert.Nil(t, out.Players[i].Card, "seat %d must stay face down", i)
		// **強さも送らない。**送ると伏せた意味が無い。
		assert.Equal(t, 0, out.Players[i].Strength, "seat %d leaks its strength", i)
	}
}

func TestKilleWebPresenter_RevealsAtShowdown(t *testing.T) {
	m, _ := setupKilleWebMock(domain.KillePhaseShowdown)
	out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, nil))

	for i := range out.Players {
		assert.NotNil(t, out.Players[i].Card, "seat %d must be revealed at the showdown", i)
	}
	assert.Equal(t, int(domain.KilleHarlequin), out.Players[2].Strength)
}

// 専用デッキなので手続き描画のフィールドが要る (ADR-0033)。
func TestKilleWebPresenter_SendsProceduralCardFields(t *testing.T) {
	m, _ := setupKilleWebMock(domain.KillePhaseShowdown)
	out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, nil))

	pig := out.Players[1].Card
	assert.NotNil(t, pig)
	assert.Equal(t, domain.KilleDeckID, pig.Deck, "a non-52 card must switch the frontend to the procedural path")
	assert.Equal(t, "Pig", pig.Label)
	assert.NotEmpty(t, pig.Glyph)
	assert.Equal(t, "gold", pig.Color)

	// 効果を持たない札は黒。
	assert.Equal(t, "black", out.Players[0].Card.Color)
}

func TestKilleWebPresenter_TopLevelFields(t *testing.T) {
	m, _ := setupKilleWebMock(domain.KillePhaseExchange)
	out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, nil))

	assert.Equal(t, int(domain.KillePhaseExchange), out.Phase)
	assert.Equal(t, 1, out.RoundNumber)
	assert.Equal(t, 3, out.DealerIdx)
	assert.Equal(t, 38, out.StockCount)
	assert.Equal(t, 4, out.Pot)
	assert.Equal(t, 1, out.Config.Stake)
	assert.Equal(t, []int{}, out.LoserIdxs)
	// nil イベントは落として送る。
	assert.Len(t, out.Events, 1)
	assert.Equal(t, "swap", out.Events[0].Kind)
	assert.Equal(t, 1, out.Events[0].Target)
	assert.True(t, out.Players[0].IsCurrentTurn)
	assert.False(t, out.Players[1].IsCurrentTurn)
}

func TestKilleWebPresenter_Messages(t *testing.T) {
	t.Run("an error wins over any phase message", func(t *testing.T) {
		m, _ := setupKilleWebMock(domain.KillePhaseExchange)
		out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})

	t.Run("the dealer gets its own prompt", func(t *testing.T) {
		m := new(interfaces.MockKilleGame)
		players := makeKillePlayers(domain.KilleNum5, domain.KilleNum9)
		m.On("GetPhase").Return(domain.KillePhaseExchange)
		m.On("GetRoundNumber").Return(0)
		m.On("GetCurrentPlayerIdx").Return(1)
		m.On("GetDealerIdx").Return(1)
		m.On("GetStockCount").Return(40)
		m.On("GetPot").Return(2)
		m.On("GetGameEndFlag").Return(false)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetLoserIdxs").Return([]int{})
		m.On("GetEvents").Return([]*domain.KilleEvent{})
		m.On("GetConfig").Return(domain.DefaultKilleConfig())
		m.On("GetPlayers").Return(players)
		for i := range players {
			m.On("GetPlayer", i).Return(players[i])
			m.On("KilleStrength", i).Return(0)
			m.On("KilleReentryCost", i).Return(1)
		}
		out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, nil))
		assert.Equal(t, "kille.dealerTurn", out.MessageCode)
	})

	t.Run("showdown", func(t *testing.T) {
		m, _ := setupKilleWebMock(domain.KillePhaseShowdown)
		out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, nil))
		assert.Equal(t, "kille.showdown", out.MessageCode)
	})

	t.Run("exchange", func(t *testing.T) {
		m, _ := setupKilleWebMock(domain.KillePhaseExchange)
		out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, nil))
		assert.Equal(t, "kille.exchangePhase", out.MessageCode)
	})
}

func TestKilleWebPresenter_GameEnd(t *testing.T) {
	for _, tc := range []struct {
		name    string
		winner  int
		wantKey string
	}{
		{"the human survives", 0, "kille.result.humanWin"},
		{"a CPU survives", 2, "kille.result.cpuWin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := new(interfaces.MockKilleGame)
			players := makeKillePlayers(domain.KilleNum5, domain.KillePig, domain.KilleHarlequin, domain.KilleNum9)
			m.On("GetPhase").Return(domain.KillePhaseGameEnd)
			m.On("GetRoundNumber").Return(4)
			m.On("GetCurrentPlayerIdx").Return(0)
			m.On("GetDealerIdx").Return(3)
			m.On("GetStockCount").Return(38)
			m.On("GetPot").Return(0)
			m.On("GetGameEndFlag").Return(true)
			m.On("GetWinnerIdx").Return(tc.winner)
			m.On("GetLoserIdxs").Return([]int{1, 3})
			m.On("GetEvents").Return([]*domain.KilleEvent{})
			m.On("GetConfig").Return(domain.DefaultKilleConfig())
			m.On("GetPlayers").Return(players)
			for i := range players {
				m.On("GetPlayer", i).Return(players[i])
				m.On("KilleStrength", i).Return(0)
				m.On("KilleReentryCost", i).Return(0)
			}

			out := parseKilleOutput(t, new(presenter.KilleWebPresenter).Output(m, nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
			assert.Equal(t, []int{1, 3}, out.LoserIdxs)
		})
	}
}

func TestKilleWebPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupKilleWebMock(domain.KillePhaseExchange)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.KilleWebPresenter).ActionLogOutput(m))
}
