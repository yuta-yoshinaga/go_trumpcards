//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KarnoffelGame カルニッフェル (Karnöffel) ゲームインタフェース
type KarnoffelGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次の局を配る
	NextHand() error
	// PlayCard 手札を1枚出す
	PlayCard(player, idx int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// KarnoffelValidPlays 出せる手札インデックスを返す
	KarnoffelValidPlays(player int) []int
	// KarnoffelTeamTricks チームが取ったトリック数を返す
	KarnoffelTeamTricks(team int) int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KarnoffelConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KarnoffelConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KarnoffelPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx 親を取得する
	GetDealerIdx() int
	// GetChosenSuit 選ばれたスートを取得する
	GetChosenSuit() int
	// GetUpCard 席に表向きで配られた札を取得する
	GetUpCard(idx int) *domain.Card
	// GetTrick 場に出ている札を取得する
	GetTrick() []*domain.Card
	// GetTrickLeaderIdx このトリックのリード席を取得する
	GetTrickLeaderIdx() int
	// GetTrickNumber 済んだトリック数を取得する
	GetTrickNumber() int
	// GetTricksWon 席が取ったトリック数を取得する
	GetTricksWon(idx int) int
	// GetHandsWon チームが取った局数を取得する
	GetHandsWon(team int) int
	// GetLastResult 直前の局の結果を取得する
	GetLastResult() *domain.KarnoffelHandResult
	// GetHandNumber 現在の局番号を取得する
	GetHandNumber() int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.KarnoffelPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.KarnoffelPlayer
}
