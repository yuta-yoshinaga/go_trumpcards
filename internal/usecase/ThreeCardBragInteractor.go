//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ThreeCardBragInteractorIF スリーカード・ブラグのインタラクターインタフェース
type ThreeCardBragInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ThreeCardBragConfig) string
	// See 手札を見て Seen に昇格する
	See() string
	// Bet 現在の賭け単位をコールする
	Bet() string
	// Raise 賭け単位を newStake へ引き上げる
	Raise(newStake int) string
	// Fold 降りる
	Fold() string
	// Show 勝負を要求する
	Show() string
	// NextRound 次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ThreeCardBragConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// threeCardBragMaxCpuTurns はCPU自動進行ループの安全網上限。
const threeCardBragMaxCpuTurns = 2000

// ThreeCardBragInteractor スリーカード・ブラグのインタラクタークラス
type ThreeCardBragInteractor struct {
	GameBase[interfaces.ThreeCardBragGame]
	sp presenter.ThreeCardBragPresenter
}

// NewThreeCardBragInteractor コンストラクタ
func NewThreeCardBragInteractor(g interfaces.ThreeCardBragGame, sp presenter.ThreeCardBragPresenter) *ThreeCardBragInteractor {
	mustNotNil("ThreeCardBragInteractor", map[string]any{"g": g, "sp": sp})
	return &ThreeCardBragInteractor{GameBase: GameBase[interfaces.ThreeCardBragGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (ti *ThreeCardBragInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.sp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *ThreeCardBragInteractor) ResetWithConfig(cfg domain.ThreeCardBragConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// See 手札を見て Seen に昇格する
func (ti *ThreeCardBragInteractor) See() string {
	return ti.runAction(ti.Game.PlayerSee)
}

// Bet 現在の賭け単位をコールする
func (ti *ThreeCardBragInteractor) Bet() string {
	return ti.runAction(ti.Game.PlayerBet)
}

// Raise 賭け単位を newStake へ引き上げる
func (ti *ThreeCardBragInteractor) Raise(newStake int) string {
	return ti.runAction(func() error { return ti.Game.PlayerRaise(newStake) })
}

// Fold 降りる
func (ti *ThreeCardBragInteractor) Fold() string {
	return ti.runAction(ti.Game.PlayerFold)
}

// Show 勝負を要求する
func (ti *ThreeCardBragInteractor) Show() string {
	return ti.runAction(ti.Game.PlayerShow)
}

// runAction はゲーム終了ガード→アクション実行→CPU自動進行→出力の共通フローを実行する。
func (ti *ThreeCardBragInteractor) runAction(action func() error) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.sp.Output(ti.Game, nil)
}

// NextRound 次のディールへ進む
func (ti *ThreeCardBragInteractor) NextRound() string {
	ti.Game.NextRound()
	ti.runCpuTurns()
	return ti.sp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *ThreeCardBragInteractor) GetConfig() domain.ThreeCardBragConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *ThreeCardBragInteractor) Hint() string {
	return ti.sp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *ThreeCardBragInteractor) ActionLog() string {
	return ti.sp.ActionLogOutput(ti.Game)
}

// runCpuTurns はベッティングフェーズの CPU 手番を自動で進める。
// Showdown はアクション内部で RoundEnd/GameEnd へ遷移する transient フェーズのため、
// 通常はループ間で Betting / RoundEnd / GameEnd のいずれかしか観測されない。
// RoundEnd では明示的な NextRound を待つため自動進行しない。
func (ti *ThreeCardBragInteractor) runCpuTurns() {
	for guard := 0; guard < threeCardBragMaxCpuTurns && !ti.Game.GetGameEndFlag(); guard++ {
		if ti.Game.GetPhase() == domain.ThreeCardBragPhaseBetting && !ti.Game.IsHumanTurn() {
			ti.Game.CpuAct()
			continue
		}
		// 人間の手番 / RoundEnd / GameEnd → 停止。
		break
	}
}

// RestoreThreeCardBragInteractor deserialises JSON into a ThreeCardBragInteractor.
func RestoreThreeCardBragInteractor(data []byte, sp presenter.ThreeCardBragPresenter) (*ThreeCardBragInteractor, error) {
	return restoreAndBuild[domain.ThreeCardBrag](data, func(g *domain.ThreeCardBrag) *ThreeCardBragInteractor {
		return &ThreeCardBragInteractor{GameBase: GameBase[interfaces.ThreeCardBragGame]{Game: g}, sp: sp}
	})
}
