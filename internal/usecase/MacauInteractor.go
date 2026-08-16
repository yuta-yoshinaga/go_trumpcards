//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MacauInteractorIF マカオインタラクターインタフェース
type MacauInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MacauConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// ChooseSuit スートを選択 (8を出した後)
	ChooseSuit(suit int) string
	// Draw カードを引く (ペナルティ中はスタックを引き受ける)
	Draw() string
	// Declare 「マカオ！」と宣言する
	Declare() string
	// SkipDeclare 宣言をスキップしてペナルティを受ける
	SkipDeclare() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MacauConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
	// IsHumanChooseSuitTurn reports whether the human just played an 8 and the
	// game is now waiting for them to pick a suit.
	IsHumanChooseSuitTurn() bool
	// IsHumanDeclareTurn reports whether the human just reached one card and the
	// game is now waiting for them to declare "Macau!".
	IsHumanDeclareTurn() bool
}

// MacauInteractor マカオインタラクタークラス
type MacauInteractor struct {
	GameBase[interfaces.MacauGame]
	gp presenter.MacauPresenter
}

// NewMacauInteractor コンストラクタ
func NewMacauInteractor(g interfaces.MacauGame, gp presenter.MacauPresenter) *MacauInteractor {
	mustNotNil("MacauInteractor", map[string]any{"g": g, "gp": gp})
	return &MacauInteractor{GameBase: GameBase[interfaces.MacauGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *MacauInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *MacauInteractor) ResetWithConfig(cfg domain.MacauConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードをプレイ
func (ci *MacauInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ChooseSuit スートを選択 (8を出した後)
func (ci *MacauInteractor) ChooseSuit(suit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerChooseSuit(suit)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Draw カードを引く
func (ci *MacauInteractor) Draw() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDraw()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Declare 「マカオ！」と宣言する
func (ci *MacauInteractor) Declare() string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDeclare()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// SkipDeclare 宣言をスキップしてペナルティを受ける
func (ci *MacauInteractor) SkipDeclare() string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerSkipDeclare()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *MacauInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *MacauInteractor) GetConfig() domain.MacauConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *MacauInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// Hint ヒントを出力する
func (ci *MacauInteractor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// IsHumanChooseSuitTurn reports whether the game is currently waiting for the
// human to pick a suit (i.e. the human just played an 8).
func (ci *MacauInteractor) IsHumanChooseSuitTurn() bool {
	return ci.Game.GetPhase() == domain.MacauPhaseChooseSuit && ci.Game.IsHumanTurn()
}

// IsHumanDeclareTurn reports whether the game is currently waiting for the human
// to declare "Macau!" (i.e. the human just reached one card).
func (ci *MacauInteractor) IsHumanDeclareTurn() bool {
	return ci.Game.GetPhase() == domain.MacauPhaseMustDeclare && ci.Game.IsHumanTurn()
}

// runCpuTurns ゲームが終わるか人間の手番またはラウンド/ゲーム終了になるまでCPUターンを実行
func (ci *MacauInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase == MacauPhaseRoundEnd || phase == MacauPhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		switch phase {
		case domain.MacauPhaseChooseSuit:
			ci.Game.CpuChooseSuit()
		case domain.MacauPhaseMustDeclare:
			ci.Game.CpuDeclare()
		case domain.MacauPhasePlay:
			ci.Game.CpuPlay()
		default:
			return
		}
	}
}

const (
	// MacauPhaseRoundEnd ラウンド終了フェーズ (domain からの再エクスポート)
	MacauPhaseRoundEnd = domain.MacauPhaseRoundEnd
	// MacauPhaseGameEnd ゲーム終了フェーズ (domain からの再エクスポート)
	MacauPhaseGameEnd = domain.MacauPhaseGameEnd
)

// RestoreMacauInteractor deserialises JSON into a MacauInteractor.
func RestoreMacauInteractor(data []byte, gp presenter.MacauPresenter) (*MacauInteractor, error) {
	return restoreAndBuild[domain.Macau](data, func(g *domain.Macau) *MacauInteractor {
		return &MacauInteractor{GameBase: GameBase[interfaces.MacauGame]{Game: g}, gp: gp}
	})
}
