//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ViraInteractorIF ヴィーラのインタラクターインタフェース
type ViraInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ViraConfig) string
	// Bid 入札する (0=Pass,1=Gask,2=Solo,3=Misère,4=Vira)
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ViraConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ViraInteractor ヴィーラのインタラクタークラス
type ViraInteractor struct {
	GameBase[interfaces.ViraGame]
	sp presenter.ViraPresenter
}

// NewViraInteractor コンストラクタ
func NewViraInteractor(g interfaces.ViraGame, sp presenter.ViraPresenter) *ViraInteractor {
	mustNotNil("ViraInteractor", map[string]any{"g": g, "sp": sp})
	return &ViraInteractor{GameBase: GameBase[interfaces.ViraGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (pi *ViraInteractor) Reset() string {
	pi.Game.Reset()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *ViraInteractor) ResetWithConfig(cfg domain.ViraConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.sp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Bid 入札する (0=Pass,1=Gask,2=Solo,3=Misère,4=Vira)
func (pi *ViraInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.PlayerBid(domain.ViraBid(bid)); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// Play カードをプレイ
func (pi *ViraInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.PlayerPlay(cardIndex); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if pi.Game.GetPhase() == domain.ViraPhaseTrickEnd {
		pi.Game.ResolveTrick()
	}
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (pi *ViraInteractor) NextTrick() string {
	pi.Game.NextTrick()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (pi *ViraInteractor) NextRound() string {
	pi.Game.ScoreRound()
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	pi.Game.NextRound()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *ViraInteractor) GetConfig() domain.ViraConfig {
	return pi.Game.GetConfig()
}

// Hint ヒント取得
func (pi *ViraInteractor) Hint() string {
	return pi.sp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *ViraInteractor) ActionLog() string {
	return pi.sp.ActionLogOutput(pi.Game)
}

// advanceCpu 入札 → プレイ の順に CPU を自動進行させる。
// 入札フェーズでは人間の手番になるまで CPU に入札させ、入札が締まって
// プレイフェーズに入ったら人間の手番までトリックを進める。
func (pi *ViraInteractor) advanceCpu() {
	runCpuBidsLoop(pi.Game, domain.ViraPhaseBid)
	pi.runCpuTurns()
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する。
func (pi *ViraInteractor) runCpuTurns() {
	runCpuTurnsLoop(pi.Game, trickPhases[domain.ViraPhase]{
		play:     domain.ViraPhasePlay,
		trickEnd: domain.ViraPhaseTrickEnd,
		roundEnd: domain.ViraPhaseRoundEnd,
		gameEnd:  domain.ViraPhaseGameEnd,
	})
}

// RestoreViraInteractor deserialises JSON into a ViraInteractor.
func RestoreViraInteractor(data []byte, sp presenter.ViraPresenter) (*ViraInteractor, error) {
	return restoreAndBuild[domain.Vira](data, func(g *domain.Vira) *ViraInteractor {
		return &ViraInteractor{GameBase: GameBase[interfaces.ViraGame]{Game: g}, sp: sp}
	})
}
