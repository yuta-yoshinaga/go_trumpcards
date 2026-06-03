package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SchnapsenInteractorIF シュナプセンインタラクターインタフェース
type SchnapsenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SchnapsenConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// DeclareMarriage マリアージュを宣言してその K/Q をリードする
	DeclareMarriage(cardIndex int) string
	// NextTrick 次のトリックへ進む (補充ドロー + ゲーム終了検出)
	NextTrick() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SchnapsenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SchnapsenInteractor シュナプセンインタラクタークラス
type SchnapsenInteractor struct {
	GameBase[interfaces.SchnapsenGame]
	sp presenter.SchnapsenPresenter
}

// NewSchnapsenInteractor コンストラクタ
func NewSchnapsenInteractor(s interfaces.SchnapsenGame, sp presenter.SchnapsenPresenter) *SchnapsenInteractor {
	mustNotNil("SchnapsenInteractor", map[string]any{"s": s, "sp": sp})
	return &SchnapsenInteractor{GameBase: GameBase[interfaces.SchnapsenGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化
func (si *SchnapsenInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SchnapsenInteractor) ResetWithConfig(cfg domain.SchnapsenConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Play カードをプレイ
func (si *SchnapsenInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPlay(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.afterHumanAction()
	return si.sp.Output(si.Game, nil)
}

// DeclareMarriage マリアージュを宣言してその K/Q をリードする
func (si *SchnapsenInteractor) DeclareMarriage(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerDeclareMarriage(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.afterHumanAction()
	return si.sp.Output(si.Game, nil)
}

// afterHumanAction 人間のプレイ/宣言後の共通処理。トリックが揃ったら解決し、CPUターンを進める。
func (si *SchnapsenInteractor) afterHumanAction() {
	if si.Game.GetPhase() == domain.SchnapsenPhaseTrickEnd {
		si.Game.ResolveTrick()
	}
	si.runCpuTurns()
}

// NextTrick 次のトリックへ進む
func (si *SchnapsenInteractor) NextTrick() string {
	si.Game.NextTrick()
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SchnapsenInteractor) GetConfig() domain.SchnapsenConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *SchnapsenInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SchnapsenInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック終了になるまでCPUターンを実行。
// Schnapsen は単一ラウンドのため、roundEnd は gameEnd と同一視する。
func (si *SchnapsenInteractor) runCpuTurns() {
	runCpuTurnsLoop(si.Game, trickPhases[domain.SchnapsenPhase]{
		play:     domain.SchnapsenPhasePlay,
		trickEnd: domain.SchnapsenPhaseTrickEnd,
		roundEnd: domain.SchnapsenPhaseGameEnd,
		gameEnd:  domain.SchnapsenPhaseGameEnd,
	})
}

// RestoreSchnapsenInteractor deserialises JSON into a SchnapsenInteractor.
func RestoreSchnapsenInteractor(data []byte, sp presenter.SchnapsenPresenter) (*SchnapsenInteractor, error) {
	return restoreAndBuild[domain.Schnapsen](data, func(g *domain.Schnapsen) *SchnapsenInteractor {
		return &SchnapsenInteractor{GameBase: GameBase[interfaces.SchnapsenGame]{Game: g}, sp: sp}
	})
}
