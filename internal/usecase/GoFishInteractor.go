package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GoFishInteractorIF Go Fishインタラクターインタフェース
type GoFishInteractorIF interface {
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
	gf  interfaces.GoFishGame
	gfp presenter.GoFishPresenter
}

// NewGoFishInteractor コンストラクタ
func NewGoFishInteractor(gf interfaces.GoFishGame, gfp presenter.GoFishPresenter) *GoFishInteractor {
	mustNotNil("GoFishInteractor", map[string]any{"gf": gf, "gfp": gfp})
	return &GoFishInteractor{gf: gf, gfp: gfp}
}

// GetConfig 現在の設定を返す
func (gi *GoFishInteractor) GetConfig() domain.GoFishConfig {
	return gi.gf.GetConfig()
}

// Reset ゲーム初期化
func (gi *GoFishInteractor) Reset(config domain.GoFishConfig) string {
	if err := config.Validate(); err != nil {
		return gi.gfp.Output(gi.gf, err)
	}
	gi.gf.SetConfig(config)
	gi.gf.Reset()
	gi.runCpuTurns()
	return gi.gfp.Output(gi.gf, nil)
}

// Ask 人間プレイヤーが相手にランクを要求する
func (gi *GoFishInteractor) Ask(targetIdx, rank int) string {
	if out, blocked := guardNotPlayable(gi.gf, gi.gfp); blocked {
		return out
	}
	err := gi.gf.PlayerAsk(targetIdx, rank)
	if err == nil && !gi.gf.GetGameEndFlag() {
		gi.runCpuTurns()
	}
	return gi.gfp.Output(gi.gf, err)
}

// ActionLog 棋譜を出力する
func (gi *GoFishInteractor) ActionLog() string {
	return gi.gfp.ActionLogOutput(gi.gf)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (gi *GoFishInteractor) runCpuTurns() {
	for !gi.gf.GetGameEndFlag() && !gi.gf.IsHumanTurn() {
		_ = gi.gf.CpuAsk()
	}
}

// Snapshot serialises the game state to JSON for KV persistence.
func (gi *GoFishInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(gi.gf)
}

// RestoreGoFishInteractor deserialises JSON into a GoFishInteractor.
func RestoreGoFishInteractor(data []byte, gfp presenter.GoFishPresenter) (*GoFishInteractor, error) {
	gf, err := restoreGame[domain.GoFish](data)
	if err != nil {
		return nil, err
	}
	return &GoFishInteractor{gf: gf, gfp: gfp}, nil
}
