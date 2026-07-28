//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MightyInteractorIF マイティインタラクターインタフェース
type MightyInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MightyConfig) string
	// Bid ビッドを宣言 (0 = パス、noTrump = ノートランプ宣言)
	Bid(bid int, noTrump bool) string
	// DeclareTrumpAndFriend 切り札とパートナーを宣言
	DeclareTrumpAndFriend(suit int, partnerSuit int, partnerVal int) string
	// ExchangeKitty 場札を交換 (捨て札を3枚指定)
	ExchangeKitty(discardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// PlayJokerLead ジョーカーをリードする (要求スート指定)
	PlayJokerLead(cardIndex int, demandSuit int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MightyConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MightyInteractor マイティインタラクタークラス
type MightyInteractor struct {
	GameBase[interfaces.MightyGame]
	mp presenter.MightyPresenter
}

// NewMightyInteractor コンストラクタ
func NewMightyInteractor(m interfaces.MightyGame, mp presenter.MightyPresenter) *MightyInteractor {
	mustNotNil("MightyInteractor", map[string]any{"m": m, "mp": mp})
	return &MightyInteractor{GameBase: GameBase[interfaces.MightyGame]{Game: m}, mp: mp}
}

// Reset ゲーム初期化
func (mi *MightyInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuBids()
	mi.runCpuDeclareAndExchange()
	if mi.Game.GetPhase() == domain.MightyPhasePlay {
		mi.runCpuTurns()
	}
	return mi.mp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *MightyInteractor) ResetWithConfig(cfg domain.MightyConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.mp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Bid ビッドを宣言
func (mi *MightyInteractor) Bid(bid int, noTrump bool) string {
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	if err := mi.Game.PlayerBid(bid, noTrump); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	mi.runCpuBids()
	mi.runCpuDeclareAndExchange()
	if mi.Game.GetPhase() == domain.MightyPhasePlay {
		mi.runCpuTurns()
	}
	return mi.mp.Output(mi.Game, nil)
}

// DeclareTrumpAndFriend 切り札とパートナーを宣言
func (mi *MightyInteractor) DeclareTrumpAndFriend(suit int, partnerSuit int, partnerVal int) string {
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	if err := mi.Game.PlayerDeclareTrumpAndFriend(suit, partnerSuit, partnerVal); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	mi.runCpuDeclareAndExchange()
	if mi.Game.GetPhase() == domain.MightyPhasePlay {
		mi.runCpuTurns()
	}
	return mi.mp.Output(mi.Game, nil)
}

// ExchangeKitty 場札を交換
func (mi *MightyInteractor) ExchangeKitty(discardIndices []int) string {
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	if err := mi.Game.PlayerExchangeKitty(discardIndices); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	if mi.Game.GetPhase() == domain.MightyPhasePlay {
		mi.runCpuTurns()
	}
	return mi.mp.Output(mi.Game, nil)
}

// Play カードをプレイ
func (mi *MightyInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	if err := mi.Game.PlayerPlay(cardIndex); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// PlayJokerLead ジョーカーをリードする
func (mi *MightyInteractor) PlayJokerLead(cardIndex int, demandSuit int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	if err := mi.Game.PlayerPlayJokerLead(cardIndex, demandSuit); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (mi *MightyInteractor) NextTrick() string {
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.NextTrick()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (mi *MightyInteractor) NextRound() string {
	mi.Game.ScoreRound()
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.NextRound()
	mi.runCpuBids()
	mi.runCpuDeclareAndExchange()
	if mi.Game.GetPhase() == domain.MightyPhasePlay {
		mi.runCpuTurns()
	}
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *MightyInteractor) GetConfig() domain.MightyConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *MightyInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *MightyInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// runCpuBids CPUビッドを実行
func (mi *MightyInteractor) runCpuBids() {
	runCpuBidsLoop(mi.Game, domain.MightyPhaseBid)
}

// runCpuDeclareAndExchange CPUの切り札宣言と場札交換を自動処理
func (mi *MightyInteractor) runCpuDeclareAndExchange() {
	if mi.Game.GetPhase() == domain.MightyPhaseTrumpAndFriend && !mi.Game.IsHumanDeclareTurn() {
		mi.Game.CpuDeclareTrumpAndFriend()
	}
	if mi.Game.GetPhase() == domain.MightyPhaseKittyExchange && !mi.Game.IsHumanExchangeTurn() {
		mi.Game.CpuExchangeKitty()
	}
}

// runCpuTurns CPUターンを実行
func (mi *MightyInteractor) runCpuTurns() {
	runCpuTurnsLoop(mi.Game, trickPhases[domain.MightyPhase]{
		play:     domain.MightyPhasePlay,
		trickEnd: domain.MightyPhaseTrickEnd,
		roundEnd: domain.MightyPhaseRoundEnd,
		gameEnd:  domain.MightyPhaseGameEnd,
	})
}

// RestoreMightyInteractor deserialises JSON into a MightyInteractor.
func RestoreMightyInteractor(data []byte, mp presenter.MightyPresenter) (*MightyInteractor, error) {
	return restoreAndBuild[domain.Mighty](data, func(g *domain.Mighty) *MightyInteractor {
		return &MightyInteractor{GameBase: GameBase[interfaces.MightyGame]{Game: g}, mp: mp}
	})
}
