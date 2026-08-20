//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupThirtyOneCuiMock() (*interfaces.MockThirtyOneGame, []*domain.ThirtyOnePlayer) {
	m := new(interfaces.MockThirtyOneGame)
	players := makeThirtyOnePlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(39)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ThirtyOnePhaseDraw)
	m.On("GetCpuKnockThreshold").Return(domain.ThirtyOneKnockThresholdNormal).Maybe()
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetKnockerIdx").Return(-1)
	m.On("GetThirtyOneIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestThirtyOneCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ThirtyOneCuiPresenter)

	t.Run("initial draw phase", func(t *testing.T) {
		m, players := setupThirtyOneCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "サーティワン")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 39")
		assert.Contains(t, result, "あなた: ライフ 3")
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "dd")
		assert.Contains(t, result, "k:") // knock help shown when no knocker
	})

	t.Run("human best-suit total shown in prompt", func(t *testing.T) {
		m, players := setupThirtyOneCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		assert.Contains(t, p.Output(m, nil), "あなたのベストスート合計:")
	})

	t.Run("human suit-score breakdown shown under the hand", func(t *testing.T) {
		m, players := setupThirtyOneCuiMock()
		// ♠A(11) + ♥5(5): best suit is spades.
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "スート別: ♠11 ♣0 ♥5 ♦0（ベスト: ♠）")
	})

	t.Run("knock notice shown when someone has knocked", func(t *testing.T) {
		m, _ := setupThirtyOneCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerIdx")
		m.On("GetKnockerIdx").Return(1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "最終ターン")
		// Knock help is no longer offered once a knock is in progress.
		assert.NotContains(t, result, "k:")
	})

	t.Run("discard top shown", func(t *testing.T) {
		m, _ := setupThirtyOneCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		assert.Contains(t, p.Output(m, nil), "捨て札: HEART 7")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupThirtyOneCuiMock()
		assert.Contains(t, p.Output(m, errors.New("invalid card index")), "invalid card index")
	})

	t.Run("eliminated player shown OUT", func(t *testing.T) {
		m, players := setupThirtyOneCuiMock()
		players[3].SetLives(-1)
		assert.Contains(t, p.Output(m, nil), "OUT")
	})

	t.Run("discard phase commands", func(t *testing.T) {
		m, _ := setupThirtyOneCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThirtyOnePhaseDiscard)
		m.On("GetCpuKnockThreshold").Return(domain.ThirtyOneKnockThresholdNormal).Maybe()
		result := p.Output(m, nil)
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "d <idx>")
	})

	t.Run("round end with thirty-one banner", func(t *testing.T) {
		m, _ := setupThirtyOneCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetThirtyOneIdx")
		m.On("GetPhase").Return(domain.ThirtyOnePhaseRoundEnd)
		m.On("GetCpuKnockThreshold").Return(domain.ThirtyOneKnockThresholdNormal).Maybe()
		m.On("GetThirtyOneIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "31を達成")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("game ended human winner", func(t *testing.T) {
		m, _ := setupThirtyOneCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		assert.Contains(t, p.Output(m, nil), "ゲーム終了")
	})
}

func TestThirtyOneCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ThirtyOneCuiPresenter)

	m := new(interfaces.MockThirtyOneGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "knock", Detail: "You knock"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewThirtyOnePlayer(true)).Maybe()
	assert.Contains(t, p.ActionLogOutput(m), "knock")
}

// **ネイティブ CUI にヒントが無かった (#4806)。**Web はクライアント側ヒントを
// 持っていたが、go run ./cmd/trumpcards thirtyone では何の補助も無かった。
func TestThirtyOneCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ThirtyOneCuiPresenter)

	t.Run("recommends a discard with its index", func(t *testing.T) {
		g := domain.NewDefaultThirtyOne()
		g.Reset()
		g.SetPhase(domain.ThirtyOnePhaseDiscard)
		g.SetCurrentPlayerIdx(0)

		out := p.HintOutput(g)
		assert.Contains(t, out, "[HINT]")
		assert.Contains(t, out, "捨てましょう")
	})

	t.Run("recommends an action in the draw phase", func(t *testing.T) {
		g := domain.NewDefaultThirtyOne()
		g.Reset()
		g.SetPhase(domain.ThirtyOnePhaseDraw)
		g.SetCurrentPlayerIdx(0)

		out := p.HintOutput(g)
		assert.Contains(t, out, "[HINT]")
		// ドロー / ノックのいずれか。生の識別子が漏れていないこと。
		assert.NotContains(t, out, "draw_stock")
		assert.NotContains(t, out, "knock_ready")
	})

	t.Run("says nothing on a CPU turn", func(t *testing.T) {
		g := domain.NewDefaultThirtyOne()
		g.Reset()
		g.SetCurrentPlayerIdx(1)
		assert.Contains(t, p.HintOutput(g), i18n.T("thirtyone.hintNone"))
	})
}
