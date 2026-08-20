//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FrenchTarotInteractorIF フレンチタロット (French Tarot) のインタラクターインタフェース
type FrenchTarotInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.FrenchTarotConfig) string
	// Bid 入札する
	Bid(bid domain.FrenchTarotBid) string
	// Pass パスする
	Pass() string
	// Discard シアン交換で 6 枚を捨てる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.FrenchTarotConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// FrenchTarotInteractor フレンチタロットのインタラクタークラス
type FrenchTarotInteractor struct {
	GameBase[interfaces.FrenchTarotGame]
	tp presenter.FrenchTarotPresenter
}

// NewFrenchTarotInteractor コンストラクタ
func NewFrenchTarotInteractor(g interfaces.FrenchTarotGame, tp presenter.FrenchTarotPresenter) *FrenchTarotInteractor {
	mustNotNil("FrenchTarotInteractor", map[string]any{"g": g, "tp": tp})
	return &FrenchTarotInteractor{GameBase: GameBase[interfaces.FrenchTarotGame]{Game: g}, tp: tp}
}

// frenchTarotTrickPhases French Tarot のトリックフェーズ定数
func frenchTarotTrickPhases() trickPhases[domain.FrenchTarotPhase] {
	return trickPhases[domain.FrenchTarotPhase]{
		play:     domain.FrenchTarotPhasePlay,
		trickEnd: domain.FrenchTarotPhaseTrickEnd,
		roundEnd: domain.FrenchTarotPhaseRoundEnd,
		gameEnd:  domain.FrenchTarotPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *FrenchTarotInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *FrenchTarotInteractor) ResetWithConfig(cfg domain.FrenchTarotConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid 入札する
func (ci *FrenchTarotInteractor) Bid(bid domain.FrenchTarotBid) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(bid); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Pass パスする
func (ci *FrenchTarotInteractor) Pass() string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPass(); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Discard シアン交換で 6 枚を捨てる
func (ci *FrenchTarotInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndices); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *FrenchTarotInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.FrenchTarotPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *FrenchTarotInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *FrenchTarotInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *FrenchTarotInteractor) GetConfig() domain.FrenchTarotConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *FrenchTarotInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *FrenchTarotInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU の入札・シアン交換・プレイを、人間の手番もしくはトリック/ラウンド終了になるまで
// 自動実行する。
func (ci *FrenchTarotInteractor) advance() {
	runCpuBidsLoop[domain.FrenchTarotPhase](ci.Game, domain.FrenchTarotPhaseBid)
	// CPU デクレアラーのシアン交換を自動実行する。
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.FrenchTarotPhaseChien || ci.Game.IsHumanDiscardTurn()
	}, ci.Game.CpuDiscard)
	runCpuTurnsLoop(ci.Game, frenchTarotTrickPhases())
}

// RestoreFrenchTarotInteractor deserialises JSON into a FrenchTarotInteractor.
func RestoreFrenchTarotInteractor(data []byte, tp presenter.FrenchTarotPresenter) (*FrenchTarotInteractor, error) {
	return restoreAndBuild[domain.FrenchTarot](data, func(g *domain.FrenchTarot) *FrenchTarotInteractor {
		return &FrenchTarotInteractor{GameBase: GameBase[interfaces.FrenchTarotGame]{Game: g}, tp: tp}
	})
}
