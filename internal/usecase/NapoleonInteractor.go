package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NapoleonInteractorIF ナポレオンインタラクターインタフェース
type NapoleonInteractorIF interface {
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
	n  interfaces.NapoleonGame
	np presenter.NapoleonPresenter
}

// NewNapoleonInteractor コンストラクタ
func NewNapoleonInteractor(n interfaces.NapoleonGame, np presenter.NapoleonPresenter) *NapoleonInteractor {
	mustNotNil("NapoleonInteractor", map[string]any{"n": n, "np": np})
	return &NapoleonInteractor{n: n, np: np}
}

// Reset ゲーム初期化
func (ni *NapoleonInteractor) Reset() string {
	ni.n.Reset()
	ni.runCpuBids()
	ni.runCpuDeclareAndExchange()
	if ni.n.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.n, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ni *NapoleonInteractor) ResetWithConfig(cfg domain.NapoleonConfig) string {
	if err := cfg.Validate(); err != nil {
		return ni.np.Output(ni.n, err)
	}
	ni.n.SetConfig(cfg)
	return ni.Reset()
}

// Bid ビッドを宣言
func (ni *NapoleonInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ni.n, ni.np); blocked {
		return out
	}
	err := ni.n.PlayerBid(bid)
	if err != nil {
		return ni.np.Output(ni.n, err)
	}
	ni.runCpuBids()
	ni.runCpuDeclareAndExchange()
	if ni.n.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.n, nil)
}

// DeclareTrump 切り札と副官を宣言
func (ni *NapoleonInteractor) DeclareTrump(suit int, adjSuit int, adjVal int) string {
	if out, blocked := guardGameEnd(ni.n, ni.np); blocked {
		return out
	}
	err := ni.n.PlayerDeclareTrump(suit, adjSuit, adjVal)
	if err != nil {
		return ni.np.Output(ni.n, err)
	}
	// 場札交換フェーズに移行。CPU napoelonの場合は自動処理
	ni.runCpuDeclareAndExchange()
	if ni.n.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.n, nil)
}

// ExchangeKitty 場札を交換
func (ni *NapoleonInteractor) ExchangeKitty(discardIndex int) string {
	if out, blocked := guardGameEnd(ni.n, ni.np); blocked {
		return out
	}
	err := ni.n.PlayerExchangeKitty(discardIndex)
	if err != nil {
		return ni.np.Output(ni.n, err)
	}
	if ni.n.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.n, nil)
}

// Play カードをプレイ
func (ni *NapoleonInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ni.n, ni.np); blocked {
		return out
	}
	err := ni.n.PlayerPlay(cardIndex)
	if err != nil {
		return ni.np.Output(ni.n, err)
	}
	ni.runCpuTurns()
	return ni.np.Output(ni.n, nil)
}

// NextTrick 次のトリックへ進む
func (ni *NapoleonInteractor) NextTrick() string {
	ni.n.NextTrick()
	ni.runCpuTurns()
	return ni.np.Output(ni.n, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ni *NapoleonInteractor) NextRound() string {
	ni.n.ScoreRound()
	if out, blocked := guardGameEnd(ni.n, ni.np); blocked {
		return out
	}
	ni.n.NextRound()
	ni.runCpuBids()
	ni.runCpuDeclareAndExchange()
	if ni.n.GetPhase() == domain.NapoleonPhasePlay {
		ni.runCpuTurns()
	}
	return ni.np.Output(ni.n, nil)
}

// GetConfig 現在の設定を取得
func (ni *NapoleonInteractor) GetConfig() domain.NapoleonConfig {
	return ni.n.GetConfig()
}

// Hint ヒント取得
func (ni *NapoleonInteractor) Hint() string {
	return ni.np.HintOutput(ni.n)
}

// ActionLog 棋譜を出力する
func (ni *NapoleonInteractor) ActionLog() string {
	return ni.np.ActionLogOutput(ni.n)
}

// runCpuBids CPUビッドを実行
func (ni *NapoleonInteractor) runCpuBids() {
	for !ni.n.GetGameEndFlag() {
		if ni.n.GetPhase() != domain.NapoleonPhaseBid {
			break
		}
		if ni.n.IsHumanBidTurn() {
			break
		}
		ni.n.CpuBid()
	}
}

// runCpuDeclareAndExchange CPUの切り札宣言と場札交換を自動処理
func (ni *NapoleonInteractor) runCpuDeclareAndExchange() {
	if ni.n.GetPhase() == domain.NapoleonPhaseTrumpDeclaration && !ni.n.IsHumanDeclareTurn() {
		ni.n.CpuDeclareTrump()
	}
	if ni.n.GetPhase() == domain.NapoleonPhaseKittyExchange && !ni.n.IsHumanExchangeTurn() {
		ni.n.CpuExchangeKitty()
	}
}

// runCpuTurns CPUターンを実行
func (ni *NapoleonInteractor) runCpuTurns() {
	for !ni.n.GetGameEndFlag() {
		phase := ni.n.GetPhase()
		if phase == domain.NapoleonPhaseTrickEnd || phase == domain.NapoleonPhaseRoundEnd || phase == domain.NapoleonPhaseGameEnd {
			break
		}
		if phase != domain.NapoleonPhasePlay {
			break
		}
		if ni.n.IsHumanTurn() {
			break
		}
		ni.n.CpuPlay()
		if ni.n.GetPhase() == domain.NapoleonPhaseTrickEnd {
			ni.n.ResolveTrick()
			if ni.n.GetPhase() == domain.NapoleonPhaseRoundEnd {
				break
			}
			ni.n.NextTrick()
		}
	}
}
