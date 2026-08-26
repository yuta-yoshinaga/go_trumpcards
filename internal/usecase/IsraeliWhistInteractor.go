//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// IsraeliWhistInteractorIF イスラエリホイストインタラクターインタフェース
type IsraeliWhistInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.IsraeliWhistConfig) string
	// AuctionBid 1 段階目のオークションで入札する
	AuctionBid(bid, suit int) string
	// AuctionPass オークションを降りる
	AuctionPass() string
	// Bid 2 段階目で目標トリック数を宣言する
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.IsraeliWhistConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// IsraeliWhistInteractor イスラエリホイストインタラクタークラス
type IsraeliWhistInteractor struct {
	GameBase[interfaces.IsraeliWhistGame]
	wp presenter.IsraeliWhistPresenter
}

// NewIsraeliWhistInteractor コンストラクタ
func NewIsraeliWhistInteractor(w interfaces.IsraeliWhistGame, wp presenter.IsraeliWhistPresenter) *IsraeliWhistInteractor {
	mustNotNil("IsraeliWhistInteractor", map[string]any{"w": w, "wp": wp})
	return &IsraeliWhistInteractor{GameBase: GameBase[interfaces.IsraeliWhistGame]{Game: w}, wp: wp}
}

// Reset ゲーム初期化。配り終えたら人間の番まで進める。
//
// **オークション・宣言・プレイの 3 段を順に進める。** 入札が 2 段階あるぶん、
// どこで止まっても人間の手番が来ない盤面になりうる。
func (wi *IsraeliWhistInteractor) Reset() string {
	wi.Game.Reset()
	wi.advanceToHuman()
	return wi.wp.Output(wi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (wi *IsraeliWhistInteractor) ResetWithConfig(cfg domain.IsraeliWhistConfig) string {
	return resetWithValidatedConfig(wi.Game, wi.wp, cfg, wi.Game.SetConfig, wi.Reset)
}

// AuctionBid 1 段階目のオークションで入札する
func (wi *IsraeliWhistInteractor) AuctionBid(bid, suit int) string {
	return wi.act(func() error { return wi.Game.PlayerAuctionBid(bid, suit) })
}

// AuctionPass オークションを降りる
func (wi *IsraeliWhistInteractor) AuctionPass() string {
	return wi.act(wi.Game.PlayerAuctionPass)
}

// Bid 2 段階目で目標トリック数を宣言する
func (wi *IsraeliWhistInteractor) Bid(bid int) string {
	return wi.act(func() error { return wi.Game.PlayerBid(bid) })
}

// act 入札系コマンドの共通処理
func (wi *IsraeliWhistInteractor) act(fn func() error) string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	if err := fn(); err != nil {
		return wi.wp.Output(wi.Game, err)
	}
	wi.advanceToHuman()
	return wi.wp.Output(wi.Game, nil)
}

// Play カードをプレイ
func (wi *IsraeliWhistInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(wi.Game, wi.wp); blocked {
		return out
	}
	if err := wi.Game.PlayerPlay(cardIndex); err != nil {
		return wi.wp.Output(wi.Game, err)
	}
	wi.runCpuTurns()
	return wi.wp.Output(wi.Game, nil)
}

// NextRound 次のラウンドへ進む
func (wi *IsraeliWhistInteractor) NextRound() string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	wi.Game.NextRound()
	// 次のラウンドもオークションから始まるので、人間の番まで進める。
	wi.advanceToHuman()
	return wi.wp.Output(wi.Game, nil)
}

// GiveUp 投了する
func (wi *IsraeliWhistInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	wi.Game.GiveUp()
	return wi.wp.Output(wi.Game, nil)
}

// GetConfig 現在の設定を取得
func (wi *IsraeliWhistInteractor) GetConfig() domain.IsraeliWhistConfig { return wi.Game.GetConfig() }

// Hint ヒント取得
func (wi *IsraeliWhistInteractor) Hint() string { return wi.wp.HintOutput(wi.Game) }

// ActionLog 棋譜を出力する
func (wi *IsraeliWhistInteractor) ActionLog() string { return wi.wp.ActionLogOutput(wi.Game) }

// advanceToHuman オークション → 宣言 → プレイ の順に、人間の番まで CPU を進める
func (wi *IsraeliWhistInteractor) advanceToHuman() {
	wi.runCpuAuction()
	wi.runCpuBids()
	wi.runCpuTurns()
}

// runCpuAuction オークションのあいだ、人間の番になるまで CPU に決めさせる。
//
// **降りた席は飛ばされるので 4 手では終わらない。** 競り上げが続くぶんだけ
// 回るので、上限を置いてハングを防ぐ (#4607 と同じ理由)。
func (wi *IsraeliWhistInteractor) runCpuAuction() {
	for turns := 0; !wi.Game.GetGameEndFlag(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if wi.Game.GetPhase() != domain.IsraeliWhistPhaseAuction || wi.Game.IsHumanAuctionTurn() {
			return
		}
		wi.Game.CpuAuction()
	}
}

// runCpuBids 宣言フェーズのあいだ、人間の番になるまで CPU に宣言させる
func (wi *IsraeliWhistInteractor) runCpuBids() {
	for turns := 0; !wi.Game.GetGameEndFlag(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if wi.Game.GetPhase() != domain.IsraeliWhistPhaseBid || wi.Game.IsHumanBidTurn() {
			return
		}
		wi.Game.CpuBid()
	}
}

// runCpuTurns 人間の手番になるかラウンド／ゲームが終わるまで CPU を進める
func (wi *IsraeliWhistInteractor) runCpuTurns() {
	for turns := 0; !wi.Game.GetGameEndFlag() && !wi.Game.IsHumanTurn(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if wi.Game.GetPhase() != domain.IsraeliWhistPhasePlay {
			return
		}
		wi.Game.CpuPlay()
	}
}

// RestoreIsraeliWhistInteractor deserialises JSON into an IsraeliWhistInteractor.
func RestoreIsraeliWhistInteractor(data []byte, wp presenter.IsraeliWhistPresenter) (*IsraeliWhistInteractor, error) {
	return restoreAndBuild[domain.IsraeliWhist](data, func(g *domain.IsraeliWhist) *IsraeliWhistInteractor {
		return &IsraeliWhistInteractor{GameBase: GameBase[interfaces.IsraeliWhistGame]{Game: g}, wp: wp}
	})
}
