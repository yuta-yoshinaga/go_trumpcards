//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BourreInteractorIF ブーレインタラクターインタフェース
type BourreInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Decide 人間プレイヤーが参加(true)/フォールド(false)を決める
	Decide(play bool) string
	// Draw 人間プレイヤーが手札を交換する
	Draw(indices []int) string
	// Play 人間プレイヤーがカードをプレイする
	Play(cardIndex int) string
	// NextHand 次のハンドへ進む
	NextHand() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.BourreConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.BourreConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BourreInteractor ブーレインタラクタークラス
type BourreInteractor struct {
	GameBase[interfaces.BourreGame]
	bp presenter.BourrePresenter
}

// NewBourreInteractor コンストラクタ
func NewBourreInteractor(bg interfaces.BourreGame, bp presenter.BourrePresenter) *BourreInteractor {
	mustNotNil("BourreInteractor", map[string]any{"bg": bg, "bp": bp})
	return &BourreInteractor{
		GameBase: GameBase[interfaces.BourreGame]{Game: bg},
		bp:       bp,
	}
}

// Reset ゲーム初期化
func (bi *BourreInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化
func (bi *BourreInteractor) ResetWithConfig(config domain.BourreConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, config, bi.Game.SetConfig, bi.Reset)
}

// Decide 人間プレイヤーが参加/フォールドを決める
func (bi *BourreInteractor) Decide(play bool) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerDecide(play)
	if err == nil {
		bi.runCpuTurns()
	}
	return bi.bp.Output(bi.Game, err)
}

// Draw 人間プレイヤーが手札を交換する
func (bi *BourreInteractor) Draw(indices []int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerDraw(indices)
	if err == nil {
		bi.runCpuTurns()
	}
	return bi.bp.Output(bi.Game, err)
}

// Play 人間プレイヤーがカードをプレイする
func (bi *BourreInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerPlay(cardIndex)
	if err == nil {
		bi.runCpuTurns()
	}
	return bi.bp.Output(bi.Game, err)
}

// NextHand 次のハンドへ進む
func (bi *BourreInteractor) NextHand() string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.NextHand()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を返す
func (bi *BourreInteractor) GetConfig() domain.BourreConfig {
	return bi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (bi *BourreInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// runCpuTurns ゲーム終了・人間の手番・ハンド終了のいずれかになるまでCPUを進める
func (bi *BourreInteractor) runCpuTurns() {
	g := bi.Game
	for i := 0; i < MaxCpuIterations; i++ {
		if g.GetGameEndFlag() {
			return
		}
		phase := g.GetPhase()
		if phase == domain.BourrePhaseRoundEnd || phase == domain.BourrePhaseGameEnd {
			return
		}
		if g.IsHumanTurn() {
			return
		}
		g.CpuPlay()
	}
}

// RestoreBourreInteractor deserialises JSON into a BourreInteractor.
func RestoreBourreInteractor(data []byte, bp presenter.BourrePresenter) (*BourreInteractor, error) {
	return restoreAndBuild[domain.Bourre](data, func(g *domain.Bourre) *BourreInteractor {
		return &BourreInteractor{GameBase: GameBase[interfaces.BourreGame]{Game: g}, bp: bp}
	})
}
