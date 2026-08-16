//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TarocchiniInteractorIF タロッキーニのインタラクターインタフェース
type TarocchiniInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TarocchiniConfig) string
	// Discard スカルトで余剰札を捨てる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TarocchiniConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TarocchiniInteractor タロッキーニのインタラクタークラス
type TarocchiniInteractor struct {
	GameBase[interfaces.TarocchiniGame]
	tp presenter.TarocchiniPresenter
}

// NewTarocchiniInteractor コンストラクタ
func NewTarocchiniInteractor(g interfaces.TarocchiniGame, tp presenter.TarocchiniPresenter) *TarocchiniInteractor {
	mustNotNil("TarocchiniInteractor", map[string]any{"g": g, "tp": tp})
	return &TarocchiniInteractor{GameBase: GameBase[interfaces.TarocchiniGame]{Game: g}, tp: tp}
}

// tarocchiniTrickPhases タロッキーニのトリックフェーズ定数
func tarocchiniTrickPhases() trickPhases[domain.TarocchiniPhase] {
	return trickPhases[domain.TarocchiniPhase]{
		play:     domain.TarocchiniPhasePlay,
		trickEnd: domain.TarocchiniPhaseTrickEnd,
		roundEnd: domain.TarocchiniPhaseRoundEnd,
		gameEnd:  domain.TarocchiniPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *TarocchiniInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *TarocchiniInteractor) ResetWithConfig(cfg domain.TarocchiniConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Discard スカルトで余剰札を捨てる
func (ci *TarocchiniInteractor) Discard(cardIndices []int) string {
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
func (ci *TarocchiniInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリックが揃った場合、即座に解決する。
	if ci.Game.GetPhase() == domain.TarocchiniPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *TarocchiniInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *TarocchiniInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *TarocchiniInteractor) GetConfig() domain.TarocchiniConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *TarocchiniInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *TarocchiniInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU のスカルトとプレイを、人間の手番かトリック/ラウンド終了まで自動実行する。
//
// **スカルトのループを先に回す。**ディーラーが CPU の局は、捨て終わるまで
// プレイフェーズに入らないので、トリックのループだけでは前に進まない。
func (ci *TarocchiniInteractor) advance() {
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.TarocchiniPhaseScarto || ci.Game.IsHumanScartoTurn()
	}, ci.Game.CpuScarto)
	runCpuTurnsLoop(ci.Game, tarocchiniTrickPhases())
}

// RestoreTarocchiniInteractor deserialises JSON into a TarocchiniInteractor.
func RestoreTarocchiniInteractor(data []byte, tp presenter.TarocchiniPresenter) (*TarocchiniInteractor, error) {
	return restoreAndBuild[domain.Tarocchini](data, func(g *domain.Tarocchini) *TarocchiniInteractor {
		return &TarocchiniInteractor{GameBase: GameBase[interfaces.TarocchiniGame]{Game: g}, tp: tp}
	})
}
