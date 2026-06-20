//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// OpenFaceChineseGame オープンフェイス・チャイニーズポーカー (OFC) のゲームインタフェース
type OpenFaceChineseGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlace 保留カードを指定段に置く (0=front,1=middle,2=back)
	PlayerPlace(row int) error
	// CpuPlay CPUプレイヤーが1枚配置する
	CpuPlay()
	// ScoreRound 全段が埋まっていればラウンドを採点する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.OpenFaceChineseConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.OpenFaceChineseConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.OpenFaceChinesePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetWinnerIdx 勝者インデックスを取得する (-1=未確定/引き分け)
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.OpenFaceChinesePlayer
	// GetCurrentCard 現在の手番プレイヤーが置こうとしている保留カードを取得する
	GetCurrentCard() *domain.Card
	// GetHint ヒントを取得する
	GetHint() *domain.OpenFaceChineseHint
}
