package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TarneebInteractorIF Tarneeb インタラクターインタフェース
type TarneebInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TarneebConfig) string
	// Bid ビッドを宣言 (0 = パス、7-13 = ビッド)
	Bid(bid int) string
	// DeclareTrump トランプスートを宣言
	DeclareTrump(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TarneebConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TarneebInteractor Tarneeb インタラクタークラス
type TarneebInteractor struct {
	GameBase[interfaces.TarneebGame]
	tp presenter.TarneebPresenter
}

// NewTarneebInteractor コンストラクタ
func NewTarneebInteractor(t interfaces.TarneebGame, tp presenter.TarneebPresenter) *TarneebInteractor {
	mustNotNil("TarneebInteractor", map[string]any{"t": t, "tp": tp})
	return &TarneebInteractor{GameBase: GameBase[interfaces.TarneebGame]{Game: t}, tp: tp}
}

// Reset ゲーム初期化
func (ti *TarneebInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuBids()
	ti.runCpuTrumpDeclaration()
	if ti.Game.GetPhase() == domain.TarneebPhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TarneebInteractor) ResetWithConfig(cfg domain.TarneebConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Bid ビッドを宣言
func (ti *TarneebInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerBid(bid); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuBids()
	ti.runCpuTrumpDeclaration()
	if ti.Game.GetPhase() == domain.TarneebPhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// DeclareTrump トランプスートを宣言
func (ti *TarneebInteractor) DeclareTrump(suit int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerDeclareTrump(suit); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	if ti.Game.GetPhase() == domain.TarneebPhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// Play カードをプレイ
func (ti *TarneebInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *TarneebInteractor) NextTrick() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextTrick()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *TarneebInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.runCpuBids()
	ti.runCpuTrumpDeclaration()
	if ti.Game.GetPhase() == domain.TarneebPhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TarneebInteractor) GetConfig() domain.TarneebConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TarneebInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TarneebInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuBids ビッドフェーズが続く限り CPU をループ実行する。
// 再配布で再度ビッドフェーズに戻った場合も適切に進められるよう、
// 最大回数を制限してデッドロックを防ぐ。
func (ti *TarneebInteractor) runCpuBids() {
	const maxIterations = 4 * 10 // 4プレイヤー × 最大10回再配布
	for i := 0; i < maxIterations; i++ {
		if ti.Game.GetPhase() != domain.TarneebPhaseBid {
			return
		}
		if ti.Game.IsHumanBidTurn() {
			return
		}
		ti.Game.CpuBid()
	}
}

// runCpuTrumpDeclaration CPUの場合はトランプ宣言を自動処理
func (ti *TarneebInteractor) runCpuTrumpDeclaration() {
	if ti.Game.GetPhase() == domain.TarneebPhaseTrumpDeclaration && !ti.Game.IsHumanTrumpTurn() {
		ti.Game.CpuDeclareTrump()
	}
}

// runCpuTurns ヒューマンの手番またはトリック / ラウンド終了まで CPU ターンを実行
func (ti *TarneebInteractor) runCpuTurns() {
	runCpuTurnsLoop(ti.Game, trickPhases[domain.TarneebPhase]{
		play:     domain.TarneebPhasePlay,
		trickEnd: domain.TarneebPhaseTrickEnd,
		roundEnd: domain.TarneebPhaseRoundEnd,
		gameEnd:  domain.TarneebPhaseGameEnd,
	})
}

// RestoreTarneebInteractor deserialises JSON into a TarneebInteractor.
func RestoreTarneebInteractor(data []byte, tp presenter.TarneebPresenter) (*TarneebInteractor, error) {
	return restoreAndBuild[domain.Tarneeb](data, func(g *domain.Tarneeb) *TarneebInteractor {
		return &TarneebInteractor{GameBase: GameBase[interfaces.TarneebGame]{Game: g}, tp: tp}
	})
}
