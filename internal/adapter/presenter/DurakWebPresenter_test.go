//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestDurakWebPresenter_Output(t *testing.T) {
	p := new(presenter.DurakWebPresenter)

	setupGame := func() *domain.Durak {
		players := []*domain.DurakPlayer{
			domain.NewDurakPlayer(true),
			domain.NewDurakPlayer(false),
			domain.NewDurakPlayer(false),
			domain.NewDurakPlayer(false),
		}
		tc := domain.NewTrumpCardsShortDeck()
		d := domain.NewDurak(tc, players)
		d.Reset()
		return d
	}

	t.Run("success initial state", func(t *testing.T) {
		d := setupGame()
		result := p.Output(d, nil)
		assert.NotEmpty(t, result)

		var out controller.DurakWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(out.Players))
		assert.False(t, out.GameEndFlag)
		assert.Equal(t, -1, out.LoserIdx)
		assert.NotEmpty(t, out.TrumpSuit)
	})

	t.Run("with error", func(t *testing.T) {
		d := setupGame()
		result := p.Output(d, errors.New("test error"))

		var out controller.DurakWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, "test error", out.Message)
	})

	t.Run("game end human loses", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(0)
		d.SetPhase(domain.DurakPhaseGameEnd)

		result := p.Output(d, nil)
		var out controller.DurakWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.True(t, out.GameEndFlag)
		assert.Contains(t, out.Message, "ドゥラーク")
		assert.Equal(t, "durak.result", out.MessageCode)
	})

	t.Run("game end CPU loses", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(1)
		d.SetPhase(domain.DurakPhaseGameEnd)

		result := p.Output(d, nil)
		var out controller.DurakWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Contains(t, out.Message, "CPU 1")
	})

	t.Run("game end draw", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(-1)
		d.SetPhase(domain.DurakPhaseGameEnd)

		result := p.Output(d, nil)
		var out controller.DurakWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Contains(t, out.Message, "引き分け")
	})

	t.Run("with table pairs", func(t *testing.T) {
		d := setupGame()
		d.SetTablePairs([]*domain.DurakTablePair{
			{Attack: domain.NewCard(domain.CardDesignSpade, 7, false)},
			{Attack: domain.NewCard(domain.CardDesignHeart, 8, false), Defense: domain.NewCard(domain.CardDesignHeart, 10, false)},
		})

		result := p.Output(d, nil)
		var out controller.DurakWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(out.TablePairs))
		assert.Nil(t, out.TablePairs[0].Defense)
		assert.NotNil(t, out.TablePairs[1].Defense)
	})
}

func TestDurakWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.DurakWebPresenter)
	gameMock := new(interfaces.MockDurakGame)
	gameMock.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	gameMock.On("GetGameEndFlag").Return(false)

	result := p.ActionLogOutput(gameMock)
	assert.NotEmpty(t, result)
}

// **`command: "hint"` 専用のレスポンスも返す (#4740)。**Web はクライアント側でも
// 簡易ヒントを出すが、これはサーバー計算のもの。
func TestDurakWebPresenter_HintOutput(t *testing.T) {
	d := domain.NewDefaultDurak()
	d.Reset()

	out := new(presenter.DurakWebPresenter).HintOutput(d)
	assert.True(t, json.Valid([]byte(out)), "JSON として妥当")
	assert.Contains(t, out, `"players"`, "状態も一緒に返る")
}
