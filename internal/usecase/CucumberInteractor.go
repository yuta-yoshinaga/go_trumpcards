//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CucumberInteractorIF キューカンバーインタラクターインタフェース
type CucumberInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CucumberConfig) string
	// Play カードを出す
	Play(cardIndex int) string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CucumberConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CucumberInteractor キューカンバーインタラクタークラス
type CucumberInteractor struct {
	GameBase[interfaces.CucumberGame]
	sp presenter.CucumberPresenter
}

// NewCucumberInteractor コンストラクタ
func NewCucumberInteractor(s interfaces.CucumberGame, sp presenter.CucumberPresenter) *CucumberInteractor {
	mustNotNil("CucumberInteractor", map[string]any{"s": s, "sp": sp})
	return &CucumberInteractor{GameBase: GameBase[interfaces.CucumberGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (ci *CucumberInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.sp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CucumberInteractor) ResetWithConfig(cfg domain.CucumberConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.sp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードを出す
func (ci *CucumberInteractor) Play(cardIndex int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.sp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.sp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.sp.Output(ci.Game, nil)
}

// NextRound 次のラウンドを配る
func (ci *CucumberInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ci.Game, ci.sp); blocked {
		return out
	}
	if err := ci.Game.NextRound(); err != nil {
		return ci.sp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.sp.Output(ci.Game, nil)
}

// GiveUp 投了する
func (ci *CucumberInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ci.Game, ci.sp); blocked {
		return out
	}
	ci.Game.GiveUp()
	return ci.sp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *CucumberInteractor) GetConfig() domain.CucumberConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *CucumberInteractor) Hint() string { return ci.sp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *CucumberInteractor) ActionLog() string { return ci.sp.ActionLogOutput(ci.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **ラウンドの区切りでも止まります。** 失点が付くのはラウンドに 1 回だけの出来事で、
// 盤面に痕跡が残らないので、読む前に配り直してはいけません。
func (ci *CucumberInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if ci.Game.GetGameEndFlag() || ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestoreCucumberInteractor deserialises JSON into a CucumberInteractor.
func RestoreCucumberInteractor(data []byte, sp presenter.CucumberPresenter) (*CucumberInteractor, error) {
	return restoreAndBuild[domain.Cucumber](data, func(g *domain.Cucumber) *CucumberInteractor {
		return NewCucumberInteractor(g, sp)
	})
}
