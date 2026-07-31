//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BostonInteractorIF ボストン (Boston) のインタラクターインタフェース
type BostonInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BostonConfig) string
	// Bid 宣言する
	Bid(level domain.BostonBidLevel, suit int) string
	// PassBid 宣言を見送る
	PassBid() string
	// CallPartner パートナーを指名する (-1 なら単独)
	CallPartner(partner int) string
	// PlayCard 手札を1枚出す
	PlayCard(idx int) string
	// NextHand 次の局へ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BostonConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BostonInteractor ボストン (Boston) のインタラクタークラス
type BostonInteractor struct {
	GameBase[interfaces.BostonGame]
	gp presenter.BostonPresenter
}

// NewBostonInteractor コンストラクタ
func NewBostonInteractor(g interfaces.BostonGame, gp presenter.BostonPresenter) *BostonInteractor {
	mustNotNil("BostonInteractor", map[string]any{"g": g, "gp": gp})
	return &BostonInteractor{GameBase: GameBase[interfaces.BostonGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (bi *BostonInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.gp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BostonInteractor) ResetWithConfig(cfg domain.BostonConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.gp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Bid 宣言する
func (bi *BostonInteractor) Bid(level domain.BostonBidLevel, suit int) string {
	return bi.act(func() error { return bi.Game.Bid(bi.Game.GetBidPlayerIdx(), level, suit) })
}

// PassBid 宣言を見送る
func (bi *BostonInteractor) PassBid() string {
	return bi.act(func() error { return bi.Game.PassBid(bi.Game.GetBidPlayerIdx()) })
}

// CallPartner パートナーを指名する
//
// **落札者の席で呼ぶ。**宣言手番はもう進んでいるので bidIdx は使えない。
func (bi *BostonInteractor) CallPartner(partner int) string {
	return bi.act(func() error { return bi.Game.CallPartner(bi.Game.GetDeclarerIdx(), partner) })
}

// PlayCard 手札を1枚出す
func (bi *BostonInteractor) PlayCard(idx int) string {
	return bi.act(func() error { return bi.Game.PlayCard(bi.Game.GetCurrentPlayerIdx(), idx) })
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (bi *BostonInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return bi.gp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.gp.Output(bi.Game, nil)
}

// NextHand 次の局へ進む
func (bi *BostonInteractor) NextHand() string {
	if out, blocked := guardGameEnd(bi.Game, bi.gp); blocked {
		return out
	}
	if err := bi.Game.NextHand(); err != nil {
		return bi.gp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.gp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BostonInteractor) GetConfig() domain.BostonConfig {
	return bi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (bi *BostonInteractor) ActionLog() string {
	return bi.gp.ActionLogOutput(bi.Game)
}

// bostonMaxCpuSteps bounds runCpuTurns so a malformed state can never spin the
// CPU loop forever (defensive — normal play always reaches a human turn, the
// settlement, or game end well within this limit).
const bostonMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (bi *BostonInteractor) runCpuTurns() {
	for step := 0; step < bostonMaxCpuSteps && !bi.Game.GetGameEndFlag(); step++ {
		phase := bi.Game.GetPhase()
		if phase == domain.BostonPhaseHandEnd || phase == domain.BostonPhaseGameEnd {
			break
		}
		if bi.Game.IsHumanTurn() {
			break
		}
		bi.Game.CpuPlay()
	}
}

// RestoreBostonInteractor deserialises JSON into a BostonInteractor.
func RestoreBostonInteractor(data []byte, gp presenter.BostonPresenter) (*BostonInteractor, error) {
	return restoreAndBuild[domain.Boston](data, func(g *domain.Boston) *BostonInteractor {
		return &BostonInteractor{GameBase: GameBase[interfaces.BostonGame]{Game: g}, gp: gp}
	})
}
