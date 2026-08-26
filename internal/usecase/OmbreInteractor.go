//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OmbreInteractorIF オンブル (Ombre) のインタラクターインタフェース
type OmbreInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.OmbreConfig) string
	// Bid ビッドする (pass/entrar/solo + 切り札スート)
	Bid(bid domain.OmbreBid, trumpSuit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.OmbreConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// OmbreInteractor オンブルのインタラクタークラス
type OmbreInteractor struct {
	GameBase[interfaces.OmbreGame]
	tp presenter.OmbrePresenter
}

// NewOmbreInteractor コンストラクタ
func NewOmbreInteractor(g interfaces.OmbreGame, tp presenter.OmbrePresenter) *OmbreInteractor {
	mustNotNil("OmbreInteractor", map[string]any{"g": g, "tp": tp})
	return &OmbreInteractor{GameBase: GameBase[interfaces.OmbreGame]{Game: g}, tp: tp}
}

// ombreTrickPhases Ombre のトリックフェーズ定数
func ombreTrickPhases() trickPhases[domain.OmbrePhase] {
	return trickPhases[domain.OmbrePhase]{
		play:     domain.OmbrePhasePlay,
		trickEnd: domain.OmbrePhaseTrickEnd,
		roundEnd: domain.OmbrePhaseRoundEnd,
		gameEnd:  domain.OmbrePhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *OmbreInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *OmbreInteractor) ResetWithConfig(cfg domain.OmbreConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid ビッドする
func (ci *OmbreInteractor) Bid(bid domain.OmbreBid, trumpSuit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(bid, trumpSuit); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *OmbreInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.OmbrePhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *OmbreInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *OmbreInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *OmbreInteractor) GetConfig() domain.OmbreConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *OmbreInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *OmbreInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU のビッド・プレイを人間の手番もしくはトリック/ラウンド終了になるまで自動実行する。
func (ci *OmbreInteractor) advance() {
	runCpuBidsLoop[domain.OmbrePhase](ci.Game, domain.OmbrePhaseBid)
	runCpuTurnsLoop(ci.Game, ombreTrickPhases())
}

// RestoreOmbreInteractor deserialises JSON into an OmbreInteractor.
func RestoreOmbreInteractor(data []byte, tp presenter.OmbrePresenter) (*OmbreInteractor, error) {
	return restoreAndBuild[domain.Ombre](data, func(g *domain.Ombre) *OmbreInteractor {
		return &OmbreInteractor{GameBase: GameBase[interfaces.OmbreGame]{Game: g}, tp: tp}
	})
}
