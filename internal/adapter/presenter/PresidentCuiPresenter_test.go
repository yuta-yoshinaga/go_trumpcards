package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makePresidentPlayersForPresenter() []*domain.PresidentPlayer {
	return []*domain.PresidentPlayer{
		domain.NewPresidentPlayer(true),
		domain.NewPresidentPlayer(false),
		domain.NewPresidentPlayer(false),
		domain.NewPresidentPlayer(false),
	}
}

func TestPresidentCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PresidentCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(pg, nil)
		assert.Contains(t, result, "President (プレジデント)")
		assert.Contains(t, result, "あなた: 1枚")
		assert.Contains(t, result, "場: なし")
		assert.Contains(t, result, "手番: あなた")
	})

	t.Run("revolution active", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		for i := 1; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		pg.SetRevolutionActive(true)
		result := p.Output(pg, nil)
		assert.Contains(t, result, "革命中")
	})

	t.Run("with table cards and last play info", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		pg.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		pg.SetLastPlayPlayerIdx(2)
		result := p.Output(pg, nil)
		assert.Contains(t, result, "場:")
		assert.Contains(t, result, "CPU 2")
	})

	t.Run("error message rendered", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		result := p.Output(pg, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("exchange actions rendered", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		pg.SetExchangeActions([]*domain.PresidentExchangeAction{
			{FromPlayerIdx: 3, ToPlayerIdx: 0, Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)}},
		})
		result := p.Output(pg, nil)
		assert.Contains(t, result, "カード交換")
	})

	t.Run("cpu actions rendered", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		pg.SetCurrentTurn(1)
		pg.CpuPlay()
		result := p.Output(pg, nil)
		assert.Contains(t, result, "CPUの行動")
	})

	t.Run("human action pass rendered", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		pg.SetHumanAction(&domain.PresidentCpuAction{PlayerIdx: 0, PlayedCards: nil})
		result := p.Output(pg, nil)
		assert.Contains(t, result, "パス")
	})

	// Game-end + rank-line rendering is exercised through the mock since
	// the real President domain has no SetGameEndFlag. The mock lets us
	// drive each rank value (1-4) plus an out-of-range value through
	// presidentRankName via the gameEnd loop.
	t.Run("game end with all four ranks + unknown rank", func(t *testing.T) {
		m := new(interfaces.MockPresidentGame)
		players := makePresidentPlayersForPresenter()
		players[0].SetRank(1)
		players[0].SetIsFinished(true)
		players[1].SetRank(2)
		players[1].SetIsFinished(true)
		players[2].SetRank(3)
		players[2].SetIsFinished(true)
		players[3].SetRank(4)
		players[3].SetIsFinished(true)

		m.On("GetPlayerCnt").Return(4)
		for i := range 4 {
			m.On("GetPlayer", i).Return(players[i])
		}
		m.On("GetRevolutionActive").Return(false)
		m.On("GetExchangeActions").Return(([]*domain.PresidentExchangeAction)(nil))
		m.On("GetTableCards").Return(([]*domain.Card)(nil))
		m.On("GetHumanAction").Return((*domain.PresidentCpuAction)(nil))
		m.On("GetCpuActions").Return(([]*domain.PresidentCpuAction)(nil))
		m.On("GetGameEndFlag").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "大統領")
		assert.Contains(t, result, "副大統領")
		assert.Contains(t, result, "副スカム")
		assert.Contains(t, result, "スカム")
	})

	t.Run("rank entry with out-of-range rank shows unknown label", func(t *testing.T) {
		m := new(interfaces.MockPresidentGame)
		players := makePresidentPlayersForPresenter()
		// Single finished player with rank 0 — exercises the default branch
		// of presidentRankName which the other test does not.
		players[0].SetRank(0)
		players[0].SetIsFinished(true)

		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetRevolutionActive").Return(false)
		m.On("GetExchangeActions").Return(([]*domain.PresidentExchangeAction)(nil))
		m.On("GetTableCards").Return(([]*domain.Card)(nil))
		m.On("GetHumanAction").Return((*domain.PresidentCpuAction)(nil))
		m.On("GetCpuActions").Return(([]*domain.PresidentCpuAction)(nil))
		m.On("GetGameEndFlag").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "不明")
	})
}

func TestPresidentCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.PresidentCuiPresenter)

	t.Run("recommends the weakest legal play", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		m := new(interfaces.MockPresidentGame)
		m.On("GetGameEndFlag").Return(false)
		m.On("IsHumanTurn").Return(true)
		m.On("GetCurrentTurn").Return(0)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("SuggestWeakestPlay", 0).Return([]int{0})
		assert.Contains(t, p.HintOutput(m), "を出す")
	})

	t.Run("recommends passing when no legal play exists", func(t *testing.T) {
		m := new(interfaces.MockPresidentGame)
		m.On("GetGameEndFlag").Return(false)
		m.On("IsHumanTurn").Return(true)
		m.On("GetCurrentTurn").Return(0)
		m.On("GetPlayer", 0).Return(makePresidentPlayersForPresenter()[0])
		m.On("SuggestWeakestPlay", 0).Return(([]int)(nil))
		assert.Contains(t, p.HintOutput(m), "パス")
	})

	t.Run("declines when it is not the human's turn", func(t *testing.T) {
		m := new(interfaces.MockPresidentGame)
		m.On("GetGameEndFlag").Return(false)
		m.On("IsHumanTurn").Return(false)
		assert.Contains(t, p.HintOutput(m), "あなたの番ではありません")
	})

	t.Run("declines when the game is over", func(t *testing.T) {
		m := new(interfaces.MockPresidentGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("IsHumanTurn").Return(true)
		assert.Contains(t, p.HintOutput(m), "あなたの番ではありません")
	})
}

func TestPresidentCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PresidentCuiPresenter)
	players := makePresidentPlayersForPresenter()
	pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
	out := p.ActionLogOutput(pg)
	assert.NotEmpty(t, out)
}
