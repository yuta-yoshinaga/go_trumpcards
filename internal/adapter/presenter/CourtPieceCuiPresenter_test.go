//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newCourtPieceForCuiTest() *domain.CourtPiece {
	cp := domain.NewDefaultCourtPiece()
	cp.Reset()
	return cp
}

func TestCourtPieceCuiPresenter_Output_PhaseLabels(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CourtPieceCuiPresenter)

	t.Run("trump phase prompt", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseTrumpDeclaration)
		cp.SetCallerIdx(0)
		out := p.Output(cp, nil)
		assert.Contains(t, out, i18n.T("courtpiece.promptTrumpHelp"))
		assert.NotContains(t, out, "courtpiece.", "a raw i18n key reached the screen")
	})

	t.Run("play phase prompt", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetTrumpSuit(domain.CardDesignSpade)
		out := p.Output(cp, nil)
		assert.Contains(t, out, i18n.T("courtpiece.promptPlayHelp"))
		assert.NotContains(t, out, "courtpiece.", "a raw i18n key reached the screen")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseTrickEnd)
		out := p.Output(cp, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("round end prompt", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)
		out := p.Output(cp, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("error block included", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		out := p.Output(cp, errors.New("bad input"))
		assert.Contains(t, out, "bad input")
	})

	t.Run("game end banner", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseGameEnd)
		cp.SetTeamScore(0, domain.CourtPieceDefaultPointLimit)
		out := p.Output(cp, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("caller and player lines render with team labels", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetTrumpSuit(domain.CardDesignSpade)
		cp.SetCallerIdx(0)
		out := p.Output(cp, nil)
		// The caller line carries the caller's team, and one player line per seat
		// is rendered. Asserted through i18n rather than on the key names: the
		// previous version matched "courtpiece.callerLine", which only held
		// because the key had no translation (issue #5380).
		assert.Contains(t, out, i18n.Tf("courtpiece.callerLine", "name", "あなた", "team", "A"))
		assert.Contains(t, out, i18n.Tf("courtpiece.playerLine",
			"name", "あなた", "team", "A", "tricks", "0", "cards", "5"))
		assert.NotContains(t, out, "courtpiece.", "a raw i18n key reached the screen")
	})
}

// **Web は合法手をリング表示しているのに、CUI は素の一覧だけだった。**
// 番号を入力してエラーを踏むまで、マストフォローで何が出せるか分からない。
func TestCourtPieceCuiPresenter_MarksPlayableCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CourtPieceCuiPresenter)

	// 人間が持っているスートをリードさせると、そのスートだけが合法になる。
	setup := func(t *testing.T) (*domain.CourtPiece, int) {
		t.Helper()
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetCurrentPlayerIdx(0)
		human := cp.GetPlayer(0)
		if human.GetCardsSize() == 0 {
			t.Fatal("前提: 人間に手札が配られていること")
		}
		leadSuit := human.GetCard(0).GetDesign()
		cp.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(leadSuit, 5, false)},
		})
		return cp, leadSuit
	}

	t.Run("human play turn marks only the legal cards", func(t *testing.T) {
		cp, _ := setup(t)
		playable := cp.GetPlayableIndices(0)
		// **配りに賭けてはいない。**leadSuit は人間の1枚目のスートなので合法手は
		// 必ず1枚以上あり、「全部合法」になるのは13枚が同一スートの配りのときだけ
		// (P ≈ 6e-12)。実測でも 200 回中 0 回。この分岐は保険で、通常は必ず本体を通る。
		if len(playable) == 0 || len(playable) == cp.GetPlayer(0).GetCardsSize() {
			t.Skipf("13枚同一スートの配り (%d/%d) -- 目印の有無を区別できない",
				len(playable), cp.GetPlayer(0).GetCardsSize())
		}

		out := p.Output(cp, nil)
		assert.Contains(t, out, "*", "合法手には目印が付く")
		// 非合法手には付かない: 目印の数が合法手の数と一致する。
		assert.Equal(t, len(playable), strings.Count(out, presenter.CuiLegalMark),
			"目印の数が合法手の数と一致する")
	})

	// **目印を出さない側も踏む。**トランプ宣言中は制限そのものが決まっていない。
	t.Run("trump declaration phase leaves the hand unmarked", func(t *testing.T) {
		cp, _ := setup(t)
		cp.SetPhase(domain.CourtPiecePhaseTrumpDeclaration)

		assert.NotContains(t, p.Output(cp, nil), presenter.CuiLegalMark,
			"宣言中は目印を出さない")
	})

	t.Run("cpu turn leaves the human hand unmarked", func(t *testing.T) {
		cp, _ := setup(t)
		cp.SetCurrentPlayerIdx(1)

		assert.NotContains(t, p.Output(cp, nil), presenter.CuiLegalMark,
			"相手の手番では目印を出さない")
	})
}

func TestCourtPieceCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CourtPieceCuiPresenter)

	t.Run("trump hint", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhaseTrumpDeclaration)
		cp.SetCallerIdx(0)
		out := p.HintOutput(cp)
		assert.NotEmpty(t, out)
	})

	t.Run("play hint", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetTrumpSuit(domain.CardDesignSpade)
		cp.SetCurrentPlayerIdx(0)
		out := p.HintOutput(cp)
		assert.NotEmpty(t, out)
	})

	t.Run("no hint when out of turn", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetCurrentPlayerIdx(1) // CPU's turn
		out := p.HintOutput(cp)
		assert.Contains(t, out, i18n.T("courtpiece.hintNone"))
	})
}

func TestCourtPieceCuiPresenter_ActionLogOutput(t *testing.T) {
	cp := newCourtPieceForCuiTest()
	p := new(presenter.CourtPieceCuiPresenter)
	out := p.ActionLogOutput(cp)
	assert.NotNil(t, out)
}

// #5656: 13トリック総取り、または連続でラウンドを取ると Court ボーナスで +2 点
// 入る。Web はラウンド結果に roundResult.court を出しているのに、CUI は汎用の
// 「次のラウンドへ」しか出さず、**スコアだけが 2 動く理由が分からなかった**。
func TestCourtPieceCuiPresenter_RoundEndCallsOutTheCourtBonus(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.CourtPieceCuiPresenter)

	// **集計 (ScoreRound) は画面より後に走る。** 実際の呼び順は
	// ResolveTrick → Output → (nr を打ってから) ScoreRound なので、
	// テストも ScoreRound を呼ばずに描く。
	t.Run("announces a clean sweep", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetCallerIdx(0)
		for i := 0; i < 13; i++ {
			cp.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		}
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)

		out := p.Output(cp, nil)

		assert.Contains(t, out, i18n.T("courtpiece.roundEndCourt"))
		// 既存の案内文は残す。
		assert.Contains(t, out, i18n.T("courtpiece.promptRoundEndHelp"))
	})

	// **8-5 で勝った初回は Court ではない。**ここで出してしまうと +1 のラウンドに
	// +2 の説明が付く。
	t.Run("stays quiet on an ordinary round win", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetCallerIdx(0)
		for i := 0; i < 8; i++ {
			cp.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		}
		for i := 0; i < 5; i++ {
			cp.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		}
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)

		out := p.Output(cp, nil)

		assert.NotContains(t, out, i18n.T("courtpiece.roundEndCourt"))
	})

	// **前のラウンドの Court を持ち越さない。** 集計済みのフラグを読むと、
	// 直前が Court だったせいで今回の +1 のラウンドにも説明が付く。
	t.Run("does not carry over the previous round's Court", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetCallerIdx(0)
		// 1 ラウンド目: 13 トリック総取り = Court。ここで集計まで走らせる。
		for i := 0; i < 13; i++ {
			cp.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		}
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)
		cp.ScoreRound()
		require.True(t, cp.IsLastRoundCourt(), "前提: 1 ラウンド目は Court")

		// 2 ラウンド目: 相手チームが 8-5 で勝つ。連勝ではないので Court ではない。
		for _, seat := range []int{0, 1, 2, 3} {
			cp.GetPlayer(seat).ResetRound()
		}
		for i := 0; i < 8; i++ {
			cp.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		}
		for i := 0; i < 5; i++ {
			cp.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false)})
		}
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)

		assert.NotContains(t, p.Output(cp, nil), i18n.T("courtpiece.roundEndCourt"))
	})

	// **連勝は集計前でも読める。** 同じチームが 2 ラウンド続けて勝てば Court。
	t.Run("announces back-to-back wins before scoring", func(t *testing.T) {
		cp := newCourtPieceForCuiTest()
		cp.SetCallerIdx(0)
		for i := 0; i < 8; i++ {
			cp.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		}
		for i := 0; i < 5; i++ {
			cp.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		}
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)
		cp.ScoreRound() // 1 勝目 (Court ではない)
		require.False(t, cp.IsLastRoundCourt())

		for _, seat := range []int{0, 1, 2, 3} {
			cp.GetPlayer(seat).ResetRound()
		}
		for i := 0; i < 7; i++ {
			cp.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		}
		for i := 0; i < 6; i++ {
			cp.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 6, false)})
		}
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)

		assert.Contains(t, p.Output(cp, nil), i18n.T("courtpiece.roundEndCourt"))
	})
}
