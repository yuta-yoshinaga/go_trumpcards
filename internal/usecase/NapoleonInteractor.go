//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NapoleonInteractorIF ナポレオンインタラクターインタフェース
type NapoleonInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.NapoleonConfig) string
	// Bid ビッドを宣言
	Bid(bid int) string
	// DeclareTrump 切り札と副官を宣言
	DeclareTrump(suit int, adjSuit int, adjVal int) string
	// ExchangeKitty 場札を交換
	ExchangeKitty(discardIndex int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.NapoleonConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// NapoleonInteractor ナポレオンインタラクタークラス
type NapoleonInteractor struct {
	GameBase[interfaces.NapoleonGame]
	np presenter.NapoleonPresenter
}

// NewNapoleonInteractor コンストラクタ
func NewNapoleonInteractor(n interfaces.NapoleonGame, np presenter.NapoleonPresenter) *NapoleonInteractor {
	mustNotNil("NapoleonInteractor", map[string]any{"n": n, "np": np})
	return &NapoleonInteractor{GameBase: GameBase[interfaces.NapoleonGame]{Game: n}, np: np}
}

// Reset ゲーム初期化
func (ni *NapoleonInteractor) Reset() string {
	ni.Game.Reset()
	ni.runCpuBids()
	ni.runCpuDeclareAndExchange()
	if ni.Game.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ni *NapoleonInteractor) ResetWithConfig(cfg domain.NapoleonConfig) string {
	return resetWithValidatedConfig(ni.Game, ni.np, cfg, ni.Game.SetConfig, ni.Reset)
}

// Bid ビッドを宣言
func (ni *NapoleonInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ni.Game, ni.np); blocked {
		return out
	}
	err := ni.Game.PlayerBid(bid)
	if err != nil {
		return ni.np.Output(ni.Game, err)
	}
	ni.runCpuBids()
	ni.runCpuDeclareAndExchange()
	if ni.Game.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.Game, nil)
}

// DeclareTrump 切り札と副官を宣言
func (ni *NapoleonInteractor) DeclareTrump(suit int, adjSuit int, adjVal int) string {
	if out, blocked := guardGameEnd(ni.Game, ni.np); blocked {
		return out
	}
	err := ni.Game.PlayerDeclareTrump(suit, adjSuit, adjVal)
	if err != nil {
		return ni.np.Output(ni.Game, err)
	}
	// 場札交換フェーズに移行。CPU napoelonの場合は自動処理
	ni.runCpuDeclareAndExchange()
	if ni.Game.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.Game, nil)
}

// ExchangeKitty 場札を交換
func (ni *NapoleonInteractor) ExchangeKitty(discardIndex int) string {
	if out, blocked := guardGameEnd(ni.Game, ni.np); blocked {
		return out
	}
	err := ni.Game.PlayerExchangeKitty(discardIndex)
	if err != nil {
		return ni.np.Output(ni.Game, err)
	}
	if ni.Game.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.Game, nil)
}

// Play カードをプレイ
func (ni *NapoleonInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ni.Game, ni.np); blocked {
		return out
	}
	err := ni.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ni.np.Output(ni.Game, err)
	}
	ni.runCpuTurns()
	return ni.np.Output(ni.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ni *NapoleonInteractor) NextTrick() string {
	if out, blocked := guardGameEnd(ni.Game, ni.np); blocked {
		return out
	}
	ni.Game.NextTrick()
	ni.runCpuTurns()
	return ni.np.Output(ni.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ni *NapoleonInteractor) NextRound() string {
	ni.Game.ScoreRound()
	if out, blocked := guardGameEnd(ni.Game, ni.np); blocked {
		return out
	}
	ni.Game.NextRound()
	ni.runCpuBids()
	ni.runCpuDeclareAndExchange()
	if ni.Game.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.Game, nil)
}

// GetConfig 現在の設定を取得
func (ni *NapoleonInteractor) GetConfig() domain.NapoleonConfig {
	return ni.Game.GetConfig()
}

// Hint ヒント取得
func (ni *NapoleonInteractor) Hint() string {
	return ni.np.HintOutput(ni.Game)
}

// ActionLog 棋譜を出力する
func (ni *NapoleonInteractor) ActionLog() string {
	return ni.np.ActionLogOutput(ni.Game)
}

// runCpuBids CPUビッドを実行
func (ni *NapoleonInteractor) runCpuBids() {
	runCpuBidsLoop(ni.Game, domain.NapoleonPhaseBid)
}

// runCpuDeclareAndExchange CPUの切り札宣言と場札交換を自動処理
func (ni *NapoleonInteractor) runCpuDeclareAndExchange() {
	if ni.Game.GetPhase() == domain.NapoleonPhaseTrumpDeclaration && !ni.Game.IsHumanDeclareTurn() {
		ni.Game.CpuDeclareTrump()
	}
	if ni.Game.GetPhase() == domain.NapoleonPhaseKittyExchange && !ni.Game.IsHumanExchangeTurn() {
		ni.Game.CpuExchangeKitty()
	}
}

// runCpuTurns CPUターンを実行
func (ni *NapoleonInteractor) runCpuTurns() {
	runCpuTurnsLoop(ni.Game, trickPhases[domain.NapoleonPhase]{
		play:     domain.NapoleonPhasePlay,
		trickEnd: domain.NapoleonPhaseTrickEnd,
		roundEnd: domain.NapoleonPhaseRoundEnd,
		gameEnd:  domain.NapoleonPhaseGameEnd,
	})
}

// RestoreNapoleonInteractor deserialises JSON into a NapoleonInteractor.
func RestoreNapoleonInteractor(data []byte, np presenter.NapoleonPresenter) (*NapoleonInteractor, error) {
	return restoreAndBuild[domain.Napoleon](data, func(g *domain.Napoleon) *NapoleonInteractor {
		return &NapoleonInteractor{GameBase: GameBase[interfaces.NapoleonGame]{Game: g}, np: np}
	})
}
