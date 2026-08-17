//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ThreeCardInteractorIF スリーカードポーカーインタラクターインタフェース
type ThreeCardInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベット
	Bet(ante, pairPlus int) string
	// Rebet 直前のラウンドと同じ額で賭け直す
	Rebet() string
	// Play プレイ
	Play() string
	// Fold フォールド
	Fold() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋��を出力する
	ActionLog() string
}

// ThreeCardInteractor スリーカードポーカーインタラクタークラス
type ThreeCardInteractor struct {
	GameBase[interfaces.ThreeCardGame]
	tp presenter.ThreeCardPresenter
}

// NewThreeCardInteractor コンストラクタ
func NewThreeCardInteractor(tc interfaces.ThreeCardGame, tp presenter.ThreeCardPresenter) *ThreeCardInteractor {
	mustNotNil("ThreeCardInteractor", map[string]any{"tc": tc, "tp": tp})
	return &ThreeCardInteractor{
		GameBase: GameBase[interfaces.ThreeCardGame]{Game: tc},
		tp:       tp,
	}
}

// Reset ゲーム初期化
func (ti *ThreeCardInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.tp, ti.Game.Reset)
}

// Bet アンテベット
func (ti *ThreeCardInteractor) Bet(ante, pairPlus int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.Bet(ante, pairPlus) })
}

// Rebet 直前のラウンドと同じ額で賭け直す
func (ti *ThreeCardInteractor) Rebet() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Rebet)
}

// Play プレイ
func (ti *ThreeCardInteractor) Play() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Play)
}

// Fold フォールド
func (ti *ThreeCardInteractor) Fold() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Fold)
}

// ActionLog 棋譜を出力する
func (ti *ThreeCardInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// Hint ヒント取得
func (ti *ThreeCardInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// RestoreThreeCardInteractor deserialises JSON into a ThreeCardInteractor.
func RestoreThreeCardInteractor(data []byte, tp presenter.ThreeCardPresenter) (*ThreeCardInteractor, error) {
	return restoreAndBuild[domain.ThreeCard](data, func(g *domain.ThreeCard) *ThreeCardInteractor {
		return &ThreeCardInteractor{GameBase: GameBase[interfaces.ThreeCardGame]{Game: g}, tp: tp}
	})
}
