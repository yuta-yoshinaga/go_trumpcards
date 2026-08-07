package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func buildPlayedScopa(t *testing.T) *domain.Scopa {
	t.Helper()
	s := domain.NewDefaultScopa()
	s.Reset()
	return s
}

func TestScopaCuiPresenter_Output(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	if out := p.Output(s, nil); out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestScopaCuiPresenter_Error(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	out := p.Output(s, scopaAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type scopaAssertErr struct{}

func (scopaAssertErr) Error() string { return "boom" }

func TestScopaCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	build := func(handVal, handSuit int, table []*domain.Card) *domain.Scopa {
		s := domain.ScopaTestNew(domain.DefaultScopaConfig())
		s.ScopaTestSetPhase(domain.ScopaPhasePlayerTurn)
		s.ScopaTestSetCurrentTurn(0)
		s.GetPlayer(0).AddCard(domain.NewCard(handSuit, handVal, false))
		s.ScopaTestSetTable(table)
		return s
	}

	t.Run("scopa sweep", func(t *testing.T) {
		s := build(7, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignClover, 4, false),
		})
		if out := p.HintOutput(s); !strings.Contains(out, "スコパ") {
			t.Errorf("expected scopa hint, got: %s", out)
		}
	})

	t.Run("plain capture", func(t *testing.T) {
		s := build(7, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		})
		if out := p.HintOutput(s); !strings.Contains(out, "捕獲") {
			t.Errorf("expected capture hint, got: %s", out)
		}
	})

	t.Run("no capture", func(t *testing.T) {
		s := build(2, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		})
		if out := p.HintOutput(s); !strings.Contains(out, "捕獲できる手はありません") {
			t.Errorf("expected no-capture hint, got: %s", out)
		}
	})

	t.Run("none outside player turn", func(t *testing.T) {
		s := domain.ScopaTestNew(domain.DefaultScopaConfig())
		s.ScopaTestSetPhase(domain.ScopaPhaseRoundEnd)
		if out := p.HintOutput(s); !strings.Contains(out, "ヒントはありません") {
			t.Errorf("expected none hint, got: %s", out)
		}
	})
}

func TestScopaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	s := buildPlayedScopa(t)
	if out := p.ActionLogOutput(s); out == "" {
		t.Error("expected non-empty action log output")
	}
}

// **なぜその点数になったのかを CUI は一切出していなかった (#4756)。**Web は
// カルテ/デナリ/プリミエラ/セッテベッロごとに誰が取ったかを出している。
func TestScopaCuiPresenter_Breakdown(t *testing.T) {
	p := &presenter.ScopaCuiPresenter{}
	players := []*domain.ScopaPlayer{domain.NewScopaPlayer(true), domain.NewScopaPlayer(false)}
	withDetail := func(det *domain.ScopaScoreDetail) *interfaces.MockScopaGame {
		m := new(interfaces.MockScopaGame)
		m.On("GetLastRoundDetail").Return(det)
		m.On("GetPlayerCnt").Return(2).Maybe()
		m.On("GetPlayer", 0).Return(players[0]).Maybe()
		m.On("GetPlayer", 1).Return(players[1]).Maybe()
		m.On("GetTableCards").Return(([]*domain.Card)(nil)).Maybe()
		m.On("GetActions").Return(([]*domain.ScopaAction)(nil)).Maybe()
		m.On("GetHumanAction").Return((*domain.ScopaAction)(nil)).Maybe()
		m.On("GetCpuActions").Return(([]*domain.ScopaAction)(nil)).Maybe()
		m.On("GetPhase").Return(domain.ScopaPhasePlayerTurn).Maybe()
		m.On("GetGameEndFlag").Return(false).Maybe()
		m.On("GetCurrentTurn").Return(0).Maybe()
		m.On("GetDeckCount").Return(0).Maybe()
		m.On("GetRoundNumber").Return(1).Maybe()
		return m
	}

	t.Run("names who took each category and for how much", func(t *testing.T) {
		out := p.Output(withDetail(&domain.ScopaScoreDetail{
			Cards:         map[int]int{0: 21, 1: 19},
			Diamonds:      map[int]int{0: 6, 1: 4},
			Sevens:        map[int]int{0: 1, 1: 3},
			HasSetteBello: 1,
			Scopas:        map[int]int{0: 2},
			Gained:        map[int]int{},
		}), nil)
		assert.Contains(t, out, "カルテ")
		assert.Contains(t, out, "プリミエラ")
		assert.Contains(t, out, "セッテベッロ")
		assert.Contains(t, out, "スコパ")
	})

	// **同点は行を消さずに「なし」と書く。**行が消えると「誰かが取った」と読める。
	t.Run("says nobody took a tied category", func(t *testing.T) {
		out := p.Output(withDetail(&domain.ScopaScoreDetail{
			Cards:         map[int]int{0: 20, 1: 20},
			Diamonds:      map[int]int{0: 5, 1: 5},
			Sevens:        map[int]int{0: 2, 1: 2},
			HasSetteBello: -1,
			Scopas:        map[int]int{},
			Gained:        map[int]int{},
		}), nil)
		assert.Contains(t, out, "なし(同点)")
		assert.Contains(t, out, "カルテ")
	})

	// **スコパ0回のプレイヤーは並べない。**「× 0」は情報にならない。
	t.Run("lists sweeps only for players who made one", func(t *testing.T) {
		out := p.Output(withDetail(&domain.ScopaScoreDetail{
			Cards: map[int]int{0: 21, 1: 19}, Diamonds: map[int]int{0: 6},
			Sevens: map[int]int{0: 4}, HasSetteBello: 0,
			Scopas: map[int]int{0: 2, 1: 0}, Gained: map[int]int{},
		}), nil)
		assert.Equal(t, 1, strings.Count(out, "スコパ:"))
	})

	t.Run("shows nothing before the first round ends", func(t *testing.T) {
		assert.NotContains(t, p.Output(withDetail(nil), nil), "前ラウンドの内訳")
	})
}
