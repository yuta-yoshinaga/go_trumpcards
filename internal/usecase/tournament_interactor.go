package usecase

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TournamentInteractorIF トーナメントアクション共通インタラクターインタフェース
// リバイ・アドオン・マック・ハンド公開の委譲メソッドを定義する。
type TournamentInteractorIF interface {
	// Rebuy リバイ実行
	Rebuy() string
	// SkipRebuy リバイ辞退
	SkipRebuy() string
	// Addon アドオン実行
	Addon() string
	// SkipAddon アドオン辞退
	SkipAddon() string
	// Muck マック (ハンドを伏せる)
	Muck() string
	// ShowHand ハンドを公開する
	ShowHand() string
}

// tournamentActions はトーナメントアクションの共通委譲ロジックを提供する。
// 各ゲームのインタラクター構造体に埋め込んで使用する。
type tournamentActions[G interfaces.TournamentActionGame] struct {
	game G
	pres outputPresenter[G]
}

// newTournamentActions コンストラクタ
func newTournamentActions[G interfaces.TournamentActionGame](game G, pres outputPresenter[G]) tournamentActions[G] {
	return tournamentActions[G]{game: game, pres: pres}
}

// Rebuy リバイ実行
func (t tournamentActions[G]) Rebuy() string {
	return execAndPresent(t.game, t.pres, t.game.Rebuy)
}

// SkipRebuy リバイ辞退
func (t tournamentActions[G]) SkipRebuy() string {
	return execAndPresent(t.game, t.pres, t.game.SkipRebuy)
}

// Addon アドオン実行
func (t tournamentActions[G]) Addon() string {
	return execAndPresent(t.game, t.pres, t.game.Addon)
}

// SkipAddon アドオン辞退
func (t tournamentActions[G]) SkipAddon() string {
	return execAndPresent(t.game, t.pres, t.game.SkipAddon)
}

// Muck マック (ハンドを伏せる)
func (t tournamentActions[G]) Muck() string {
	return execAndPresent(t.game, t.pres, t.game.Muck)
}

// ShowHand ハンドを公開する
func (t tournamentActions[G]) ShowHand() string {
	return execAndPresent(t.game, t.pres, t.game.ShowHand)
}
