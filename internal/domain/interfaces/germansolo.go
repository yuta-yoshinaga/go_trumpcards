//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GermanSoloGame ジャーマン・ソロ (GermanSolo) のゲームインタフェース
type GermanSoloGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間がビッドする (pass/frage/solo/tout + 切り札スート)
	PlayerBid(bid domain.GermanSoloBid, trumpSuit int) error
	// CpuBid CPUプレイヤーが1回ビッドする
	CpuBid()
	// DeclareAce 落札者が味方を呼ぶエースを指名する
	DeclareAce(playerIdx, suit int) error
	// CpuDeclareAce CPU の落札者にエースを呼ばせる
	CpuDeclareAce()
	// IsHumanAceCallTurn 人間のエース呼び待ちかを返す
	IsHumanAceCallTurn() bool
	// GetCalledAceSuit 呼ばれたエースのスート (-1=未指名)
	GetCalledAceSuit() int
	// GetPartnerIdx 味方の席 (-1=まだ伏せられている / 単独)
	GetPartnerIdx() int
	// IsPlayingAlone 落札者が単独で戦っているか
	IsPlayingAlone() bool
	// GetCallableAceSuits 落札者が呼べるエースのスート (画面の選択肢)
	GetCallableAceSuits() []int
	// GetSideTrickCounts (落札者側, 相手側) の獲得トリック数
	GetSideTrickCounts() (int, int)
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.GermanSoloConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.GermanSoloConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GermanSoloPhase
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
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetForehandIdx forehand インデックスを取得する
	GetForehandIdx() int
	// GetDeclarerIdx ジャーマン・ソロインデックスを取得する (-1=未確定)
	GetDeclarerIdx() int
	// GetWinningBid 確定ビッドを取得する
	GetWinningBid() domain.GermanSoloBid
	// GetHighestBid 競り中の最高宣言を取得する
	GetHighestBid() domain.GermanSoloBid
	// GetBiddableBids 今の競り状況で宣言できるビッドを取得する (画面の選択肢)
	GetBiddableBids() []int
	// RequiredTricks 確定した契約の成功に必要なトリック数を取得する
	RequiredTricks() int
	// GetDeclarerSideSize 宣言側の人数を取得する (単独=1, 味方あり=2)
	GetDeclarerSideSize() int
	// GetTrumpSuit 切り札スートを取得する (-1=未確定, 1..4)
	GetTrumpSuit() int
	// GetCurrentBidderIdx 現在のビッド手番インデックスを取得する
	GetCurrentBidderIdx() int
	// GetPlayerScores プレイヤー別累積点を取得する
	GetPlayerScores() [domain.GermanSoloPlayerCnt]int
	// GetOutcome 直近ディールの結果を取得する
	GetOutcome() domain.GermanSoloOutcome
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.GermanSoloResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.GermanSoloPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.GermanSoloHint
}
