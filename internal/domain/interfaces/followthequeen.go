//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FollowTheQueenGame フォロー・ザ・クイーンゲームインタフェース
type FollowTheQueenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset() error
	// PlayerAction プレイヤーのベッティングアクションを実行する
	PlayerAction(action, amount, humanPlayMs int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetPlayers プレイヤー一覧を取得する
	GetPlayers() []*domain.FollowTheQueenPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.FollowTheQueenPlayer
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetCommunityCard 共有カードを取得する (カード不足時のみ)
	GetCommunityCard() *domain.Card
	// GetWildRank は現在ワイルドになっているランクを返す。0 はワイルド無し。
	GetWildRank() int
	// IsWild はそのカードが現在ワイルドかを返す。**判定はここ 1 か所** ——
	// クイーンが常時ワイルドである点を presenter 側で書き直すと二重定義になる。
	IsWild(card *domain.Card) bool
	// GetPot ポット額を取得する
	GetPot() int
	// GetSidePots サイドポット一覧を取得する
	GetSidePots() []domain.SidePot
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTurn 現在のターンを取得する
	GetCurrentTurn() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetLastBet 最後のベット額を取得する
	GetLastBet() int
	// GetMinRaise 最小レイズ額を取得する
	GetMinRaise() int
	// GetRaiseCount 現在のレイズ回数を取得する
	GetRaiseCount() int
	// GetRoundResults ラウンド結果を取得する
	GetRoundResults() []domain.FollowTheQueenResult
	// GetCpuActions CPU行動記録を取得する
	GetCpuActions() []domain.FollowTheQueenCpuAction
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.FollowTheQueenConfig
	// SetConfig ゲーム設定を変更する
	SetConfig(cfg domain.FollowTheQueenConfig)
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetActedFlags 各プレイヤーのアクション済みフラグを取得する
	GetActedFlags() []bool
	// GetHandCount ハンド数を取得する
	GetHandCount() int
	// Resize プレイヤー数を変更する
	Resize(players []*domain.FollowTheQueenPlayer)
	TournamentActionGame
	// GetHumanProfile メタAIプロファイルを取得する
	GetHumanProfile() *domain.BettingHumanProfile
	// ResetProfile メタAIプロファイルをリセットする
	ResetProfile()
	// ExportProfile メタAIプロファイルをエクスポートする
	ExportProfile() interface{}
	// ImportProfile JSONバイトからメタAIプロファイルをインポートする
	ImportProfile(data []byte) error
	// GetBringInPlayerIdx ブリングインプレイヤーインデックスを取得する
	GetBringInPlayerIdx() int
}
