//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// Rummy500InteractorIF Rummy 500インタラクターインタフェース
type Rummy500InteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.Rummy500Config) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札からカードを引く（インデックス指定）
	DrawFromDiscard(idx int) string
	// Meld 手札のカードでメルドを場に出す
	Meld(cardIndices []int) string
	// Layoff 既存メルドにカードを追加する
	Layoff(meldOwner, meldIdx, cardIndex int) string
	// Discard カードを捨ててターンを終える
	Discard(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.Rummy500Config
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// Rummy500Interactor Rummy 500インタラクタークラス
type Rummy500Interactor struct {
	GameBase[interfaces.Rummy500Game]
	gp presenter.Rummy500Presenter
}

// NewRummy500Interactor コンストラクタ
func NewRummy500Interactor(g interfaces.Rummy500Game, gp presenter.Rummy500Presenter) *Rummy500Interactor {
	mustNotNil("Rummy500Interactor", map[string]any{"g": g, "gp": gp})
	return &Rummy500Interactor{GameBase: GameBase[interfaces.Rummy500Game]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *Rummy500Interactor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *Rummy500Interactor) ResetWithConfig(cfg domain.Rummy500Config) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *Rummy500Interactor) DrawFromStock() string {
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
func (ci *Rummy500Interactor) DrawFromDiscard(idx int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDrawFromDiscard(idx); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Meld メルド
func (ci *Rummy500Interactor) Meld(cardIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerMeld(cardIndices); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	return ci.gp.Output(ci.Game, nil)
}

// Layoff レイオフ
func (ci *Rummy500Interactor) Layoff(meldOwner, meldIdx, cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerLayoff(meldOwner, meldIdx, cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨ててターンを終える
func (ci *Rummy500Interactor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndex); err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *Rummy500Interactor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *Rummy500Interactor) GetConfig() domain.Rummy500Config {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *Rummy500Interactor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// Hint ヒントを出力する
func (ci *Rummy500Interactor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// runCpuTurns CPUターンを連続実行
func (ci *Rummy500Interactor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase == domain.Rummy500PhaseRoundEnd || phase == domain.Rummy500PhaseGameEnd {
			return
		}
		if ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestoreRummy500Interactor deserialises JSON into a Rummy500Interactor.
func RestoreRummy500Interactor(data []byte, gp presenter.Rummy500Presenter) (*Rummy500Interactor, error) {
	return restoreAndBuild[domain.Rummy500](data, func(g *domain.Rummy500) *Rummy500Interactor {
		return &Rummy500Interactor{GameBase: GameBase[interfaces.Rummy500Game]{Game: g}, gp: gp}
	})
}
