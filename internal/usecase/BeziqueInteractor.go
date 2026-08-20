//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// beziqueCpuTurnGuard は CPU ターン自動進行ループの最大反復回数。
const beziqueCpuTurnGuard = 1000

// BeziqueInteractorIF ベジークインタラクターインタフェース
type BeziqueInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BeziqueConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// DeclareMeld 役を宣言する
	DeclareMeld(meldIndex int) string
	// SkipMeld 役宣言をパスする
	SkipMeld() string
	// NextRound 次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BeziqueConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BeziqueInteractor ベジークインタラクタークラス
type BeziqueInteractor struct {
	GameBase[interfaces.BeziqueGame]
	bp presenter.BeziquePresenter
}

// NewBeziqueInteractor コンストラクタ
func NewBeziqueInteractor(b interfaces.BeziqueGame, bp presenter.BeziquePresenter) *BeziqueInteractor {
	mustNotNil("BeziqueInteractor", map[string]any{"b": b, "bp": bp})
	return &BeziqueInteractor{GameBase: GameBase[interfaces.BeziqueGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化
func (bi *BeziqueInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BeziqueInteractor) ResetWithConfig(cfg domain.BeziqueConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Play カードをプレイ
func (bi *BeziqueInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	if err := bi.Game.PlayerPlay(cardIndex); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// DeclareMeld 役を宣言する
func (bi *BeziqueInteractor) DeclareMeld(meldIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	if err := bi.Game.PlayerDeclareMeld(meldIndex); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// SkipMeld 役宣言をパスする
func (bi *BeziqueInteractor) SkipMeld() string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	if err := bi.Game.PlayerSkipMeld(); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextRound 次のディールへ進む
func (bi *BeziqueInteractor) NextRound() string {
	bi.Game.NextRound()
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BeziqueInteractor) GetConfig() domain.BeziqueConfig {
	return bi.Game.GetConfig()
}

// Hint ヒント取得
func (bi *BeziqueInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BeziqueInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// runCpuTurns ゲームが終わるか、人間の手番 (プレイ/役宣言) か、ディール終了になるまで
// CPU ターンを自動進行する。プレイフェーズでは CpuPlay を、役宣言フェーズでは
// (現在の手番が CPU の場合のみ) CpuMeld を呼ぶ。ディール終了 (RoundEnd) では停止し、
// プレイヤーの NextRound コマンドを待つ。
func (bi *BeziqueInteractor) runCpuTurns() {
	for guard := 0; guard < beziqueCpuTurnGuard; guard++ {
		if bi.Game.GetGameEndFlag() {
			return
		}
		switch bi.Game.GetPhase() {
		case domain.BeziquePhasePlay:
			if bi.Game.IsHumanTurn() {
				return
			}
			bi.Game.CpuPlay()
		case domain.BeziquePhaseMeld:
			// 役宣言フェーズはトリック勝者のもの。人間なら停止して入力を待つ。
			if bi.Game.IsHumanTurn() {
				return
			}
			bi.Game.CpuMeld()
		default:
			// RoundEnd / GameEnd は自動進行しない。
			return
		}
	}
}

// RestoreBeziqueInteractor deserialises JSON into a BeziqueInteractor.
func RestoreBeziqueInteractor(data []byte, bp presenter.BeziquePresenter) (*BeziqueInteractor, error) {
	return restoreAndBuild[domain.Bezique](data, func(g *domain.Bezique) *BeziqueInteractor {
		return &BeziqueInteractor{GameBase: GameBase[interfaces.BeziqueGame]{Game: g}, bp: bp}
	})
}
