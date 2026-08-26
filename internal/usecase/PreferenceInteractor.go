//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PreferenceInteractorIF プレフェランスのインタラクターインタフェース
type PreferenceInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PreferenceConfig) string
	// Bid 入札する (0=Pass,1=Six,2=Misère,3=Seven,4=Eight)
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PreferenceConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PreferenceInteractor プレフェランスのインタラクタークラス
type PreferenceInteractor struct {
	GameBase[interfaces.PreferenceGame]
	sp presenter.PreferencePresenter
}

// NewPreferenceInteractor コンストラクタ
func NewPreferenceInteractor(g interfaces.PreferenceGame, sp presenter.PreferencePresenter) *PreferenceInteractor {
	mustNotNil("PreferenceInteractor", map[string]any{"g": g, "sp": sp})
	return &PreferenceInteractor{GameBase: GameBase[interfaces.PreferenceGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (pi *PreferenceInteractor) Reset() string {
	pi.Game.Reset()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PreferenceInteractor) ResetWithConfig(cfg domain.PreferenceConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.sp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Bid 入札する (0=Pass,1=Six,2=Misère,3=Seven,4=Eight)
func (pi *PreferenceInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.PlayerBid(domain.PreferenceBid(bid)); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// Play カードをプレイ
func (pi *PreferenceInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.PlayerPlay(cardIndex); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if pi.Game.GetPhase() == domain.PreferencePhaseTrickEnd {
		pi.Game.ResolveTrick()
	}
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (pi *PreferenceInteractor) NextTrick() string {
	pi.Game.NextTrick()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (pi *PreferenceInteractor) NextRound() string {
	pi.Game.ScoreRound()
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	pi.Game.NextRound()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *PreferenceInteractor) GetConfig() domain.PreferenceConfig {
	return pi.Game.GetConfig()
}

// Hint ヒント取得
func (pi *PreferenceInteractor) Hint() string {
	return pi.sp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *PreferenceInteractor) ActionLog() string {
	return pi.sp.ActionLogOutput(pi.Game)
}

// advanceCpu 入札 → プレイ の順に CPU を自動進行させる。
// 入札フェーズでは人間の手番になるまで CPU に入札させ、入札が締まって
// プレイフェーズに入ったら人間の手番までトリックを進める。
func (pi *PreferenceInteractor) advanceCpu() {
	runCpuBidsLoop(pi.Game, domain.PreferencePhaseBid)
	pi.runCpuTurns()
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する。
func (pi *PreferenceInteractor) runCpuTurns() {
	runCpuTurnsLoop(pi.Game, trickPhases[domain.PreferencePhase]{
		play:     domain.PreferencePhasePlay,
		trickEnd: domain.PreferencePhaseTrickEnd,
		roundEnd: domain.PreferencePhaseRoundEnd,
		gameEnd:  domain.PreferencePhaseGameEnd,
	})
}

// RestorePreferenceInteractor deserialises JSON into a PreferenceInteractor.
func RestorePreferenceInteractor(data []byte, sp presenter.PreferencePresenter) (*PreferenceInteractor, error) {
	return restoreAndBuild[domain.Preference](data, func(g *domain.Preference) *PreferenceInteractor {
		return &PreferenceInteractor{GameBase: GameBase[interfaces.PreferenceGame]{Game: g}, sp: sp}
	})
}
