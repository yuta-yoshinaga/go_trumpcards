//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TeenPattiWebInput ティーン・パティのWebインプット
type TeenPattiWebInput struct {
	BaseWebInput
	// RaiseStake raise コマンドで引き上げる賭け単位
	RaiseStake *int `json:"raiseStake,omitempty"`
	// Accept respond コマンドでサイドショーを受諾するか
	Accept *bool `json:"accept,omitempty"`
	// Config ゲーム設定
	Config *TeenPattiWebConfig `json:"config,omitempty"`
}

// TeenPattiWebConfig ティーン・パティのWeb設定
type TeenPattiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	Ante          *int `json:"ante,omitempty"`
	StartingChips *int `json:"startingChips,omitempty"`
}

// TeenPattiWebOutputPlayer ティーン・パティのWebアウトプットプレイヤー
type TeenPattiWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Chips     int              `json:"chips"`
	Seen      bool             `json:"seen"`
	Folded    bool             `json:"folded"`
	Out       bool             `json:"out"`
	RoundBet  int              `json:"roundBet"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// HandName ショーダウンで公開された手の役名 (非公開時は空文字)
	HandName string `json:"handName,omitempty"`
}

// TeenPattiWebOutputHint ヒント出力
type TeenPattiWebOutputHint struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// TeenPattiWebOutputSideShowHand サイドショー参加者 1 人分の公開手札
type TeenPattiWebOutputSideShowHand struct {
	// PlayerIdx 参加者の席インデックス
	PlayerIdx int `json:"playerIdx"`
	// HandName 役名 i18n キー
	HandName string `json:"handName"`
	// Cards 公開された 3 枚のカード
	Cards []*WebOutputCard `json:"cards"`
}

// TeenPattiWebOutputSideShow 直近で成立したサイドショーの比較結果 (人間が当事者のときのみ設定)
type TeenPattiWebOutputSideShow struct {
	// RequesterIdx サイドショー申請者の席インデックス
	RequesterIdx int `json:"requesterIdx"`
	// TargetIdx サイドショー対象者の席インデックス
	TargetIdx int `json:"targetIdx"`
	// WinnerIdx 手比べの勝者インデックス
	WinnerIdx int `json:"winnerIdx"`
	// LoserIdx 手比べの敗者 (フォールドした) インデックス
	LoserIdx int `json:"loserIdx"`
	// Requester 申請者の公開手札
	Requester *TeenPattiWebOutputSideShowHand `json:"requester"`
	// Target 対象者の公開手札
	Target *TeenPattiWebOutputSideShowHand `json:"target"`
}

// TeenPattiWebOutputConfig ティーン・パティの設定アウトプット
type TeenPattiWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	Ante          int `json:"ante"`
	StartingChips int `json:"startingChips"`
}

// TeenPattiWebOutput ティーン・パティのWebアウトプット
type TeenPattiWebOutput struct {
	Players            []*TeenPattiWebOutputPlayer `json:"players"`
	Pot                int                         `json:"pot"`
	Stake              int                         `json:"stake"`
	Phase              int                         `json:"phase"`
	RoundNumber        int                         `json:"roundNumber"`
	DealerIdx          int                         `json:"dealerIdx"`
	CurrentPlayerIdx   int                         `json:"currentPlayerIdx"`
	RoundWinnerIdx     int                         `json:"roundWinnerIdx"`
	MatchWinnerIdx     int                         `json:"matchWinnerIdx"`
	IsShowdown         bool                        `json:"isShowdown"`
	CanShow            bool                        `json:"canShow"`
	CanRequestSideShow bool                        `json:"canRequestSideShow"`
	SideShowRequester  int                         `json:"sideShowRequester"`
	SideShowTarget     int                         `json:"sideShowTarget"`
	GameEndFlag        bool                        `json:"gameEndFlag"`
	IsHumanTurn        bool                        `json:"isHumanTurn"`
	Hint               *TeenPattiWebOutputHint     `json:"hint,omitempty"`
	LastSideShow       *TeenPattiWebOutputSideShow `json:"lastSideShow,omitempty"`
	Config             TeenPattiWebOutputConfig    `json:"config"`
	WebOutputBase
}

// ToConfig builds a TeenPattiConfig from the nested web config, applying bounds checking.
func (c *TeenPattiWebConfig) ToConfig() domain.TeenPattiConfig {
	cfg := domain.DefaultTeenPattiConfig()
	cfg.CpuDifficulty = domain.TeenPattiCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TeenPattiCpuDifficultyEasy), int(domain.TeenPattiCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, 1, 1000)
	webutil.ApplyBoundedInt(&cfg.StartingChips, c.StartingChips, 2, domain.TeenPattiMaxStartingChips)
	return cfg
}

// ToConfig builds a TeenPattiConfig from the web input.
func (p TeenPattiWebInput) ToConfig() domain.TeenPattiConfig {
	return configOrDefault(p.Config, (*TeenPattiWebConfig).ToConfig, domain.DefaultTeenPattiConfig())
}

// TeenPattiWebController ティーン・パティのWebコントローラークラス
type TeenPattiWebController = GameWebController[usecase.TeenPattiInteractorIF, TeenPattiWebInput, *TeenPattiWebOutput]

// NewTeenPattiWebController and NewTeenPattiWebControllerWithProvider are
// the standard and provider-backed constructors for TeenPattiWebController.
var NewTeenPattiWebController, NewTeenPattiWebControllerWithProvider = webControllerPair[usecase.TeenPattiInteractorIF, TeenPattiWebInput, *TeenPattiWebOutput](
	newTeenPattiDefaultOutput, teenPattiDispatch,
)

func newTeenPattiDefaultOutput(msg string) *TeenPattiWebOutput {
	return &TeenPattiWebOutput{
		Players:           make([]*TeenPattiWebOutputPlayer, 0),
		RoundWinnerIdx:    -1,
		MatchWinnerIdx:    -1,
		SideShowRequester: -1,
		SideShowTarget:    -1,
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

func teenPattiDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TeenPattiInteractorIF, param TeenPattiWebInput, newDefault func(string) *TeenPattiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "see":
		bc.writePresenterResponse(w, ti.See())
	case "bet":
		bc.writePresenterResponse(w, ti.Bet())
	case "raise":
		if !requireParam(bc, w, newDefault, param.RaiseStake == nil, "param error: raiseStake is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Raise(*param.RaiseStake))
	case "fold":
		bc.writePresenterResponse(w, ti.Fold())
	case "show":
		bc.writePresenterResponse(w, ti.Show())
	case "sideshow":
		bc.writePresenterResponse(w, ti.RequestSideShow())
	case "respond":
		if !requireParam(bc, w, newDefault, param.Accept == nil, "param error: accept is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.RespondSideShow(*param.Accept))
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
