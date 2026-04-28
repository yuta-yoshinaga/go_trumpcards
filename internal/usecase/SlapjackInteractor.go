package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SlapjackInteractorIF スラップジャックインタラクターインタフェース
type SlapjackInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SlapjackConfig) string
	// Step 現手番プレイヤーがストック先頭1枚を場に出す
	Step() string
	// Slap 指定プレイヤーがスラップを試みる
	Slap(playerIdx int) string
	// Tick 保留中の CPU アクションを進行させる
	Tick() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SlapjackConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SlapjackInteractor スラップジャックインタラクタークラス
type SlapjackInteractor struct {
	GameBase[interfaces.SlapjackGame]
	sp presenter.SlapjackPresenter
}

// NewSlapjackInteractor コンストラクタ
func NewSlapjackInteractor(s interfaces.SlapjackGame, sp presenter.SlapjackPresenter) *SlapjackInteractor {
	mustNotNil("SlapjackInteractor", map[string]any{"s": s, "sp": sp})
	return &SlapjackInteractor{GameBase: GameBase[interfaces.SlapjackGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化
func (si *SlapjackInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SlapjackInteractor) ResetWithConfig(cfg domain.SlapjackConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Step 現手番プレイヤーがストック先頭1枚を場に出す
func (si *SlapjackInteractor) Step() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.Step(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	return si.sp.Output(si.Game, nil)
}

// Slap 指定プレイヤーがスラップを試みる
func (si *SlapjackInteractor) Slap(playerIdx int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.Slap(playerIdx); err != nil {
		return si.sp.Output(si.Game, err)
	}
	return si.sp.Output(si.Game, nil)
}

// Tick 保留中の CPU アクションを進行させる
func (si *SlapjackInteractor) Tick() string {
	return runAndPresent(si.Game, si.sp, func() { si.Game.Tick() })
}

// GetConfig 現在の設定を取得
func (si *SlapjackInteractor) GetConfig() domain.SlapjackConfig {
	return si.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *SlapjackInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreSlapjackInteractor deserialises JSON into a SlapjackInteractor.
func RestoreSlapjackInteractor(data []byte, sp presenter.SlapjackPresenter) (*SlapjackInteractor, error) {
	return restoreAndBuild[domain.Slapjack](data, func(g *domain.Slapjack) *SlapjackInteractor {
		return &SlapjackInteractor{GameBase: GameBase[interfaces.SlapjackGame]{Game: g}, sp: sp}
	})
}
