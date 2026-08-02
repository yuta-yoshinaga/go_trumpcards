//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newCourtPieceForWebTest() *domain.CourtPiece {
	cp := domain.NewDefaultCourtPiece()
	cp.Reset()
	return cp
}

type cpWebOutPartial struct {
	Phase       int               `json:"phase"`
	TrumpSuit   int               `json:"trumpSuit"`
	CallerIdx   int               `json:"callerIdx"`
	TeamScores  []int             `json:"teamScores"`
	Message     string            `json:"message,omitempty"`
	MessageCode string            `json:"messageCode,omitempty"`
	Players     []json.RawMessage `json:"players"`
	GameEndFlag bool              `json:"gameEndFlag"`
	WinnerTeam  int               `json:"winnerTeam"`
}

func TestCourtPieceWebPresenter_Output(t *testing.T) {
	p := new(presenter.CourtPieceWebPresenter)

	t.Run("default trump declaration phase", func(t *testing.T) {
		cp := newCourtPieceForWebTest()
		raw := p.Output(cp, nil)
		var got cpWebOutPartial
		require.NoError(t, json.Unmarshal([]byte(raw), &got))
		assert.Equal(t, int(domain.CourtPiecePhaseTrumpDeclaration), got.Phase)
		assert.Equal(t, "courtpiece.trumpPhase", got.MessageCode)
		assert.Equal(t, []int{0, 0}, got.TeamScores)
		assert.False(t, got.GameEndFlag)
	})

	t.Run("play phase lead vs follow", func(t *testing.T) {
		cp := newCourtPieceForWebTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		var got cpWebOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(cp, nil)), &got))
		assert.Equal(t, "courtpiece.playPhase.lead", got.MessageCode)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		cp := newCourtPieceForWebTest()
		cp.SetPhase(domain.CourtPiecePhasePlay)
		cp.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
		})
		var got cpWebOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(cp, nil)), &got))
		assert.Equal(t, "courtpiece.playPhase.follow", got.MessageCode)
		// 4 players always in a default Court Piece game.
		assert.Equal(t, 4, len(got.Players))
	})

	t.Run("error returned as message", func(t *testing.T) {
		cp := newCourtPieceForWebTest()
		raw := p.Output(cp, errors.New("boom"))
		var got cpWebOutPartial
		require.NoError(t, json.Unmarshal([]byte(raw), &got))
		assert.Equal(t, "boom", got.Message)
	})

	t.Run("trick end + round end", func(t *testing.T) {
		cp := newCourtPieceForWebTest()
		cp.SetPhase(domain.CourtPiecePhaseTrickEnd)
		var got cpWebOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(cp, nil)), &got))
		assert.Equal(t, "courtpiece.trickEnd", got.MessageCode)

		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)
		require.NoError(t, json.Unmarshal([]byte(p.Output(cp, nil)), &got))
		assert.Equal(t, "courtpiece.roundEnd", got.MessageCode)
	})

	t.Run("game end human-team win", func(t *testing.T) {
		cp := newCourtPieceForWebTest()
		// Bring team 0 to one point below the limit, then have team 0 win the round.
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)
		cp.SetTeamScore(0, domain.CourtPieceDefaultPointLimit-1)
		cp.SetTrickNumber(domain.CourtPieceHandSize)
		// Human is at idx 0 (team 0). Give team 0 the majority of tricks.
		for i := 0; i < 7; i++ {
			cp.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		}
		for i := 0; i < 6; i++ {
			cp.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		}
		cp.ScoreRound()
		require.True(t, cp.GetGameEndFlag())
		var got cpWebOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(cp, nil)), &got))
		assert.Equal(t, "courtpiece.gameEndHumanWin", got.MessageCode)
		assert.True(t, got.GameEndFlag)
		assert.Equal(t, 0, got.WinnerTeam)
	})

	t.Run("game end CPU-team win", func(t *testing.T) {
		cp := newCourtPieceForWebTest()
		cp.SetPhase(domain.CourtPiecePhaseRoundEnd)
		cp.SetTeamScore(1, domain.CourtPieceDefaultPointLimit-1)
		cp.SetTrickNumber(domain.CourtPieceHandSize)
		// CPU team (1) takes the majority of tricks.
		for i := 0; i < 7; i++ {
			cp.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		}
		for i := 0; i < 6; i++ {
			cp.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		}
		cp.ScoreRound()
		require.True(t, cp.GetGameEndFlag())
		var got cpWebOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(cp, nil)), &got))
		assert.Equal(t, "courtpiece.gameEndCpuWin", got.MessageCode)
		assert.Equal(t, 1, got.WinnerTeam)
	})
}

func TestCourtPieceWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CourtPieceWebPresenter)
	cp := newCourtPieceForWebTest()
	cp.SetPhase(domain.CourtPiecePhaseTrumpDeclaration)
	cp.SetCallerIdx(0)
	raw := p.HintOutput(cp)
	assert.Contains(t, raw, "hint")
}

func TestCourtPieceWebPresenter_HintOutput_Empty(t *testing.T) {
	p := new(presenter.CourtPieceWebPresenter)
	cp := newCourtPieceForWebTest()
	cp.SetPhase(domain.CourtPiecePhasePlay)
	cp.SetCurrentPlayerIdx(1) // not the human's turn
	raw := p.HintOutput(cp)
	// no hint should be emitted but the structure should still be valid JSON
	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(raw), &got))
	_, hasHint := got["hint"]
	assert.False(t, hasHint)
}

func TestCourtPieceWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CourtPieceWebPresenter)
	cp := newCourtPieceForWebTest()
	out := p.ActionLogOutput(cp)
	assert.NotEmpty(t, out)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。CourtPiece.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestCourtPieceWebPresenterOutputCarriesTheHint(t *testing.T) {
	cp := newCourtPieceForWebTest()
	cp.SetPhase(domain.CourtPiecePhaseTrumpDeclaration)
	cp.SetCallerIdx(0)

	result := new(presenter.CourtPieceWebPresenter).Output(cp, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
// ページは `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
// 付いていないとヒントを押しても画面に何も出ない。
func TestCourtPieceWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultCourtPiece()
	g.Reset()
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")
	assert.Contains(t, new(presenter.CourtPieceWebPresenter).HintOutput(g), "courtPiece.hintRequested")
}
