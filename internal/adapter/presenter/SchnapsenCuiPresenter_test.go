//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupSchnapsenCuiMock(trumpCard *domain.Card) *interfaces.MockSchnapsenGame {
	m := new(interfaces.MockSchnapsenGame)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SchnapsenPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTrumpCard").Return(trumpCard)
	m.On("GetStockRemaining").Return(9)
	m.On("IsEndgame").Return(false)
	m.On("GetMarriageIndices", 0).Return([]int(nil))
	// 合法手の目印 (#5627)。既定は「制限なし」= 印を出さない状態。
	m.On("GetValidPlayIndices", mock.Anything).Return([]int(nil)).Maybe()
	m.On("GetWinnerIdx").Return(-1)
	return m
}

func setupSchnapsenCuiMockWithPlayers(trumpCard *domain.Card) (*interfaces.MockSchnapsenGame, []*domain.SchnapsenPlayer) {
	m := setupSchnapsenCuiMock(trumpCard)
	players := []*domain.SchnapsenPlayer{
		domain.NewSchnapsenPlayer(true),
		domain.NewSchnapsenPlayer(false),
	}
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayerPoints", 0).Return(18)
	m.On("GetPlayerPoints", 1).Return(5)
	return m, players
}

func TestSchnapsenCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.SchnapsenCuiPresenter)

	t.Run("initial state shows header, trump, points", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, players := setupSchnapsenCuiMockWithPlayers(trump)
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))

		out := p.Output(m, nil)
		assert.Contains(t, out, "Schnapsen")
		assert.Contains(t, out, "トリック: 1")
		assert.Contains(t, out, "山札: 9枚")
		assert.Contains(t, out, "得点: あなた=18  CPU=5")
		assert.Contains(t, out, "play <idx>")
	})

	t.Run("marriage hint surfaces when available", func(t *testing.T) {
		trump := domain.NewCard(domain.CardDesignSpade, 13, false)
		m, players := setupSchnapsenCuiMockWithPlayers(trump)
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 12, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMarriageIndices")
		m.On("GetMarriageIndices", 0).Return([]int{0, 1})

		out := p.Output(m, nil)
		assert.Contains(t, out, "マリアージュ宣言可能")
	})

	t.Run("trump card exhausted (endgame)", func(t *testing.T) {
		m, _ := setupSchnapsenCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsEndgame")
		m.On("IsEndgame").Return(true)
		out := p.Output(m, nil)
		assert.Contains(t, out, "第2フェーズ")
	})

	t.Run("error is rendered", func(t *testing.T) {
		m, _ := setupSchnapsenCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		out := p.Output(m, errors.New("kaboom"))
		assert.Contains(t, out, "kaboom")
	})

	t.Run("trick-end prompt", func(t *testing.T) {
		m, _ := setupSchnapsenCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SchnapsenPhaseTrickEnd)
		out := p.Output(m, nil)
		assert.Contains(t, out, "トリック終了")
	})

	t.Run("game end p0 banner", func(t *testing.T) {
		m, _ := setupSchnapsenCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.On("GetPlayerPoints", 0).Return(70)
		m.On("GetPlayerPoints", 1).Return(40)

		out := p.Output(m, nil)
		assert.Contains(t, out, "あなたの勝利")
		assert.Contains(t, out, "(70-40)")
	})

	t.Run("game end p1 banner", func(t *testing.T) {
		m, _ := setupSchnapsenCuiMockWithPlayers(nil)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerPoints")
		m.On("GetPlayerPoints", 0).Return(40)
		m.On("GetPlayerPoints", 1).Return(70)

		out := p.Output(m, nil)
		assert.Contains(t, out, "CPUの勝利")
	})
}

func TestSchnapsenCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SchnapsenCuiPresenter)

	t.Run("hint shows card and reason", func(t *testing.T) {
		m, players := setupSchnapsenCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.SchnapsenHint{CardIndex: &idx, Reason: "follow_cut"})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
		assert.Contains(t, out, "切り札でカット")
	})

	t.Run("marriage hint", func(t *testing.T) {
		m, players := setupSchnapsenCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 12, false))
		idx := 0
		m.On("GetHint").Return(&domain.SchnapsenHint{CardIndex: &idx, Reason: "marriage", IsMarriage: true})

		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
	})

	t.Run("hint nil falls back to hintNone", func(t *testing.T) {
		m, _ := setupSchnapsenCuiMockWithPlayers(nil)
		m.On("GetHint").Return((*domain.SchnapsenHint)(nil))
		out := p.HintOutput(m)
		assert.Contains(t, out, "ヒントはありません")
	})

	t.Run("hint with unknown reason falls back to shared lookup", func(t *testing.T) {
		m, players := setupSchnapsenCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		idx := 0
		m.On("GetHint").Return(&domain.SchnapsenHint{CardIndex: &idx, Reason: "unknown_reason"})
		out := p.HintOutput(m)
		assert.NotEmpty(t, out)
	})
}

func TestSchnapsenCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SchnapsenCuiPresenter)
	m, _ := setupSchnapsenCuiMockWithPlayers(nil)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotNil(t, out)
}

// #5627: 第2フェーズ (山札切れ) は「同スートを出す → 勝てるなら勝つ → 無ければ
// 切り札」という3段のマストフォローが効く、このゲームで最も間違えやすい局面。
// Web は合法手に緑のリングを出しているのに、CUI は素の一覧だけだった。
func TestSchnapsenCuiPresenterMarksTheLegalCardsInTheEndgame(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SchnapsenCuiPresenter)

	// **既定の期待値は先に消す。**testify は最初に一致した期待値を返すので、
	// あとから `m.On(...)` を足しても既定が勝ち、上書きしたつもりのケースが
	// 何も確かめない (実際この形で 2 ケースが空振りしていた)。
	setup := func(t *testing.T, endgame bool, valid []int, turn int) *interfaces.MockSchnapsenGame {
		t.Helper()
		m, players := setupSchnapsenCuiMockWithPlayers(domain.NewCard(domain.CardDesignSpade, 13, false))
		for _, c := range []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 10, false),
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignClover, 11, false),
		} {
			players[0].AddCard(c)
		}
		m.ExpectedCalls = filterSchnapsenCall(m.ExpectedCalls, "IsEndgame")
		m.ExpectedCalls = filterSchnapsenCall(m.ExpectedCalls, "GetValidPlayIndices")
		m.ExpectedCalls = filterSchnapsenCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("IsEndgame").Return(endgame)
		m.On("GetValidPlayIndices", mock.Anything).Return(valid)
		m.On("GetCurrentPlayerIdx").Return(turn)
		return m
	}

	t.Run("marks only the cards the follow rule allows", func(t *testing.T) {
		m := setup(t, true, []int{0, 1}, 0)

		out := p.Output(m, nil)
		assert.Equal(t, 2, strings.Count(out, presenter.CuiLegalMark), "合法手の数だけ印が付く")
	})

	// **第1フェーズは自由に出せる。**印を出すと「これ以外は出せない」と読める。
	t.Run("marks nothing in the first phase", func(t *testing.T) {
		m := setup(t, false, []int{0, 1}, 0)

		assert.NotContains(t, p.Output(m, nil), presenter.CuiLegalMark)
	})

	t.Run("marks nothing on the opponent's turn", func(t *testing.T) {
		m := setup(t, true, []int{0, 1}, 1)

		assert.NotContains(t, p.Output(m, nil), presenter.CuiLegalMark)
	})
}

// filterSchnapsenCall removes an existing expectation so a test can override it.
func filterSchnapsenCall(calls []*mock.Call, method string) []*mock.Call {
	out := calls[:0]
	for _, c := range calls {
		if c.Method != method {
			out = append(out, c)
		}
	}
	return out
}
