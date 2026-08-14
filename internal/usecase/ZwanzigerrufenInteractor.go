//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ZwanzigerrufenInteractorIF ツヴァンツィガールーフェンのインタラクターインタフェース。
type ZwanzigerrufenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ZwanzigerrufenConfig) string
	// Bid 入札する
	Bid(bid domain.ZwanzigerrufenBid) string
	// Pass パスする
	Pass() string
	// Discard 場札交換で 6 枚を伏せる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound 次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ZwanzigerrufenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ZwanzigerrufenInteractor ツヴァンツィガールーフェンのインタラクター。
type ZwanzigerrufenInteractor struct {
	GameBase[interfaces.ZwanzigerrufenGame]
	tp presenter.ZwanzigerrufenPresenter
}

// NewZwanzigerrufenInteractor コンストラクタ。
func NewZwanzigerrufenInteractor(g interfaces.ZwanzigerrufenGame, tp presenter.ZwanzigerrufenPresenter) *ZwanzigerrufenInteractor {
	mustNotNil("ZwanzigerrufenInteractor", map[string]any{"g": g, "tp": tp})
	return &ZwanzigerrufenInteractor{GameBase: GameBase[interfaces.ZwanzigerrufenGame]{Game: g}, tp: tp}
}

// Reset ゲーム初期化。
func (zi *ZwanzigerrufenInteractor) Reset() string {
	zi.Game.Reset()
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化。
func (zi *ZwanzigerrufenInteractor) ResetWithConfig(cfg domain.ZwanzigerrufenConfig) string {
	return resetWithValidatedConfig(zi.Game, zi.tp, cfg, zi.Game.SetConfig, zi.Reset)
}

// Bid 入札する。
func (zi *ZwanzigerrufenInteractor) Bid(bid domain.ZwanzigerrufenBid) string {
	if out, blocked := guardGameEnd(zi.Game, zi.tp); blocked {
		return out
	}
	if err := zi.Game.PlayerBid(bid); err != nil {
		return zi.tp.Output(zi.Game, err)
	}
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// Pass パスする。
func (zi *ZwanzigerrufenInteractor) Pass() string {
	if out, blocked := guardGameEnd(zi.Game, zi.tp); blocked {
		return out
	}
	if err := zi.Game.PlayerPass(); err != nil {
		return zi.tp.Output(zi.Game, err)
	}
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// Discard 場札交換で 6 枚を伏せる。
func (zi *ZwanzigerrufenInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardGameEnd(zi.Game, zi.tp); blocked {
		return out
	}
	if err := zi.Game.PlayerDiscard(cardIndices); err != nil {
		return zi.tp.Output(zi.Game, err)
	}
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// Play カードをプレイ。
func (zi *ZwanzigerrufenInteractor) Play(cardIndex int) string {
	if out, blocked := guardGameEnd(zi.Game, zi.tp); blocked {
		return out
	}
	if err := zi.Game.PlayerPlayCard(cardIndex); err != nil {
		return zi.tp.Output(zi.Game, err)
	}
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// NextTrick 次のトリックへ進む。
func (zi *ZwanzigerrufenInteractor) NextTrick() string {
	zi.Game.NextTrick()
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// NextRound 次のディールへ進む。
func (zi *ZwanzigerrufenInteractor) NextRound() string {
	zi.Game.NextRound()
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// GetConfig 現在の設定を取得。
func (zi *ZwanzigerrufenInteractor) GetConfig() domain.ZwanzigerrufenConfig {
	return zi.Game.GetConfig()
}

// Hint ヒント取得。
func (zi *ZwanzigerrufenInteractor) Hint() string { return zi.tp.HintOutput(zi.Game) }

// ActionLog 棋譜を出力する。
func (zi *ZwanzigerrufenInteractor) ActionLog() string { return zi.tp.ActionLogOutput(zi.Game) }

// zwanzigerrufenMaxCpuSteps advance の防御的な反復上限。
const zwanzigerrufenMaxCpuSteps = 500

// advance 人間の入力が必要になるまで CPU を進める。
//
// **トリック終了では止める。** 出揃った 4 枚を見せずに次を配ると、何が起きたのか
// 分からないまま盤面が変わる。次へ進めるのは人間の操作 (NextTrick)。
func (zi *ZwanzigerrufenInteractor) advance() {
	for range zwanzigerrufenMaxCpuSteps {
		if zi.Game.GetGameEndFlag() || zi.Game.IsHumanTurn() {
			return
		}
		switch zi.Game.GetPhase() {
		case domain.ZwanzigerrufenPhaseBid:
			zi.Game.CpuBid()
		case domain.ZwanzigerrufenPhaseTalon:
			zi.Game.CpuDiscard()
		case domain.ZwanzigerrufenPhasePlay:
			zi.Game.CpuPlayCard()
		default:
			// TrickEnd / RoundEnd / GameEnd は人間の操作待ち。
			return
		}
	}
}

// RestoreZwanzigerrufenInteractor deserialises JSON into a ZwanzigerrufenInteractor.
func RestoreZwanzigerrufenInteractor(data []byte, tp presenter.ZwanzigerrufenPresenter) (*ZwanzigerrufenInteractor, error) {
	return restoreAndBuild[domain.Zwanzigerrufen](data, func(g *domain.Zwanzigerrufen) *ZwanzigerrufenInteractor {
		return &ZwanzigerrufenInteractor{GameBase: GameBase[interfaces.ZwanzigerrufenGame]{Game: g}, tp: tp}
	})
}
