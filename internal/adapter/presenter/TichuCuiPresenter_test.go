package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestTichuCuiPresenter_Output(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuCuiPresenter)

	declare := p.Output(tg, nil)
	assert.NotEmpty(t, declare)

	for tg.GetPhase() == domain.TichuPhaseDeclare {
		tg.CpuPlay()
	}
	play := p.Output(tg, nil)
	assert.Contains(t, play, "----------")

	for !tg.GetGameEndFlag() {
		tg.CpuPlay()
	}
	end := p.Output(tg, nil)
	assert.NotEmpty(t, end)

	withErr := p.Output(tg, errors.New("boom"))
	assert.True(t, strings.Contains(withErr, "boom"))
}

func TestTichuCuiPresenter_ActionLog(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(tg))
}

// **得点差もボムの使用状況も終局まで分からなかった。**Web は常時スコアバーに
// 出している (#4888)。
func TestTichuCuiPresenter_ShowsRunningScoreAndBombs(t *testing.T) {
	tg := newTichuAllCpu()
	tg.Reset()
	p := new(presenter.TichuCuiPresenter)

	// 宣言フェーズから既にスコアが出る。
	declare := p.Output(tg, nil)
	assert.Contains(t, declare, "チームA (P0/P2):")
	assert.Contains(t, declare, "チームB (P1/P3):")
	// まだボムは使われていないので行は出さない。
	assert.NotContains(t, declare, "ボム使用")

	for tg.GetPhase() == domain.TichuPhaseDeclare {
		tg.CpuPlay()
	}
	assert.Contains(t, p.Output(tg, nil), "チームA (P0/P2):")

	// ボムが使われたら回数が出る。
	tg.SetBombCountForTest(2)
	assert.Contains(t, p.Output(tg, nil), "ボム使用: 2回")

	// ワンツーが成立したらその旨も出る。
	tg.SetIsOneTwoForTest(true)
	assert.Contains(t, p.Output(tg, nil), "ワンツー成立")

	// **終局時に二重に出さない。**下の gameEnd ブロックが出す。
	for !tg.GetGameEndFlag() {
		tg.CpuPlay()
	}
	end := p.Output(tg, nil)
	assert.Equal(t, 1, strings.Count(end, "チームA (P0/P2):"))
	assert.Contains(t, end, "ディール終了")
}

// #5635: ボムは得点を左右する重要な役なのに、CUI は手札を無印で並べるだけで、
// どの札が構成できるのかは目視で数えるしかなかった。
func TestTichuCuiPresenterMarksTheBombCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TichuCuiPresenter)

	build := func(t *testing.T, cards []*domain.Card) *domain.Tichu {
		t.Helper()
		players := []*domain.TichuPlayer{
			domain.NewTichuPlayer(true),
			domain.NewTichuPlayer(false),
			domain.NewTichuPlayer(false),
			domain.NewTichuPlayer(false),
		}
		g := domain.NewTichu(domain.NewTrumpCards(domain.TichuJokerCount), players, domain.DefaultTichuConfig())
		for _, c := range cards {
			players[0].AddCard(c)
		}
		return g
	}

	t.Run("marks the four of a kind", func(t *testing.T) {
		g := build(t, []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignSpade, 2, false),
		})

		out := p.Output(g, nil)
		assert.Equal(t, 4, strings.Count(out, presenter.CuiBombMark), "ボムの4枚だけに印が付く")
	})

	t.Run("marks nothing without a bomb", func(t *testing.T) {
		g := build(t, []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignClover, 9, false),
		})

		assert.NotContains(t, p.Output(g, nil), presenter.CuiBombMark)
	})

	// **他の印と別の記号。**同じ画面で意味の違う印が同じ形だと読み分けられない。
	t.Run("uses a mark of its own", func(t *testing.T) {
		assert.NotEqual(t, presenter.CuiLegalMark, presenter.CuiBombMark)
		assert.NotEqual(t, presenter.CuiKittyMark, presenter.CuiBombMark)
	})
}
