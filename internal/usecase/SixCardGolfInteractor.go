package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SixCardGolfInteractorIF シックスカードゴルフインタラクターインタフェース
type SixCardGolfInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SixCardGolfConfig) string
	// FlipInitial セットアップ時にカードをめくる
	FlipInitial(pos int) string
	// DrawStock 山札から引く
	DrawStock() string
	// DrawDiscard 捨て札から引く
	DrawDiscard() string
	// SwapCard 引いたカードとグリッド位置を交換
	SwapCard(pos int) string
	// DiscardDrawn 引いたカードを捨てる
	DiscardDrawn() string
	// FlipCard 捨て後に伏せ札をめくる
	FlipCard(pos int) string
	// SkipFlip めくりスキップ
	SkipFlip() string
	// NextRound 次のラウンドへ
	NextRound() string
	// GetConfig 設定取得
	GetConfig() domain.SixCardGolfConfig
	// ActionLog 棋譜
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// SixCardGolfInteractor シックスカードゴルフインタラクター
type SixCardGolfInteractor struct {
	GameBase[interfaces.SixCardGolfGame]
	gp presenter.SixCardGolfPresenter
}

// NewSixCardGolfInteractor コンストラクタ
func NewSixCardGolfInteractor(g interfaces.SixCardGolfGame, gp presenter.SixCardGolfPresenter) *SixCardGolfInteractor {
	mustNotNil("SixCardGolfInteractor", map[string]any{"g": g, "gp": gp})
	return &SixCardGolfInteractor{GameBase: GameBase[interfaces.SixCardGolfGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *SixCardGolfInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *SixCardGolfInteractor) ResetWithConfig(cfg domain.SixCardGolfConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// FlipInitial セットアップ時にカードをめくる
func (ci *SixCardGolfInteractor) FlipInitial(pos int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.FlipInitial(pos)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawStock 山札から引く
func (ci *SixCardGolfInteractor) DrawStock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.DrawStock()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	return ci.gp.Output(ci.Game, nil)
}

// DrawDiscard 捨て札から引く
func (ci *SixCardGolfInteractor) DrawDiscard() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.DrawDiscard()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	return ci.gp.Output(ci.Game, nil)
}

// SwapCard 引いたカードとグリッド位置を交換
func (ci *SixCardGolfInteractor) SwapCard(pos int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.SwapCard(pos)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DiscardDrawn 引いたカードを捨てる
func (ci *SixCardGolfInteractor) DiscardDrawn() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.DiscardDrawn()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	if ci.Game.GetCanFlip() {
		return ci.gp.Output(ci.Game, nil)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// FlipCard 捨て後に伏せ札をめくる
func (ci *SixCardGolfInteractor) FlipCard(pos int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.FlipCard(pos)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// SkipFlip めくりスキップ
func (ci *SixCardGolfInteractor) SkipFlip() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.SkipFlip()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ
func (ci *SixCardGolfInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 設定取得
func (ci *SixCardGolfInteractor) GetConfig() domain.SixCardGolfConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜
func (ci *SixCardGolfInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// Hint ヒントを出力する
func (ci *SixCardGolfInteractor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// runCpuTurns CPUターンループ
func (ci *SixCardGolfInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() {
			return
		}
		phase := ci.Game.GetPhase()
		if phase == domain.SixCardGolfPhaseRoundOver || phase == domain.SixCardGolfPhaseGameOver {
			return
		}
		if ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
	}
}

// RestoreSixCardGolfInteractor JSON復元
func RestoreSixCardGolfInteractor(data []byte, gp presenter.SixCardGolfPresenter) (*SixCardGolfInteractor, error) {
	return restoreAndBuild[domain.SixCardGolf](data, func(g *domain.SixCardGolf) *SixCardGolfInteractor {
		return &SixCardGolfInteractor{GameBase: GameBase[interfaces.SixCardGolfGame]{Game: g}, gp: gp}
	})
}
