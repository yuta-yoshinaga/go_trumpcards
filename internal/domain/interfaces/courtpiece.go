//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CourtPieceGame Court Piece ゲームインタフェース
type CourtPieceGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDeclareTrump プレイヤー (呼び手) がトランプを宣言する
	PlayerDeclareTrump(suit int) error
	// CpuDeclareTrump CPU プレイヤー (呼び手) がトランプを宣言する
	CpuDeclareTrump()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CourtPieceConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CourtPieceConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CourtPiecePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanTrumpTurn 現在のトランプ宣言手番が人間かを返す
	IsHumanTrumpTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrumpSuit トランプスートを取得する (0 = 未宣言)
	GetTrumpSuit() int
	// GetCallerIdx 呼び手 (Hakim) インデックスを取得する
	GetCallerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetConsecutiveWins 同一チームの連続ラウンド勝利数を取得する
	GetConsecutiveWins() int
	// GetLastWinnerTeam 直前ラウンドの勝利チームを取得する (-1 = なし)
	GetLastWinnerTeam() int
	// IsLastRoundCourt 直前ラウンドが Court ボーナスだったかを返す
	IsLastRoundCourt() bool
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetWinnerTeam 勝利チームを取得する (-1 = 未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CourtPiecePlayer
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを取得する
	GetValidPlayIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.CourtPieceHint
}
