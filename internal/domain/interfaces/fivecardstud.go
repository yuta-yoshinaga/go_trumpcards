//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FiveCardStudGame ファイブカードスタッドゲームインタフェース
type FiveCardStudGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset() error
	// PlayerAction プレイヤーのベッティングアクションを実行する
	PlayerAction(action, amount, humanPlayMs int) error
	// GetIsSoko は Soko (Canadian Stud) モードかを返す。
	// プレゼンターは役名の表を切り替えるためにこれを見る必要がある:
	// Soko のランクスケールで PokerHandNames を引くと別の役名が出る。
	GetIsSoko() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetPlayers プレイヤー一覧を取得する
	GetPlayers() []*domain.FiveCardStudPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.FiveCardStudPlayer
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetCommunityCard 共有カードを取得する (カード不足時のみ)
	GetCommunityCard() *domain.Card
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
	GetRoundResults() []domain.FiveCardStudResult
	// GetCpuActions CPU行動記録を取得する
	GetCpuActions() []domain.FiveCardStudCpuAction
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.FiveCardStudConfig
	// SetConfig ゲーム設定を変更する
	SetConfig(cfg domain.FiveCardStudConfig)
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetActedFlags 各プレイヤーのアクション済みフラグを取得する
	GetActedFlags() []bool
	// GetHandCount ハンド数を取得する
	GetHandCount() int
	// Resize プレイヤー数を変更する
	Resize(players []*domain.FiveCardStudPlayer)
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
