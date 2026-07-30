//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ToepenGame トゥーペンゲームインタフェース
type ToepenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayCard プレイヤーが手札の札を出す
	PlayCard(player, handIdx int) error
	// Toep 賭け点を吊り上げる
	Toep(player int) error
	// Respond toep に追随(true)か降参(false)かを答える
	Respond(player int, stay bool) error
	// Redeal 貧民の手札を捨てて配り直す
	Redeal(player int) error
	// CanRedeal 配り直しを要求できるかを返す
	CanRedeal(player int) bool
	// NextHand 次のハンドを開始する
	NextHand() error
	// ToepenCpuDecide CPU が取る手を決める
	ToepenCpuDecide(idx int) domain.ToepenCpuAction

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ToepenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ToepenConfig)

	// GetGameEndFlag 終局しているかを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ToepenPhase
	// GetCurrentPlayerIdx 手番のプレイヤー添字を取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リード側の添字を取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx 親の添字を取得する
	GetDealerIdx() int
	// GetCurrentTrick 場に出ている札を取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadSuit リードスートを取得する (未決は -1)
	GetLeadSuit() int
	// GetTrickNumber 完了したトリック数を取得する
	GetTrickNumber() int
	// GetStake 現在の賭け点を取得する
	GetStake() int
	// GetKnockerIdx toep を宣言した者を取得する (応答フェーズ外は -1)
	GetKnockerIdx() int
	// GetPendingRespondent 応答待ちのプレイヤーを取得する (無ければ -1)
	GetPendingRespondent() int
	// GetLastTrickWinner 最後にトリックを取った者を取得する
	GetLastTrickWinner() int
	// GetHandNumber 現在のハンド番号を取得する
	GetHandNumber() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.ToepenPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.ToepenPlayer
	// GetLives 累計失点を取得する
	GetLives(idx int) int
	// IsFolded このハンドから降りたかを取得する
	IsFolded(idx int) bool
	// IsEliminated 脱落したかを取得する
	IsEliminated(idx int) bool
	// GetValidPlayIndices 出せる手札の添字を取得する
	GetValidPlayIndices(player int) []int
	// GetWinnerIdx 勝者の添字を取得する (-1: 未確定)
	GetWinnerIdx() int
}
