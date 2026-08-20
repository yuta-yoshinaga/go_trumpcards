//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TienLenInteractorIF Tien Lenインタラクターインタフェース
type TienLenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Play 人間プレイヤーがカードを出す (または パスする)
	Play(indices []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.TienLenConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.TienLenConfig
	// Hint 推奨手を出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TienLenInteractor Tien Lenインタラクタークラス
type TienLenInteractor struct {
	GameBase[interfaces.TienLenGame]
	tlp presenter.TienLenPresenter
}

// NewTienLenInteractor コンストラクタ
func NewTienLenInteractor(tg interfaces.TienLenGame, tlp presenter.TienLenPresenter) *TienLenInteractor {
	mustNotNil("TienLenInteractor", map[string]any{"tg": tg, "tlp": tlp})
	return &TienLenInteractor{
		GameBase: GameBase[interfaces.TienLenGame]{Game: tg},
		tlp:      tlp,
	}
}

// Reset ゲーム初期化
func (ti *TienLenInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tlp.Output(ti.Game, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
func (ti *TienLenInteractor) Play(indices []int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tlp); blocked {
		return out
	}
	err := ti.Game.PlayerPlay(indices)
	if err == nil && !ti.Game.GetGameEndFlag() && !ti.Game.HasPendingAction() {
		ti.runCpuTurns()
	}
	return ti.tlp.Output(ti.Game, err)
}

// GetConfig 現在の設定を返す
func (ti *TienLenInteractor) GetConfig() domain.TienLenConfig {
	return ti.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (ti *TienLenInteractor) ResetWithConfig(config domain.TienLenConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tlp, config, ti.Game.SetConfig, ti.Reset)
}

// Hint 推奨手を出力する
func (ti *TienLenInteractor) Hint() string {
	return ti.tlp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TienLenInteractor) ActionLog() string {
	return ti.tlp.ActionLogOutput(ti.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (ti *TienLenInteractor) runCpuTurns() {
	runCpuTurnsCapped(ti.Game, ti.Game.CpuPlay)
}

// RestoreTienLenInteractor deserialises JSON into a TienLenInteractor.
func RestoreTienLenInteractor(data []byte, tlp presenter.TienLenPresenter) (*TienLenInteractor, error) {
	return restoreAndBuild[domain.TienLen](data, func(g *domain.TienLen) *TienLenInteractor {
		return &TienLenInteractor{GameBase: GameBase[interfaces.TienLenGame]{Game: g}, tlp: tlp}
	})
}
