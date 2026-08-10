//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKlaberjassCuiMock(phase domain.KlaberjassPhase, trump int, turnUp *domain.Card) *interfaces.MockKlaberjassGame {
	m := new(interfaces.MockKlaberjassGame)
	players := makeKlaberjassPlayers(
		[]*domain.Card{kjTestCard(domain.CardDesignSpade, 11), kjTestCard(domain.CardDesignHeart, 1)},
		[]*domain.Card{kjTestCard(domain.CardDesignClover, 7), kjTestCard(domain.CardDesignClover, 8)},
	)
	m.On("GetPhase").Return(phase)
	m.On("GetLastTrickWinner").Return(-1)
	m.On("GetDealNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetMakerIdx").Return(1)
	m.On("GetTrumpSuit").Return(trump)
	m.On("GetTurnUpCard").Return(turnUp)
	m.On("GetTrick").Return([]*domain.Card{kjTestCard(domain.CardDesignHeart, 10)})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("IsBete").Return(false)
	m.On("GetConfig").Return(domain.DefaultKlaberjassConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("KlaberjassValidPlays", 0).Return([]int{1})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetHandPoints", i).Return(0)
		m.On("GetScore", i).Return(0)
	}
	return m
}

func TestKlaberjassCuiPresenter_HidesTheOpponentsHand(t *testing.T) {
	m := setupKlaberjassCuiMock(domain.KlaberjassPhasePlay, domain.CardDesignSpade, nil)
	out := new(presenter.KlaberjassCuiPresenter).Output(m, nil)

	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "非公開 2枚")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "[宣言]")
	assert.Contains(t, out, "場:")
}

// **出せる札を出さないと操作できない。**追随・切札・上乗せが全部強制なので、
// 手札の並びだけでは何が合法か判らない。
func TestKlaberjassCuiPresenter_ListsThePlayableIndexes(t *testing.T) {
	m := setupKlaberjassCuiMock(domain.KlaberjassPhasePlay, domain.CardDesignSpade, nil)
	out := new(presenter.KlaberjassCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "出せる札: 1")
}

// 表向きカードはビッド中だけ出す。切札が決まったあとは意味を持たない。
func TestKlaberjassCuiPresenter_ShowsTheTurnUpOnlyWhileBidding(t *testing.T) {
	bidding := setupKlaberjassCuiMock(domain.KlaberjassPhaseBidTurnUp, 0, kjTestCard(domain.CardDesignHeart, 13))
	assert.Contains(t, new(presenter.KlaberjassCuiPresenter).Output(bidding, nil), "表向き:")

	playing := setupKlaberjassCuiMock(domain.KlaberjassPhasePlay, domain.CardDesignSpade, kjTestCard(domain.CardDesignHeart, 13))
	assert.NotContains(t, new(presenter.KlaberjassCuiPresenter).Output(playing, nil), "表向き:")
}

func TestKlaberjassCuiPresenter_PhasePrompts(t *testing.T) {
	for _, tc := range []struct {
		phase domain.KlaberjassPhase
		want  string
	}{
		{domain.KlaberjassPhaseBidTurnUp, "表向きのスートを取る"},
		{domain.KlaberjassPhaseBidFree, "好きなスートを指名"},
		{domain.KlaberjassPhaseSchmeiss, "投げの提案"},
		{domain.KlaberjassPhasePlay, "追随・切札・上乗せは強制"},
		{domain.KlaberjassPhaseHandEnd, "次のディールへ"},
	} {
		m := setupKlaberjassCuiMock(tc.phase, domain.CardDesignSpade, nil)
		assert.Contains(t, new(presenter.KlaberjassCuiPresenter).Output(m, nil), tc.want)
	}
}

// ベートは通常の精算と字面で見分けられる必要がある。
func TestKlaberjassCuiPresenter_NamesBete(t *testing.T) {
	m := new(interfaces.MockKlaberjassGame)
	players := makeKlaberjassPlayers([]*domain.Card{}, []*domain.Card{})
	m.On("GetPhase").Return(domain.KlaberjassPhaseHandEnd)
	m.On("GetLastTrickWinner").Return(1)
	m.On("GetDealNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetMakerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTurnUpCard").Return((*domain.Card)(nil))
	m.On("GetTrick").Return([]*domain.Card{})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("IsBete").Return(true)
	m.On("GetConfig").Return(domain.DefaultKlaberjassConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(false)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetHandPoints", i).Return(0)
		m.On("GetScore", i).Return(0)
	}
	assert.Contains(t, new(presenter.KlaberjassCuiPresenter).Output(m, nil), "ベート")
}

func TestKlaberjassCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	m := setupKlaberjassCuiMock(domain.KlaberjassPhasePlay, domain.CardDesignSpade, nil)
	assert.Contains(t, new(presenter.KlaberjassCuiPresenter).Output(m, errors.New("boom")), "boom")

	end := new(interfaces.MockKlaberjassGame)
	players := makeKlaberjassPlayers([]*domain.Card{}, []*domain.Card{})
	end.On("GetPhase").Return(domain.KlaberjassPhaseGameEnd)
	end.On("GetDealNumber").Return(9)
	end.On("GetCurrentPlayerIdx").Return(0)
	end.On("GetBidPlayerIdx").Return(1)
	end.On("GetDealerIdx").Return(0)
	end.On("GetMakerIdx").Return(0)
	end.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	end.On("GetTurnUpCard").Return((*domain.Card)(nil))
	end.On("GetTrick").Return([]*domain.Card{})
	end.On("GetGameEndFlag").Return(true)
	end.On("GetWinnerIdx").Return(0)
	end.On("IsBete").Return(false)
	end.On("GetConfig").Return(domain.DefaultKlaberjassConfig())
	end.On("GetPlayers").Return(players)
	end.On("IsHumanTurn").Return(false)
	for i := range players {
		end.On("GetPlayer", i).Return(players[i])
		end.On("GetHandPoints", i).Return(0)
		end.On("GetScore", i).Return(0)
	}
	assert.Contains(t, new(presenter.KlaberjassCuiPresenter).Output(end, nil), "ゲーム終了")
}

func TestKlaberjassCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupKlaberjassCuiMock(domain.KlaberjassPhasePlay, domain.CardDesignSpade, nil)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.KlaberjassCuiPresenter).ActionLogOutput(m))
}

// **最終トリックには 10 点が付く。**書かないと、ベラや宣言点を足しても
// handPoints と合わない理由が説明できない (#4937)。
func TestKlaberjassCuiPresenter_NamesTheLastTrickBonus(t *testing.T) {
	build := func(lastTrick int) *interfaces.MockKlaberjassGame {
		m := new(interfaces.MockKlaberjassGame)
		players := makeKlaberjassPlayers([]*domain.Card{}, []*domain.Card{})
		m.On("GetPhase").Return(domain.KlaberjassPhaseHandEnd)
		m.On("GetLastTrickWinner").Return(lastTrick)
		m.On("GetDealNumber").Return(1)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetBidPlayerIdx").Return(1)
		m.On("GetDealerIdx").Return(0)
		m.On("GetMakerIdx").Return(0)
		m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
		m.On("GetTurnUpCard").Return((*domain.Card)(nil))
		m.On("GetTrick").Return([]*domain.Card{})
		m.On("GetGameEndFlag").Return(false)
		m.On("GetWinnerIdx").Return(-1)
		m.On("IsBete").Return(false)
		m.On("GetConfig").Return(domain.DefaultKlaberjassConfig())
		m.On("GetPlayers").Return(players)
		m.On("IsHumanTurn").Return(false)
		for i := range players {
			m.On("GetPlayer", i).Return(players[i])
			m.On("GetHandPoints", i).Return(0)
			m.On("GetScore", i).Return(0)
		}
		return m
	}
	p := new(presenter.KlaberjassCuiPresenter)

	out := p.Output(build(1), nil)
	assert.Contains(t, out, "最終トリック")
	assert.Contains(t, out, "+"+strconv.Itoa(domain.KlaberjassLastTrickBonus)+"点")

	// まだ誰も取っていなければ出さない。
	assert.NotContains(t, p.Output(build(-1), nil), "最終トリック")
}
