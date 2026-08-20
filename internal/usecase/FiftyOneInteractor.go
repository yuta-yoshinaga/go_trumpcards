package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FiftyOneInteractorIF フィフティワンインタラクターインタフェース
type FiftyOneInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset(config domain.FiftyOneConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.FiftyOneConfig
	// ExchangeOne 手札1枚と場札1枚を交換する
	ExchangeOne(handIdx, tableIdx int) string
	// ExchangeAll 手札5枚と場札5枚を全交換する
	ExchangeAll() string
	// Stop ストップ宣言する
	Stop() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// FiftyOneInteractor フィフティワンインタラクタークラス
type FiftyOneInteractor struct {
	GameBase[interfaces.FiftyOneGame]
	fop presenter.FiftyOnePresenter
}

// NewFiftyOneInteractor コンストラクタ
func NewFiftyOneInteractor(fo interfaces.FiftyOneGame, fop presenter.FiftyOnePresenter) *FiftyOneInteractor {
	mustNotNil("FiftyOneInteractor", map[string]any{"fo": fo, "fop": fop})
	return &FiftyOneInteractor{GameBase: GameBase[interfaces.FiftyOneGame]{Game: fo}, fop: fop}
}

// GetConfig 現在の設定を返す
func (fi *FiftyOneInteractor) GetConfig() domain.FiftyOneConfig {
	return fi.Game.GetConfig()
}

// Reset ゲーム初期化
func (fi *FiftyOneInteractor) Reset(config domain.FiftyOneConfig) string {
	if err := config.Validate(); err != nil {
		return fi.fop.Output(fi.Game, err)
	}
	fi.Game.SetConfig(config)
	fi.Game.Reset()
	fi.runCpuTurns()
	return fi.fop.Output(fi.Game, nil)
}

// ExchangeOne 手札1枚と場札1枚を交換する
func (fi *FiftyOneInteractor) ExchangeOne(handIdx, tableIdx int) string {
	if out, blocked := guardNotPlayable(fi.Game, fi.fop); blocked {
		return out
	}
	err := fi.Game.ExchangeOne(handIdx, tableIdx)
	if err == nil && !fi.Game.GetGameEndFlag() {
		fi.runCpuTurns()
	}
	return fi.fop.Output(fi.Game, err)
}

// ExchangeAll 手札5枚と場札5枚を全交換する
func (fi *FiftyOneInteractor) ExchangeAll() string {
	if out, blocked := guardNotPlayable(fi.Game, fi.fop); blocked {
		return out
	}
	err := fi.Game.ExchangeAll()
	if err == nil && !fi.Game.GetGameEndFlag() {
		fi.runCpuTurns()
	}
	return fi.fop.Output(fi.Game, err)
}

// Stop ストップ宣言する
func (fi *FiftyOneInteractor) Stop() string {
	if out, blocked := guardNotPlayable(fi.Game, fi.fop); blocked {
		return out
	}
	err := fi.Game.Stop()
	if err == nil && !fi.Game.GetGameEndFlag() {
		fi.runCpuTurns()
	}
	return fi.fop.Output(fi.Game, err)
}

// ActionLog 棋譜を出力する
func (fi *FiftyOneInteractor) ActionLog() string {
	return fi.fop.ActionLogOutput(fi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (fi *FiftyOneInteractor) runCpuTurns() {
	runCpuTurnsCapped(fi.Game, func() { _ = fi.Game.CpuPlay() })
}

// RestoreFiftyOneInteractor deserialises JSON into a FiftyOneInteractor.
func RestoreFiftyOneInteractor(data []byte, fop presenter.FiftyOnePresenter) (*FiftyOneInteractor, error) {
	return restoreAndBuild[domain.FiftyOne](data, func(g *domain.FiftyOne) *FiftyOneInteractor {
		return &FiftyOneInteractor{GameBase: GameBase[interfaces.FiftyOneGame]{Game: g}, fop: fop}
	})
}
