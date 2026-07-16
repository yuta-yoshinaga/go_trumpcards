//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HighCardFlushInteractorIF ハイカードフラッシュインタラクターインタフェース
type HighCardFlushInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテ＋オプションサイドベット
	Bet(ante, flushBonus, straightFlush int) string
	// Raise レイズ
	Raise(multiplier int) string
	// Fold フォールド
	Fold() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// HighCardFlushInteractor ハイカードフラッシュインタラクタークラス
type HighCardFlushInteractor struct {
	GameBase[interfaces.HighCardFlushGame]
	hp presenter.HighCardFlushPresenter
}

// NewHighCardFlushInteractor コンストラクタ
func NewHighCardFlushInteractor(hcf interfaces.HighCardFlushGame, hp presenter.HighCardFlushPresenter) *HighCardFlushInteractor {
	mustNotNil("HighCardFlushInteractor", map[string]any{"hcf": hcf, "hp": hp})
	return &HighCardFlushInteractor{
		GameBase: GameBase[interfaces.HighCardFlushGame]{Game: hcf},
		hp:       hp,
	}
}

// Reset ゲーム初期化
func (hi *HighCardFlushInteractor) Reset() string {
	return runAndPresent(hi.Game, hi.hp, hi.Game.Reset)
}

// Bet アンテ＋オプションサイドベット
func (hi *HighCardFlushInteractor) Bet(ante, flushBonus, straightFlush int) string {
	return execAndPresent(hi.Game, hi.hp, func() error { return hi.Game.Bet(ante, flushBonus, straightFlush) })
}

// Raise レイズ
func (hi *HighCardFlushInteractor) Raise(multiplier int) string {
	return execAndPresent(hi.Game, hi.hp, func() error { return hi.Game.Raise(multiplier) })
}

// Fold フォールド
func (hi *HighCardFlushInteractor) Fold() string {
	return execAndPresent(hi.Game, hi.hp, hi.Game.Fold)
}

// ActionLog 棋譜を出力する
func (hi *HighCardFlushInteractor) ActionLog() string {
	return hi.hp.ActionLogOutput(hi.Game)
}

// Hint ヒントを出力する
func (hi *HighCardFlushInteractor) Hint() string {
	return hi.hp.HintOutput(hi.Game)
}

// RestoreHighCardFlushInteractor deserialises JSON into a HighCardFlushInteractor.
func RestoreHighCardFlushInteractor(data []byte, hp presenter.HighCardFlushPresenter) (*HighCardFlushInteractor, error) {
	return restoreAndBuild[domain.HighCardFlush](data, func(g *domain.HighCardFlush) *HighCardFlushInteractor {
		return &HighCardFlushInteractor{GameBase: GameBase[interfaces.HighCardFlushGame]{Game: g}, hp: hp}
	})
}
