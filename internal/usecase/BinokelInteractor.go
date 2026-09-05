package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BinokelInteractorIF ビノクルインタラクターインタフェース
type BinokelInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BinokelConfig) string
	// Bid ビッドする
	Bid(amount int) string
	// Pass パスする
	Pass() string
	// DiscardToDabb Dabbへカードを捨てる
	DiscardToDabb(cardIndices []int) string
	// CallTrump トランプスートを宣言する
	CallTrump(suit int) string
	// ConfirmMelds メルドを確認
	ConfirmMelds() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BinokelConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BinokelInteractor ビノクルインタラクタークラス
type BinokelInteractor struct {
	GameBase[interfaces.BinokelGame]
	pp presenter.BinokelPresenter
}

// NewBinokelInteractor コンストラクタ
func NewBinokelInteractor(p interfaces.BinokelGame, pp presenter.BinokelPresenter) *BinokelInteractor {
	mustNotNil("BinokelInteractor", map[string]any{"p": p, "pp": pp})
	return &BinokelInteractor{GameBase: GameBase[interfaces.BinokelGame]{Game: p}, pp: pp}
}

// Reset ゲーム初期化
func (pi *BinokelInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *BinokelInteractor) ResetWithConfig(cfg domain.BinokelConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.pp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Bid ビッドする
func (pi *BinokelInteractor) Bid(amount int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerBid(amount)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// Pass パスする
func (pi *BinokelInteractor) Pass() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerPass()
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// DiscardToDabb Dabbへカードを捨てる
func (pi *BinokelInteractor) DiscardToDabb(cardIndices []int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerDiscardToDabb(cardIndices)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// CallTrump トランプスートを宣言する
func (pi *BinokelInteractor) CallTrump(suit int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerCallTrump(suit)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	return pi.pp.Output(pi.Game, nil)
}

// ConfirmMelds メルドを確認してプレイフェーズに進む
func (pi *BinokelInteractor) ConfirmMelds() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.ConfirmMelds()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// Play カードをプレイ
func (pi *BinokelInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (pi *BinokelInteractor) NextTrick() string {
	pi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.NextTrick()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// NextRound 次のラウンドへ進む
func (pi *BinokelInteractor) NextRound() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.NextRound()
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *BinokelInteractor) GetConfig() domain.BinokelConfig {
	return pi.Game.GetConfig()
}

// Hint ヒント取得
func (pi *BinokelInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *BinokelInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// runCpuBids ビッド、Dabb交換、およびトランプ宣言フェーズでCPUを自動実行する
func (pi *BinokelInteractor) runCpuBids() {
	for i := 0; i < MaxCpuIterations; i++ {
		if pi.Game.GetGameEndFlag() {
			return
		}
		phase := pi.Game.GetPhase()
		if phase == domain.BinokelPhaseBid {
			if pi.Game.IsHumanBidTurn() {
				break
			}
			pi.Game.CpuBid()
		} else if phase == domain.BinokelPhaseDabb {
			if pi.Game.IsHumanDabbTurn() {
				break
			}
			pi.Game.CpuDiscardToDabb()
		} else if phase == domain.BinokelPhaseTrump {
			if pi.Game.IsHumanTurn() {
				break
			}
			pi.Game.CpuCallTrump()
		} else {
			break
		}
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (pi *BinokelInteractor) runCpuTurns() {
	runCpuTurnsLoop(pi.Game, trickPhases[domain.BinokelPhase]{
		play:     domain.BinokelPhasePlay,
		trickEnd: domain.BinokelPhaseTrickEnd,
		roundEnd: domain.BinokelPhaseRoundEnd,
		gameEnd:  domain.BinokelPhaseGameEnd,
	})
}

// RestoreBinokelInteractor deserialises JSON into a BinokelInteractor.
func RestoreBinokelInteractor(data []byte, pp presenter.BinokelPresenter) (*BinokelInteractor, error) {
	return restoreAndBuild[domain.Binokel](data, func(g *domain.Binokel) *BinokelInteractor {
		return &BinokelInteractor{GameBase: GameBase[interfaces.BinokelGame]{Game: g}, pp: pp}
	})
}
