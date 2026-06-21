//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TwentyNineInteractorIF トゥエンティナイン (29) のインタラクターインタフェース
type TwentyNineInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TwentyNineConfig) string
	// Bid 入札する (0=Pass,16,20,24,28)
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TwentyNineConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TwentyNineInteractor トゥエンティナイン (29) のインタラクタークラス
type TwentyNineInteractor struct {
	GameBase[interfaces.TwentyNineGame]
	sp presenter.TwentyNinePresenter
}

// NewTwentyNineInteractor コンストラクタ
func NewTwentyNineInteractor(g interfaces.TwentyNineGame, sp presenter.TwentyNinePresenter) *TwentyNineInteractor {
	mustNotNil("TwentyNineInteractor", map[string]any{"g": g, "sp": sp})
	return &TwentyNineInteractor{GameBase: GameBase[interfaces.TwentyNineGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (ti *TwentyNineInteractor) Reset() string {
	ti.Game.Reset()
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TwentyNineInteractor) ResetWithConfig(cfg domain.TwentyNineConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Bid 入札する (0=Pass,16,20,24,28)
func (ti *TwentyNineInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	if err := ti.Game.PlayerBid(domain.TwentyNineBid(bid)); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// Play カードをプレイ
func (ti *TwentyNineInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.sp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ti.Game.GetPhase() == domain.TwentyNinePhaseTrickEnd {
		ti.Game.ResolveTrick()
	}
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *TwentyNineInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *TwentyNineInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TwentyNineInteractor) GetConfig() domain.TwentyNineConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TwentyNineInteractor) Hint() string {
	return ti.sp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TwentyNineInteractor) ActionLog() string {
	return ti.sp.ActionLogOutput(ti.Game)
}

// advanceCpu 入札 → プレイ の順に CPU を自動進行させる。
// 入札フェーズでは人間の手番になるまで CPU に入札させ、入札が締まって
// プレイフェーズに入ったら人間の手番までトリックを進める。
func (ti *TwentyNineInteractor) advanceCpu() {
	runCpuBidsLoop(ti.Game, domain.TwentyNinePhaseBid)
	ti.runCpuTurns()
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する。
func (ti *TwentyNineInteractor) runCpuTurns() {
	runCpuTurnsLoop(ti.Game, trickPhases[domain.TwentyNinePhase]{
		play:     domain.TwentyNinePhasePlay,
		trickEnd: domain.TwentyNinePhaseTrickEnd,
		roundEnd: domain.TwentyNinePhaseRoundEnd,
		gameEnd:  domain.TwentyNinePhaseGameEnd,
	})
}

// RestoreTwentyNineInteractor deserialises JSON into a TwentyNineInteractor.
func RestoreTwentyNineInteractor(data []byte, sp presenter.TwentyNinePresenter) (*TwentyNineInteractor, error) {
	return restoreAndBuild[domain.TwentyNine](data, func(g *domain.TwentyNine) *TwentyNineInteractor {
		return &TwentyNineInteractor{GameBase: GameBase[interfaces.TwentyNineGame]{Game: g}, sp: sp}
	})
}
