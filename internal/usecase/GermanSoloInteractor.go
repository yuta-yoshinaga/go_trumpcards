//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GermanSoloInteractorIF ジャーマン・ソロ (GermanSolo) のインタラクターインタフェース
type GermanSoloInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.GermanSoloConfig) string
	// Bid ビッドする (pass/frage/solo/tout + 切り札スート)
	Bid(bid domain.GermanSoloBid, trumpSuit int) string
	// CallAce 落札者が味方を呼ぶエースを指名する
	CallAce(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.GermanSoloConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GermanSoloInteractor ジャーマン・ソロのインタラクタークラス
type GermanSoloInteractor struct {
	GameBase[interfaces.GermanSoloGame]
	tp presenter.GermanSoloPresenter
}

// NewGermanSoloInteractor コンストラクタ
func NewGermanSoloInteractor(g interfaces.GermanSoloGame, tp presenter.GermanSoloPresenter) *GermanSoloInteractor {
	mustNotNil("GermanSoloInteractor", map[string]any{"g": g, "tp": tp})
	return &GermanSoloInteractor{GameBase: GameBase[interfaces.GermanSoloGame]{Game: g}, tp: tp}
}

// germanSoloTrickPhases GermanSolo のトリックフェーズ定数
func germanSoloTrickPhases() trickPhases[domain.GermanSoloPhase] {
	return trickPhases[domain.GermanSoloPhase]{
		play:     domain.GermanSoloPhasePlay,
		trickEnd: domain.GermanSoloPhaseTrickEnd,
		roundEnd: domain.GermanSoloPhaseRoundEnd,
		gameEnd:  domain.GermanSoloPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *GermanSoloInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *GermanSoloInteractor) ResetWithConfig(cfg domain.GermanSoloConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid ビッドする
func (ci *GermanSoloInteractor) Bid(bid domain.GermanSoloBid, trumpSuit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(bid, trumpSuit); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// CallAce は人間の落札者が味方を呼ぶエースを指名する。
//
// **これが無いと盤面は落札の直後で固まる。** エース呼びフェーズを抜けるまで
// PlayerPlay は「フェーズが違う」で弾かれ続ける。
func (ci *GermanSoloInteractor) CallAce(suit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.DeclareAce(ci.Game.GetDeclarerIdx(), suit); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *GermanSoloInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.GermanSoloPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *GermanSoloInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *GermanSoloInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *GermanSoloInteractor) GetConfig() domain.GermanSoloConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *GermanSoloInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *GermanSoloInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU のビッド・プレイを人間の手番もしくはトリック/ラウンド終了になるまで自動実行する。
func (ci *GermanSoloInteractor) advance() {
	runCpuBidsLoop[domain.GermanSoloPhase](ci.Game, domain.GermanSoloPhaseBid)
	ci.runAceCall()
	runCpuTurnsLoop(ci.Game, germanSoloTrickPhases())
}

// runAceCall は CPU が落札していればエース呼びを進める。
//
// 人間が落札していれば CallAce を待って止まる。Solo / Tout は単独契約なので
// このフェーズを通らず、呼べるエースが 1 枚も無い Frage も startAceCall が
// そのままプレイへ進めるので、どちらもここには来ない。
func (ci *GermanSoloInteractor) runAceCall() {
	if ci.Game.GetPhase() != domain.GermanSoloPhaseAceCall || ci.Game.IsHumanAceCallTurn() {
		return
	}
	ci.Game.CpuDeclareAce()
}

// RestoreGermanSoloInteractor deserialises JSON into an GermanSoloInteractor.
func RestoreGermanSoloInteractor(data []byte, tp presenter.GermanSoloPresenter) (*GermanSoloInteractor, error) {
	return restoreAndBuild[domain.GermanSolo](data, func(g *domain.GermanSolo) *GermanSoloInteractor {
		return &GermanSoloInteractor{GameBase: GameBase[interfaces.GermanSoloGame]{Game: g}, tp: tp}
	})
}
