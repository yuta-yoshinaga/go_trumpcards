package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GoFishInteractorIF Go Fishインタラクターインタフェース
type GoFishInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset(config domain.GoFishConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.GoFishConfig
	// Ask 人間プレイヤーが相手にランクを要求する
	Ask(targetIdx, rank int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GoFishInteractor Go Fishインタラクタークラス
type GoFishInteractor struct {
	GameBase[interfaces.GoFishGame]
	gfp presenter.GoFishPresenter
}

// NewGoFishInteractor コンストラクタ
func NewGoFishInteractor(gf interfaces.GoFishGame, gfp presenter.GoFishPresenter) *GoFishInteractor {
	mustNotNil("GoFishInteractor", map[string]any{"gf": gf, "gfp": gfp})
	return &GoFishInteractor{GameBase: GameBase[interfaces.GoFishGame]{Game: gf}, gfp: gfp}
}

// GetConfig 現在の設定を返す
func (gi *GoFishInteractor) GetConfig() domain.GoFishConfig {
	return gi.Game.GetConfig()
}

// Reset ゲーム初期化
func (gi *GoFishInteractor) Reset(config domain.GoFishConfig) string {
	if err := config.Validate(); err != nil {
		return gi.gfp.Output(gi.Game, err)
	}
	gi.Game.SetConfig(config)
	gi.Game.Reset()
	gi.runCpuTurns()
	return gi.gfp.Output(gi.Game, nil)
}

// Ask 人間プレイヤーが相手にランクを要求する
func (gi *GoFishInteractor) Ask(targetIdx, rank int) string {
	if out, blocked := guardNotPlayable(gi.Game, gi.gfp); blocked {
		return out
	}
	err := gi.Game.PlayerAsk(targetIdx, rank)
	if err == nil && !gi.Game.GetGameEndFlag() {
		gi.runCpuTurns()
	}
	return gi.gfp.Output(gi.Game, err)
}

// ActionLog 棋譜を出力する
func (gi *GoFishInteractor) ActionLog() string {
	return gi.gfp.ActionLogOutput(gi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (gi *GoFishInteractor) runCpuTurns() {
	runCpuTurnsCapped(gi.Game, func() { _ = gi.Game.CpuAsk() })
}

// RestoreGoFishInteractor deserialises JSON into a GoFishInteractor.
func RestoreGoFishInteractor(data []byte, gfp presenter.GoFishPresenter) (*GoFishInteractor, error) {
	return restoreAndBuild[domain.GoFish](data, func(g *domain.GoFish) *GoFishInteractor {
		return &GoFishInteractor{GameBase: GameBase[interfaces.GoFishGame]{Game: g}, gfp: gfp}
	})
}
