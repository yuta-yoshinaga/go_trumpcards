package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CanastaInteractorIF カナスタインタラクターインタフェース
type CanastaInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CanastaConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札の山を取る
	DrawFromDiscard(naturalPairIndices []int) string
	// Meld メルドを出す
	Meld(meldGroups [][]int) string
	// SkipMeld メルドフェーズをスキップ
	SkipMeld() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// GoOut 上がる
	GoOut() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CanastaConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CanastaInteractor カナスタインタラクタークラス
type CanastaInteractor struct {
	g  interfaces.CanastaGame
	gp presenter.CanastaPresenter
}

// NewCanastaInteractor コンストラクタ
func NewCanastaInteractor(g interfaces.CanastaGame, gp presenter.CanastaPresenter) *CanastaInteractor {
	mustNotNil("CanastaInteractor", map[string]any{"g": g, "gp": gp})
	return &CanastaInteractor{g: g, gp: gp}
}

// Reset ゲーム初期化
func (ci *CanastaInteractor) Reset() string {
	ci.g.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CanastaInteractor) ResetWithConfig(cfg domain.CanastaConfig) string {
	return resetWithValidatedConfig(ci.g, ci.gp, cfg, ci.g.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *CanastaInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerDrawFromStock()
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// DrawFromDiscard 捨て札の山を取る
func (ci *CanastaInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerDrawFromDiscard(naturalPairIndices)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// Meld メルドを出す
func (ci *CanastaInteractor) Meld(meldGroups [][]int) string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerMeld(meldGroups)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// SkipMeld メルドフェーズをスキップ
func (ci *CanastaInteractor) SkipMeld() string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerSkipMeld()
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// Discard カードを捨てる
func (ci *CanastaInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerDiscard(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// GoOut 上がる
func (ci *CanastaInteractor) GoOut() string {
	if out, blocked := guardNotPlayable(ci.g, ci.gp); blocked {
		return out
	}
	err := ci.g.PlayerGoOut()
	if err != nil {
		return ci.gp.Output(ci.g, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// NextRound 次のラウンドへ進む
func (ci *CanastaInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ci.g, ci.gp); blocked {
		return out
	}
	ci.g.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.g, nil)
}

// GetConfig 現在の設定を取得
func (ci *CanastaInteractor) GetConfig() domain.CanastaConfig {
	return ci.g.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *CanastaInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.g)
}

// runCpuTurns CPUターンを実行
func (ci *CanastaInteractor) runCpuTurns() {
	for !ci.g.GetGameEndFlag() {
		phase := ci.g.GetPhase()
		if phase == domain.CanastaPhaseRoundEnd || phase == domain.CanastaPhaseGameEnd {
			break
		}
		if ci.g.IsHumanTurn() {
			break
		}
		ci.g.CpuPlay()
	}
}

// Snapshot serialises the game state to JSON for KV persistence.
func (ci *CanastaInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(ci.g)
}

// RestoreCanastaInteractor deserialises JSON into a CanastaInteractor.
func RestoreCanastaInteractor(data []byte, gp presenter.CanastaPresenter) (*CanastaInteractor, error) {
	g, err := restoreGame[domain.Canasta](data)
	if err != nil {
		return nil, err
	}
	return &CanastaInteractor{g: g, gp: gp}, nil
}
