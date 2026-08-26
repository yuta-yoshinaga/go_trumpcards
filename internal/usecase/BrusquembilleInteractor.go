//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BrusquembilleInteractorIF ブリュスカンビーユインタラクターインタフェース
type BrusquembilleInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BrusquembilleConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む (補充ドロー + ゲーム終了検出)
	NextTrick() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BrusquembilleConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BrusquembilleInteractor ブリュスカンビーユインタラクタークラス
type BrusquembilleInteractor struct {
	GameBase[interfaces.BrusquembilleGame]
	bp presenter.BrusquembillePresenter
}

// NewBrusquembilleInteractor コンストラクタ
func NewBrusquembilleInteractor(b interfaces.BrusquembilleGame, bp presenter.BrusquembillePresenter) *BrusquembilleInteractor {
	mustNotNil("BrusquembilleInteractor", map[string]any{"b": b, "bp": bp})
	return &BrusquembilleInteractor{GameBase: GameBase[interfaces.BrusquembilleGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化
func (bi *BrusquembilleInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BrusquembilleInteractor) ResetWithConfig(cfg domain.BrusquembilleConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Play カードをプレイ
func (bi *BrusquembilleInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	if err := bi.Game.PlayerPlay(cardIndex); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決
	if bi.Game.GetPhase() == domain.BrusquembillePhaseTrickEnd {
		bi.Game.ResolveTrick()
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (bi *BrusquembilleInteractor) NextTrick() string {
	bi.Game.NextTrick()
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BrusquembilleInteractor) GetConfig() domain.BrusquembilleConfig {
	return bi.Game.GetConfig()
}

// Hint ヒント取得
func (bi *BrusquembilleInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BrusquembilleInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック終了になるまでCPUターンを実行。
// Brusquembille は単一ハンドのため、roundEnd は gameEnd と同一視する。
func (bi *BrusquembilleInteractor) runCpuTurns() {
	runCpuTurnsLoop(bi.Game, trickPhases[domain.BrusquembillePhase]{
		play:     domain.BrusquembillePhasePlay,
		trickEnd: domain.BrusquembillePhaseTrickEnd,
		roundEnd: domain.BrusquembillePhaseGameEnd,
		gameEnd:  domain.BrusquembillePhaseGameEnd,
	})
}

// RestoreBrusquembilleInteractor deserialises JSON into a BrusquembilleInteractor.
func RestoreBrusquembilleInteractor(data []byte, bp presenter.BrusquembillePresenter) (*BrusquembilleInteractor, error) {
	return restoreAndBuild[domain.Brusquembille](data, func(g *domain.Brusquembille) *BrusquembilleInteractor {
		return &BrusquembilleInteractor{GameBase: GameBase[interfaces.BrusquembilleGame]{Game: g}, bp: bp}
	})
}
