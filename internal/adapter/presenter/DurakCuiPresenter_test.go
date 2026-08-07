//go:build test

package presenter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestDurakCuiPresenter_Output(t *testing.T) {
	p := new(presenter.DurakCuiPresenter)

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

	t.Run("initial state", func(t *testing.T) {
		d := setupGame()
		result := p.Output(d, nil)
		assert.Contains(t, result, "Durak")
		assert.Contains(t, result, "切り札")
	})

	t.Run("with error", func(t *testing.T) {
		d := setupGame()
		result := p.Output(d, domain.ErrInvalidCard)
		assert.Contains(t, result, "invalid card")
	})

	t.Run("game end human loses", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(0)
		d.SetPhase(domain.DurakPhaseGameEnd)
		result := p.Output(d, nil)
		assert.Contains(t, result, "ドゥラーク")
	})

	t.Run("game end CPU loses", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(1)
		d.SetPhase(domain.DurakPhaseGameEnd)
		result := p.Output(d, nil)
		assert.Contains(t, result, "CPU 1")
	})

	t.Run("game end draw", func(t *testing.T) {
		d := setupGame()
		d.SetGameEndFlag(true)
		d.SetLoserIdx(-1)
		d.SetPhase(domain.DurakPhaseGameEnd)
		result := p.Output(d, nil)
		assert.Contains(t, result, "引き分け")
	})

	t.Run("with table pairs", func(t *testing.T) {
		d := setupGame()
		d.SetTablePairs([]*domain.DurakTablePair{
			{Attack: domain.NewCard(domain.CardDesignSpade, 7, false)},
		})
		result := p.Output(d, nil)
		assert.Contains(t, result, "テーブル")
	})

	t.Run("defend phase shows banner", func(t *testing.T) {
		d := setupGame()
		d.SetPhase(domain.DurakPhaseDefend)
		result := p.Output(d, nil)
		assert.Contains(t, result, "防御")
	})

	t.Run("defended pair on table", func(t *testing.T) {
		d := setupGame()
		d.SetTablePairs([]*domain.DurakTablePair{
			{
				Attack:  domain.NewCard(domain.CardDesignSpade, 7, false),
				Defense: domain.NewCard(domain.CardDesignSpade, 8, false),
			},
		})
		result := p.Output(d, nil)
		assert.Contains(t, result, "SPADE 7")
		assert.Contains(t, result, "SPADE 8")
	})

	t.Run("trump cards highlighted in human hand", func(t *testing.T) {
		origNoColor := color.NoColor()
		color.SetNoColor(false)
		defer color.SetNoColor(origNoColor)

		d := setupGame()
		d.SetTrumpSuit(domain.CardDesignHeart)
		human := d.GetPlayer(0)
		human.Reset()
		human.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false)) // trump
		human.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false)) // non-trump

		result := p.Output(d, nil)
		// The card text itself carries a suit color, so assert on the bold-yellow
		// opening code that wraps the trump card's index marker.
		boldStart := strings.Split(color.BoldYellow("X"), "X")[0]
		assert.Contains(t, result, boldStart+"[0]")    // trump card highlighted
		assert.NotContains(t, result, boldStart+"[1]") // non-trump left plain
		assert.Contains(t, result, "SPADE 7")
	})

	t.Run("finished player rendered", func(t *testing.T) {
		d := setupGame()
		// Mark CPU 1 as finished — exercises durakPlayerStr's finished branch.
		d.GetPlayer(1).SetIsFinished(true)
		result := p.Output(d, nil)
		assert.Contains(t, result, "上がり")
	})
}

func TestDurakCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.DurakCuiPresenter)
	gameMock := new(interfaces.MockDurakGame)
	gameMock.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	gameMock.On("GetGameEndFlag").Return(false)

	result := p.ActionLogOutput(gameMock)
	assert.NotEmpty(t, result)
}

// **他のトリック系はサーバー計算の理由付きヒントを持つのに、Durak は CUI に
// hint コマンドすら無かった (#4740)。**
func TestDurakCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.DurakCuiPresenter)

	newMock := func(hint *domain.DurakHint) *interfaces.MockDurakGame {
		m := new(interfaces.MockDurakGame)
		m.On("GetHint").Return(hint)
		m.On("GetCurrentTurn").Return(0).Maybe()
		pl := domain.NewDurakPlayer(true)
		pl.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		m.On("GetPlayer", 0).Return(pl).Maybe()
		return m
	}

	t.Run("attack hint names the card and the reason", func(t *testing.T) {
		idx := 0
		out := p.HintOutput(newMock(&domain.DurakHint{CardIndex: &idx, Reason: "attack_weakest"}))
		assert.Contains(t, out, "SPADE 6")
		assert.Contains(t, out, "最弱の非切り札で攻める")
	})

	t.Run("defend hint names which attack to beat", func(t *testing.T) {
		idx, atk := 0, 1
		out := p.HintOutput(newMock(&domain.DurakHint{CardIndex: &idx, AttackIdx: &atk, Reason: "defend_beat"}))
		assert.Contains(t, out, "SPADE 6")
		assert.Contains(t, out, "この札で返せる")
	})

	// **「引き取る」「パス」も助言のうち。**カードを勧められない局面で黙ると、
	// プレイヤーは手が無いのか判断が付かない。
	t.Run("take hint says to pick the cards up", func(t *testing.T) {
		out := p.HintOutput(newMock(&domain.DurakHint{TakeCards: true, Reason: "take_cannot_beat"}))
		assert.Contains(t, out, "返せる札が無い")
	})

	t.Run("pass hint says to pass", func(t *testing.T) {
		out := p.HintOutput(newMock(&domain.DurakHint{Reason: "pass_no_addition"}))
		assert.Contains(t, out, "追撃できる札が無い")
	})

	t.Run("no hint says so explicitly", func(t *testing.T) {
		assert.Contains(t, p.HintOutput(newMock(nil)), "ヒントはありません")
	})
}
