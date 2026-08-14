//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HasenpfefferInteractorIF ハーゼンプフェファーインタラクターインタフェース
type HasenpfefferInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.HasenpfefferConfig) string
	// Bid 宣言する (0: 降りる)
	Bid(bid int) string
	// Discard 切り札を宣言して1枚捨てる
	Discard(cardIndex, suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextHand 次のハンドへ進む
	NextHand() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.HasenpfefferConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// HasenpfefferInteractor ハーゼンプフェファーインタラクタークラス
type HasenpfefferInteractor struct {
	GameBase[interfaces.HasenpfefferGame]
	hp presenter.HasenpfefferPresenter
}

// NewHasenpfefferInteractor コンストラクタ
func NewHasenpfefferInteractor(h interfaces.HasenpfefferGame, hp presenter.HasenpfefferPresenter) *HasenpfefferInteractor {
	mustNotNil("HasenpfefferInteractor", map[string]any{"h": h, "hp": hp})
	return &HasenpfefferInteractor{GameBase: GameBase[interfaces.HasenpfefferGame]{Game: h}, hp: hp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (hi *HasenpfefferInteractor) Reset() string {
	hi.Game.Reset()
	hi.advance()
	return hi.hp.Output(hi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HasenpfefferInteractor) ResetWithConfig(cfg domain.HasenpfefferConfig) string {
	return resetWithValidatedConfig(hi.Game, hi.hp, cfg, hi.Game.SetConfig, hi.Reset)
}

// Bid 宣言する
func (hi *HasenpfefferInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	if err := hi.Game.PlayerBid(bid); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	hi.advance()
	return hi.hp.Output(hi.Game, nil)
}

// Discard 切り札を宣言して1枚捨てる
func (hi *HasenpfefferInteractor) Discard(cardIndex, suit int) string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	if err := hi.Game.PlayerDiscard(cardIndex, suit); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	hi.advance()
	return hi.hp.Output(hi.Game, nil)
}

// Play カードをプレイ
func (hi *HasenpfefferInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(hi.Game, hi.hp); blocked {
		return out
	}
	if err := hi.Game.PlayerPlay(cardIndex); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	hi.advance()
	return hi.hp.Output(hi.Game, nil)
}

// NextHand 次のハンドへ進む
func (hi *HasenpfefferInteractor) NextHand() string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	hi.Game.NextHand()
	hi.advance()
	return hi.hp.Output(hi.Game, nil)
}

// GiveUp 投了する
func (hi *HasenpfefferInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	hi.Game.GiveUp()
	return hi.hp.Output(hi.Game, nil)
}

// GetConfig 現在の設定を取得
func (hi *HasenpfefferInteractor) GetConfig() domain.HasenpfefferConfig { return hi.Game.GetConfig() }

// Hint ヒント取得
func (hi *HasenpfefferInteractor) Hint() string { return hi.hp.HintOutput(hi.Game) }

// ActionLog 棋譜を出力する
func (hi *HasenpfefferInteractor) ActionLog() string { return hi.hp.ActionLogOutput(hi.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **競り・捨て札・プレイの 3 段すべてを回す。** どれか 1 つで止めると、人間が
// 操作できない盤面を返してしまう。ハンド終了では止める（next は明示的に）。
func (hi *HasenpfefferInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if hi.Game.GetGameEndFlag() {
			return
		}
		switch hi.Game.GetPhase() {
		case domain.HasenpfefferPhaseBid:
			if hi.Game.IsHumanBidTurn() {
				return
			}
			hi.Game.CpuBid()
		case domain.HasenpfefferPhaseDiscard:
			if hi.Game.IsHumanDiscardTurn() {
				return
			}
			hi.Game.CpuDiscard()
		case domain.HasenpfefferPhasePlay:
			if hi.Game.IsHumanTurn() {
				return
			}
			hi.Game.CpuPlay()
		default:
			return
		}
	}
}

// RestoreHasenpfefferInteractor deserialises JSON into a HasenpfefferInteractor.
func RestoreHasenpfefferInteractor(data []byte, hp presenter.HasenpfefferPresenter) (*HasenpfefferInteractor, error) {
	return restoreAndBuild[domain.Hasenpfeffer](data, func(g *domain.Hasenpfeffer) *HasenpfefferInteractor {
		return NewHasenpfefferInteractor(g, hp)
	})
}
