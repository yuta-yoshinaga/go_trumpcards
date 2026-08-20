package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func escobaBuildCuiGame(t *testing.T) *domain.Escoba {
	t.Helper()
	e := domain.NewDefaultEscoba()
	e.Reset()
	return e
}

func TestEscobaCuiPresenter_HintOutput(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	build := func(handVal, handSuit int, table []*domain.Card) *domain.Escoba {
		e := domain.NewDefaultEscoba()
		e.SetPhase(domain.EscobaPhasePlayerTurn)
		e.SetCurrentTurn(0)
		e.GetPlayer(0).AddCard(domain.NewCard(handSuit, handVal, false))
		e.SetTableCards(table)
		return e
	}

	t.Run("escoba sweep", func(t *testing.T) {
		// played 5 (target 10) + table 3+7 = 10, clears the table.
		e := build(5, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		})
		if out := p.HintOutput(e); !strings.Contains(out, "エスコバ") {
			t.Errorf("expected escoba hint, got: %s", out)
		}
	})

	t.Run("plain capture", func(t *testing.T) {
		// played 5 (target 10) + table 3+7 = 10, but a 9 remains on the table.
		e := build(5, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		})
		if out := p.HintOutput(e); !strings.Contains(out, "捕獲") {
			t.Errorf("expected capture hint, got: %s", out)
		}
	})

	t.Run("no capture", func(t *testing.T) {
		// played 2 (target 13); no subset of {3,7} sums to 13.
		e := build(2, domain.CardDesignHeart, []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		})
		if out := p.HintOutput(e); !strings.Contains(out, "捕獲できる手はありません") {
			t.Errorf("expected no-capture hint, got: %s", out)
		}
	})

	t.Run("none outside player turn", func(t *testing.T) {
		e := domain.NewDefaultEscoba()
		e.SetPhase(domain.EscobaPhaseRoundEnd)
		if out := p.HintOutput(e); !strings.Contains(out, "ヒントはありません") {
			t.Errorf("expected none hint, got: %s", out)
		}
	})
}

func TestEscobaCuiPresenter_Output(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	if out := p.Output(e, nil); out == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestEscobaCuiPresenter_Error(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	out := p.Output(e, escobaAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type escobaAssertErr struct{}

func (escobaAssertErr) Error() string { return "boom" }

func TestEscobaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaBuildCuiGame(t)
	if out := p.ActionLogOutput(e); out == "" {
		t.Error("expected non-empty action log output")
	}
}

func TestEscobaCuiPresenter_OutputGameEnd(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}
	e := escobaPlayedOutGame(t) // all-CPU game driven to game end (defined in the web presenter test)
	out := p.Output(e, nil)
	if out == "" {
		t.Fatal("expected non-empty output at game end")
	}
}

// #5662: Web は captured-viewer で獲得札の実カードをいつでも開けるのに、CUI は
// 枚数の数字だけだった。「7 は取れているか」「espadas は何枚か」は得点計算に
// 直結するのに、数字からは読み取れない。
func TestEscobaCuiPresenter_ListsTheCapturedCards(t *testing.T) {
	p := &presenter.EscobaCuiPresenter{}

	t.Run("lists the human's captured cards", func(t *testing.T) {
		e := escobaBuildCuiGame(t)
		captured := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignClover, 1, false),
		}
		e.GetPlayer(0).AddCaptured(captured)

		out := p.Output(e, nil)

		// 色付けは cuiCardStr の担当。スート名と数字だけを見る。
		prefix, _, ok := strings.Cut(i18n.Tf("escoba.capturedLine", "cards", "\x00"), "\x00")
		require.True(t, ok)
		assert.Contains(t, out, prefix)
		assert.Contains(t, out, "SPADE 7")
		assert.Contains(t, out, "CLOVER 1")
		_ = captured
	})

	// 取り札が増えたら表示も増える (ラウンド途中でも都度更新される)。
	t.Run("keeps up as more cards are captured", func(t *testing.T) {
		e := escobaBuildCuiGame(t)
		e.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})
		before := p.Output(e, nil)

		e.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})
		after := p.Output(e, nil)

		assert.NotEqual(t, before, after)
		assert.NotContains(t, before, "HEART 3")
		assert.Contains(t, after, "HEART 3")
		assert.Contains(t, after, "SPADE 7", "先に取った札も残る")
	})

	// **1枚も取っていないうちは出さない。**空の一覧は「取れていない」と
	// 「表示が壊れている」の区別が付かない。
	t.Run("says nothing before anything is captured", func(t *testing.T) {
		e := escobaBuildCuiGame(t)

		out := p.Output(e, nil)

		prefix, _, ok := strings.Cut(i18n.Tf("escoba.capturedLine", "cards", "\x00"), "\x00")
		require.True(t, ok)
		require.NotEmpty(t, strings.TrimSpace(prefix))
		assert.NotContains(t, out, prefix)
	})

	// CPU の獲得札は伏せたまま (手札と同じ扱い)。
	t.Run("does not reveal a CPU's captured cards", func(t *testing.T) {
		e := escobaBuildCuiGame(t)
		e.GetPlayer(1).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)})

		out := p.Output(e, nil)

		prefix, _, ok := strings.Cut(i18n.Tf("escoba.capturedLine", "cards", "\x00"), "\x00")
		require.True(t, ok)
		assert.NotContains(t, out, prefix)
	})
}
