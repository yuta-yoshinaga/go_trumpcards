package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// EgyptianRatscrewInteractorIF エジプシャン・ラットスクリューインタラクターインタフェース
type EgyptianRatscrewInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.EgyptianRatscrewConfig) string
	// Step 現手番プレイヤーがストック先頭1枚を場に出す
	Step() string
	// Slap 指定プレイヤーがスラップを試みる
	Slap(playerIdx int) string
	// Tick 保留中の CPU アクションを進行させる
	Tick() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.EgyptianRatscrewConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// EgyptianRatscrewInteractor エジプシャン・ラットスクリューインタラクタークラス
type EgyptianRatscrewInteractor struct {
	GameBase[interfaces.EgyptianRatscrewGame]
	ep presenter.EgyptianRatscrewPresenter
}

// NewEgyptianRatscrewInteractor コンストラクタ
func NewEgyptianRatscrewInteractor(e interfaces.EgyptianRatscrewGame, ep presenter.EgyptianRatscrewPresenter) *EgyptianRatscrewInteractor {
	mustNotNil("EgyptianRatscrewInteractor", map[string]any{"e": e, "ep": ep})
	return &EgyptianRatscrewInteractor{GameBase: GameBase[interfaces.EgyptianRatscrewGame]{Game: e}, ep: ep}
}

// Reset ゲーム初期化
func (ei *EgyptianRatscrewInteractor) Reset() string {
	return runAndPresent(ei.Game, ei.ep, ei.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ei *EgyptianRatscrewInteractor) ResetWithConfig(cfg domain.EgyptianRatscrewConfig) string {
	return resetWithValidatedConfig(ei.Game, ei.ep, cfg, ei.Game.SetConfig, ei.Reset)
}

// Step 現手番プレイヤーがストック先頭1枚を場に出す
func (ei *EgyptianRatscrewInteractor) Step() string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	if err := ei.Game.Step(); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	return ei.ep.Output(ei.Game, nil)
}

// Slap 指定プレイヤーがスラップを試みる
func (ei *EgyptianRatscrewInteractor) Slap(playerIdx int) string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	if err := ei.Game.Slap(playerIdx); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	return ei.ep.Output(ei.Game, nil)
}

// Tick 保留中の CPU アクションを進行させる
func (ei *EgyptianRatscrewInteractor) Tick() string {
	return runAndPresent(ei.Game, ei.ep, func() { ei.Game.Tick() })
}

// GetConfig 現在の設定を取得
func (ei *EgyptianRatscrewInteractor) GetConfig() domain.EgyptianRatscrewConfig {
	return ei.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ei *EgyptianRatscrewInteractor) ActionLog() string {
	return ei.ep.ActionLogOutput(ei.Game)
}

// RestoreEgyptianRatscrewInteractor deserialises JSON into an EgyptianRatscrewInteractor.
func RestoreEgyptianRatscrewInteractor(data []byte, ep presenter.EgyptianRatscrewPresenter) (*EgyptianRatscrewInteractor, error) {
	return restoreAndBuild[domain.EgyptianRatscrew](data, func(g *domain.EgyptianRatscrew) *EgyptianRatscrewInteractor {
		return &EgyptianRatscrewInteractor{GameBase: GameBase[interfaces.EgyptianRatscrewGame]{Game: g}, ep: ep}
	})
}
