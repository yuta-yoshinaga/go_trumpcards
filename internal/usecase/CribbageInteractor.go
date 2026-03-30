package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CribbageInteractorIF クリベッジインタラクターインタフェース
type CribbageInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CribbageConfig) string
	// Discard クリブに2枚捨てる
	Discard(cardIndices []int) string
	// Peg ペギングでカードを出す
	Peg(cardIndex int) string
	// Go Goを宣言する
	Go() string
	// ShowNext ショーフェーズの次のスコア計算
	ShowNext() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CribbageConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CribbageInteractor クリベッジインタラクタークラス
type CribbageInteractor struct {
	g  interfaces.CribbageGame
	gp presenter.CribbagePresenter
}

// NewCribbageInteractor コンストラクタ
func NewCribbageInteractor(g interfaces.CribbageGame, gp presenter.CribbagePresenter) *CribbageInteractor {
	mustNotNil("CribbageInteractor", map[string]any{"g": g, "gp": gp})
	return &CribbageInteractor{g: g, gp: gp}
}

// Reset ゲーム初期化
func (ci *CribbageInteractor) Reset() string {
	ci.g.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CribbageInteractor) ResetWithConfig(cfg domain.CribbageConfig) string {
	return resetWithValidatedConfig(ci.g, ci.gp, cfg, ci.g.SetConfig, ci.Reset)
}

// Discard クリブに2枚捨てる
func (ci *CribbageInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerDiscard(cardIndices)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// Peg ペギングでカードを出す
func (ci *CribbageInteractor) Peg(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerPeg(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// Go Goを宣言する
func (ci *CribbageInteractor) Go() string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerGo()
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// ShowNext ショーフェーズの次のスコア計算
func (ci *CribbageInteractor) ShowNext() string {
	if out, blocked := guardGameEnd(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.ShowNext()
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	return ci.gp.Output(ci.g, nil)
}

// NextRound 次のラウンドへ進む
func (ci *CribbageInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ci.g, ci.gp); blocked {
		return out
	}
	ci.g.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// GetConfig 現在の設定を取得
func (ci *CribbageInteractor) GetConfig() domain.CribbageConfig {
	return ci.g.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *CribbageInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.g)
}

// runCpuTurns CPUターンを実行
func (ci *CribbageInteractor) runCpuTurns() {
	for !ci.g.GetGameEndFlag() {
		phase := ci.g.GetPhase()
		if phase == domain.CribbagePhaseRoundEnd || phase == domain.CribbagePhaseGameEnd ||
			phase == domain.CribbagePhaseShow {
			break
		}
		if ci.g.IsHumanTurn() {
			break
		}
		ci.g.CpuPlay()
	}
}

// Snapshot serialises the game state to JSON for KV persistence.
func (ci *CribbageInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(ci.g)
}

// RestoreCribbageInteractor deserialises JSON into a CribbageInteractor.
func RestoreCribbageInteractor(data []byte, gp presenter.CribbagePresenter) (*CribbageInteractor, error) {
	g, err := restoreGame[domain.Cribbage](data)
	if err != nil {
		return nil, err
	}
	return &CribbageInteractor{g: g, gp: gp}, nil
}
