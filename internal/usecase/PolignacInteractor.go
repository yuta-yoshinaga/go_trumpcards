//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PolignacInteractorIF ポリニャックインタラクターインタフェース
type PolignacInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PolignacConfig) string
	// DeclareCapot capot を宣言する
	DeclareCapot() string
	// Pass 宣言せずにプレイへ進む
	Pass() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PolignacConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PolignacInteractor ポリニャックインタラクタークラス
type PolignacInteractor struct {
	GameBase[interfaces.PolignacGame]
	pp presenter.PolignacPresenter
}

// NewPolignacInteractor コンストラクタ
func NewPolignacInteractor(p interfaces.PolignacGame, pp presenter.PolignacPresenter) *PolignacInteractor {
	mustNotNil("PolignacInteractor", map[string]any{"p": p, "pp": pp})
	return &PolignacInteractor{GameBase: GameBase[interfaces.PolignacGame]{Game: p}, pp: pp}
}

// Reset ゲーム初期化。**配り終えた時点で宣言フェーズなので CPU は動かさない。**
func (pi *PolignacInteractor) Reset() string {
	pi.Game.Reset()
	return pi.pp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PolignacInteractor) ResetWithConfig(cfg domain.PolignacConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.pp, cfg, pi.Game.SetConfig, pi.Reset)
}

// DeclareCapot capot を宣言してプレイに入る
func (pi *PolignacInteractor) DeclareCapot() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	if err := pi.Game.DeclareCapot(); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// Pass 宣言せずにプレイに入る
func (pi *PolignacInteractor) Pass() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	if err := pi.Game.PassDeclaration(); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// Play カードをプレイ
func (pi *PolignacInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.pp); blocked {
		return out
	}
	if err := pi.Game.PlayerPlay(cardIndex); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// NextRound 次のラウンドへ進む
func (pi *PolignacInteractor) NextRound() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.NextRound()
	// 次のラウンドも宣言フェーズから始まるので、ここでは CPU を回さない。
	return pi.pp.Output(pi.Game, nil)
}

// GiveUp 投了する
func (pi *PolignacInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.GiveUp()
	return pi.pp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *PolignacInteractor) GetConfig() domain.PolignacConfig { return pi.Game.GetConfig() }

// Hint ヒント取得
func (pi *PolignacInteractor) Hint() string { return pi.pp.HintOutput(pi.Game) }

// ActionLog 棋譜を出力する
func (pi *PolignacInteractor) ActionLog() string { return pi.pp.ActionLogOutput(pi.Game) }

// runCpuTurns 人間の手番になるか、ラウンド／ゲームが終わるまで CPU を進める。
//
// **宣言フェーズとラウンド終了では必ず止まる。** 宣言は人間しか行わず、
// ラウンド終了は失点を確認させるための停止点なので、勝手に進めない。
func (pi *PolignacInteractor) runCpuTurns() {
	for turns := 0; !pi.Game.GetGameEndFlag() && !pi.Game.IsHumanTurn(); turns++ {
		// 進まない CpuPlay でハングしないための上限 (#4607 と同じ理由)。
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if pi.Game.GetPhase() != domain.PolignacPhasePlay {
			return
		}
		pi.Game.CpuPlay()
	}
}

// RestorePolignacInteractor deserialises JSON into a PolignacInteractor.
func RestorePolignacInteractor(data []byte, pp presenter.PolignacPresenter) (*PolignacInteractor, error) {
	return restoreAndBuild[domain.Polignac](data, func(g *domain.Polignac) *PolignacInteractor {
		return &PolignacInteractor{GameBase: GameBase[interfaces.PolignacGame]{Game: g}, pp: pp}
	})
}
