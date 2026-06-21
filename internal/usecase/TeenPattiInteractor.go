//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TeenPattiInteractorIF ティーン・パティのインタラクターインタフェース
type TeenPattiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TeenPattiConfig) string
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
	// RequestSideShow サイドショーを申請する
	RequestSideShow() string
	// RespondSideShow サイドショー申請に応答する (accept=受諾)
	RespondSideShow(accept bool) string
	// NextRound 次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TeenPattiConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// teenPattiMaxCpuTurns はCPU自動進行ループの安全網上限。
const teenPattiMaxCpuTurns = 2000

// TeenPattiInteractor ティーン・パティのインタラクタークラス
type TeenPattiInteractor struct {
	GameBase[interfaces.TeenPattiGame]
	sp presenter.TeenPattiPresenter
}

// NewTeenPattiInteractor コンストラクタ
func NewTeenPattiInteractor(g interfaces.TeenPattiGame, sp presenter.TeenPattiPresenter) *TeenPattiInteractor {
	mustNotNil("TeenPattiInteractor", map[string]any{"g": g, "sp": sp})
	return &TeenPattiInteractor{GameBase: GameBase[interfaces.TeenPattiGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (ti *TeenPattiInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.sp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TeenPattiInteractor) ResetWithConfig(cfg domain.TeenPattiConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// See 手札を見て Seen に昇格する
func (ti *TeenPattiInteractor) See() string {
	return ti.runAction(ti.Game.PlayerSee)
}

// Bet 現在の賭け単位をコールする
func (ti *TeenPattiInteractor) Bet() string {
	return ti.runAction(ti.Game.PlayerBet)
}

// Raise 賭け単位を newStake へ引き上げる
func (ti *TeenPattiInteractor) Raise(newStake int) string {
	return ti.runAction(func() error { return ti.Game.PlayerRaise(newStake) })
}

// Fold 降りる
func (ti *TeenPattiInteractor) Fold() string {
	return ti.runAction(ti.Game.PlayerFold)
}

// Show 勝負を要求する
func (ti *TeenPattiInteractor) Show() string {
	return ti.runAction(ti.Game.PlayerShow)
}

// RequestSideShow サイドショーを申請する
func (ti *TeenPattiInteractor) RequestSideShow() string {
	return ti.runAction(ti.Game.PlayerRequestSideShow)
}

// RespondSideShow サイドショー申請に応答する
func (ti *TeenPattiInteractor) RespondSideShow(accept bool) string {
	return ti.runAction(func() error { return ti.Game.PlayerRespondSideShow(accept) })
}

// runAction はゲーム終了ガード→アクション実行→CPU自動進行→出力の共通フローを実行する。
func (ti *TeenPattiInteractor) runAction(action func() error) string {
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
func (ti *TeenPattiInteractor) NextRound() string {
	ti.Game.NextRound()
	ti.runCpuTurns()
	return ti.sp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TeenPattiInteractor) GetConfig() domain.TeenPattiConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TeenPattiInteractor) Hint() string {
	return ti.sp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TeenPattiInteractor) ActionLog() string {
	return ti.sp.ActionLogOutput(ti.Game)
}

// runCpuTurns はベッティング/サイドショーフェーズの CPU 手番を自動で進める。
// CpuAct は Betting と SideShow の両方を処理する。Showdown はアクション内部で
// RoundEnd/GameEnd へ遷移する transient フェーズのため、通常はループ間で
// Betting / SideShow / RoundEnd / GameEnd のいずれかしか観測されない。
// RoundEnd では明示的な NextRound を待つため自動進行しない。
func (ti *TeenPattiInteractor) runCpuTurns() {
	for guard := 0; guard < teenPattiMaxCpuTurns && !ti.Game.GetGameEndFlag(); guard++ {
		ph := ti.Game.GetPhase()
		if (ph == domain.TeenPattiPhaseBetting || ph == domain.TeenPattiPhaseSideShow) && !ti.Game.IsHumanTurn() {
			ti.Game.CpuAct()
			continue
		}
		// 人間の手番 / RoundEnd / GameEnd → 停止。
		break
	}
}

// RestoreTeenPattiInteractor deserialises JSON into a TeenPattiInteractor.
func RestoreTeenPattiInteractor(data []byte, sp presenter.TeenPattiPresenter) (*TeenPattiInteractor, error) {
	return restoreAndBuild[domain.TeenPatti](data, func(g *domain.TeenPatti) *TeenPattiInteractor {
		return &TeenPattiInteractor{GameBase: GameBase[interfaces.TeenPattiGame]{Game: g}, sp: sp}
	})
}
