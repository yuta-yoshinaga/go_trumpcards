package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PinochleGame ピノクルゲームインタフェース
type PinochleGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間プレイヤーがビッドする
	PlayerBid(amount int) error
	// PlayerPass 人間プレイヤーがパスする
	PlayerPass() error
	// CpuBid CPUがビッドまたはパスする
	CpuBid()
	// PlayerCallTrump 人間プレイヤーがトランプスートを宣言する
	PlayerCallTrump(suit int) error
	// CpuCallTrump CPUがトランプを宣言する
	CpuCallTrump()
	// ConfirmMelds メルドを確認してプレイフェーズに進む
	ConfirmMelds()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUがカードをプレイする
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PinochleConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PinochleConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PinochlePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx ビッドプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int
	// GetHighestBid 最高ビッド額を取得する
	GetHighestBid() int
	// GetHighestBidder 最高ビッダーインデックスを取得する
	GetHighestBidder() int
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PinochlePlayer
	// GetPlayerMelds プレイヤーのメルドを取得する
	GetPlayerMelds() [domain.PinochlePlayerCnt][]*domain.PinochleMeld
	// GetHint ヒントを取得する
	GetHint() *domain.PinochleHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// SortHand 手札をソートする
	SortHand(playerIdx int)
}
