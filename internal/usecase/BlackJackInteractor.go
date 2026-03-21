package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BlackJackInteractorIF ブラックジャックインタラクターインタフェース
type BlackJackInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Hit ヒット
	Hit() string
	// Stand スタンド
	Stand() string
	// Bet ベット
	Bet(amount, ppBet, t3Bet, handCount int) string
	// DoubleDown ダブルダウン
	DoubleDown() string
	// Split スプリット
	Split() string
	// Insurance インシュランス
	Insurance() string
	// DeclineInsurance インシュランス辞退
	DeclineInsurance() string
	// Surrender サレンダー
	Surrender() string
	// EarlySurrender アーリーサレンダー
	EarlySurrender() string
	// DeclineEarlySurrender アーリーサレンダー辞退
	DeclineEarlySurrender() string
	// SetDeckCount デッキ数設定
	SetDeckCount(count int) string
	// ToggleHint ヒント表示切り替え
	ToggleHint() string
	// ToggleSoft17 ソフト17ルール切り替え
	ToggleSoft17() string
	// ToggleCounting カウンティング表示切り替え
	ToggleCounting() string
	// ToggleDAS スプリット後ダブルダウン許可切り替え
	ToggleDAS() string
	// SetCountingSystem カウンティングシステム変更
	SetCountingSystem(system int) string
	// SetDeckPenetration デッキペネトレーション率設定
	SetDeckPenetration(penetration int) string
	// SetCpuPlayerCount CPUプレイヤー数変更
	SetCpuPlayerCount(count int) string
	// SetSurrenderRule サレンダールール変更
	SetSurrenderRule(rule int) string
	// ResetWithConfig 設定付きリセット
	ResetWithConfig(cfg domain.BlackJackConfig) string
	// ActionLog 棋譜を出力する
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
	return runAndPresent(bi.bj, bi.bjp, bi.bj.Reset)
}

// Hit ヒット
func (bi *BlackJackInteractor) Hit() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerHit)
}

// Stand スタンド
func (bi *BlackJackInteractor) Stand() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerStand)
}

// Bet ベット
func (bi *BlackJackInteractor) Bet(amount, ppBet, t3Bet, handCount int) string {
	return execAndPresent(bi.bj, bi.bjp, func() error { return bi.bj.PlayerBet(amount, ppBet, t3Bet, handCount) })
}

// DoubleDown ダブルダウン
func (bi *BlackJackInteractor) DoubleDown() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerDoubleDown)
}

// Split スプリット
func (bi *BlackJackInteractor) Split() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerSplit)
}

// Insurance インシュランス
func (bi *BlackJackInteractor) Insurance() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerInsurance)
}

// DeclineInsurance インシュランス辞退
func (bi *BlackJackInteractor) DeclineInsurance() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerDeclineInsurance)
}

// Surrender サレンダー
func (bi *BlackJackInteractor) Surrender() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerSurrender)
}

// EarlySurrender アーリーサレンダー
func (bi *BlackJackInteractor) EarlySurrender() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerEarlySurrender)
}

// DeclineEarlySurrender アーリーサレンダー辞退
func (bi *BlackJackInteractor) DeclineEarlySurrender() string {
	return execAndPresent(bi.bj, bi.bjp, bi.bj.PlayerDeclineEarlySurrender)
}

// SetDeckCount デッキ数設定
func (bi *BlackJackInteractor) SetDeckCount(count int) string {
	return execAndPresent(bi.bj, bi.bjp, func() error { return bi.bj.SetDeckCount(count) })
}

// ToggleHint ヒント表示切り替え
func (bi *BlackJackInteractor) ToggleHint() string {
	return runAndPresent(bi.bj, bi.bjp, bi.bj.ToggleHint)
}

// applyConfig は設定の取得・変更・保存・表示を行う共通ヘルパー
func (bi *BlackJackInteractor) applyConfig(modify func(*domain.BlackJackConfig)) string {
	config := bi.bj.GetConfig()
	modify(&config)
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// ToggleSoft17 ソフト17ルール切り替え
func (bi *BlackJackInteractor) ToggleSoft17() string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.DealerHitsSoft17 = !c.DealerHitsSoft17 })
}

// ToggleCounting カウンティング表示切り替え
func (bi *BlackJackInteractor) ToggleCounting() string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.CountingEnabled = !c.CountingEnabled })
}

// ToggleDAS スプリット後ダブルダウン許可切り替え
func (bi *BlackJackInteractor) ToggleDAS() string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.DoubleAfterSplit = !c.DoubleAfterSplit })
}

// SetCountingSystem カウンティングシステム変更
func (bi *BlackJackInteractor) SetCountingSystem(system int) string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.CountingSystem = system })
}

// SetDeckPenetration デッキペネトレーション率設定
func (bi *BlackJackInteractor) SetDeckPenetration(penetration int) string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.DeckPenetration = penetration })
}

// SetCpuPlayerCount CPUプレイヤー数変更
func (bi *BlackJackInteractor) SetCpuPlayerCount(count int) string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.CpuPlayerCount = count })
}

// SetSurrenderRule サレンダールール変更
func (bi *BlackJackInteractor) SetSurrenderRule(rule int) string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.SurrenderRule = rule })
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
