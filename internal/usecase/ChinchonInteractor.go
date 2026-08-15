//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ChinchonInteractorIF チンチョンインタラクターインタフェース
type ChinchonInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ChinchonConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札からカードを引く
	DrawFromDiscard() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// Knock 1枚捨ててノックする
	Knock(cardIndex int) string
	// Layoff ノッカーのメルドにカードをレイオフする
	Layoff(cardIndices []int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ChinchonConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ChinchonInteractor チンチョンインタラクタークラス
type ChinchonInteractor struct {
	GameBase[interfaces.ChinchonGame]
	gp presenter.ChinchonPresenter
}

// NewChinchonInteractor コンストラクタ
func NewChinchonInteractor(g interfaces.ChinchonGame, gp presenter.ChinchonPresenter) *ChinchonInteractor {
	mustNotNil("ChinchonInteractor", map[string]any{"g": g, "gp": gp})
	return &ChinchonInteractor{GameBase: GameBase[interfaces.ChinchonGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *ChinchonInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ChinchonInteractor) ResetWithConfig(cfg domain.ChinchonConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *ChinchonInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromStock(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawFromDiscard 捨て札からカードを引く
func (ci *ChinchonInteractor) DrawFromDiscard() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromDiscard(); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨てる
func (ci *ChinchonInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Knock 1枚捨ててノックする
func (ci *ChinchonInteractor) Knock(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerKnock(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Layoff ノッカーのメルドにカードをレイオフする
func (ci *ChinchonInteractor) Layoff(cardIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerLayoff(cardIndices); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *ChinchonInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *ChinchonInteractor) GetConfig() domain.ChinchonConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *ChinchonInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行する。
//
// ラウンドはノックか山札切れで必ず終了し、脱落でマッチも有限回で終わるため通常は
// 早期に脱出するが、無限ループ防止の安全上限を設けている。
func (ci *ChinchonInteractor) runCpuTurns() {
	for i := 0; i < chinchonMaxCpuTurns; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase == domain.ChinchonPhaseRoundEnd || phase == domain.ChinchonPhaseGameEnd {
			return
		}
		if ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
	}
}

// chinchonMaxCpuTurns は runCpuTurns の安全上限。
const chinchonMaxCpuTurns = 100000

// RestoreChinchonInteractor deserialises JSON into a ChinchonInteractor.
func RestoreChinchonInteractor(data []byte, gp presenter.ChinchonPresenter) (*ChinchonInteractor, error) {
	return restoreAndBuild[domain.Chinchon](data, func(g *domain.Chinchon) *ChinchonInteractor {
		return &ChinchonInteractor{GameBase: GameBase[interfaces.ChinchonGame]{Game: g}, gp: gp}
	})
}
