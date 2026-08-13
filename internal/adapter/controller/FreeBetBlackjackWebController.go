//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FreeBetBlackjackWebInput フリーベット・ブラックジャックWebインプット
type FreeBetBlackjackWebInput struct {
	BaseWebInput
	// Ante はアンティ額。
	Ante *int `json:"ante,omitempty"`
}

// FreeBetWebOutCfg はフリーベット・ブラックジャックの設定
type FreeBetWebOutCfg struct {
	InitialChips int `json:"initialChips"`
	DefaultAnte  int `json:"defaultAnte"`
}

// FreeBetWebOutputHand は 1 つの手札
type FreeBetWebOutputHand struct {
	Cards []*WebOutputCard `json:"cards"`
	Score int              `json:"score"`
	// Bet は**プレイヤー自身の金**。負ければ失う額。
	Bet int `json:"bet"`
	// FreeBet は**ハウスが出した金**。勝てば配当が付き、負けても失わない。
	FreeBet   int  `json:"freeBet"`
	IsSoft    bool `json:"isSoft"`
	Stood     bool `json:"stood"`
	Doubled   bool `json:"doubled"`
	Busted    bool `json:"busted"`
	Blackjack bool `json:"blackjack"`
	Result    int  `json:"result"`
}

// FreeBetBlackjackWebOutput フリーベット・ブラックジャックWebアウトプット
type FreeBetBlackjackWebOutput struct {
	Phase       int                     `json:"phase"`
	Hands       []*FreeBetWebOutputHand `json:"hands"`
	ActiveHand  int                     `json:"activeHand"`
	DealerCards []*WebOutputCard        `json:"dealerCards"`
	DealerScore int                     `json:"dealerScore"`
	// DealerPushed22 はディーラーが 22 でバストしたか。**無料ダブル / 無料スプリットの対価。**
	DealerPushed22 bool `json:"dealerPushed22"`
	// CanFreeDouble / CanFreeSplit はサーバが判定する。ページは再計算しないこと。
	CanFreeDouble  bool `json:"canFreeDouble"`
	CanFreeSplit   bool `json:"canFreeSplit"`
	AnteBet        int  `json:"anteBet"`
	Payout         int  `json:"payout"`
	Chips          int  `json:"chips"`
	RoundNumber    int  `json:"roundNumber"`
	RemainingCards int  `json:"remainingCards"`
	GameEndFlag    bool `json:"gameEndFlag"`

	Config *FreeBetWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// FreeBetBlackjackWebController フリーベット・ブラックジャックWebコントローラークラス
type FreeBetBlackjackWebController = GameWebController[usecase.FreeBetBlackjackInteractorIF, FreeBetBlackjackWebInput, *FreeBetBlackjackWebOutput]

// NewFreeBetBlackjackWebController and NewFreeBetBlackjackWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewFreeBetBlackjackWebController, NewFreeBetBlackjackWebControllerWithProvider = webControllerPair[usecase.FreeBetBlackjackInteractorIF, FreeBetBlackjackWebInput, *FreeBetBlackjackWebOutput](
	newFreeBetDefaultOutput, freeBetDispatch,
)

func newFreeBetDefaultOutput(msg string) *FreeBetBlackjackWebOutput {
	return &FreeBetBlackjackWebOutput{
		Hands:         make([]*FreeBetWebOutputHand, 0),
		DealerCards:   make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func freeBetDispatch(bc *baseController, w http.ResponseWriter, ci usecase.FreeBetBlackjackInteractorIF, param FreeBetBlackjackWebInput, newOut func(string) *FreeBetBlackjackWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		if !requireParam(bc, w, newOut, param.Ante == nil, "param error: ante is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.PlaceBet(*param.Ante))
	case "h", "hit":
		bc.writePresenterResponse(w, ci.Hit())
	case "s", "stand":
		bc.writePresenterResponse(w, ci.Stand())
	case "fd", "freedouble":
		bc.writePresenterResponse(w, ci.FreeDouble())
	case "fs", "freesplit":
		bc.writePresenterResponse(w, ci.FreeSplit())
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
