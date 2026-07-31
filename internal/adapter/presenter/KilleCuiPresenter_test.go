//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKilleCuiMock(phase domain.KillePhase, players []*domain.KillePlayer) *interfaces.MockKilleGame {
	m := new(interfaces.MockKilleGame)
	m.On("GetPhase").Return(phase)
	m.On("GetRoundNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(3)
	m.On("GetStockCount").Return(38)
	m.On("GetPot").Return(4)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLoserIdxs").Return([]int{1})
	m.On("GetPlayers").Return(players)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m
}

func TestKilleCuiPresenter_HidesOtherHands(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KillePig, domain.KilleCuckoo, domain.KilleNum9)
	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	out := new(presenter.KilleCuiPresenter).Output(m, nil)

	assert.Contains(t, out, "5", "the human's own card is shown")
	assert.NotContains(t, out, "Pig", "seat 1 must stay face down")
	assert.NotContains(t, out, "Cuckoo", "seat 2 must stay face down")
	assert.Contains(t, out, "非公開")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "ポット: 4")
}

func TestKilleCuiPresenter_ShowdownNamesTheReason(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KillePig, domain.KilleCuckoo, domain.KilleNum9)
	players[1].SetOut(domain.KilleKnockPig)
	players[2].SetOut(domain.KilleKnockHussar)
	players[3].SetOut(domain.KilleKnockLowest)
	m := setupKilleCuiMock(domain.KillePhaseShowdown, players)
	out := new(presenter.KilleCuiPresenter).Output(m, nil)

	// 公開されるので全員の札が出る。
	assert.Contains(t, out, "Pig")
	assert.Contains(t, out, "Cuckoo")
	// **なぜ落ちたかまで出す。**軽騎兵と豚は自分の手の強さと無関係に落ちる。
	assert.Contains(t, out, "豚に噛まれた")
	assert.Contains(t, out, "軽騎兵に返り討ち")
	assert.Contains(t, out, "[脱落]")
	assert.Contains(t, out, "1人が脱落")
}

func TestKilleCuiPresenter_DealerGetsItsOwnPrompt(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KilleNum9, domain.KilleNum7, domain.KilleNum3)
	m := new(interfaces.MockKilleGame)
	m.On("GetPhase").Return(domain.KillePhaseExchange)
	m.On("GetRoundNumber").Return(0)
	m.On("GetCurrentPlayerIdx").Return(3)
	m.On("GetDealerIdx").Return(3)
	m.On("GetStockCount").Return(38)
	m.On("GetPot").Return(4)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLoserIdxs").Return([]int{})
	m.On("GetPlayers").Return(players)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}

	out := new(presenter.KilleCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "山札から引いて交換")
}

func TestKilleCuiPresenter_SatisfiedAndEliminatedTags(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KilleNum9, domain.KilleNum7, domain.KilleNum3)
	players[1].SetSatisfied(true)
	players[2].SetIsFinished(true)
	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	out := new(presenter.KilleCuiPresenter).Output(m, nil)

	assert.Contains(t, out, "[満足]")
	assert.Contains(t, out, "退場")
}

func TestKilleCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KilleNum9, domain.KilleNum7, domain.KilleNum3)

	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	assert.Contains(t, new(presenter.KilleCuiPresenter).Output(m, errors.New("boom")), "boom")

	end := new(interfaces.MockKilleGame)
	end.On("GetPhase").Return(domain.KillePhaseGameEnd)
	end.On("GetRoundNumber").Return(9)
	end.On("GetCurrentPlayerIdx").Return(0)
	end.On("GetDealerIdx").Return(3)
	end.On("GetStockCount").Return(38)
	end.On("GetPot").Return(0)
	end.On("GetGameEndFlag").Return(true)
	end.On("GetWinnerIdx").Return(0)
	end.On("GetLoserIdxs").Return([]int{1, 2, 3})
	end.On("GetPlayers").Return(players)
	for i := range players {
		end.On("GetPlayer", i).Return(players[i])
	}
	assert.Contains(t, new(presenter.KilleCuiPresenter).Output(end, nil), "ゲーム終了")
}

func TestKilleCuiPresenter_ActionLogOutput(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5)
	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.KilleCuiPresenter).ActionLogOutput(m))
}
