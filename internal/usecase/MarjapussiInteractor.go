//go:build !js || !wasm || extra5

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MarjapussiInteractorIF マルヤプッシ (Marjapussi) のインタラクターインタフェース
type MarjapussiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MarjapussiConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MarjapussiConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MarjapussiInteractor マルヤプッシのインタラクタークラス
type MarjapussiInteractor struct {
	GameBase[interfaces.MarjapussiGame]
	tp presenter.MarjapussiPresenter
}

// NewMarjapussiInteractor コンストラクタ
func NewMarjapussiInteractor(g interfaces.MarjapussiGame, tp presenter.MarjapussiPresenter) *MarjapussiInteractor {
	mustNotNil("MarjapussiInteractor", map[string]any{"g": g, "tp": tp})
	return &MarjapussiInteractor{GameBase: GameBase[interfaces.MarjapussiGame]{Game: g}, tp: tp}
}

// marjapussiTrickPhases マルヤプッシのトリックフェーズ定数
func marjapussiTrickPhases() trickPhases[domain.MarjapussiPhase] {
	return trickPhases[domain.MarjapussiPhase]{
		play:     domain.MarjapussiPhasePlay,
		trickEnd: domain.MarjapussiPhaseTrickEnd,
		roundEnd: domain.MarjapussiPhaseRoundEnd,
		gameEnd:  domain.MarjapussiPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ti *MarjapussiInteractor) Reset() string {
	ti.Game.Reset()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *MarjapussiInteractor) ResetWithConfig(cfg domain.MarjapussiConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Play カードをプレイ
func (ti *MarjapussiInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ti.Game.GetPhase() == domain.MarjapussiPhaseTrickEnd {
		ti.Game.ResolveTrick()
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *MarjapussiInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *MarjapussiInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *MarjapussiInteractor) GetConfig() domain.MarjapussiConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *MarjapussiInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *MarjapussiInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// advance CPU のプレイを人間の手番もしくはトリック/ラウンド終了になるまで自動実行する。
func (ti *MarjapussiInteractor) advance() {
	runCpuTurnsLoop(ti.Game, marjapussiTrickPhases())
}

// RestoreMarjapussiInteractor deserialises JSON into a MarjapussiInteractor.
func RestoreMarjapussiInteractor(data []byte, tp presenter.MarjapussiPresenter) (*MarjapussiInteractor, error) {
	return restoreAndBuild[domain.Marjapussi](data, func(g *domain.Marjapussi) *MarjapussiInteractor {
		return &MarjapussiInteractor{GameBase: GameBase[interfaces.MarjapussiGame]{Game: g}, tp: tp}
	})
}
