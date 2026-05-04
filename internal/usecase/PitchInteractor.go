package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PitchInteractorIF ピッチインタラクターインタフェース
type PitchInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PitchConfig) string
	// Bid ビッドを宣言 (0=pass, 2..4)
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PitchConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PitchInteractor ピッチインタラクタークラス
type PitchInteractor struct {
	GameBase[interfaces.PitchGame]
	pp presenter.PitchPresenter
}

// NewPitchInteractor コンストラクタ
func NewPitchInteractor(p interfaces.PitchGame, pp presenter.PitchPresenter) *PitchInteractor {
	mustNotNil("PitchInteractor", map[string]any{"p": p, "pp": pp})
	return &PitchInteractor{GameBase: GameBase[interfaces.PitchGame]{Game: p}, pp: pp}
}

// Reset ゲーム初期化
func (pi *PitchInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuBids()
	if pi.Game.GetPhase() == domain.PitchPhasePlay {
		pi.runCpuTurns()
	}
	return pi.pp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PitchInteractor) ResetWithConfig(cfg domain.PitchConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.pp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Bid ビッドを宣言 (0=pass, 2..4)
func (pi *PitchInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerBid(bid)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuBids()
	if pi.Game.GetPhase() == domain.PitchPhasePlay {
		pi.runCpuTurns()
	}
	return pi.pp.Output(pi.Game, nil)
}

// Play カードをプレイ
func (pi *PitchInteractor) Play(cardIndex int) string {
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
func (pi *PitchInteractor) NextTrick() string {
	pi.Game.NextTrick()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (pi *PitchInteractor) NextRound() string {
	pi.Game.ScoreRound()
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.NextRound()
	pi.runCpuBids()
	if pi.Game.GetPhase() == domain.PitchPhasePlay {
		pi.runCpuTurns()
	}
	return pi.pp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *PitchInteractor) GetConfig() domain.PitchConfig {
	return pi.Game.GetConfig()
}

// Hint ヒント取得
func (pi *PitchInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *PitchInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// runCpuBids ヒューマンのビッド番が来るかビッドフェーズが終了するまでCPUビッドを実行
func (pi *PitchInteractor) runCpuBids() {
	runCpuBidsLoop(pi.Game, domain.PitchPhaseBid)
}

// runCpuTurns 人間の手番もしくはトリック/ラウンド終了になるまでCPUターンを実行
func (pi *PitchInteractor) runCpuTurns() {
	runCpuTurnsLoop(pi.Game, trickPhases[domain.PitchPhase]{
		play:     domain.PitchPhasePlay,
		trickEnd: domain.PitchPhaseTrickEnd,
		roundEnd: domain.PitchPhaseRoundEnd,
		gameEnd:  domain.PitchPhaseGameEnd,
	})
}

// RestorePitchInteractor deserialises JSON into a PitchInteractor.
func RestorePitchInteractor(data []byte, pp presenter.PitchPresenter) (*PitchInteractor, error) {
	return restoreAndBuild[domain.Pitch](data, func(g *domain.Pitch) *PitchInteractor {
		return &PitchInteractor{GameBase: GameBase[interfaces.PitchGame]{Game: g}, pp: pp}
	})
}
