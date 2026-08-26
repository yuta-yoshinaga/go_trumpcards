//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DehlaPakadGame はデーラ・パカドのゲームインタフェース。
type DehlaPakadGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand 次のハンドを開始する
	NextHand()
	// SelectTrump 切り札を宣言する
	SelectTrump(suit int) error
	// CpuSelectTrump CPU が切り札を宣言する
	CpuSelectTrump()
	// PlayerPlay 人間が 1 枚出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU が 1 枚出す
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DehlaPakadConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.DehlaPakadConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() string
	// IsHumanTurn 人間が決める番かを取得する
	IsHumanTurn() bool
	// GetHandNumber 現在のハンド番号を取得する
	GetHandNumber() int
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetTrumpChooserIdx 切り札を決める席を取得する
	GetTrumpChooserIdx() int
	// GetTrumpSuit 切り札スートを取得する (-1 = 未宣言)
	GetTrumpSuit() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTurn 手番の席を取得する (別名)
	GetCurrentTurn() int
	// GetLeadPlayerIdx リードの席を取得する
	GetLeadPlayerIdx() int
	// GetCurrentTrick 進行中のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLastTrick 直前に完了したトリックを取得する
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前トリックを取った席を取得する (-1 = なし)
	GetLastTrickWinner() int
	// GetPrevTrickWinner 2 連勝判定に使う直前の勝者を取得する
	GetPrevTrickWinner() int
	// GetCentrePile まだ誰も引き取っていない札を取得する
	GetCentrePile() []*domain.Card
	// GetCentrePileTens 中央の山にある 10 の枚数を取得する
	GetCentrePileTens() int
	// GetPlayableIndices いま出せる手札の位置を取得する
	GetPlayableIndices(playerIdx int) []int
	// GetTeamTens チーム別に取った 10 の枚数を取得する
	GetTeamTens() []int
	// GetTeamKots チーム別のコート数を取得する
	GetTeamKots() []int
	// GetStreakTeam 連勝中のチームを取得する (-1 = なし)
	GetStreakTeam() int
	// GetStreakCount 連勝数を取得する
	GetStreakCount() int
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定席のプレイヤーを取得する
	GetPlayer(i int) *domain.DehlaPakadPlayer
	// GetLastResult 直前ハンドの結果を取得する
	GetLastResult() *domain.DehlaPakadHandResult
	// GetHandHistory 完了した各ハンドの結果を古い順に取得する
	GetHandHistory() []*domain.DehlaPakadHandResult
	// GetWinnerTeam 勝ったチームを取得する (-1 = 未決)
	GetWinnerTeam() int
	// GetHint ヒントを取得する
	GetHint() *domain.DehlaPakadHint
}
