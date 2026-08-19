package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestTrenteEtQuaranteCuiPresenter_OutputBetPhase(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "チップ")
}

func TestTrenteEtQuaranteCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestTrenteEtQuaranteCuiPresenter_OutputResult(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Noir")
}

func TestTrenteEtQuaranteCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestTrenteEtQuaranteCuiPresenter_HintOutputNone(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	// After the round resolves (result phase) GetHint returns nil.
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestTrenteEtQuaranteCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultTrenteEtQuarante()
	g.Reset()
	if err := g.PlaceBet(domain.TrenteEtQuaranteBetNoir, 100); err != nil {
		t.Fatalf("PlaceBet: %v", err)
	}
	p := new(presenter.TrenteEtQuaranteCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// setupTeqCuiMock returns a mock parked in the Result phase with a refait.
func setupTeqCuiMock(refait bool) *interfaces.MockTrenteEtQuaranteGame {
	m := new(interfaces.MockTrenteEtQuaranteGame)
	m.On("GetPhase").Return(domain.TrenteEtQuarantePhaseResult)
	m.On("GetRoundNumber").Return(1)
	m.On("GetChips").Return(1000)
	m.On("GetRemainingDeck").Return(200)
	m.On("GetCurrentBet").Return(domain.TrenteEtQuaranteBetNoir)
	m.On("GetStake").Return(100)
	m.On("GetNoirRow").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	m.On("GetRougeRow").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 1, false)})
	m.On("GetNoirTotal").Return(31)
	m.On("GetRougeTotal").Return(31)
	m.On("GetWinningRow").Return(-1)
	m.On("GetFirstCardRed").Return(false)
	m.On("GetRefait").Return(refait)
	m.On("GetResult").Return(domain.TrenteEtQuaranteResultDraw)
	m.On("GetPayout").Return(-50)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

// #5696: Web は refait.intro/half/edge の3文でルフェを説明するのに、CUI は
// 「31 での Refait — ステークの半額が胴元に渡ります。」の1行だけだった。
// **半額を取られること自体は既に出ている**ので、足りないのは「32〜40 の同点は
// 全額返るのに 31 だけ違う」という、唯一のハウスエッジである理由のほう。
func TestTrenteEtQuaranteCuiPresenter_ExplainsRefait(t *testing.T) {
	p := new(presenter.TrenteEtQuaranteCuiPresenter)

	t.Run("explains why a 31 tie costs half", func(t *testing.T) {
		out := p.Output(setupTeqCuiMock(true), nil)

		assert.Contains(t, out, i18n.T("trenteetquarante.result.refait"))
		assert.Contains(t, out, i18n.T("trenteetquarante.result.refaitWhy"))
	})

	t.Run("stays quiet on an ordinary round", func(t *testing.T) {
		out := p.Output(setupTeqCuiMock(false), nil)

		assert.NotContains(t, out, i18n.T("trenteetquarante.result.refaitWhy"))
	})
}
