//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MinchiateInteractorIF ミンキアーテのインタラクターインタフェース
type MinchiateInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MinchiateConfig) string
	// Discard スカルトで余剰札を捨てる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MinchiateConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MinchiateInteractor ミンキアーテのインタラクタークラス
type MinchiateInteractor struct {
	GameBase[interfaces.MinchiateGame]
	tp presenter.MinchiatePresenter
}

// NewMinchiateInteractor コンストラクタ
func NewMinchiateInteractor(g interfaces.MinchiateGame, tp presenter.MinchiatePresenter) *MinchiateInteractor {
	mustNotNil("MinchiateInteractor", map[string]any{"g": g, "tp": tp})
	return &MinchiateInteractor{GameBase: GameBase[interfaces.MinchiateGame]{Game: g}, tp: tp}
}

// minchiateTrickPhases ミンキアーテのトリックフェーズ定数
func minchiateTrickPhases() trickPhases[domain.MinchiatePhase] {
	return trickPhases[domain.MinchiatePhase]{
		play:     domain.MinchiatePhasePlay,
		trickEnd: domain.MinchiatePhaseTrickEnd,
		roundEnd: domain.MinchiatePhaseRoundEnd,
		gameEnd:  domain.MinchiatePhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *MinchiateInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *MinchiateInteractor) ResetWithConfig(cfg domain.MinchiateConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Discard スカルトで余剰札を捨てる
func (ci *MinchiateInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerScarto(cardIndices); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *MinchiateInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリックが揃った場合、即座に解決する。
	if ci.Game.GetPhase() == domain.MinchiatePhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *MinchiateInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *MinchiateInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *MinchiateInteractor) GetConfig() domain.MinchiateConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *MinchiateInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *MinchiateInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU のスカルトとプレイを、人間の手番かトリック/ラウンド終了まで自動実行する。
//
// **スカルトのループを先に回す。**ディーラーが CPU の局は、捨て終わるまで
// プレイフェーズに入らないので、トリックのループだけでは前に進まない。
func (ci *MinchiateInteractor) advance() {
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.MinchiatePhaseScarto || ci.Game.IsHumanScartoTurn()
	}, ci.Game.CpuScarto)
	runCpuTurnsLoop(ci.Game, minchiateTrickPhases())
}

// RestoreMinchiateInteractor deserialises JSON into a MinchiateInteractor.
func RestoreMinchiateInteractor(data []byte, tp presenter.MinchiatePresenter) (*MinchiateInteractor, error) {
	return restoreAndBuild[domain.Minchiate](data, func(g *domain.Minchiate) *MinchiateInteractor {
		return &MinchiateInteractor{GameBase: GameBase[interfaces.MinchiateGame]{Game: g}, tp: tp}
	})
}
