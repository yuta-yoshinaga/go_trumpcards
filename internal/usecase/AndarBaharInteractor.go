//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AndarBaharInteractorIF アンダーバハールインタラクターインタフェース
type AndarBaharInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ラウンド初期化
	Reset() string
	// Bet ベットして決着まで進める
	Bet(amount, target, sideAmount, sideBand int) string
	// ClearHistory 罫線履歴をクリアする
	ClearHistory() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// AndarBaharInteractor アンダーバハールインタラクタークラス
type AndarBaharInteractor struct {
	GameBase[interfaces.AndarBaharGame]
	ap presenter.AndarBaharPresenter
}

// NewAndarBaharInteractor コンストラクタ
func NewAndarBaharInteractor(ab interfaces.AndarBaharGame, ap presenter.AndarBaharPresenter) *AndarBaharInteractor {
	mustNotNil("AndarBaharInteractor", map[string]any{"ab": ab, "ap": ap})
	return &AndarBaharInteractor{
		GameBase: GameBase[interfaces.AndarBaharGame]{Game: ab},
		ap:       ap,
	}
}

// Reset ラウンド初期化
func (ai *AndarBaharInteractor) Reset() string {
	return runAndPresent(ai.Game, ai.ap, ai.Game.Reset)
}

// Bet ベットして決着まで自動進行
//
// **配布は一気に進みます。** 基準札と同ランクが出るまで交互に配るだけで、途中に
// 人間の判断は入りません。
func (ai *AndarBaharInteractor) Bet(amount, target, sideAmount, sideBand int) string {
	return execAndPresent(ai.Game, ai.ap, func() error {
		return ai.Game.Bet(amount, target, sideAmount, sideBand)
	})
}

// ClearHistory 罫線履歴をクリアする
func (ai *AndarBaharInteractor) ClearHistory() string {
	return runAndPresent(ai.Game, ai.ap, ai.Game.ClearHistory)
}

// Hint ヒント取得
func (ai *AndarBaharInteractor) Hint() string { return ai.ap.HintOutput(ai.Game) }

// ActionLog 棋譜を出力する
func (ai *AndarBaharInteractor) ActionLog() string { return ai.ap.ActionLogOutput(ai.Game) }

// RestoreAndarBaharInteractor deserialises JSON into an AndarBaharInteractor.
func RestoreAndarBaharInteractor(data []byte, ap presenter.AndarBaharPresenter) (*AndarBaharInteractor, error) {
	return restoreAndBuild[domain.AndarBahar](data, func(g *domain.AndarBahar) *AndarBaharInteractor {
		return NewAndarBaharInteractor(g, ap)
	})
}
