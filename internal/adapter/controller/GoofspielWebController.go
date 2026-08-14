//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GoofspielWebInput ゴフスピールWebインプット
type GoofspielWebInput struct {
	BaseWebInput
	CardIndex *int                `json:"cardIndex,omitempty"`
	Config    *GoofspielWebConfig `json:"config,omitempty"`
}

// GoofspielWebConfig ゴフスピールWeb設定
type GoofspielWebConfig struct {
	PlayerCnt *int `json:"playerCnt,omitempty"`
	TieRule   *int `json:"tieRule,omitempty"`
}

// GoofspielWebOutputPlayer ゴフスピールWebアウトプットプレイヤー
type GoofspielWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は残りの入札札の枚数。
	CardCount int `json:"cardCount"`
	// Cards は残りの入札札。**CPU の分も公開します。**
	//
	// 使った札は場に出るので、残りは誰にでも数えられます。隠されているのは
	// 「今このラウンドで何を出したか」だけです。
	Cards []*WebOutputCard `json:"cards"`
	Score int              `json:"score"`
	// HasBid はこのラウンドで伏せ終えたか。
	HasBid bool `json:"hasBid"`
	// RevealedBid は直前に公開された入札 (nil = まだ公開されていない)。
	RevealedBid *WebOutputCard `json:"revealedBid,omitempty"`
}

// GoofspielWebOutputHint ヒント出力
type GoofspielWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// GoofspielWebOutput ゴフスピールWebアウトプット
type GoofspielWebOutput struct {
	Players []*GoofspielWebOutputPlayer `json:"players"`
	Phase   int                         `json:"phase"`
	// ValidPlays は入札できる手札の位置。
	ValidPlays []int `json:"validPlays"`
	// CurrentPrize はいま公開されている賞札 (nil = 決着済み)。
	CurrentPrize *WebOutputCard `json:"currentPrize,omitempty"`
	// CarriedPrizes は同点で持ち越された賞札。
	CarriedPrizes []*WebOutputCard `json:"carriedPrizes"`
	// PrizeValue はいま懸かっている得点 (持ち越しを含む)。
	PrizeValue     int `json:"prizeValue"`
	PrizeRemaining int `json:"prizeRemaining"`
	// LastWinnerIdx は直前のラウンドの勝者 (-1 = 同点で勝者なし)。
	LastWinnerIdx int                     `json:"lastWinnerIdx"`
	LastGained    int                     `json:"lastGained"`
	RoundNumber   int                     `json:"roundNumber"`
	GameEndFlag   bool                    `json:"gameEndFlag"`
	WinnerIdx     int                     `json:"winnerIdx"`
	Hint          *GoofspielWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config GoofspielWebOutputConfig `json:"config"`
}

// GoofspielWebOutputConfig ゴフスピール設定アウトプット
type GoofspielWebOutputConfig struct {
	PlayerCnt int `json:"playerCnt"`
	TieRule   int `json:"tieRule"`
}

// ToConfig builds a GoofspielConfig from the nested web config, applying bounds checking.
func (c *GoofspielWebConfig) ToConfig() domain.GoofspielConfig {
	cfg := domain.DefaultGoofspielConfig()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.GoofspielPlayerCntMin, domain.GoofspielPlayerCntMax, cfg.PlayerCnt)
	cfg.TieRule = domain.GoofspielTieRule(webutil.BoundedIntPtr(c.TieRule,
		int(domain.GoofspielTieDiscard), int(domain.GoofspielTieCarryOver), int(cfg.TieRule)))
	return cfg
}

// ToConfig builds a GoofspielConfig from the web input.
func (p GoofspielWebInput) ToConfig() domain.GoofspielConfig {
	return configOrDefault(p.Config, (*GoofspielWebConfig).ToConfig, domain.DefaultGoofspielConfig())
}

// GoofspielWebController ゴフスピールWebコントローラークラス
type GoofspielWebController = GameWebController[usecase.GoofspielInteractorIF, GoofspielWebInput, *GoofspielWebOutput]

// NewGoofspielWebController and NewGoofspielWebControllerWithProvider are the
// standard and provider-backed constructors for GoofspielWebController.
var NewGoofspielWebController, NewGoofspielWebControllerWithProvider = webControllerPair[usecase.GoofspielInteractorIF, GoofspielWebInput, *GoofspielWebOutput](
	newGoofspielDefaultOutput, goofspielDispatch,
)

func newGoofspielDefaultOutput(msg string) *GoofspielWebOutput {
	return &GoofspielWebOutput{
		Players:       make([]*GoofspielWebOutputPlayer, 0),
		ValidPlays:    make([]int, 0),
		CarriedPrizes: make([]*WebOutputCard, 0),
		LastWinnerIdx: -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func goofspielDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GoofspielInteractorIF, param GoofspielWebInput, newDefault func(string) *GoofspielWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, gi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.Bid(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, gi.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, gi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, gi.Hint, gi.ActionLog)
	}
	return true
}
