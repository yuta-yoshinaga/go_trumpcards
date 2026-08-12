//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// LingerLongerInteractorIF リンガーロンガーインタラクターインタフェース
type LingerLongerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.LingerLongerConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.LingerLongerConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// LingerLongerInteractor リンガーロンガーインタラクタークラス
type LingerLongerInteractor struct {
	GameBase[interfaces.LingerLongerGame]
	sp presenter.LingerLongerPresenter
}

// NewLingerLongerInteractor コンストラクタ
func NewLingerLongerInteractor(s interfaces.LingerLongerGame, sp presenter.LingerLongerPresenter) *LingerLongerInteractor {
	mustNotNil("LingerLongerInteractor", map[string]any{"s": s, "sp": sp})
	return &LingerLongerInteractor{GameBase: GameBase[interfaces.LingerLongerGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (li *LingerLongerInteractor) Reset() string {
	li.Game.Reset()
	li.advance()
	return li.sp.Output(li.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (li *LingerLongerInteractor) ResetWithConfig(cfg domain.LingerLongerConfig) string {
	return resetWithValidatedConfig(li.Game, li.sp, cfg, li.Game.SetConfig, li.Reset)
}

// Play カードをプレイ
func (li *LingerLongerInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(li.Game, li.sp); blocked {
		return out
	}
	if err := li.Game.PlayerPlay(cardIndex); err != nil {
		return li.sp.Output(li.Game, err)
	}
	li.advance()
	return li.sp.Output(li.Game, nil)
}

// GiveUp 投了する
func (li *LingerLongerInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(li.Game, li.sp); blocked {
		return out
	}
	li.Game.GiveUp()
	return li.sp.Output(li.Game, nil)
}

// GetConfig 現在の設定を取得
func (li *LingerLongerInteractor) GetConfig() domain.LingerLongerConfig { return li.Game.GetConfig() }

// Hint ヒント取得
func (li *LingerLongerInteractor) Hint() string { return li.sp.HintOutput(li.Game) }

// ActionLog 棋譜を出力する
func (li *LingerLongerInteractor) ActionLog() string { return li.sp.ActionLogOutput(li.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **人間が脱落しても止まらない。** 脱落した席は `IsHumanTurn` が偽のままなので、
// 残りの CPU 同士が決着するまで回り切ります。
func (li *LingerLongerInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if li.Game.GetGameEndFlag() || li.Game.IsHumanTurn() {
			return
		}
		li.Game.CpuPlay()
	}
}

// RestoreLingerLongerInteractor deserialises JSON into a LingerLongerInteractor.
func RestoreLingerLongerInteractor(data []byte, sp presenter.LingerLongerPresenter) (*LingerLongerInteractor, error) {
	return restoreAndBuild[domain.LingerLonger](data, func(g *domain.LingerLonger) *LingerLongerInteractor {
		return NewLingerLongerInteractor(g, sp)
	})
}
