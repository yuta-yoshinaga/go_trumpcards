//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GleekInteractorIF グリーク (Gleek) のインタラクターインタフェース
type GleekInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.GleekConfig) string
	// Bid ストックを競る (0=降りる)
	Bid(bid int) string
	// Discard 落札者が 7 枚捨てる
	Discard(indices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.GleekConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GleekInteractor グリークのインタラクタークラス
type GleekInteractor struct {
	GameBase[interfaces.GleekGame]
	tp presenter.GleekPresenter
}

// NewGleekInteractor コンストラクタ
func NewGleekInteractor(g interfaces.GleekGame, tp presenter.GleekPresenter) *GleekInteractor {
	mustNotNil("GleekInteractor", map[string]any{"g": g, "tp": tp})
	return &GleekInteractor{GameBase: GameBase[interfaces.GleekGame]{Game: g}, tp: tp}
}

// gleekTrickPhases Gleek のトリックフェーズ定数
func gleekTrickPhases() trickPhases[domain.GleekPhase] {
	return trickPhases[domain.GleekPhase]{
		play:     domain.GleekPhasePlay,
		trickEnd: domain.GleekPhaseTrickEnd,
		roundEnd: domain.GleekPhaseRoundEnd,
		gameEnd:  domain.GleekPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *GleekInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *GleekInteractor) ResetWithConfig(cfg domain.GleekConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid ストックを競る
func (ci *GleekInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(bid); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Discard は落札した人間が 7 枚を捨てる。
//
// **これが無いと盤面は落札の直後で固まる。** 捨て札フェーズを抜けるまで
// PlayerPlay は「フェーズが違う」で弾かれ続ける。
func (ci *GleekInteractor) Discard(indices []int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(indices); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *GleekInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.GleekPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *GleekInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *GleekInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *GleekInteractor) GetConfig() domain.GleekConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *GleekInteractor) Hint() string { return ci.tp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *GleekInteractor) ActionLog() string { return ci.tp.ActionLogOutput(ci.Game) }

// advance CPU の競り・捨て札・プレイを人間の手番になるまで自動実行する。
func (ci *GleekInteractor) advance() {
	ci.runCpuBids()
	ci.runCpuDiscard()
	runCpuTurnsLoop(ci.Game, gleekTrickPhases())
}

// runCpuBids は競りフェーズを CPU の手番が続く間だけ進める。
//
// **上限で必ず止まる。** 競りは NextBidAmount が 0 になるか、残り 1 人になった
// 時点で閉じるので、ループは有限。それでも回数の壁を置くのは、壊れた盤を復元
// したときに HTTP ハンドラごと固まらないため。
func (ci *GleekInteractor) runCpuBids() {
	for i := 0; i < gleekMaxCpuBidSteps; i++ {
		if ci.Game.GetPhase() != domain.GleekPhaseBid || ci.Game.IsHumanBidTurn() {
			return
		}
		before := ci.Game.GetCurrentBidderIdx()
		ci.Game.CpuBid()
		if ci.Game.GetPhase() == domain.GleekPhaseBid && ci.Game.GetCurrentBidderIdx() == before {
			return // 進まなくなったら抜ける
		}
	}
}

// gleekMaxCpuBidSteps CPU の競りループの上限。
const gleekMaxCpuBidSteps = 64

// runCpuDiscard は CPU が落札していれば捨て札を進める。
//
// 人間が落札していれば Discard を待って止まる。
func (ci *GleekInteractor) runCpuDiscard() {
	if ci.Game.GetPhase() != domain.GleekPhaseDiscard || ci.Game.IsHumanDiscardTurn() {
		return
	}
	ci.Game.CpuDiscard()
}

// RestoreGleekInteractor deserialises JSON into a GleekInteractor.
func RestoreGleekInteractor(data []byte, tp presenter.GleekPresenter) (*GleekInteractor, error) {
	return restoreAndBuild[domain.Gleek](data, func(g *domain.Gleek) *GleekInteractor {
		return &GleekInteractor{GameBase: GameBase[interfaces.GleekGame]{Game: g}, tp: tp}
	})
}
