//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestSeahavenTowersWebPresenterOutputPlaying(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	result := p.Output(s, nil)

	var out controller.SeahavenTowersWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, int(domain.SeahavenTowersPhasePlaying), out.Phase)
	assert.Equal(t, domain.SeahavenTowersTableauCnt, len(out.Tableau))
	assert.Equal(t, domain.SeahavenTowersCellCnt, len(out.ReservedCells))
	assert.Equal(t, domain.SeahavenTowersFoundationCnt, len(out.Foundation))
	assert.Equal(t, 0, out.MoveCount)
	assert.Equal(t, "seahaventowers.playing", out.MessageCode)
}

func TestSeahavenTowersWebPresenterOutputGameClear(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhaseGameClear)

	result := p.Output(s, nil)

	var out controller.SeahavenTowersWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "seahaventowers.gameClear", out.MessageCode)
	assert.NotNil(t, out.MessageParams)
}

func TestSeahavenTowersWebPresenterOutputGameOver(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhaseGameOver)

	result := p.Output(s, nil)

	var out controller.SeahavenTowersWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "seahaventowers.gameOver", out.MessageCode)
}

func TestSeahavenTowersWebPresenterOutputStalemate(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)
	s.SetIsStalemate(true)

	result := p.Output(s, nil)

	var out controller.SeahavenTowersWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, -1, out.UndoToEscape)
	assert.Equal(t, "seahaventowers.stalemate", out.MessageCode)
	assert.Empty(t, out.MessageParams)
}

func TestSeahavenTowersWebPresenterOutputStalemateWithEscape(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	sg := new(interfaces.MockSeahavenTowersGame)
	sg.On("GetPhase").Return(domain.SeahavenTowersPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(7).Maybe()
	sg.On("CanUndo").Return(true).Maybe()
	sg.On("IsStalemate").Return(true).Maybe()
	sg.On("UndoToEscape").Return(4).Maybe()
	var tableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	sg.On("GetTableau").Return(tableau).Maybe()
	var freeCells [domain.SeahavenTowersCellCnt]*domain.Card
	sg.On("GetFreeCells").Return(freeCells).Maybe()
	var foundation [domain.SeahavenTowersFoundationCnt][]*domain.Card
	sg.On("GetFoundation").Return(foundation).Maybe()

	result := p.Output(sg, nil)

	var out controller.SeahavenTowersWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, 4, out.UndoToEscape)
	assert.Equal(t, "seahaventowers.stalemateWithEscape", out.MessageCode)
	assert.Equal(t, "4", out.MessageParams["count"])
}

func TestSeahavenTowersWebPresenterOutputError(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()

	result := p.Output(s, errors.New("test error"))

	var out controller.SeahavenTowersWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Contains(t, out.Message, "test error")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestSeahavenTowersWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
		s.Reset()
		s.SetPhase(domain.SeahavenTowersPhasePlaying)

		// **配りに依存させない。**SetTableau で場を丸ごと置き換えるので、
		// 動かせる札が 1 枚だけ残り、ヒントが必ず出る。
		var tableau [domain.SeahavenTowersTableauCnt][]*domain.Card
		tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		s.SetTableau(tableau)

		var out controller.SeahavenTowersWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(SeahavenTowersWebPresenter).Output(s, nil)), &out))
		assert.NotNil(t, out.Hint, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not while cleared", func(t *testing.T) {
		s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
		s.Reset()
		s.SetPhase(domain.SeahavenTowersPhaseGameClear)

		var out controller.SeahavenTowersWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(SeahavenTowersWebPresenter).Output(s, nil)), &out))
		assert.Nil(t, out.Hint)
	})
}

func TestSeahavenTowersWebPresenterHintOutputWithHint(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	var tableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	s.SetTableau(tableau)

	result := p.HintOutput(s)

	var out controller.SeahavenTowersWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, "seahaventowers.hintAvailable", out.MessageCode)
}

func TestSeahavenTowersWebPresenterHintOutputNoHint(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhaseGameOver)

	result := p.HintOutput(s)

	var out controller.SeahavenTowersWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "seahaventowers.noHint", out.MessageCode)
}

func TestSeahavenTowersWebPresenterActionLogPlaying(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()
	s.SetPhase(domain.SeahavenTowersPhasePlaying)

	result := p.ActionLogOutput(s)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Empty(t, out.Entries)
}

func TestSeahavenTowersWebPresenterActionLogGameOver(t *testing.T) {
	p := new(SeahavenTowersWebPresenter)
	s := domain.NewSeahavenTowers(domain.NewTrumpCards(0))
	s.Reset()

	var tableau [domain.SeahavenTowersTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	s.SetTableau(tableau)
	s.SetPhase(domain.SeahavenTowersPhasePlaying)
	_ = s.MoveTableauToFoundation(0)
	s.SetPhase(domain.SeahavenTowersPhaseGameOver)

	result := p.ActionLogOutput(s)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotEmpty(t, out.Entries)
}
