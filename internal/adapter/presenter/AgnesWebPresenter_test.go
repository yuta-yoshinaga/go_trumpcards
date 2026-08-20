//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newWebAgnes returns a real Agnes with deterministic state for web rendering.
func newWebAgnes() *domain.Agnes {
	a := domain.NewDefaultAgnes()
	a.Reset()
	a.SetBaseRank(7)

	var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
	// Col 0: a face-down card (-> card null) and a face-up card.
	tab[0] = []*domain.AgnesTableauCard{
		{Card: domain.NewCard(domain.CardDesignSpade, 10, false), FaceUp: false},
		{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true},
	}
	for i := 1; i < domain.AgnesTableauCnt; i++ {
		tab[i] = []*domain.AgnesTableauCard{
			{Card: domain.NewCard(domain.CardDesignClover, i+1, false), FaceUp: true},
		}
	}
	a.SetTableau(tab)

	var f [domain.AgnesFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	a.SetFoundation(f)
	a.SetStock([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	return a
}

func TestAgnesWebPresenter_Output(t *testing.T) {
	t.Run("playing state", func(t *testing.T) {
		a := newWebAgnes()
		p := new(AgnesWebPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, `"baseRank":7`)
		assert.Contains(t, result, `"stockCount":1`)
		assert.Contains(t, result, `"tableau"`)
		assert.Contains(t, result, `"foundation"`)
		assert.Contains(t, result, `"agnes.playing"`)
		// Face-down card -> card null with faceUp false.
		assert.Contains(t, result, `"card":null,"faceUp":false`)
	})

	t.Run("error", func(t *testing.T) {
		a := newWebAgnes()
		p := new(AgnesWebPresenter)
		result := p.Output(a, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		a := newWebAgnes()
		a.SetPhase(domain.AgnesPhaseGameClear)
		p := new(AgnesWebPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, "agnes.gameClear")
	})

	t.Run("game over", func(t *testing.T) {
		a := newWebAgnes()
		a.SetPhase(domain.AgnesPhaseGameOver)
		p := new(AgnesWebPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, "agnes.gameOver")
	})
}

// **受動ヒントは Output() に載る。**Agnes は実ドメインオブジェクトを使うので
// モックの差し替えが要らないぶん、ヒントを検証するテストが無いまま通っていた。
// **手詰まりの判定はドメインが持つ。** 以前は載せておらず、フロントが
// agnesHasLegalMove() で同じ規則を実装し直していた (#5601)。
func TestAgnesWebPresenter_OutputCarriesTheStalemateFlag(t *testing.T) {
	t.Run("stock left is never a stalemate", func(t *testing.T) {
		a := newWebAgnes() // 山札 1 枚 = 必ず配れる
		require.False(t, a.IsStalemate())
		assert.Contains(t, new(AgnesWebPresenter).Output(a, nil), `"isStalemate":false`)
	})

	t.Run("no stock and no move is a stalemate", func(t *testing.T) {
		a := newWebAgnes()
		a.SetStock(nil)
		// place できない盤面にする: 各列 1 枚で、組札にも積めない値。
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		for i := range tab {
			tab[i] = []*domain.AgnesTableauCard{
				{Card: domain.NewCard(domain.CardDesignSpade, 2, false), FaceUp: true},
			}
		}
		a.SetTableau(tab)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)

		// ドメインがそう言うことを先に確かめる。ここが false のままだと、
		// 出力の検査は「たまたま false」を見ているだけになる。
		require.True(t, a.IsStalemate())
		assert.Contains(t, new(AgnesWebPresenter).Output(a, nil), `"isStalemate":true`)
	})
}

func TestAgnesWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("while the game is playable", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true}}
		a.SetTableau(tab)

		result := new(AgnesWebPresenter).Output(a, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// **終局では探索を走らせない。**ゲートを消したら CI で捕まえる。
	t.Run("not once the game is over", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetPhase(domain.AgnesPhaseGameOver)

		result := new(AgnesWebPresenter).Output(a, nil)
		assert.NotContains(t, result, `"hint"`, "a finished game must not run the search")
	})
}

func TestAgnesWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true}}
		a.SetTableau(tab)
		p := new(AgnesWebPresenter)
		result := p.HintOutput(a)
		assert.Contains(t, result, `"agnes.hintAvailable"`)
	})

	t.Run("no hint", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetPhase(domain.AgnesPhaseGameOver)
		p := new(AgnesWebPresenter)
		result := p.HintOutput(a)
		assert.Contains(t, result, `"agnes.noHint"`)
	})
}

func TestAgnesWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		p := new(AgnesWebPresenter)
		_ = p.ActionLogOutput(a)
	})

	t.Run("after game over", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.GiveUp()
		p := new(AgnesWebPresenter)
		_ = p.ActionLogOutput(a)
	})
}
