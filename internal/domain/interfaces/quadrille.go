//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// QuadrilleGame カドリール (Quadrille) のゲームインタフェース
type QuadrilleGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間がビッドする (pass/entrar/solo + 切り札スート)
	PlayerBid(bid domain.QuadrilleBid, trumpSuit int) error
	// CpuBid CPUプレイヤーが1回ビッドする
	CpuBid()
	// DeclareKing 落札者が味方を呼ぶ王を指名する
	DeclareKing(playerIdx, suit int) error
	// CpuDeclareKing CPU の落札者に王を呼ばせる
	CpuDeclareKing()
	// IsHumanKingCallTurn 人間の王呼び待ちかを返す
	IsHumanKingCallTurn() bool
	// GetCalledKingSuit 呼ばれた王のスート (-1=未指名)
	GetCalledKingSuit() int
	// GetPartnerIdx 味方の席 (-1=まだ伏せられている / 単独)
	GetPartnerIdx() int
	// IsRoiSeul 落札者が単独で戦っているか
	IsRoiSeul() bool
	// GetCallableKingSuits 落札者が呼べる王のスート (画面の選択肢)
	GetCallableKingSuits() []int
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
	GetConfig() domain.QuadrilleConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.QuadrilleConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.QuadrillePhase
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
	// GetQuadrilleIdx カドリールインデックスを取得する (-1=未確定)
	GetQuadrilleIdx() int
	// GetWinningBid 確定ビッドを取得する
	GetWinningBid() domain.QuadrilleBid
	// GetHighestBid 競り中の最高宣言を取得する
	GetHighestBid() domain.QuadrilleBid
	// GetTrumpSuit 切り札スートを取得する (-1=未確定, 1..4)
	GetTrumpSuit() int
	// GetCurrentBidderIdx 現在のビッド手番インデックスを取得する
	GetCurrentBidderIdx() int
	// GetPlayerScores プレイヤー別累積点を取得する
	GetPlayerScores() [domain.QuadrillePlayerCnt]int
	// GetOutcome 直近ディールの結果を取得する
	GetOutcome() domain.QuadrilleOutcome
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.QuadrilleResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.QuadrillePlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.QuadrilleHint
}
