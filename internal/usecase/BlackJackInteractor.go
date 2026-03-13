package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BlackJackInteractorIF ブラックジャックインタラクターインタフェース
type BlackJackInteractorIF interface {
	Reset() string
	Hit() string
	Stand() string
	Bet(amount, ppBet, t3Bet, handCount int) string
	DoubleDown() string
	Split() string
	Insurance() string
	DeclineInsurance() string
	Surrender() string
	EarlySurrender() string
	DeclineEarlySurrender() string
	SetDeckCount(count int) string
	ToggleHint() string
	ToggleSoft17() string
	ToggleCounting() string
	ToggleDAS() string
	SetCountingSystem(system int) string
	SetDeckPenetration(penetration int) string
	SetCpuPlayerCount(count int) string
	SetSurrenderRule(rule int) string
	ResetWithConfig(cfg domain.BlackJackConfig) string
	ActionLog() string
}

// BlackJackInteractor ブラックジャックインタラクタークラス
type BlackJackInteractor struct {
	bj  interfaces.BlackJackGame
	bjp presenter.BlackJackPresenter
}

// NewBlackJackInteractor コンストラクタ
func NewBlackJackInteractor(bj interfaces.BlackJackGame, bjp presenter.BlackJackPresenter) *BlackJackInteractor {
	mustNotNil("BlackJackInteractor", map[string]any{"bj": bj, "bjp": bjp})
	return &BlackJackInteractor{
		bj:  bj,
		bjp: bjp,
	}
}

// Reset ゲーム初期化
func (bi *BlackJackInteractor) Reset() string {
	bi.bj.Reset()
	return bi.bjp.Output(bi.bj, nil)
}

// Hit ヒット
func (bi *BlackJackInteractor) Hit() string {
	err := bi.bj.PlayerHit()
	return bi.bjp.Output(bi.bj, err)
}

// Stand スタンド
func (bi *BlackJackInteractor) Stand() string {
	err := bi.bj.PlayerStand()
	return bi.bjp.Output(bi.bj, err)
}

// Bet ベット
func (bi *BlackJackInteractor) Bet(amount, ppBet, t3Bet, handCount int) string {
	err := bi.bj.PlayerBet(amount, ppBet, t3Bet, handCount)
	return bi.bjp.Output(bi.bj, err)
}

// DoubleDown ダブルダウン
func (bi *BlackJackInteractor) DoubleDown() string {
	err := bi.bj.PlayerDoubleDown()
	return bi.bjp.Output(bi.bj, err)
}

// Split スプリット
func (bi *BlackJackInteractor) Split() string {
	err := bi.bj.PlayerSplit()
	return bi.bjp.Output(bi.bj, err)
}

// Insurance インシュランス
func (bi *BlackJackInteractor) Insurance() string {
	err := bi.bj.PlayerInsurance()
	return bi.bjp.Output(bi.bj, err)
}

// DeclineInsurance インシュランス辞退
func (bi *BlackJackInteractor) DeclineInsurance() string {
	err := bi.bj.PlayerDeclineInsurance()
	return bi.bjp.Output(bi.bj, err)
}

// Surrender サレンダー
func (bi *BlackJackInteractor) Surrender() string {
	err := bi.bj.PlayerSurrender()
	return bi.bjp.Output(bi.bj, err)
}

// EarlySurrender アーリーサレンダー
func (bi *BlackJackInteractor) EarlySurrender() string {
	err := bi.bj.PlayerEarlySurrender()
	return bi.bjp.Output(bi.bj, err)
}

// DeclineEarlySurrender アーリーサレンダー辞退
func (bi *BlackJackInteractor) DeclineEarlySurrender() string {
	err := bi.bj.PlayerDeclineEarlySurrender()
	return bi.bjp.Output(bi.bj, err)
}

// SetDeckCount デッキ数設定
func (bi *BlackJackInteractor) SetDeckCount(count int) string {
	err := bi.bj.SetDeckCount(count)
	return bi.bjp.Output(bi.bj, err)
}

// ToggleHint ヒント表示切り替え
func (bi *BlackJackInteractor) ToggleHint() string {
	bi.bj.ToggleHint()
	return bi.bjp.Output(bi.bj, nil)
}

// ToggleSoft17 ソフト17ルール切り替え
func (bi *BlackJackInteractor) ToggleSoft17() string {
	config := bi.bj.GetConfig()
	config.DealerHitsSoft17 = !config.DealerHitsSoft17
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// ToggleCounting カウンティング表示切り替え
func (bi *BlackJackInteractor) ToggleCounting() string {
	config := bi.bj.GetConfig()
	config.CountingEnabled = !config.CountingEnabled
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// ToggleDAS スプリット後ダブルダウン許可切り替え
func (bi *BlackJackInteractor) ToggleDAS() string {
	config := bi.bj.GetConfig()
	config.DoubleAfterSplit = !config.DoubleAfterSplit
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// SetCountingSystem カウンティングシステム変更
func (bi *BlackJackInteractor) SetCountingSystem(system int) string {
	config := bi.bj.GetConfig()
	config.CountingSystem = system
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// SetDeckPenetration デッキペネトレーション率設定
func (bi *BlackJackInteractor) SetDeckPenetration(penetration int) string {
	config := bi.bj.GetConfig()
	config.DeckPenetration = penetration
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// SetCpuPlayerCount CPUプレイヤー数変更
func (bi *BlackJackInteractor) SetCpuPlayerCount(count int) string {
	config := bi.bj.GetConfig()
	config.CpuPlayerCount = count
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// SetSurrenderRule サレンダールール変更
func (bi *BlackJackInteractor) SetSurrenderRule(rule int) string {
	config := bi.bj.GetConfig()
	config.SurrenderRule = rule
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// ResetWithConfig 設定付きリセット
func (bi *BlackJackInteractor) ResetWithConfig(cfg domain.BlackJackConfig) string {
	bi.bj.Reset()
	err := bi.bj.SetConfig(cfg)
	if err != nil {
		return bi.bjp.Output(bi.bj, err)
	}
	bi.bj.Reset()
	return bi.bjp.Output(bi.bj, nil)
}

// ActionLog 棋譜を出力する
func (bi *BlackJackInteractor) ActionLog() string {
	return bi.bjp.ActionLogOutput(bi.bj)
}
