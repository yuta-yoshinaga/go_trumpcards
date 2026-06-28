package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BeggarMyNeighbourInteractorIF Beggar-My-Neighbour インタラクターインタフェース
type BeggarMyNeighbourInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BeggarMyNeighbourConfig) string
	// Step 状態機械を1ステップ進める
	Step() string
	// AutoPlay 決着まで自動で進める
	AutoPlay() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BeggarMyNeighbourConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BeggarMyNeighbourInteractor Beggar-My-Neighbour インタラクタークラス
type BeggarMyNeighbourInteractor struct {
	GameBase[interfaces.BeggarMyNeighbourGame]
	wp presenter.BeggarMyNeighbourPresenter
}

// NewBeggarMyNeighbourInteractor コンストラクタ
func NewBeggarMyNeighbourInteractor(g interfaces.BeggarMyNeighbourGame, wp presenter.BeggarMyNeighbourPresenter) *BeggarMyNeighbourInteractor {
	mustNotNil("BeggarMyNeighbourInteractor", map[string]any{"g": g, "wp": wp})
	return &BeggarMyNeighbourInteractor{GameBase: GameBase[interfaces.BeggarMyNeighbourGame]{Game: g}, wp: wp}
}

// Reset ゲーム初期化
func (bi *BeggarMyNeighbourInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.wp, bi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BeggarMyNeighbourInteractor) ResetWithConfig(cfg domain.BeggarMyNeighbourConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.wp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Step 状態機械を1ステップ進める
func (bi *BeggarMyNeighbourInteractor) Step() string {
	if out, blocked := guardGameEnd(bi.Game, bi.wp); blocked {
		return out
	}
	if err := bi.Game.Step(); err != nil {
		return bi.wp.Output(bi.Game, err)
	}
	return bi.wp.Output(bi.Game, nil)
}

// AutoPlay 決着まで自動で進める
func (bi *BeggarMyNeighbourInteractor) AutoPlay() string {
	if out, blocked := guardGameEnd(bi.Game, bi.wp); blocked {
		return out
	}
	return execAndPresent(bi.Game, bi.wp, bi.Game.AutoPlay)
}

// GetConfig 現在の設定を取得
func (bi *BeggarMyNeighbourInteractor) GetConfig() domain.BeggarMyNeighbourConfig {
	return bi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (bi *BeggarMyNeighbourInteractor) ActionLog() string {
	return bi.wp.ActionLogOutput(bi.Game)
}

// RestoreBeggarMyNeighbourInteractor deserialises JSON into a BeggarMyNeighbourInteractor.
func RestoreBeggarMyNeighbourInteractor(data []byte, wp presenter.BeggarMyNeighbourPresenter) (*BeggarMyNeighbourInteractor, error) {
	return restoreAndBuild[domain.BeggarMyNeighbour](data, func(g *domain.BeggarMyNeighbour) *BeggarMyNeighbourInteractor {
		return &BeggarMyNeighbourInteractor{GameBase: GameBase[interfaces.BeggarMyNeighbourGame]{Game: g}, wp: wp}
	})
}
