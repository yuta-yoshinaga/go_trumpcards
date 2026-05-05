package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// WarInteractorIF 戦争インタラクターインタフェース
type WarInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.WarConfig) string
	// Step 状態機械を1ステップ進める
	Step() string
	// AutoPlay 決着まで自動で進める
	AutoPlay() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.WarConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// WarInteractor 戦争インタラクタークラス
type WarInteractor struct {
	GameBase[interfaces.WarGame]
	wp presenter.WarPresenter
}

// NewWarInteractor コンストラクタ
func NewWarInteractor(w interfaces.WarGame, wp presenter.WarPresenter) *WarInteractor {
	mustNotNil("WarInteractor", map[string]any{"w": w, "wp": wp})
	return &WarInteractor{GameBase: GameBase[interfaces.WarGame]{Game: w}, wp: wp}
}

// Reset ゲーム初期化
func (wi *WarInteractor) Reset() string {
	return runAndPresent(wi.Game, wi.wp, wi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (wi *WarInteractor) ResetWithConfig(cfg domain.WarConfig) string {
	return resetWithValidatedConfig(wi.Game, wi.wp, cfg, wi.Game.SetConfig, wi.Reset)
}

// Step 状態機械を1ステップ進める
func (wi *WarInteractor) Step() string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	if err := wi.Game.Step(); err != nil {
		return wi.wp.Output(wi.Game, err)
	}
	return wi.wp.Output(wi.Game, nil)
}

// AutoPlay 決着まで自動で進める
func (wi *WarInteractor) AutoPlay() string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	return execAndPresent(wi.Game, wi.wp, wi.Game.AutoPlay)
}

// GetConfig 現在の設定を取得
func (wi *WarInteractor) GetConfig() domain.WarConfig {
	return wi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (wi *WarInteractor) ActionLog() string {
	return wi.wp.ActionLogOutput(wi.Game)
}

// RestoreWarInteractor deserialises JSON into a WarInteractor.
func RestoreWarInteractor(data []byte, wp presenter.WarPresenter) (*WarInteractor, error) {
	return restoreAndBuild[domain.War](data, func(g *domain.War) *WarInteractor {
		return &WarInteractor{GameBase: GameBase[interfaces.WarGame]{Game: g}, wp: wp}
	})
}
