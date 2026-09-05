package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BatakInteractorIF Batak インタラクターインタフェース
type BatakInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BatakConfig) string
	// Bid ビッドを宣言
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BatakConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BatakInteractor Batak インタラクタークラス
type BatakInteractor struct {
	GameBase[interfaces.BatakGame]
	sp presenter.BatakPresenter
}

// NewBatakInteractor コンストラクタ
func NewBatakInteractor(cb interfaces.BatakGame, sp presenter.BatakPresenter) *BatakInteractor {
	mustNotNil("BatakInteractor", map[string]any{"cb": cb, "sp": sp})
	return &BatakInteractor{GameBase: GameBase[interfaces.BatakGame]{Game: cb}, sp: sp}
}

// Reset ゲーム初期化
func (ci *BatakInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuBids()
	if ci.Game.GetPhase() == domain.BatakPhasePlay {
		ci.runCpuTurns()
	}
	return ci.sp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *BatakInteractor) ResetWithConfig(cfg domain.BatakConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.sp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid ビッドを宣言
func (ci *BatakInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.sp); blocked {
		return out
	}
	err := ci.Game.PlayerBid(bid)
	if err != nil {
		return ci.sp.Output(ci.Game, err)
	}
	ci.runCpuBids()
	if ci.Game.GetPhase() == domain.BatakPhasePlay {
		ci.runCpuTurns()
	}
	return ci.sp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *BatakInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.sp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ci.sp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決
	if ci.Game.GetPhase() == domain.BatakPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.runCpuTurns()
	return ci.sp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *BatakInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.runCpuTurns()
	return ci.sp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *BatakInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.sp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuBids()
	return ci.sp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *BatakInteractor) GetConfig() domain.BatakConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *BatakInteractor) Hint() string {
	return ci.sp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *BatakInteractor) ActionLog() string {
	return ci.sp.ActionLogOutput(ci.Game)
}

// runCpuBids ヒューマンのビッド番もしくはビッドフェーズ終了まで CPU ビッドを実行
func (ci *BatakInteractor) runCpuBids() {
	runCpuBidsLoop(ci.Game, domain.BatakPhaseBid)
}

// runCpuTurns ヒューマンの手番またはトリック / ラウンド終了まで CPU ターンを実行
func (ci *BatakInteractor) runCpuTurns() {
	runCpuTurnsLoop(ci.Game, trickPhases[domain.BatakPhase]{
		play:     domain.BatakPhasePlay,
		trickEnd: domain.BatakPhaseTrickEnd,
		roundEnd: domain.BatakPhaseRoundEnd,
		gameEnd:  domain.BatakPhaseGameEnd,
	})
}

// RestoreBatakInteractor deserialises JSON into a BatakInteractor.
func RestoreBatakInteractor(data []byte, sp presenter.BatakPresenter) (*BatakInteractor, error) {
	return restoreAndBuild[domain.Batak](data, func(g *domain.Batak) *BatakInteractor {
		return &BatakInteractor{GameBase: GameBase[interfaces.BatakGame]{Game: g}, sp: sp}
	})
}
