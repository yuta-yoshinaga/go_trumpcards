//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CegoGame チェゴ (Cego) のゲームインタフェース
type CegoGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間プレイヤーが入札する
	PlayerBid(bid domain.CegoBid) error
	// PlayerPass 人間プレイヤーがパスする
	PlayerPass() error
	// CpuBid CPU プレイヤーが 1 回入札する
	CpuBid()
	// PlayerChooseContract 人間デクレアラーがコントラクトを選ぶ
	PlayerChooseContract(ct domain.CegoContract) error
	// CpuChooseContract CPU デクレアラーがコントラクトを自動選択する
	CpuChooseContract()
	// PlayerDiscard 人間デクレアラーが Cego 交換で残す 1 枚を選ぶ
	PlayerDiscard(keepIndices []int) error
	// CpuDiscard CPU デクレアラーが Cego 交換を自動で行う
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
	GetConfig() domain.CegoConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CegoConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CegoPhase
	// IsHumanTurn 現在のプレイ手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在の入札手番が人間かを返す
	IsHumanBidTurn() bool
	// IsHumanContractTurn 現在のコントラクト選択手番が人間かを返す
	IsHumanContractTurn() bool
	// IsHumanExchangeTurn 現在の場札交換手番が人間かを返す
	IsHumanExchangeTurn() bool
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
	GetHighestBid() domain.CegoBid
	// GetHighestBidder 最高入札者を取得する (-1=なし)
	GetHighestBidder() int
	// GetDeclarerIdx デクレアラーインデックスを取得する (-1=未確定)
	GetDeclarerIdx() int
	// GetContract コントラクト (確定入札) を取得する
	GetContract() domain.CegoBid
	// GetContractType コントラクト種別 (Cego / Handspiel) を取得する
	GetContractType() domain.CegoContract
	// GetBlindCount 場札 (Cego / blind) の枚数を取得する
	GetBlindCount() int
	// GetStashOwner stash の所有側を取得する (0=デクレアラー側, 1=対戦側)
	GetStashOwner() int
	// GetPlayerScores プレイヤー別累積得点を取得する
	GetPlayerScores() [domain.CegoPlayerCnt]int
	// GetCardPoints プレイヤー i の獲得カードポイントを取得する
	GetCardPoints(i int) int
	// GetOutcome 直近ディールの結果を取得する
	GetOutcome() domain.CegoOutcome
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.CegoResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CegoPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.CegoHint
}
