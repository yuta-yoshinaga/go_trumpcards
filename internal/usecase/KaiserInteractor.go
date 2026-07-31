//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KaiserInteractorIF カイザー (Kaiser) のインタラクターインタフェース
type KaiserInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KaiserConfig) string
	// Bid 点数を宣言する
	Bid(value int, contract domain.KaiserContract) string
	// PassBid ビッドを見送る
	PassBid() string
	// SetTrump 切札を指定する
	SetTrump(suit int) string
	// Discard キティを取り込んだあと2枚捨てる
	Discard(idxs []int) string
	// PlayCard 手札を1枚出す
	PlayCard(idx int) string
	// NextHand 次の局へ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KaiserConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KaiserInteractor カイザー (Kaiser) のインタラクタークラス
type KaiserInteractor struct {
	GameBase[interfaces.KaiserGame]
	gp presenter.KaiserPresenter
}

// NewKaiserInteractor コンストラクタ
func NewKaiserInteractor(g interfaces.KaiserGame, gp presenter.KaiserPresenter) *KaiserInteractor {
	mustNotNil("KaiserInteractor", map[string]any{"g": g, "gp": gp})
	return &KaiserInteractor{GameBase: GameBase[interfaces.KaiserGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ki *KaiserInteractor) Reset() string {
	ki.Game.Reset()
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ki *KaiserInteractor) ResetWithConfig(cfg domain.KaiserConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.gp, cfg, ki.Game.SetConfig, ki.Reset)
}

// Bid 点数を宣言する
func (ki *KaiserInteractor) Bid(value int, contract domain.KaiserContract) string {
	return ki.act(func() error { return ki.Game.Bid(ki.Game.GetBidPlayerIdx(), value, contract) })
}

// PassBid ビッドを見送る
func (ki *KaiserInteractor) PassBid() string {
	return ki.act(func() error { return ki.Game.PassBid(ki.Game.GetBidPlayerIdx()) })
}

// SetTrump 切札を指定する
//
// **落札者の席で呼ぶ。**ビッド手番はもう進んでいるので bidIdx は使えない。
func (ki *KaiserInteractor) SetTrump(suit int) string {
	return ki.act(func() error { return ki.Game.SetTrump(ki.Game.GetDeclarerIdx(), suit) })
}

// Discard キティを取り込んだあと2枚捨てる
func (ki *KaiserInteractor) Discard(idxs []int) string {
	return ki.act(func() error { return ki.Game.Discard(ki.Game.GetDeclarerIdx(), idxs) })
}

// PlayCard 手札を1枚出す
func (ki *KaiserInteractor) PlayCard(idx int) string {
	return ki.act(func() error { return ki.Game.PlayCard(ki.Game.GetCurrentPlayerIdx(), idx) })
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (ki *KaiserInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// NextHand 次の局へ進む
func (ki *KaiserInteractor) NextHand() string {
	if out, blocked := guardGameEnd(ki.Game, ki.gp); blocked {
		return out
	}
	if err := ki.Game.NextHand(); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// GetConfig 現在の設定を取得
func (ki *KaiserInteractor) GetConfig() domain.KaiserConfig {
	return ki.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ki *KaiserInteractor) ActionLog() string {
	return ki.gp.ActionLogOutput(ki.Game)
}

// kaiserMaxCpuSteps bounds runCpuTurns so a malformed state can never spin the
// CPU loop forever (defensive — normal play always reaches a human turn, the
// settlement, or game end well within this limit).
const kaiserMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (ki *KaiserInteractor) runCpuTurns() {
	for step := 0; step < kaiserMaxCpuSteps && !ki.Game.GetGameEndFlag(); step++ {
		phase := ki.Game.GetPhase()
		if phase == domain.KaiserPhaseHandEnd || phase == domain.KaiserPhaseGameEnd {
			break
		}
		if ki.Game.IsHumanTurn() {
			break
		}
		ki.Game.CpuPlay()
	}
}

// RestoreKaiserInteractor deserialises JSON into a KaiserInteractor.
func RestoreKaiserInteractor(data []byte, gp presenter.KaiserPresenter) (*KaiserInteractor, error) {
	return restoreAndBuild[domain.Kaiser](data, func(g *domain.Kaiser) *KaiserInteractor {
		return &KaiserInteractor{GameBase: GameBase[interfaces.KaiserGame]{Game: g}, gp: gp}
	})
}
