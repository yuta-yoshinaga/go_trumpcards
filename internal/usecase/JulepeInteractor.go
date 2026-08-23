//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// JulepeInteractorIF フレペインタラクターインタフェース
type JulepeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.JulepeConfig) string
	// Play 参加する
	Play() string
	// Pass 降りる
	Pass() string
	// PlayCard カードをプレイ
	PlayCard(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.JulepeConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// JulepeInteractor フレペインタラクタークラス
type JulepeInteractor struct {
	GameBase[interfaces.JulepeGame]
	rp presenter.JulepePresenter
}

// NewJulepeInteractor コンストラクタ
func NewJulepeInteractor(r interfaces.JulepeGame, rp presenter.JulepePresenter) *JulepeInteractor {
	mustNotNil("JulepeInteractor", map[string]any{"r": r, "rp": rp})
	return &JulepeInteractor{GameBase: GameBase[interfaces.JulepeGame]{Game: r}, rp: rp}
}

// Reset ゲーム初期化。**配り終えた時点は選択フェーズなので CPU は動かさない。**
func (ri *JulepeInteractor) Reset() string {
	ri.Game.Reset()
	return ri.rp.Output(ri.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ri *JulepeInteractor) ResetWithConfig(cfg domain.JulepeConfig) string {
	return resetWithValidatedConfig(ri.Game, ri.rp, cfg, ri.Game.SetConfig, ri.Reset)
}

// Play 参加する
func (ri *JulepeInteractor) Play() string { return ri.decide(true) }

// Pass 降りる
func (ri *JulepeInteractor) Pass() string { return ri.decide(false) }

// decide 参加 / 降りるの共通処理
func (ri *JulepeInteractor) decide(play bool) string {
	if out, blocked := guardGameEnd(ri.Game, ri.rp); blocked {
		return out
	}
	if err := ri.Game.Decide(play); err != nil {
		return ri.rp.Output(ri.Game, err)
	}
	ri.runCpuTurns()
	return ri.rp.Output(ri.Game, nil)
}

// PlayCard カードをプレイ
func (ri *JulepeInteractor) PlayCard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ri.Game, ri.rp); blocked {
		return out
	}
	if err := ri.Game.PlayerPlay(cardIndex); err != nil {
		return ri.rp.Output(ri.Game, err)
	}
	ri.runCpuTurns()
	return ri.rp.Output(ri.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ri *JulepeInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ri.Game, ri.rp); blocked {
		return out
	}
	ri.Game.NextRound()
	// 次のラウンドも選択フェーズから始まるので、ここでは CPU を回さない。
	return ri.rp.Output(ri.Game, nil)
}

// GiveUp 投了する
func (ri *JulepeInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ri.Game, ri.rp); blocked {
		return out
	}
	ri.Game.GiveUp()
	return ri.rp.Output(ri.Game, nil)
}

// GetConfig 現在の設定を取得
func (ri *JulepeInteractor) GetConfig() domain.JulepeConfig { return ri.Game.GetConfig() }

// Hint ヒント取得
func (ri *JulepeInteractor) Hint() string { return ri.rp.HintOutput(ri.Game) }

// ActionLog 棋譜を出力する
func (ri *JulepeInteractor) ActionLog() string { return ri.rp.ActionLogOutput(ri.Game) }

// runCpuTurns 人間の手番になるか、ラウンド／ゲームが終わるまで CPU を進める。
//
// **人間が降りたラウンドは最後まで自動で回す。** 手番が回ってこないので
// IsHumanTurn() は永久に false になり、そこで止めると進まなくなる。
func (ri *JulepeInteractor) runCpuTurns() {
	for turns := 0; !ri.Game.GetGameEndFlag() && !ri.Game.IsHumanTurn(); turns++ {
		// 進まない CpuPlay でハングしないための上限 (#4607 と同じ理由)。
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if ri.Game.GetPhase() != domain.JulepePhasePlay {
			return
		}
		ri.Game.CpuPlay()
	}
}

// RestoreJulepeInteractor deserialises JSON into a JulepeInteractor.
func RestoreJulepeInteractor(data []byte, rp presenter.JulepePresenter) (*JulepeInteractor, error) {
	return restoreAndBuild[domain.Julepe](data, func(g *domain.Julepe) *JulepeInteractor {
		return &JulepeInteractor{GameBase: GameBase[interfaces.JulepeGame]{Game: g}, rp: rp}
	})
}
