package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpadesInteractorIF スペードインタラクターインタフェース
type SpadesInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SpadesConfig) string
	// Bid ビッドを宣言
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SpadesConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SpadesInteractor スペードインタラクタークラス
type SpadesInteractor struct {
	GameBase[interfaces.SpadesGame]
	sp presenter.SpadesPresenter
}

// NewSpadesInteractor コンストラクタ
func NewSpadesInteractor(s interfaces.SpadesGame, sp presenter.SpadesPresenter) *SpadesInteractor {
	mustNotNil("SpadesInteractor", map[string]any{"s": s, "sp": sp})
	return &SpadesInteractor{GameBase: GameBase[interfaces.SpadesGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化
func (si *SpadesInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuBids()
	if si.Game.GetPhase() == domain.SpadesPhasePlay {
		si.runCpuTurns()
	}
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SpadesInteractor) ResetWithConfig(cfg domain.SpadesConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Bid ビッドを宣言
func (si *SpadesInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.PlayerBid(bid)
	if err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuBids()
	if si.Game.GetPhase() == domain.SpadesPhasePlay {
		si.runCpuTurns()
	}
	return si.sp.Output(si.Game, nil)
}

// Play カードをプレイ
func (si *SpadesInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.PlayerPlay(cardIndex)
	if err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextTrick 次のトリックへ進む
func (si *SpadesInteractor) NextTrick() string {
	si.Game.NextTrick()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (si *SpadesInteractor) NextRound() string {
	si.Game.ScoreRound()
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.runCpuBids()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SpadesInteractor) GetConfig() domain.SpadesConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *SpadesInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SpadesInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// runCpuBids ゲームが終わるかヒューマンのビッド番またはビッドフェーズが終了するまでCPUビッドを実行
func (si *SpadesInteractor) runCpuBids() {
	runCpuBidsLoop(si.Game, domain.SpadesPhaseBid)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (si *SpadesInteractor) runCpuTurns() {
	runCpuTurnsLoop(si.Game, trickPhases[domain.SpadesPhase]{
		play:     domain.SpadesPhasePlay,
		trickEnd: domain.SpadesPhaseTrickEnd,
		roundEnd: domain.SpadesPhaseRoundEnd,
		gameEnd:  domain.SpadesPhaseGameEnd,
	})
}

// RestoreSpadesInteractor deserialises JSON into a SpadesInteractor.
func RestoreSpadesInteractor(data []byte, sp presenter.SpadesPresenter) (*SpadesInteractor, error) {
	return restoreAndBuild[domain.Spades](data, func(g *domain.Spades) *SpadesInteractor {
		return &SpadesInteractor{GameBase: GameBase[interfaces.SpadesGame]{Game: g}, sp: sp}
	})
}
