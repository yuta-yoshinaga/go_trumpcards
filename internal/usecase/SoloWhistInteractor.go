//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SoloWhistInteractorIF ソロ・ホイストのインタラクターインタフェース
type SoloWhistInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SoloWhistConfig) string
	// Bid 入札する (0=Pass,1=Solo,2=Misère,3=Abundance)
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SoloWhistConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SoloWhistInteractor ソロ・ホイストのインタラクタークラス
type SoloWhistInteractor struct {
	GameBase[interfaces.SoloWhistGame]
	sp presenter.SoloWhistPresenter
}

// NewSoloWhistInteractor コンストラクタ
func NewSoloWhistInteractor(g interfaces.SoloWhistGame, sp presenter.SoloWhistPresenter) *SoloWhistInteractor {
	mustNotNil("SoloWhistInteractor", map[string]any{"g": g, "sp": sp})
	return &SoloWhistInteractor{GameBase: GameBase[interfaces.SoloWhistGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (si *SoloWhistInteractor) Reset() string {
	si.Game.Reset()
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SoloWhistInteractor) ResetWithConfig(cfg domain.SoloWhistConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Bid 入札する (0=Pass,1=Solo,2=Misère,3=Abundance)
func (si *SoloWhistInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerBid(domain.SoloWhistBid(bid)); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// Play カードをプレイ
func (si *SoloWhistInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPlay(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if si.Game.GetPhase() == domain.SoloWhistPhaseTrickEnd {
		si.Game.ResolveTrick()
	}
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// NextTrick 次のトリックへ進む
func (si *SoloWhistInteractor) NextTrick() string {
	si.Game.NextTrick()
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (si *SoloWhistInteractor) NextRound() string {
	si.Game.ScoreRound()
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SoloWhistInteractor) GetConfig() domain.SoloWhistConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *SoloWhistInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SoloWhistInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// advanceCpu 入札 → プレイ の順に CPU を自動進行させる。
// 入札フェーズでは人間の手番になるまで CPU に入札させ、入札が締まって
// プレイフェーズに入ったら人間の手番までトリックを進める。
func (si *SoloWhistInteractor) advanceCpu() {
	runCpuBidsLoop(si.Game, domain.SoloWhistPhaseBid)
	si.runCpuTurns()
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する。
func (si *SoloWhistInteractor) runCpuTurns() {
	runCpuTurnsLoop(si.Game, trickPhases[domain.SoloWhistPhase]{
		play:     domain.SoloWhistPhasePlay,
		trickEnd: domain.SoloWhistPhaseTrickEnd,
		roundEnd: domain.SoloWhistPhaseRoundEnd,
		gameEnd:  domain.SoloWhistPhaseGameEnd,
	})
}

// RestoreSoloWhistInteractor deserialises JSON into a SoloWhistInteractor.
func RestoreSoloWhistInteractor(data []byte, sp presenter.SoloWhistPresenter) (*SoloWhistInteractor, error) {
	return restoreAndBuild[domain.SoloWhist](data, func(g *domain.SoloWhist) *SoloWhistInteractor {
		return &SoloWhistInteractor{GameBase: GameBase[interfaces.SoloWhistGame]{Game: g}, sp: sp}
	})
}
