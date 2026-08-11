//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PasurInteractorIF パスールインタラクターインタフェース
type PasurInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PasurConfig) string
	// Play カードをプレイ (tableIndices が空ならトレール)
	Play(cardIndex int, tableIndices []int) string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PasurConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PasurInteractor パスールインタラクタークラス
type PasurInteractor struct {
	GameBase[interfaces.PasurGame]
	sp presenter.PasurPresenter
}

// NewPasurInteractor コンストラクタ
func NewPasurInteractor(s interfaces.PasurGame, sp presenter.PasurPresenter) *PasurInteractor {
	mustNotNil("PasurInteractor", map[string]any{"s": s, "sp": sp})
	return &PasurInteractor{GameBase: GameBase[interfaces.PasurGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (pi *PasurInteractor) Reset() string {
	pi.Game.Reset()
	pi.advance()
	return pi.sp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PasurInteractor) ResetWithConfig(cfg domain.PasurConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.sp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Play カードをプレイ
func (pi *PasurInteractor) Play(cardIndex int, tableIndices []int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.PlayerPlay(cardIndex, tableIndices); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	pi.advance()
	return pi.sp.Output(pi.Game, nil)
}

// GiveUp 投了する
func (pi *PasurInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	pi.Game.GiveUp()
	return pi.sp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *PasurInteractor) GetConfig() domain.PasurConfig { return pi.Game.GetConfig() }

// Hint ヒント取得
func (pi *PasurInteractor) Hint() string { return pi.sp.HintOutput(pi.Game) }

// ActionLog 棋譜を出力する
func (pi *PasurInteractor) ActionLog() string { return pi.sp.ActionLogOutput(pi.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **ラウンドの区切りが無い。** 手札が尽きたら自動で配り足し、山札が尽きたら
// そのまま終局するので、途中で止まる段はプレイだけです。
func (pi *PasurInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if pi.Game.GetGameEndFlag() || pi.Game.IsHumanTurn() {
			return
		}
		pi.Game.CpuPlay()
	}
}

// RestorePasurInteractor deserialises JSON into a PasurInteractor.
func RestorePasurInteractor(data []byte, sp presenter.PasurPresenter) (*PasurInteractor, error) {
	return restoreAndBuild[domain.Pasur](data, func(g *domain.Pasur) *PasurInteractor {
		return NewPasurInteractor(g, sp)
	})
}
