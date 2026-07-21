//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TeenPattiGame ティーン・パティのゲームインタフェース
type TeenPattiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerSee 手札を見て Seen に昇格する
	PlayerSee() error
	// PlayerBet 現在の賭け単位をコールする
	PlayerBet() error
	// PlayerRaise 賭け単位を newStake へ引き上げる
	PlayerRaise(newStake int) error
	// PlayerFold 降りる
	PlayerFold() error
	// PlayerShow 残り 2 人のときに勝負を要求する
	PlayerShow() error
	// PlayerRequestSideShow 直前の Seen プレイヤーにサイドショーを申請する
	PlayerRequestSideShow() error
	// PlayerRespondSideShow サイドショー申請に応答する (accept=受諾)
	PlayerRespondSideShow(accept bool) error
	// CpuAct CPU プレイヤーが 1 アクション実行する (Betting / SideShow 両方を処理)
	CpuAct()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TeenPattiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TeenPattiConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TeenPattiPhase
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在の手番プレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetPot 現在のポットを取得する
	GetPot() int
	// GetStake 現在の賭け単位を取得する
	GetStake() int
	// GetRoundWinnerIdx 直近ディールの勝者を取得する (-1=未確定)
	GetRoundWinnerIdx() int
	// IsShowdown ショーダウンが行われたかを返す
	IsShowdown() bool
	// CanRequestSideShow 現在の手番プレイヤーがサイドショーを申請できるかを返す
	CanRequestSideShow() bool
	// GetSideShowRequester サイドショー申請者インデックスを取得する (-1=なし)
	GetSideShowRequester() int
	// GetSideShowTarget サイドショー対象者インデックスを取得する (-1=なし)
	GetSideShowTarget() int
	// GetLastSideShow 直近で成立したサイドショーの申請者・対象・敗者を返す (ok=false=結果なし)
	GetLastSideShow() (requester, target, loser int, ok bool)
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetMatchWinnerIdx 試合の勝者を取得する (-1=未確定)
	GetMatchWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.TeenPattiPlayer
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// CanShow 現在の手番プレイヤーが Show を要求できるかを返す
	CanShow() bool
	// GetHint ヒントを取得する
	GetHint() *domain.TeenPattiHint
}
