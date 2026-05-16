package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BriscolaInteractorIF ブリスコラインタラクターインタフェース
type BriscolaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BriscolaConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む (補充ドロー + ゲーム終了検出)
	NextTrick() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BriscolaConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BriscolaInteractor ブリスコラインタラクタークラス
type BriscolaInteractor struct {
	GameBase[interfaces.BriscolaGame]
	bp presenter.BriscolaPresenter
}

// NewBriscolaInteractor コンストラクタ
func NewBriscolaInteractor(b interfaces.BriscolaGame, bp presenter.BriscolaPresenter) *BriscolaInteractor {
	mustNotNil("BriscolaInteractor", map[string]any{"b": b, "bp": bp})
	return &BriscolaInteractor{GameBase: GameBase[interfaces.BriscolaGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化
func (bi *BriscolaInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BriscolaInteractor) ResetWithConfig(cfg domain.BriscolaConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Play カードをプレイ
func (bi *BriscolaInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	if err := bi.Game.PlayerPlay(cardIndex); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決
	if bi.Game.GetPhase() == domain.BriscolaPhaseTrickEnd {
		bi.Game.ResolveTrick()
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (bi *BriscolaInteractor) NextTrick() string {
	bi.Game.NextTrick()
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BriscolaInteractor) GetConfig() domain.BriscolaConfig {
	return bi.Game.GetConfig()
}

// Hint ヒント取得
func (bi *BriscolaInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BriscolaInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック終了になるまでCPUターンを実行。
// Briscola は単一ハンドのため、roundEnd は gameEnd と同一視する。
func (bi *BriscolaInteractor) runCpuTurns() {
	runCpuTurnsLoop(bi.Game, trickPhases[domain.BriscolaPhase]{
		play:     domain.BriscolaPhasePlay,
		trickEnd: domain.BriscolaPhaseTrickEnd,
		roundEnd: domain.BriscolaPhaseGameEnd,
		gameEnd:  domain.BriscolaPhaseGameEnd,
	})
}

// RestoreBriscolaInteractor deserialises JSON into a BriscolaInteractor.
func RestoreBriscolaInteractor(data []byte, bp presenter.BriscolaPresenter) (*BriscolaInteractor, error) {
	return restoreAndBuild[domain.Briscola](data, func(g *domain.Briscola) *BriscolaInteractor {
		return &BriscolaInteractor{GameBase: GameBase[interfaces.BriscolaGame]{Game: g}, bp: bp}
	})
}
