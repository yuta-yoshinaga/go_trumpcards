//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TysiacInteractorIF サウザンド (Tysiąc) のインタラクターインタフェース
type TysiacInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TysiacConfig) string
	// Bid ビッドする (raise=true で +10、false でパス)
	Bid(raise bool) string
	// Discard talon 交換で1枚を相手へ渡す
	Discard(cardIndex int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TysiacConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TysiacInteractor サウザンドのインタラクタークラス
type TysiacInteractor struct {
	GameBase[interfaces.TysiacGame]
	tp presenter.TysiacPresenter
}

// NewTysiacInteractor コンストラクタ
func NewTysiacInteractor(g interfaces.TysiacGame, tp presenter.TysiacPresenter) *TysiacInteractor {
	mustNotNil("TysiacInteractor", map[string]any{"g": g, "tp": tp})
	return &TysiacInteractor{GameBase: GameBase[interfaces.TysiacGame]{Game: g}, tp: tp}
}

// tysiacTrickPhases Tysiąc のトリックフェーズ定数
func tysiacTrickPhases() trickPhases[domain.TysiacPhase] {
	return trickPhases[domain.TysiacPhase]{
		play:     domain.TysiacPhasePlay,
		trickEnd: domain.TysiacPhaseTrickEnd,
		roundEnd: domain.TysiacPhaseRoundEnd,
		gameEnd:  domain.TysiacPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ti *TysiacInteractor) Reset() string {
	ti.Game.Reset()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TysiacInteractor) ResetWithConfig(cfg domain.TysiacConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Bid ビッドする
func (ti *TysiacInteractor) Bid(raise bool) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerBid(raise); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// Discard talon 交換で1枚を相手へ渡す
func (ti *TysiacInteractor) Discard(cardIndex int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerDiscard(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// Play カードをプレイ
func (ti *TysiacInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ti.Game.GetPhase() == domain.TysiacPhaseTrickEnd {
		ti.Game.ResolveTrick()
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *TysiacInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *TysiacInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TysiacInteractor) GetConfig() domain.TysiacConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TysiacInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TysiacInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// advance CPU のビッド・プレイを人間の手番もしくはトリック/ラウンド終了になるまで自動実行する。
// まずビッドフェーズの CPU を消化し (CPU declarer の talon 交換は domain 側で自動)、続いて
// プレイフェーズの CPU ターンを共通ループで処理する。
func (ti *TysiacInteractor) advance() {
	runCpuBidsLoop[domain.TysiacPhase](ti.Game, domain.TysiacPhaseBid)
	runCpuTurnsLoop(ti.Game, tysiacTrickPhases())
}

// RestoreTysiacInteractor deserialises JSON into a TysiacInteractor.
func RestoreTysiacInteractor(data []byte, tp presenter.TysiacPresenter) (*TysiacInteractor, error) {
	return restoreAndBuild[domain.Tysiac](data, func(g *domain.Tysiac) *TysiacInteractor {
		return &TysiacInteractor{GameBase: GameBase[interfaces.TysiacGame]{Game: g}, tp: tp}
	})
}
