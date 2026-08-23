//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BauernschnapsenGame バウエルンシュナプセンゲームインタフェース
type BauernschnapsenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// DeclareContract 席 playerIdx の契約を宣言する
	DeclareContract(playerIdx int, c domain.BauernschnapsenContract, trumpSuit int) error
	// CpuDeclareContract CPU 席の契約宣言を 1 手進める
	CpuDeclareContract()
	// IsHumanContractTurn 人間の契約宣言待ちかを返す
	IsHumanContractTurn() bool
	// GetContract 採用された契約を返す
	GetContract() domain.BauernschnapsenContract
	// GetDeclarerIdx 契約を宣言した席を返す (-1 = 未確定)
	GetDeclarerIdx() int
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerDeclareMarriage プレイヤーがマリアージュを宣言する
	PlayerDeclareMarriage(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BauernschnapsenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BauernschnapsenConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BauernschnapsenPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
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
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetRoundPoints 当ラウンドのチーム別カード点数を取得する
	GetRoundPoints(team int) int
	// GetRoundMarriagePoints 当ラウンドのマリアージュ得点を取得する
	GetRoundMarriagePoints(team int) int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BauernschnapsenPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.BauernschnapsenHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetMarriageIndices マリアージュ宣言可能なカードのインデックスリストを返す
	GetMarriageIndices(playerIdx int) []int
}
