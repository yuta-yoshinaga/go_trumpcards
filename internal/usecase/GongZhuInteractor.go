package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GongZhuInteractorIF 拱猪インタラクターインタフェース
type GongZhuInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.GongZhuConfig) string
	// Expose ポイントカードを公開してプレイへ進む
	Expose(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.GongZhuConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GongZhuInteractor 拱猪インタラクタークラス
type GongZhuInteractor struct {
	GameBase[interfaces.GongZhuGame]
	gp presenter.GongZhuPresenter
}

// NewGongZhuInteractor コンストラクタ
func NewGongZhuInteractor(g interfaces.GongZhuGame, gp presenter.GongZhuPresenter) *GongZhuInteractor {
	mustNotNil("GongZhuInteractor", map[string]any{"g": g, "gp": gp})
	return &GongZhuInteractor{GameBase: GameBase[interfaces.GongZhuGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化（公開フェーズで人間の入力を待つ）
func (gi *GongZhuInteractor) Reset() string {
	gi.Game.Reset()
	return gi.gp.Output(gi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (gi *GongZhuInteractor) ResetWithConfig(cfg domain.GongZhuConfig) string {
	return resetWithValidatedConfig(gi.Game, gi.gp, cfg, gi.Game.SetConfig, gi.Reset)
}

// Expose ポイントカードを公開してプレイへ進む
func (gi *GongZhuInteractor) Expose(cardIndices []int) string {
	if out, blocked := guardGameEnd(gi.Game, gi.gp); blocked {
		return out
	}
	err := gi.Game.PlayerExpose(cardIndices)
	if err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.Game.CpuExpose()
	gi.Game.ExecuteExpose()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// Play カードをプレイ
func (gi *GongZhuInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(gi.Game, gi.gp); blocked {
		return out
	}
	err := gi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (gi *GongZhuInteractor) NextTrick() string {
	gi.Game.NextTrick()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (gi *GongZhuInteractor) NextRound() string {
	gi.Game.ScoreRound()
	if out, blocked := guardGameEnd(gi.Game, gi.gp); blocked {
		return out
	}
	gi.Game.NextRound()
	return gi.gp.Output(gi.Game, nil)
}

// GetConfig 現在の設定を取得
func (gi *GongZhuInteractor) GetConfig() domain.GongZhuConfig {
	return gi.Game.GetConfig()
}

// Hint ヒント取得
func (gi *GongZhuInteractor) Hint() string {
	return gi.gp.HintOutput(gi.Game)
}

// ActionLog 棋譜を出力する
func (gi *GongZhuInteractor) ActionLog() string {
	return gi.gp.ActionLogOutput(gi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (gi *GongZhuInteractor) runCpuTurns() {
	runCpuTurnsLoop(gi.Game, trickPhases[domain.GongZhuPhase]{
		play:     domain.GongZhuPhasePlay,
		trickEnd: domain.GongZhuPhaseTrickEnd,
		roundEnd: domain.GongZhuPhaseRoundEnd,
		gameEnd:  domain.GongZhuPhaseGameEnd,
	})
}

// RestoreGongZhuInteractor deserialises JSON into a GongZhuInteractor.
func RestoreGongZhuInteractor(data []byte, gp presenter.GongZhuPresenter) (*GongZhuInteractor, error) {
	return restoreAndBuild[domain.GongZhu](data, func(g *domain.GongZhu) *GongZhuInteractor {
		return &GongZhuInteractor{GameBase: GameBase[interfaces.GongZhuGame]{Game: g}, gp: gp}
	})
}
