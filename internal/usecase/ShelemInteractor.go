//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ShelemInteractorIF シェレムインタラクターインタフェース
type ShelemInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ShelemConfig) string
	// Bid 点数で入札する
	Bid(bid int) string
	// BidShelem Shelem（全トリック独占）を宣言する
	BidShelem() string
	// Pass 競りを降りる
	Pass() string
	// Discard 落札者が 4 枚捨てて切り札を決める
	Discard(indices []int, suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ShelemConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ShelemInteractor シェレムインタラクタークラス
type ShelemInteractor struct {
	GameBase[interfaces.ShelemGame]
	sp presenter.ShelemPresenter
}

// NewShelemInteractor コンストラクタ
func NewShelemInteractor(s interfaces.ShelemGame, sp presenter.ShelemPresenter) *ShelemInteractor {
	mustNotNil("ShelemInteractor", map[string]any{"s": s, "sp": sp})
	return &ShelemInteractor{GameBase: GameBase[interfaces.ShelemGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。配り終えたら人間の番まで進める。
//
// **競り・捨て札・プレイの 3 段を順に進める。** CPU が落札するとウィドウ交換と
// 切り札決定までその場で終わるので、そこで止めると誰も打たない盤面を返す。
func (si *ShelemInteractor) Reset() string {
	si.Game.Reset()
	si.advanceToHuman()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *ShelemInteractor) ResetWithConfig(cfg domain.ShelemConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Bid 点数で入札する
func (si *ShelemInteractor) Bid(bid int) string {
	return si.act(func() error { return si.Game.PlayerBid(bid) })
}

// BidShelem Shelem（全トリック独占）を宣言する
func (si *ShelemInteractor) BidShelem() string { return si.act(si.Game.PlayerBidShelem) }

// Pass 競りを降りる
func (si *ShelemInteractor) Pass() string { return si.act(si.Game.PlayerPass) }

// Discard 落札者が 4 枚捨てて切り札を決める
func (si *ShelemInteractor) Discard(indices []int, suit int) string {
	return si.act(func() error { return si.Game.PlayerDiscard(indices, suit) })
}

// act 競り・捨て札系コマンドの共通処理
func (si *ShelemInteractor) act(fn func() error) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := fn(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.advanceToHuman()
	return si.sp.Output(si.Game, nil)
}

// Play カードをプレイ
func (si *ShelemInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPlay(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound 次のラウンドへ進む
func (si *ShelemInteractor) NextRound() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	// 次のラウンドも競りから始まるので、人間の番まで進める。
	si.advanceToHuman()
	return si.sp.Output(si.Game, nil)
}

// GiveUp 投了する
func (si *ShelemInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.GiveUp()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *ShelemInteractor) GetConfig() domain.ShelemConfig { return si.Game.GetConfig() }

// Hint ヒント取得
func (si *ShelemInteractor) Hint() string { return si.sp.HintOutput(si.Game) }

// ActionLog 棋譜を出力する
func (si *ShelemInteractor) ActionLog() string { return si.sp.ActionLogOutput(si.Game) }

// advanceToHuman 競り → プレイ の順に、人間の番まで CPU を進める。
//
// **捨て札は CPU の落札なら競りの締めで済んでいる。** 人間が落札したときだけ
// 捨て札フェーズで止まり、そこで待つ。
func (si *ShelemInteractor) advanceToHuman() {
	si.runCpuBids()
	si.runCpuTurns()
}

// runCpuBids 競りのあいだ、人間の番になるまで CPU に決めさせる。
//
// **降りた席は飛ばされるので 4 手では終わらない。** 競り上げが続くぶんだけ
// 回るので、上限を置いてハングを防ぐ (#4607 と同じ理由)。
func (si *ShelemInteractor) runCpuBids() {
	for turns := 0; !si.Game.GetGameEndFlag(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if si.Game.GetPhase() != domain.ShelemPhaseBid || si.Game.IsHumanBidTurn() {
			return
		}
		si.Game.CpuBid()
	}
}

// runCpuTurns 人間の手番になるかラウンド／ゲームが終わるまで CPU を進める
func (si *ShelemInteractor) runCpuTurns() {
	for turns := 0; !si.Game.GetGameEndFlag() && !si.Game.IsHumanTurn(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if si.Game.GetPhase() != domain.ShelemPhasePlay {
			return
		}
		si.Game.CpuPlay()
	}
}

// RestoreShelemInteractor deserialises JSON into a ShelemInteractor.
func RestoreShelemInteractor(data []byte, sp presenter.ShelemPresenter) (*ShelemInteractor, error) {
	return restoreAndBuild[domain.Shelem](data, func(g *domain.Shelem) *ShelemInteractor {
		return &ShelemInteractor{GameBase: GameBase[interfaces.ShelemGame]{Game: g}, sp: sp}
	})
}
