//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ecarteCpuTurnGuard は CPU ターン自動進行ループの最大反復回数。
const ecarteCpuTurnGuard = 1000

// EcarteInteractorIF エカルテインタラクターインタフェース
type EcarteInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.EcarteConfig) string
	// Propose elder が交換を提案する
	Propose() string
	// Stand elder が交換せずに勝負する
	Stand() string
	// Respond 親が提案に承諾 (accept=true) か拒否 (accept=false) する
	Respond(accept bool) string
	// Discard 現在の手番プレイヤーが捨て札を選んで引き直す
	Discard(indices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.EcarteConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// EcarteInteractor エカルテインタラクタークラス
type EcarteInteractor struct {
	GameBase[interfaces.EcarteGame]
	ep presenter.EcartePresenter
}

// NewEcarteInteractor コンストラクタ
func NewEcarteInteractor(b interfaces.EcarteGame, ep presenter.EcartePresenter) *EcarteInteractor {
	mustNotNil("EcarteInteractor", map[string]any{"b": b, "ep": ep})
	return &EcarteInteractor{GameBase: GameBase[interfaces.EcarteGame]{Game: b}, ep: ep}
}

// Reset ゲーム初期化
func (ei *EcarteInteractor) Reset() string {
	ei.Game.Reset()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ei *EcarteInteractor) ResetWithConfig(cfg domain.EcarteConfig) string {
	return resetWithValidatedConfig(ei.Game, ei.ep, cfg, ei.Game.SetConfig, ei.Reset)
}

// Propose elder が交換を提案する
func (ei *EcarteInteractor) Propose() string {
	if out, blocked := guardNotPlayable(ei.Game, ei.ep); blocked {
		return out
	}
	if err := ei.Game.PlayerPropose(); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// Stand elder が交換せずに勝負する
func (ei *EcarteInteractor) Stand() string {
	if out, blocked := guardNotPlayable(ei.Game, ei.ep); blocked {
		return out
	}
	if err := ei.Game.PlayerStand(); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// Respond 親が提案に承諾/拒否する
func (ei *EcarteInteractor) Respond(accept bool) string {
	if out, blocked := guardNotPlayable(ei.Game, ei.ep); blocked {
		return out
	}
	if err := ei.Game.PlayerRespond(accept); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// Discard 捨て札を選んで引き直す
func (ei *EcarteInteractor) Discard(indices []int) string {
	if out, blocked := guardNotPlayable(ei.Game, ei.ep); blocked {
		return out
	}
	if err := ei.Game.PlayerDiscard(indices); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// Play カードをプレイ
func (ei *EcarteInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ei.Game, ei.ep); blocked {
		return out
	}
	if err := ei.Game.PlayerPlay(cardIndex); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// NextRound 次のディールへ進む
func (ei *EcarteInteractor) NextRound() string {
	ei.Game.NextRound()
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// GetConfig 現在の設定を取得
func (ei *EcarteInteractor) GetConfig() domain.EcarteConfig {
	return ei.Game.GetConfig()
}

// Hint ヒント取得
func (ei *EcarteInteractor) Hint() string {
	return ei.ep.HintOutput(ei.Game)
}

// ActionLog 棋譜を出力する
func (ei *EcarteInteractor) ActionLog() string {
	return ei.ep.ActionLogOutput(ei.Game)
}

// runCpuTurns ゲームが終わるか、人間の手番 (交換/プレイ) か、ディール終了になるまで
// CPU ターンを自動進行する。交換フェーズでは現在の手番が CPU の場合に CpuExchange を、
// プレイフェーズでは CpuPlay を呼ぶ。ディール終了 (RoundEnd) では停止し、プレイヤーの
// NextRound コマンドを待つ (NextRound が再度このループを起動する)。交換フェーズの手番は
// elder と親の間で交互に移るため、現在の手番が CPU の場合のみ CpuExchange を呼ぶ。
func (ei *EcarteInteractor) runCpuTurns() {
	for guard := 0; guard < ecarteCpuTurnGuard; guard++ {
		if ei.Game.GetGameEndFlag() {
			return
		}
		switch ei.Game.GetPhase() {
		case domain.EcartePhaseExchange:
			if ei.Game.IsHumanTurn() {
				return
			}
			ei.Game.CpuExchange()
		case domain.EcartePhasePlay:
			if ei.Game.IsHumanTurn() {
				return
			}
			ei.Game.CpuPlay()
		default:
			// RoundEnd / GameEnd は自動進行しない。
			return
		}
	}
}

// RestoreEcarteInteractor deserialises JSON into an EcarteInteractor.
func RestoreEcarteInteractor(data []byte, ep presenter.EcartePresenter) (*EcarteInteractor, error) {
	return restoreAndBuild[domain.Ecarte](data, func(g *domain.Ecarte) *EcarteInteractor {
		return &EcarteInteractor{GameBase: GameBase[interfaces.EcarteGame]{Game: g}, ep: ep}
	})
}
