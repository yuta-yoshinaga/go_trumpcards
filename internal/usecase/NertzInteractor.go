package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NertzInteractorIF Nertz / Pounce インタラクターインタフェース
type NertzInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を適用してゲーム初期化
	ResetWithConfig(cfg domain.NertzConfig) string
	// NextRound 次ラウンドを開始する
	NextRound() string
	// Draw 指定プレイヤーがストックを引く (人間プレイヤー専用エントリ)
	Draw(playerIdx int) string
	// MoveNertzToFoundation ナッツパイル → ファウンデーション
	MoveNertzToFoundation(playerIdx, foundationIdx int) string
	// MoveNertzToTableau ナッツパイル → タブロー
	MoveNertzToTableau(playerIdx, toCol int) string
	// MoveWasteToFoundation ウェイスト → ファウンデーション
	MoveWasteToFoundation(playerIdx, foundationIdx int) string
	// MoveWasteToTableau ウェイスト → タブロー
	MoveWasteToTableau(playerIdx, toCol int) string
	// MoveTableauToFoundation タブロー → ファウンデーション
	MoveTableauToFoundation(playerIdx, fromCol, foundationIdx int) string
	// MoveTableauToTableau タブロー → タブロー
	MoveTableauToTableau(playerIdx, fromCol, fromIdx, toCol int) string
	// Tick CPU を 1tick 進める (リアルタイム駆動)
	Tick() string
	// Hint ヒントを出力する
	Hint() string
	// Undo 直前の人間アクションを取り消す
	Undo() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.NertzConfig
}

// NertzInteractor Nertz インタラクター
type NertzInteractor struct {
	GameBase[interfaces.NertzGame]
	np presenter.NertzPresenter
}

// NewNertzInteractor コンストラクタ
func NewNertzInteractor(g interfaces.NertzGame, np presenter.NertzPresenter) *NertzInteractor {
	mustNotNil("NertzInteractor", map[string]any{"g": g, "np": np})
	return &NertzInteractor{GameBase: GameBase[interfaces.NertzGame]{Game: g}, np: np}
}

// Reset ゲーム初期化
func (ni *NertzInteractor) Reset() string {
	return runAndPresent(ni.Game, ni.np, ni.Game.Reset)
}

// ResetWithConfig 設定を適用してゲーム初期化
func (ni *NertzInteractor) ResetWithConfig(cfg domain.NertzConfig) string {
	if err := cfg.Validate(); err != nil {
		return ni.np.Output(ni.Game, err)
	}
	return runAndPresent(ni.Game, ni.np, func() { ni.Game.ResetWithConfig(cfg) })
}

// NextRound 次ラウンドを開始する
func (ni *NertzInteractor) NextRound() string {
	return runAndPresent(ni.Game, ni.np, ni.Game.NextRound)
}

// Draw 指定プレイヤーがストックを引く
func (ni *NertzInteractor) Draw(playerIdx int) string {
	return execAndPresent(ni.Game, ni.np, func() error { return ni.Game.DrawStock(playerIdx) })
}

// MoveNertzToFoundation ナッツパイル → ファウンデーション
func (ni *NertzInteractor) MoveNertzToFoundation(playerIdx, foundationIdx int) string {
	return execAndPresent(ni.Game, ni.np, func() error {
		return ni.Game.MoveNertzToFoundation(playerIdx, foundationIdx)
	})
}

// MoveNertzToTableau ナッツパイル → タブロー
func (ni *NertzInteractor) MoveNertzToTableau(playerIdx, toCol int) string {
	return execAndPresent(ni.Game, ni.np, func() error {
		return ni.Game.MoveNertzToTableau(playerIdx, toCol)
	})
}

// MoveWasteToFoundation ウェイスト → ファウンデーション
func (ni *NertzInteractor) MoveWasteToFoundation(playerIdx, foundationIdx int) string {
	return execAndPresent(ni.Game, ni.np, func() error {
		return ni.Game.MoveWasteToFoundation(playerIdx, foundationIdx)
	})
}

// MoveWasteToTableau ウェイスト → タブロー
func (ni *NertzInteractor) MoveWasteToTableau(playerIdx, toCol int) string {
	return execAndPresent(ni.Game, ni.np, func() error {
		return ni.Game.MoveWasteToTableau(playerIdx, toCol)
	})
}

// MoveTableauToFoundation タブロー → ファウンデーション
func (ni *NertzInteractor) MoveTableauToFoundation(playerIdx, fromCol, foundationIdx int) string {
	return execAndPresent(ni.Game, ni.np, func() error {
		return ni.Game.MoveTableauToFoundation(playerIdx, fromCol, foundationIdx)
	})
}

// MoveTableauToTableau タブロー → タブロー
func (ni *NertzInteractor) MoveTableauToTableau(playerIdx, fromCol, fromIdx, toCol int) string {
	return execAndPresent(ni.Game, ni.np, func() error {
		return ni.Game.MoveTableauToTableau(playerIdx, fromCol, fromIdx, toCol)
	})
}

// Tick CPU を 1tick 進める。
//
// 注: ドメインの Tick() は適用済み NertzAction のリストを返すが、現在のフロ
// ントエンドは presenter が出すフルスナップショットから派生するためここでは
// 破棄する。アニメーションキューを将来導入する際にこのリストを Web 出力に
// 載せる予定 (PR #1528 レビュー指摘)。
func (ni *NertzInteractor) Tick() string {
	return runAndPresent(ni.Game, ni.np, func() { ni.Game.Tick() })
}

// Hint ヒントを出力する
func (ni *NertzInteractor) Hint() string {
	return ni.np.HintOutput(ni.Game)
}

// Undo 直前の人間アクションを取り消す
func (ni *NertzInteractor) Undo() string {
	return execAndPresent(ni.Game, ni.np, ni.Game.Undo)
}

// ActionLog 棋譜を出力する
func (ni *NertzInteractor) ActionLog() string {
	return ni.np.ActionLogOutput(ni.Game)
}

// GetConfig 現在の設定を取得
func (ni *NertzInteractor) GetConfig() domain.NertzConfig {
	return ni.Game.GetConfig()
}

// RestoreNertzInteractor deserialises JSON into a NertzInteractor.
func RestoreNertzInteractor(data []byte, np presenter.NertzPresenter) (*NertzInteractor, error) {
	return restoreAndBuild[domain.Nertz](data, func(g *domain.Nertz) *NertzInteractor {
		return &NertzInteractor{GameBase: GameBase[interfaces.NertzGame]{Game: g}, np: np}
	})
}
