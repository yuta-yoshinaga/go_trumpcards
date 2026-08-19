//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BidEuchreInteractorIF ビッド・ユーカー (Bid Euchre) のインタラクターインタフェース
type BidEuchreInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BidEuchreConfig) string
	// Bid 宣言する
	Bid(value int) string
	// PassBid 宣言を見送る
	PassBid() string
	// ChooseTrump 切札を宣言する
	ChooseTrump(t int) string
	// PlayCard 手札を1枚出す
	PlayCard(idx int) string
	// NextHand 次の局へ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BidEuchreConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BidEuchreInteractor ビッド・ユーカー (Bid Euchre) のインタラクタークラス
type BidEuchreInteractor struct {
	GameBase[interfaces.BidEuchreGame]
	gp presenter.BidEuchrePresenter
}

// NewBidEuchreInteractor コンストラクタ
func NewBidEuchreInteractor(g interfaces.BidEuchreGame, gp presenter.BidEuchrePresenter) *BidEuchreInteractor {
	mustNotNil("BidEuchreInteractor", map[string]any{"g": g, "gp": gp})
	return &BidEuchreInteractor{GameBase: GameBase[interfaces.BidEuchreGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (bi *BidEuchreInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.gp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BidEuchreInteractor) ResetWithConfig(cfg domain.BidEuchreConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.gp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Bid 宣言する
func (bi *BidEuchreInteractor) Bid(value int) string {
	return bi.act(func() error { return bi.Game.Bid(bi.Game.GetBidPlayerIdx(), value) })
}

// PassBid 宣言を見送る
func (bi *BidEuchreInteractor) PassBid() string {
	return bi.act(func() error { return bi.Game.PassBid(bi.Game.GetBidPlayerIdx()) })
}

// ChooseTrump 切札を宣言する
func (bi *BidEuchreInteractor) ChooseTrump(t int) string {
	return bi.act(func() error {
		return bi.Game.ChooseTrump(bi.Game.GetDeclarerIdx(), domain.BidEuchreTrump(t))
	})
}

// PlayCard 手札を1枚出す
func (bi *BidEuchreInteractor) PlayCard(idx int) string {
	return bi.act(func() error { return bi.Game.PlayCard(bi.Game.GetCurrentPlayerIdx(), idx) })
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (bi *BidEuchreInteractor) act(action func() error) string {
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
func (bi *BidEuchreInteractor) NextHand() string {
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
func (bi *BidEuchreInteractor) GetConfig() domain.BidEuchreConfig {
	return bi.Game.GetConfig()
}

// Hint ヒントを出力する
func (bi *BidEuchreInteractor) Hint() string { return bi.gp.HintOutput(bi.Game) }

// ActionLog 棋譜を出力する
func (bi *BidEuchreInteractor) ActionLog() string {
	return bi.gp.ActionLogOutput(bi.Game)
}

// bidEuchreMaxCpuSteps bounds runCpuTurns so a malformed state can never spin
// the CPU loop forever (defensive — normal play always reaches a human turn,
// the settlement, or game end well within this limit).
const bidEuchreMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (bi *BidEuchreInteractor) runCpuTurns() {
	for step := 0; step < bidEuchreMaxCpuSteps && !bi.Game.GetGameEndFlag(); step++ {
		phase := bi.Game.GetPhase()
		if phase == domain.BidEuchrePhaseHandEnd || phase == domain.BidEuchrePhaseGameEnd {
			break
		}
		if bi.Game.IsHumanTurn() {
			break
		}
		bi.Game.CpuPlay()
	}
}

// RestoreBidEuchreInteractor deserialises JSON into a BidEuchreInteractor.
func RestoreBidEuchreInteractor(data []byte, gp presenter.BidEuchrePresenter) (*BidEuchreInteractor, error) {
	return restoreAndBuild[domain.BidEuchre](data, func(g *domain.BidEuchre) *BidEuchreInteractor {
		return &BidEuchreInteractor{GameBase: GameBase[interfaces.BidEuchreGame]{Game: g}, gp: gp}
	})
}
