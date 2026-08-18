//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DoppelkopfGame ドッペルコップのゲームインタフェース
type DoppelkopfGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerAnnounce 人間プレイヤーが Re/Kontra を宣言する
	PlayerAnnounce() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DoppelkopfConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.DoppelkopfConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.DoppelkopfPhase
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
	// IsRe playerIdx が Re チームかを返す
	IsRe(playerIdx int) bool
	// IsSoloRe ソロ Re かどうかを返す
	IsSoloRe() bool
	// AreTeamsRevealed チームが公開済みかを返す
	AreTeamsRevealed() bool
	// IsReAnnounced Re 宣言済みかを返す
	IsReAnnounced() bool
	// IsKontraAnnounced Kontra 宣言済みかを返す
	IsKontraAnnounced() bool
	// CanHumanAnnounce 人間プレイヤーが今宣言できるかを返す
	CanHumanAnnounce() bool
	// GetRoundRePoints 直近ラウンドの Re チーム得点を取得する
	GetRoundRePoints() int
	// GetRoundReWon 直近ラウンドで Re が勝ったかを取得する
	GetRoundReWon() bool
	// GetRoundGamePoints 直近ラウンドのゲームポイントを取得する
	GetRoundGamePoints() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.DoppelkopfPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetTrumpIndices 指定プレイヤーの手札のうち切り札の位置を取得する
	GetTrumpIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.DoppelkopfHint
}
