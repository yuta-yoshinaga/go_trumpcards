//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// buraTestGame returns a dealt game with a known trump so the fixtures below
// do not depend on the shuffle.
func buraTestGame(t *testing.T) *domain.Bura {
	t.Helper()
	b := domain.NewDefaultBura()
	b.Reset()
	return b
}

func buraDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

func TestBuraWebPresenter_HidesTheCpuHandButKeepsItsCount(t *testing.T) {
	// Workers hand this JSON straight to the browser: anything not dropped
	// here is the opponent's hand on the wire.
	p := new(BuraWebPresenter)
	out := buraDecode(t, p.Output(buraTestGame(t), nil))

	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, 2)

	human := players[0].(map[string]any)
	assert.Equal(t, false, human["hidden"])
	assert.Len(t, human["cards"], domain.BuraHandSize)

	cpu := players[1].(map[string]any)
	assert.Equal(t, true, cpu["hidden"])
	assert.Empty(t, cpu["cards"], "the CPU's cards must not reach the client")
	assert.Equal(t, float64(domain.BuraHandSize), cpu["cardCount"],
		"the count is public -- without it the UI cannot draw the right number of backs")
}

func TestBuraWebPresenter_RevealsEveryHandOnceTheRoundIsOver(t *testing.T) {
	b := buraTestGame(t)
	require.NoError(t, b.Claim(0)) // a short claim ends the round

	out := buraDecode(t, new(BuraWebPresenter).Output(b, nil))
	players := out["players"].([]any)
	cpu := players[1].(map[string]any)
	assert.Equal(t, false, cpu["hidden"])
	assert.Len(t, cpu["cards"], domain.BuraHandSize)
}

func TestBuraWebPresenter_ReportsTheOutcomeCodes(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*domain.Bura)
		want  string
	}{
		{
			name:  "a true claim wins",
			setup: func(b *domain.Bura) { b.SetPlayerPoints(0, domain.BuraWinThreshold); _ = b.Claim(0) },
			want:  "bura.win",
		},
		{
			name:  "a short claim loses",
			setup: func(b *domain.Bura) { _ = b.Claim(0) },
			want:  "bura.lose",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := buraTestGame(t)
			tc.setup(b)
			out := buraDecode(t, new(BuraWebPresenter).Output(b, nil))
			assert.Equal(t, tc.want, out["messageCode"])
		})
	}
}

func TestBuraWebPresenter_SendsNoMessageWhileTheRoundRuns(t *testing.T) {
	out := buraDecode(t, new(BuraWebPresenter).Output(buraTestGame(t), nil))
	assert.Empty(t, out["messageCode"])
}

func TestBuraWebPresenter_SurfacesAnError(t *testing.T) {
	out := buraDecode(t, new(BuraWebPresenter).Output(buraTestGame(t), assert.AnError))
	assert.Equal(t, assert.AnError.Error(), out["message"])
	assert.Empty(t, out["messageCode"])
}

func TestBuraWebPresenter_ShipsTheHintOnAnOrdinaryResponse(t *testing.T) {
	// Every other game sets Hint only in HintOutput, which no page calls, so
	// their hint toggles show nothing. Bura's arrives on the state response.
	out := buraDecode(t, new(BuraWebPresenter).Output(buraTestGame(t), nil))
	hint, ok := out["hint"].(map[string]any)
	require.True(t, ok, "the hint must ride along with ordinary state")
	assert.NotEmpty(t, hint["reason"])
}

func TestBuraWebPresenter_OmitsTheHintOnceTheRoundIsOver(t *testing.T) {
	b := buraTestGame(t)
	require.NoError(t, b.Claim(0))
	out := buraDecode(t, new(BuraWebPresenter).Output(b, nil))
	assert.Nil(t, out["hint"])
}

func TestBuraWebPresenter_HintOutputCarriesTheSameSuggestion(t *testing.T) {
	// A fresh deal can hold a combination, in which case the suggestion is to
	// declare rather than to lead -- so this cannot assert one fixed reason
	// without becoming a shuffle-dependent flake. What must always hold is the
	// pairing: a card suggestion carries indices, a declare/claim does not.
	for range 200 {
		b := buraTestGame(t)
		out := buraDecode(t, new(BuraWebPresenter).HintOutput(b))
		hint, ok := out["hint"].(map[string]any)
		require.True(t, ok)

		switch hint["reason"] {
		case "bura.hint.lead", "bura.hint.respond":
			assert.NotEmpty(t, hint["cardIndices"], "a card suggestion must name the cards")
		case "bura.hint.declare":
			// **役はどれも手札 3 枚すべてで成る。**一番重要な瞬間に何も
			// 光らないのが元の不具合だった (#4909)。
			idx, ok := hint["cardIndices"].([]any)
			require.True(t, ok, "declaring must name the cards that form the combination")
			assert.Len(t, idx, domain.BuraHandSize)
		case "bura.hint.claim":
			// クレームは得点の主張で、特定の札の話ではない。
			assert.Empty(t, hint["cardIndices"], "a points claim is not about particular cards")
		default:
			t.Fatalf("unexpected hint reason on a live round: %v", hint["reason"])
		}
	}
}

func TestBuraWebPresenter_ReportsTheDrawDistinctly(t *testing.T) {
	// Play the round out card by card WITHOUT ever claiming or declaring --
	// that is precisely the state under test. Taking the CPU's suggestion
	// would sometimes end the round early on a claim and never reach a draw.
	b := domain.NewDefaultBura()
	b.Reset()
	for range 200 {
		if b.GetGameEndFlag() {
			break
		}
		idx := b.GetCurrentPlayerIdx()
		if idx < 0 {
			break
		}
		n := 1
		if lead := b.GetCurrentLead(); len(lead) > 0 {
			n = len(lead)
		}
		indices := make([]int, n)
		for i := range indices {
			indices[i] = i
		}
		if err := b.PlayCards(idx, indices); err != nil {
			// A multi-card lead can mix suits; a single card always works.
			if err := b.PlayCards(idx, []int{0}); err != nil {
				break
			}
		}
	}
	require.True(t, b.IsDraw(), "expected an unclaimed round to run out")

	out := buraDecode(t, new(BuraWebPresenter).Output(b, nil))
	assert.Equal(t, "bura.draw", out["messageCode"])
	assert.Equal(t, true, out["isDraw"])
	assert.Equal(t, float64(-1), out["winnerIdx"])
}

func TestBuraWebPresenter_ActionLogIsEmptyUntilTheRoundEnds(t *testing.T) {
	p := new(BuraWebPresenter)
	b := buraTestGame(t)
	assert.NotEmpty(t, p.ActionLogOutput(b))

	require.NoError(t, b.Claim(0))
	assert.NotEmpty(t, p.ActionLogOutput(b))
}

func TestBuraWebPresenter_HintDeclinesWhenItIsNotTheHumansTurn(t *testing.T) {
	b := buraTestGame(t)
	// Hand the turn to the CPU without ending the round.
	b.SetCurrentPlayerIdx(1)

	out := buraDecode(t, new(BuraWebPresenter).HintOutput(b))
	hint := out["hint"].(map[string]any)
	assert.Equal(t, "bura.hint.not_your_turn", hint["reason"])
	assert.Empty(t, hint["cardIndices"])

	// ...and the ordinary state response carries no hint at all in that state.
	state := buraDecode(t, new(BuraWebPresenter).Output(b, nil))
	assert.Nil(t, state["hint"])
}

func TestBuraWebPresenter_HintReportsAFinishedRound(t *testing.T) {
	b := buraTestGame(t)
	require.NoError(t, b.Claim(0))
	out := buraDecode(t, new(BuraWebPresenter).HintOutput(b))
	hint := out["hint"].(map[string]any)
	assert.Equal(t, "bura.hint.game_end", hint["reason"])
}

// **役があるときは必ず 3 枚すべてが挙がる。**ランダムな配りに任せると declare の
// 分岐を一度も踏まないまま通ることがあるので、役のある手札を直接組んで踏む (#4909)。
func TestBuraWebPresenter_DeclareHintNamesTheWholeHand(t *testing.T) {
	b := buraTestGame(t)
	// 全部切札 = ブラ。手札全体で成る役なので、挙がるのは 3 枚すべて。
	trump := b.GetTrumpSuit()
	p := b.GetPlayer(0)
	p.Reset()
	for _, v := range []int{1, 10, 13} {
		p.AddCard(domain.NewCard(trump, v, false))
	}
	hand := make([]*domain.Card, p.GetCardsSize())
	for i := range hand {
		hand[i] = p.GetCard(i)
	}
	require.Equal(t, domain.BuraCombinationBura, domain.BuraDetectCombination(hand, trump))

	out := buraDecode(t, new(BuraWebPresenter).HintOutput(b))
	hint, ok := out["hint"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "bura.hint.declare", hint["reason"])
	idx, ok := hint["cardIndices"].([]any)
	require.True(t, ok)
	assert.Len(t, idx, domain.BuraHandSize)
	assert.Equal(t, []any{float64(0), float64(1), float64(2)}, idx)
}
