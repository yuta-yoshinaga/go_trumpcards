//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// UnsunKarutaInteractorIF はうんすんカルタのインタラクターインタフェース。
type UnsunKarutaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.UnsunKarutaConfig) string
	// Play カードを出す (リードなら declare でメリ / モンチ を宣言)
	Play(cardIndex int, declare bool) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールを集計して次へ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.UnsunKarutaConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// UnsunKarutaInteractor はうんすんカルタのインタラクター。
type UnsunKarutaInteractor struct {
	GameBase[interfaces.UnsunKarutaGame]
	tp presenter.UnsunKarutaPresenter
}

// NewUnsunKarutaInteractor コンストラクタ。
func NewUnsunKarutaInteractor(
	g interfaces.UnsunKarutaGame, tp presenter.UnsunKarutaPresenter,
) *UnsunKarutaInteractor {
	mustNotNil("UnsunKarutaInteractor", map[string]any{"g": g, "tp": tp})
	return &UnsunKarutaInteractor{GameBase: GameBase[interfaces.UnsunKarutaGame]{Game: g}, tp: tp}
}

// unsunKarutaTrickPhases はトリック進行のフェーズ定数。
func unsunKarutaTrickPhases() trickPhases[domain.UnsunKarutaPhase] {
	return trickPhases[domain.UnsunKarutaPhase]{
		play:     domain.UnsunKarutaPhasePlay,
		trickEnd: domain.UnsunKarutaPhaseTrickEnd,
		roundEnd: domain.UnsunKarutaPhaseRoundEnd,
		gameEnd:  domain.UnsunKarutaPhaseGameEnd,
	}
}

// Reset ゲーム初期化。
func (ci *UnsunKarutaInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化。
func (ci *UnsunKarutaInteractor) ResetWithConfig(cfg domain.UnsunKarutaConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードを出す。
func (ci *UnsunKarutaInteractor) Play(cardIndex int, declare bool) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex, declare); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が 8 枚目を出してトリックが揃ったら、その場で解決する。
	if ci.Game.GetPhase() == domain.UnsunKarutaPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む。
func (ci *UnsunKarutaInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールを集計して次へ進む。
func (ci *UnsunKarutaInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得。
func (ci *UnsunKarutaInteractor) GetConfig() domain.UnsunKarutaConfig { return ci.Game.GetConfig() }

// Hint ヒント取得。
func (ci *UnsunKarutaInteractor) Hint() string { return ci.tp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する。
func (ci *UnsunKarutaInteractor) ActionLog() string { return ci.tp.ActionLogOutput(ci.Game) }

// advance は CPU の手番を人間の番かトリック終了まで進める。
func (ci *UnsunKarutaInteractor) advance() {
	runCpuTurnsLoop(ci.Game, unsunKarutaTrickPhases())
}

// RestoreUnsunKarutaInteractor deserialises JSON into an interactor.
func RestoreUnsunKarutaInteractor(
	data []byte, tp presenter.UnsunKarutaPresenter,
) (*UnsunKarutaInteractor, error) {
	return restoreAndBuild[domain.UnsunKaruta](data, func(g *domain.UnsunKaruta) *UnsunKarutaInteractor {
		return &UnsunKarutaInteractor{GameBase: GameBase[interfaces.UnsunKarutaGame]{Game: g}, tp: tp}
	})
}
