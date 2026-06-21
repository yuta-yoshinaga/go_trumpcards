//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NapInteractorIF ナップのインタラクターインタフェース
type NapInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.NapConfig) string
	// Bid 入札する (0=Pass,2=Two,3=Three,4=Four,5=Nap)
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.NapConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// NapInteractor ナップのインタラクタークラス
type NapInteractor struct {
	GameBase[interfaces.NapGame]
	sp presenter.NapPresenter
}

// NewNapInteractor コンストラクタ
func NewNapInteractor(g interfaces.NapGame, sp presenter.NapPresenter) *NapInteractor {
	mustNotNil("NapInteractor", map[string]any{"g": g, "sp": sp})
	return &NapInteractor{GameBase: GameBase[interfaces.NapGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (ni *NapInteractor) Reset() string {
	ni.Game.Reset()
	ni.advanceCpu()
	return ni.sp.Output(ni.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ni *NapInteractor) ResetWithConfig(cfg domain.NapConfig) string {
	return resetWithValidatedConfig(ni.Game, ni.sp, cfg, ni.Game.SetConfig, ni.Reset)
}

// Bid 入札する (0=Pass,2=Two,3=Three,4=Four,5=Nap)
func (ni *NapInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ni.Game, ni.sp); blocked {
		return out
	}
	if err := ni.Game.PlayerBid(domain.NapBid(bid)); err != nil {
		return ni.sp.Output(ni.Game, err)
	}
	ni.advanceCpu()
	return ni.sp.Output(ni.Game, nil)
}

// Play カードをプレイ
func (ni *NapInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ni.Game, ni.sp); blocked {
		return out
	}
	if err := ni.Game.PlayerPlay(cardIndex); err != nil {
		return ni.sp.Output(ni.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ni.Game.GetPhase() == domain.NapPhaseTrickEnd {
		ni.Game.ResolveTrick()
	}
	ni.advanceCpu()
	return ni.sp.Output(ni.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ni *NapInteractor) NextTrick() string {
	ni.Game.NextTrick()
	ni.advanceCpu()
	return ni.sp.Output(ni.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ni *NapInteractor) NextRound() string {
	ni.Game.ScoreRound()
	if out, blocked := guardGameEnd(ni.Game, ni.sp); blocked {
		return out
	}
	ni.Game.NextRound()
	ni.advanceCpu()
	return ni.sp.Output(ni.Game, nil)
}

// GetConfig 現在の設定を取得
func (ni *NapInteractor) GetConfig() domain.NapConfig {
	return ni.Game.GetConfig()
}

// Hint ヒント取得
func (ni *NapInteractor) Hint() string {
	return ni.sp.HintOutput(ni.Game)
}

// ActionLog 棋譜を出力する
func (ni *NapInteractor) ActionLog() string {
	return ni.sp.ActionLogOutput(ni.Game)
}

// advanceCpu 入札 → プレイ の順に CPU を自動進行させる。
// 入札フェーズでは人間の手番になるまで CPU に入札させ、入札が締まって
// プレイフェーズに入ったら人間の手番までトリックを進める。
func (ni *NapInteractor) advanceCpu() {
	runCpuBidsLoop(ni.Game, domain.NapPhaseBid)
	ni.runCpuTurns()
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する。
func (ni *NapInteractor) runCpuTurns() {
	runCpuTurnsLoop(ni.Game, trickPhases[domain.NapPhase]{
		play:     domain.NapPhasePlay,
		trickEnd: domain.NapPhaseTrickEnd,
		roundEnd: domain.NapPhaseRoundEnd,
		gameEnd:  domain.NapPhaseGameEnd,
	})
}

// RestoreNapInteractor deserialises JSON into a NapInteractor.
func RestoreNapInteractor(data []byte, sp presenter.NapPresenter) (*NapInteractor, error) {
	return restoreAndBuild[domain.Nap](data, func(g *domain.Nap) *NapInteractor {
		return &NapInteractor{GameBase: GameBase[interfaces.NapGame]{Game: g}, sp: sp}
	})
}
