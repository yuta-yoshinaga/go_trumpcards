//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KoenigrufenGame ケーニッヒルーフェン (Königrufen) のゲームインタフェース
type KoenigrufenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間プレイヤーが入札する
	PlayerBid(bid domain.KoenigrufenBid) error
	// PlayerPass 人間プレイヤーがパスする
	PlayerPass() error
	// CpuBid CPU プレイヤーが 1 回入札する
	CpuBid()
	// PlayerCallKing 人間デクレアラーが呼ぶキングのスートを指名する
	PlayerCallKing(suit int) error
	// CpuCallKing CPU デクレアラーが呼ぶキングを自動選択する
	CpuCallKing()
	// PlayerDiscard 人間デクレアラーが場札交換で 6 枚を捨てる
	PlayerDiscard(cardIndices []int) error
	// CpuDiscard CPU デクレアラーが場札交換で 6 枚を捨てる
	CpuDiscard()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KoenigrufenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KoenigrufenConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KoenigrufenPhase
	// IsHumanTurn 現在のプレイ手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在の入札手番が人間かを返す
	IsHumanBidTurn() bool
	// IsHumanCallTurn 現在の王呼び手番が人間かを返す
	IsHumanCallTurn() bool
	// IsHumanDiscardTurn 現在の場札交換手番が人間かを返す
	IsHumanDiscardTurn() bool
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
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetBidPlayerIdx 入札手番インデックスを取得する
	GetBidPlayerIdx() int
	// GetHighestBid 現在の最高入札を取得する
	GetHighestBid() domain.KoenigrufenBid
	// GetHighestBidder 最高入札者を取得する (-1=なし)
	GetHighestBidder() int
	// GetDeclarerIdx デクレアラーインデックスを取得する (-1=未確定)
	GetDeclarerIdx() int
	// GetContract コントラクト (確定入札) を取得する
	GetContract() domain.KoenigrufenBid
	// GetCalledKing 呼ばれたキングのスートを取得する (-1=未呼び/単独)
	GetCalledKing() int
	// GetPartnerIdx 秘密のパートナーインデックスを取得する (サーバー内部専用)
	GetPartnerIdx() int
	// GetPartnerRevealed パートナーが公開済みかを取得する
	GetPartnerRevealed() bool
	// GetTalonCount 場札の枚数を取得する
	GetTalonCount() int
	// GetTalon 場札を取得する
	GetTalon() []*domain.Card
	// GetStashOwner stash の所有側を取得する (常に 0=デクレアラー側)
	GetStashOwner() int
	// GetPlayerScores プレイヤー別累積得点を取得する
	GetPlayerScores() [domain.KoenigrufenPlayerCnt]int
	// GetCardPoints プレイヤー i の獲得カードポイントを取得する
	GetCardPoints(i int) int
	// GetOutcome 直近ディールの結果を取得する
	GetOutcome() domain.KoenigrufenOutcome
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.KoenigrufenResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.KoenigrufenPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.KoenigrufenHint
}
