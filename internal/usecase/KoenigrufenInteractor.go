//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KoenigrufenInteractorIF ケーニッヒルーフェン (Königrufen) のインタラクターインタフェース
type KoenigrufenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KoenigrufenConfig) string
	// Bid 入札する
	Bid(bid domain.KoenigrufenBid) string
	// Pass パスする
	Pass() string
	// CallKing 呼ぶキングのスートを指名する
	CallKing(suit int) string
	// Discard 場札交換で 6 枚を捨てる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KoenigrufenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KoenigrufenInteractor ケーニッヒルーフェンのインタラクタークラス
type KoenigrufenInteractor struct {
	GameBase[interfaces.KoenigrufenGame]
	tp presenter.KoenigrufenPresenter
}

// NewKoenigrufenInteractor コンストラクタ
func NewKoenigrufenInteractor(g interfaces.KoenigrufenGame, tp presenter.KoenigrufenPresenter) *KoenigrufenInteractor {
	mustNotNil("KoenigrufenInteractor", map[string]any{"g": g, "tp": tp})
	return &KoenigrufenInteractor{GameBase: GameBase[interfaces.KoenigrufenGame]{Game: g}, tp: tp}
}

// koenigrufenTrickPhases Königrufen のトリックフェーズ定数
func koenigrufenTrickPhases() trickPhases[domain.KoenigrufenPhase] {
	return trickPhases[domain.KoenigrufenPhase]{
		play:     domain.KoenigrufenPhasePlay,
		trickEnd: domain.KoenigrufenPhaseTrickEnd,
		roundEnd: domain.KoenigrufenPhaseRoundEnd,
		gameEnd:  domain.KoenigrufenPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *KoenigrufenInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *KoenigrufenInteractor) ResetWithConfig(cfg domain.KoenigrufenConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid 入札する
func (ci *KoenigrufenInteractor) Bid(bid domain.KoenigrufenBid) string {
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
func (ci *KoenigrufenInteractor) Pass() string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPass(); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// CallKing 呼ぶキングのスートを指名する
func (ci *KoenigrufenInteractor) CallKing(suit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerCallKing(suit); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Discard 場札交換で 6 枚を捨てる
func (ci *KoenigrufenInteractor) Discard(cardIndices []int) string {
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
func (ci *KoenigrufenInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.KoenigrufenPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *KoenigrufenInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *KoenigrufenInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *KoenigrufenInteractor) GetConfig() domain.KoenigrufenConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *KoenigrufenInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *KoenigrufenInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU の入札・王呼び・場札交換・プレイを、人間の手番もしくはトリック/ラウンド終了に
// なるまで自動実行する。
func (ci *KoenigrufenInteractor) advance() {
	runCpuBidsLoop[domain.KoenigrufenPhase](ci.Game, domain.KoenigrufenPhaseBid)
	// CPU デクレアラーの王呼びを自動実行する。
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.KoenigrufenPhaseCall || ci.Game.IsHumanCallTurn()
	}, ci.Game.CpuCallKing)
	// CPU デクレアラーの場札交換を自動実行する。
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.KoenigrufenPhaseTalon || ci.Game.IsHumanDiscardTurn()
	}, ci.Game.CpuDiscard)
	runCpuTurnsLoop(ci.Game, koenigrufenTrickPhases())
}

// RestoreKoenigrufenInteractor deserialises JSON into a KoenigrufenInteractor.
func RestoreKoenigrufenInteractor(data []byte, tp presenter.KoenigrufenPresenter) (*KoenigrufenInteractor, error) {
	return restoreAndBuild[domain.Koenigrufen](data, func(g *domain.Koenigrufen) *KoenigrufenInteractor {
		return &KoenigrufenInteractor{GameBase: GameBase[interfaces.KoenigrufenGame]{Game: g}, tp: tp}
	})
}
