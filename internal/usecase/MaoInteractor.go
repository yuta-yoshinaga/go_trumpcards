//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MaoInteractorIF マオインタラクターインタフェース
type MaoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MaoConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// ChooseSuit スートを選択 (8を出した後)
	ChooseSuit(suit int) string
	// Draw カードを引く (ペナルティ中はスタックを引き受ける)
	Draw() string
	// Declare 「マオ！」と宣言する
	Declare() string
	// SkipDeclare 宣言をスキップしてペナルティを受ける
	SkipDeclare() string
	// DeclareWord 隠しルールに従って言葉を宣言する
	DeclareWord(word string) string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MaoConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
	// IsHumanChooseSuitTurn reports whether the human just played an 8 and the
	// game is now waiting for them to pick a suit.
	IsHumanChooseSuitTurn() bool
	// IsHumanDeclareTurn reports whether the human just reached one card and the
	// game is now waiting for them to declare "Mao!".
	IsHumanDeclareTurn() bool
	// IsHumanAwaitingWord reports whether the human must now declare the secret
	// rule's word before play can continue.
	IsHumanAwaitingWord() bool
}

// MaoInteractor マオインタラクタークラス
type MaoInteractor struct {
	GameBase[interfaces.MaoGame]
	gp presenter.MaoPresenter
}

// NewMaoInteractor コンストラクタ
func NewMaoInteractor(g interfaces.MaoGame, gp presenter.MaoPresenter) *MaoInteractor {
	mustNotNil("MaoInteractor", map[string]any{"g": g, "gp": gp})
	return &MaoInteractor{GameBase: GameBase[interfaces.MaoGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *MaoInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *MaoInteractor) ResetWithConfig(cfg domain.MaoConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードをプレイ
func (ci *MaoInteractor) Play(cardIndex int) string {
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
func (ci *MaoInteractor) ChooseSuit(suit int) string {
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
func (ci *MaoInteractor) Draw() string {
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

// Declare 「マオ！」と宣言する
func (ci *MaoInteractor) Declare() string {
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
func (ci *MaoInteractor) SkipDeclare() string {
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

// DeclareWord 隠しルールに従って言葉を宣言する
func (ci *MaoInteractor) DeclareWord(word string) string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDeclareWord(word)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *MaoInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *MaoInteractor) GetConfig() domain.MaoConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *MaoInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// IsHumanChooseSuitTurn reports whether the game is currently waiting for the
// human to pick a suit (i.e. the human just played an 8).
func (ci *MaoInteractor) IsHumanChooseSuitTurn() bool {
	return ci.Game.GetPhase() == domain.MaoPhaseChooseSuit && ci.Game.IsHumanTurn()
}

// IsHumanDeclareTurn reports whether the game is currently waiting for the human
// to declare "Mao!" (i.e. the human just reached one card).
func (ci *MaoInteractor) IsHumanDeclareTurn() bool {
	return ci.Game.GetPhase() == domain.MaoPhaseMustDeclare && ci.Game.IsHumanTurn()
}

// IsHumanAwaitingWord reports whether the game is currently waiting for the
// human to declare the secret rule's word.
func (ci *MaoInteractor) IsHumanAwaitingWord() bool {
	return ci.Game.GetAwaitingWord()
}

// runCpuTurns ゲームが終わるか人間の手番/宣言待ち/ラウンド・ゲーム終了になるまで
// CPUターンを実行する。隠しルールの宣言待ち (awaitingWord) のときは、人間が
// 言葉を宣言するまでCPUを進めない。
func (ci *MaoInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		// 人間が隠しルールの言葉を宣言すべき状態なら一旦停止する。
		if ci.Game.GetAwaitingWord() {
			break
		}
		phase := ci.Game.GetPhase()
		if phase == MaoPhaseRoundEnd || phase == MaoPhaseGameEnd {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		switch phase {
		case domain.MaoPhaseChooseSuit:
			ci.Game.CpuChooseSuit()
		case domain.MaoPhaseMustDeclare:
			ci.Game.CpuDeclare()
		case domain.MaoPhasePlay:
			ci.Game.CpuPlay()
		default:
			return
		}
	}
}

const (
	// MaoPhaseRoundEnd ラウンド終了フェーズ (domain からの再エクスポート)
	MaoPhaseRoundEnd = domain.MaoPhaseRoundEnd
	// MaoPhaseGameEnd ゲーム終了フェーズ (domain からの再エクスポート)
	MaoPhaseGameEnd = domain.MaoPhaseGameEnd
)

// RestoreMaoInteractor deserialises JSON into a MaoInteractor.
func RestoreMaoInteractor(data []byte, gp presenter.MaoPresenter) (*MaoInteractor, error) {
	return restoreAndBuild[domain.Mao](data, func(g *domain.Mao) *MaoInteractor {
		return &MaoInteractor{GameBase: GameBase[interfaces.MaoGame]{Game: g}, gp: gp}
	})
}
