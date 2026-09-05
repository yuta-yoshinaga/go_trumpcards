//go:build !js || !wasm || extra5

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MarjapussiGame マルヤプッシ (Marjapussi) のゲームインタフェース
type MarjapussiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MarjapussiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MarjapussiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MarjapussiPhase
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
	// GetTrumpSuit 切り札スートを取得する (0=未設定)
	GetTrumpSuit() int
	// GetPussi ベリー袋 (pussi) のカード一覧を取得する
	GetPussi() []*domain.Card
	// GetTeamScores チーム別累積点を取得する
	GetTeamScores() [domain.MarjapussiTeamCnt]int
	// GetPlayerScores プレイヤー別累積点を取得する
	GetPlayerScores() [domain.MarjapussiPlayerCnt]int
	// GetRoundCardPoints 現ラウンドのチーム別カード得点を取得する
	GetRoundCardPoints() [domain.MarjapussiTeamCnt]int
	// GetRoundMarriage 現ラウンドのチーム別結婚点を取得する
	GetRoundMarriage() [domain.MarjapussiTeamCnt]int
	// GetMarriageOptions いま結婚を宣言できるスートとその点を取得する
	GetMarriageOptions(playerIdx int) []domain.MarjapussiMarriageOption
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetWinnerTeam 勝利チームを取得する (-1=未確定)
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MarjapussiPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.MarjapussiHint
}
