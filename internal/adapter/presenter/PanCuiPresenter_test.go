//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupPanCuiMock() (*interfaces.MockPanGame, []*domain.PanPlayer) {
	m := new(interfaces.MockPanGame)
	players := []*domain.PanPlayer{
		domain.NewPanPlayer(true),
		domain.NewPanPlayer(false),
		domain.NewPanPlayer(false),
	}
	m.On("GetRoundNumber").Return(1)
	m.On("GetTargetRounds").Return(3)
	m.On("GetDrawPileCount").Return(279)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PanPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestPanCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.PanCuiPresenter)

	t.Run("draw phase shows header and players", func(t *testing.T) {
		m, players := setupPanCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "パングインゲ")
		assert.Contains(t, result, "ラウンド: 1/3")
		assert.Contains(t, result, "山札: 279枚")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "手番: あなた")
		assert.Contains(t, result, "ドローフェーズ")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupPanCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札")
	})

	t.Run("play phase prompts include command examples and a meld note", func(t *testing.T) {
		m, _ := setupPanCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PanPhasePlay)
		result := p.Output(m, nil)
		assert.Contains(t, result, "プレイフェーズ")
		assert.Contains(t, result, "例: m 0 1 2")  // meld command format example
		assert.Contains(t, result, "例: lo 1 0 3") // layoff command format example
		assert.Contains(t, result, "valle")       // legal-meld / chip note
		assert.Contains(t, result, "最低3枚")        // min meld size note
	})

	t.Run("round end shows prompt", func(t *testing.T) {
		m, _ := setupPanCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PanPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupPanCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("player with melds shown", func(t *testing.T) {
		m, players := setupPanCuiMock()
		players[0].AddLaidMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "メルド[0]")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupPanCuiMock()
		result := p.Output(m, errors.New("oops"))
		assert.Contains(t, result, "oops")
	})
}

func TestPanCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PanCuiPresenter)
	m := new(interfaces.MockPanGame)
	m.On("GetGameEndFlag").Return(false)
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}

// **チップが動いた理由を出す。**バジェ (3/5/7 のセット) は各プレイヤーにチップを
// 配る特別ルールなのに、盤面のどのメルドがそれなのか出ていなかった (#4853)。
func TestPanCuiPresenter_MarksValleMelds(t *testing.T) {
	set := func(v int) []*domain.Card {
		return []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, v, false),
			domain.NewCard(domain.CardDesignHeart, v, false),
			domain.NewCard(domain.CardDesignClover, v, false),
		}
	}
	for _, v := range []int{3, 5, 7} {
		assert.True(t, domain.PanIsValleMeld(set(v)))
	}
	for _, v := range []int{1, 4, 6, 12} {
		assert.False(t, domain.PanIsValleMeld(set(v)))
	}
	// ラン (3-4-5) はバジェではない。
	run := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignSpade, 4, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
	}
	assert.False(t, domain.PanIsValleMeld(run))

	// 出力側: バジェにだけ印が付く。
	m, players := setupPanCuiMock()
	players[0].SetLaidMelds([][]*domain.Card{set(5), set(4), run})
	out := new(presenter.PanCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "★バジェ")
	assert.Equal(t, 1, strings.Count(out, "★バジェ"))
}
