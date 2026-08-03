//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SixBidSoloInteractorIF シックスビッド・ソロ (Six-Bid Solo) のインタラクターインタフェース
type SixBidSoloInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SixBidSoloConfig) string
	// Bid 宣言する
	Bid(kind int) string
	// PassBid 宣言を見送る
	PassBid() string
	// Declare 切札 (とコール・ソロの指名札) を決める
	Declare(suit, calledSuit, calledValue int) string
	// PlayCard 手札を1枚出す
	PlayCard(idx int) string
	// NextHand 次の局へ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SixBidSoloConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SixBidSoloInteractor シックスビッド・ソロ (Six-Bid Solo) のインタラクタークラス
type SixBidSoloInteractor struct {
	GameBase[interfaces.SixBidSoloGame]
	gp presenter.SixBidSoloPresenter
}

// NewSixBidSoloInteractor コンストラクタ
func NewSixBidSoloInteractor(g interfaces.SixBidSoloGame, gp presenter.SixBidSoloPresenter) *SixBidSoloInteractor {
	mustNotNil("SixBidSoloInteractor", map[string]any{"g": g, "gp": gp})
	return &SixBidSoloInteractor{GameBase: GameBase[interfaces.SixBidSoloGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (si *SixBidSoloInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.gp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SixBidSoloInteractor) ResetWithConfig(cfg domain.SixBidSoloConfig) string {
	return resetWithValidatedConfig(si.Game, si.gp, cfg, si.Game.SetConfig, si.Reset)
}

// Bid 宣言する
func (si *SixBidSoloInteractor) Bid(kind int) string {
	return si.act(func() error {
		return si.Game.Bid(si.Game.GetBidPlayerIdx(), domain.SixBidSoloBidKind(kind))
	})
}

// PassBid 宣言を見送る
func (si *SixBidSoloInteractor) PassBid() string {
	return si.act(func() error { return si.Game.PassBid(si.Game.GetBidPlayerIdx()) })
}

// Declare 切札 (とコール・ソロの指名札) を決める
//
// **コール・ソロ以外では指名札を無視する。**calledSuit が 0 なら札は送られて
// いないものとして扱う。
func (si *SixBidSoloInteractor) Declare(suit, calledSuit, calledValue int) string {
	return si.act(func() error {
		var called *domain.Card
		if calledSuit > 0 {
			called = domain.NewCard(calledSuit, calledValue, true)
		}
		return si.Game.Declare(si.Game.GetDeclarerIdx(), suit, called)
	})
}

// PlayCard 手札を1枚出す
func (si *SixBidSoloInteractor) PlayCard(idx int) string {
	return si.act(func() error { return si.Game.PlayCard(si.Game.GetCurrentPlayerIdx(), idx) })
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (si *SixBidSoloInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(si.Game, si.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return si.gp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.gp.Output(si.Game, nil)
}

// NextHand 次の局へ進む
func (si *SixBidSoloInteractor) NextHand() string {
	if out, blocked := guardGameEnd(si.Game, si.gp); blocked {
		return out
	}
	if err := si.Game.NextHand(); err != nil {
		return si.gp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.gp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SixBidSoloInteractor) GetConfig() domain.SixBidSoloConfig {
	return si.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *SixBidSoloInteractor) ActionLog() string {
	return si.gp.ActionLogOutput(si.Game)
}

// sixBidSoloMaxCpuSteps bounds runCpuTurns so a malformed state can never spin
// the CPU loop forever (defensive — normal play always reaches a human turn,
// the settlement, or game end well within this limit).
const sixBidSoloMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (si *SixBidSoloInteractor) runCpuTurns() {
	for step := 0; step < sixBidSoloMaxCpuSteps && !si.Game.GetGameEndFlag(); step++ {
		phase := si.Game.GetPhase()
		if phase == domain.SixBidSoloPhaseHandEnd || phase == domain.SixBidSoloPhaseGameEnd {
			break
		}
		if si.Game.IsHumanTurn() {
			break
		}
		si.Game.CpuPlay()
	}
}

// RestoreSixBidSoloInteractor deserialises JSON into a SixBidSoloInteractor.
func RestoreSixBidSoloInteractor(data []byte, gp presenter.SixBidSoloPresenter) (*SixBidSoloInteractor, error) {
	return restoreAndBuild[domain.SixBidSolo](data, func(g *domain.SixBidSolo) *SixBidSoloInteractor {
		return &SixBidSoloInteractor{GameBase: GameBase[interfaces.SixBidSoloGame]{Game: g}, gp: gp}
	})
}
