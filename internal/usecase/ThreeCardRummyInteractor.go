//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ThreeCardRummyInteractorIF スリーカード・ラミーインタラクターインタフェース
type ThreeCardRummyInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベット
	Bet(ante, lowBonus int) string
	// Rebet 直前のラウンドと同じ額で賭け直す
	Rebet() string
	// Play プレイ
	Play() string
	// Fold フォールド
	Fold() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ThreeCardRummyInteractor スリーカード・ラミーインタラクタークラス
type ThreeCardRummyInteractor struct {
	GameBase[interfaces.ThreeCardRummyGame]
	tp presenter.ThreeCardRummyPresenter
}

// NewThreeCardRummyInteractor コンストラクタ
func NewThreeCardRummyInteractor(tc interfaces.ThreeCardRummyGame, tp presenter.ThreeCardRummyPresenter) *ThreeCardRummyInteractor {
	mustNotNil("ThreeCardRummyInteractor", map[string]any{"tc": tc, "tp": tp})
	return &ThreeCardRummyInteractor{
		GameBase: GameBase[interfaces.ThreeCardRummyGame]{Game: tc},
		tp:       tp,
	}
}

// Reset ゲーム初期化
func (ti *ThreeCardRummyInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.tp, ti.Game.Reset)
}

// Bet アンテベット
func (ti *ThreeCardRummyInteractor) Bet(ante, lowBonus int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.Bet(ante, lowBonus) })
}

// Rebet 直前のラウンドと同じ額で賭け直す
func (ti *ThreeCardRummyInteractor) Rebet() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Rebet)
}

// Play プレイ
func (ti *ThreeCardRummyInteractor) Play() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Play)
}

// Fold フォールド
func (ti *ThreeCardRummyInteractor) Fold() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Fold)
}

// ActionLog 棋譜を出力する
func (ti *ThreeCardRummyInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// Hint ヒント取得
func (ti *ThreeCardRummyInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// RestoreThreeCardRummyInteractor deserialises JSON into a ThreeCardRummyInteractor.
func RestoreThreeCardRummyInteractor(data []byte, tp presenter.ThreeCardRummyPresenter) (*ThreeCardRummyInteractor, error) {
	return restoreAndBuild[domain.ThreeCardRummy](data, func(g *domain.ThreeCardRummy) *ThreeCardRummyInteractor {
		return &ThreeCardRummyInteractor{GameBase: GameBase[interfaces.ThreeCardRummyGame]{Game: g}, tp: tp}
	})
}
