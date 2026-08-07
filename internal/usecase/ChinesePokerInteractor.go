//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ChinesePokerInteractorIF チャイニーズポーカーインタラクターインタフェース
type ChinesePokerInteractorIF interface {
	Snapshot() ([]byte, error)
	Reset() string
	Bet(amount int) string
	SetHands(frontIndices []int, middleIndices []int) string
	// Hint ヒント取得
	Hint() string
	ActionLog() string
}

// ChinesePokerInteractor チャイニーズポーカーインタラクター
type ChinesePokerInteractor struct {
	GameBase[interfaces.ChinesePokerGame]
	pp presenter.ChinesePokerPresenter
}

// NewChinesePokerInteractor コンストラクタ
func NewChinesePokerInteractor(cp interfaces.ChinesePokerGame, pp presenter.ChinesePokerPresenter) *ChinesePokerInteractor {
	mustNotNil("ChinesePokerInteractor", map[string]any{"cp": cp, "pp": pp})
	return &ChinesePokerInteractor{
		GameBase: GameBase[interfaces.ChinesePokerGame]{Game: cp},
		pp:       pp,
	}
}

// Reset ゲーム初期化
func (ci *ChinesePokerInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.pp, ci.Game.Reset)
}

// Bet ベット
func (ci *ChinesePokerInteractor) Bet(amount int) string {
	return execAndPresent(ci.Game, ci.pp, func() error { return ci.Game.Bet(amount) })
}

// SetHands ハンド設定
func (ci *ChinesePokerInteractor) SetHands(frontIndices []int, middleIndices []int) string {
	return execAndPresent(ci.Game, ci.pp, func() error { return ci.Game.SetHands(frontIndices, middleIndices) })
}

// Hint ヒント取得
func (ci *ChinesePokerInteractor) Hint() string {
	return ci.pp.HintOutput(ci.Game)
}

// ActionLog 棋譜出力
func (ci *ChinesePokerInteractor) ActionLog() string {
	return ci.pp.ActionLogOutput(ci.Game)
}

// RestoreChinesePokerInteractor JSON復元
func RestoreChinesePokerInteractor(data []byte, pp presenter.ChinesePokerPresenter) (*ChinesePokerInteractor, error) {
	return restoreAndBuild[domain.ChinesePoker](data, func(g *domain.ChinesePoker) *ChinesePokerInteractor {
		return &ChinesePokerInteractor{GameBase: GameBase[interfaces.ChinesePokerGame]{Game: g}, pp: pp}
	})
}
