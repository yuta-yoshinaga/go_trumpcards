//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// WattenInteractorIF ヴァッテンインタラクターインタフェース
type WattenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.WattenConfig) string
	// Declare 宣言 (Schlag ランク + 切り札スート)
	Declare(rank, suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// Raise ステークを引き上げる
	Raise() string
	// Respond レイズに応答する (true=hold / false=fold)
	Respond(hold bool) string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.WattenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// WattenInteractor ヴァッテンインタラクタークラス
type WattenInteractor struct {
	GameBase[interfaces.WattenGame]
	wp presenter.WattenPresenter
}

// NewWattenInteractor コンストラクタ
func NewWattenInteractor(g interfaces.WattenGame, wp presenter.WattenPresenter) *WattenInteractor {
	mustNotNil("WattenInteractor", map[string]any{"g": g, "wp": wp})
	return &WattenInteractor{GameBase: GameBase[interfaces.WattenGame]{Game: g}, wp: wp}
}

// Reset ゲーム初期化
func (wi *WattenInteractor) Reset() string {
	wi.Game.Reset()
	wi.runCpu()
	return wi.wp.Output(wi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (wi *WattenInteractor) ResetWithConfig(cfg domain.WattenConfig) string {
	return resetWithValidatedConfig(wi.Game, wi.wp, cfg, wi.Game.SetConfig, wi.Reset)
}

// Declare 宣言 (Schlag ランク + 切り札スート)
func (wi *WattenInteractor) Declare(rank, suit int) string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	if err := wi.Game.PlayerDeclare(rank, suit); err != nil {
		return wi.wp.Output(wi.Game, err)
	}
	wi.runCpu()
	return wi.wp.Output(wi.Game, nil)
}

// Play カードをプレイ
func (wi *WattenInteractor) Play(cardIndex int) string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	if wi.Game.GetPhase() != domain.WattenPhasePlay || !wi.Game.IsHumanTurn() {
		return wi.wp.Output(wi.Game, nil)
	}
	if err := wi.Game.PlayerPlay(cardIndex); err != nil {
		return wi.wp.Output(wi.Game, err)
	}
	wi.runCpu()
	return wi.wp.Output(wi.Game, nil)
}

// Raise ステークを引き上げる
func (wi *WattenInteractor) Raise() string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	if err := wi.Game.PlayerRaise(); err != nil {
		return wi.wp.Output(wi.Game, err)
	}
	wi.runCpu()
	return wi.wp.Output(wi.Game, nil)
}

// Respond レイズに応答する (true=hold / false=fold)
func (wi *WattenInteractor) Respond(hold bool) string {
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	if err := wi.Game.PlayerRespond(hold); err != nil {
		return wi.wp.Output(wi.Game, err)
	}
	wi.runCpu()
	return wi.wp.Output(wi.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (wi *WattenInteractor) NextRound() string {
	wi.Game.ScoreRound()
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	wi.Game.NextRound()
	wi.runCpu()
	return wi.wp.Output(wi.Game, nil)
}

// GetConfig 現在の設定を取得
func (wi *WattenInteractor) GetConfig() domain.WattenConfig {
	return wi.Game.GetConfig()
}

// Hint ヒント取得
func (wi *WattenInteractor) Hint() string {
	return wi.wp.HintOutput(wi.Game)
}

// ActionLog 棋譜を出力する
func (wi *WattenInteractor) ActionLog() string {
	return wi.wp.ActionLogOutput(wi.Game)
}

// runCpu 宣言 / プレイ / レイズ応答 / トリック解決を、人間の手番になるまで自動進行する。
// 生成的な runCpuTurnsLoop では Respond フェーズを扱えないため専用ループを持つ。
func (wi *WattenInteractor) runCpu() {
	g := wi.Game
	for i := 0; i < MaxCpuIterations; i++ {
		if g.GetGameEndFlag() {
			return
		}
		switch g.GetPhase() {
		case domain.WattenPhaseDeclare:
			if g.IsHumanDeclareTurn() {
				return
			}
			g.CpuDeclare()
		case domain.WattenPhasePlay:
			if g.IsHumanTurn() {
				return
			}
			g.CpuPlay()
		case domain.WattenPhaseRespond:
			if g.IsHumanRespondTurn() {
				return
			}
			g.CpuRespond()
		case domain.WattenPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		default: // RoundEnd / GameEnd
			return
		}
	}
}

// RestoreWattenInteractor deserialises JSON into a WattenInteractor.
func RestoreWattenInteractor(data []byte, wp presenter.WattenPresenter) (*WattenInteractor, error) {
	g, err := restoreGame[domain.Watten](data)
	if err != nil {
		return nil, err
	}
	return &WattenInteractor{GameBase: GameBase[interfaces.WattenGame]{Game: g}, wp: wp}, nil
}
