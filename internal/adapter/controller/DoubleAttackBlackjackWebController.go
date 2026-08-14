//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoubleAttackBlackjackWebInput 追加ベット・ブラックジャックWebインプット
type DoubleAttackBlackjackWebInput struct {
	BaseWebInput
	// Ante はアンティ額。
	Ante *int `json:"ante,omitempty"`
	// BustIt は任意のサイドベット額。**0 は「置かない」という有効な値。**
	BustIt *int `json:"bustIt,omitempty"`
	// Amount は追加ベット額。**0 は「見送り」という有効な値**なのでポインタで受ける。
	Amount *int `json:"amount,omitempty"`
}

// DoubleAttackWebOutCfg は追加ベット・ブラックジャックの設定
type DoubleAttackWebOutCfg struct {
	InitialChips int `json:"initialChips"`
	DefaultAnte  int `json:"defaultAnte"`
}

// DoubleAttackWebOutputHand は 1 つの手札
type DoubleAttackWebOutputHand struct {
	Cards   []*WebOutputCard `json:"cards"`
	Score   int              `json:"score"`
	Bet     int              `json:"bet"`
	IsSoft  bool             `json:"isSoft"`
	Stood   bool             `json:"stood"`
	Doubled bool             `json:"doubled"`
	Busted  bool             `json:"busted"`
	// Blackjack は最初の 2 枚で 21。**配当は 1:1** で、3:2 ではない。
	Blackjack bool `json:"blackjack"`
	Result    int  `json:"result"`
}

// DoubleAttackBlackjackWebOutput 追加ベット・ブラックジャックWebアウトプット
type DoubleAttackBlackjackWebOutput struct {
	Phase int                          `json:"phase"`
	Hands []*DoubleAttackWebOutputHand `json:"hands"`
	// ActiveHand はいま操作している手札の位置。
	ActiveHand int `json:"activeHand"`
	// DealerCards は追加ベットの前は**アップカード 1 枚だけ**。
	DealerCards []*WebOutputCard `json:"dealerCards"`
	DealerScore int              `json:"dealerScore"`
	// DealerHoleDealt は 2 枚目が配られたか。false の間は情報が伏せられている。
	DealerHoleDealt bool `json:"dealerHoleDealt"`
	// MaxAttackBet は追加ベットの上限 (アンティまで)。ページはこれに従うこと。
	MaxAttackBet   int  `json:"maxAttackBet"`
	CanDouble      bool `json:"canDouble"`
	CanSplit       bool `json:"canSplit"`
	AnteBet        int  `json:"anteBet"`
	AttackBet      int  `json:"attackBet"`
	BustItBet      int  `json:"bustItBet"`
	Payout         int  `json:"payout"`
	BustItPayout   int  `json:"bustItPayout"`
	Chips          int  `json:"chips"`
	RoundNumber    int  `json:"roundNumber"`
	RemainingCards int  `json:"remainingCards"`
	GameEndFlag    bool `json:"gameEndFlag"`

	Config *DoubleAttackWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// DoubleAttackBlackjackWebController 追加ベット・ブラックジャックWebコントローラークラス
type DoubleAttackBlackjackWebController = GameWebController[usecase.DoubleAttackBlackjackInteractorIF, DoubleAttackBlackjackWebInput, *DoubleAttackBlackjackWebOutput]

// NewDoubleAttackBlackjackWebController and NewDoubleAttackBlackjackWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewDoubleAttackBlackjackWebController, NewDoubleAttackBlackjackWebControllerWithProvider = webControllerPair[usecase.DoubleAttackBlackjackInteractorIF, DoubleAttackBlackjackWebInput, *DoubleAttackBlackjackWebOutput](
	newDoubleAttackDefaultOutput, doubleAttackDispatch,
)

func newDoubleAttackDefaultOutput(msg string) *DoubleAttackBlackjackWebOutput {
	return &DoubleAttackBlackjackWebOutput{
		Hands:         make([]*DoubleAttackWebOutputHand, 0),
		DealerCards:   make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func doubleAttackDispatch(bc *baseController, w http.ResponseWriter, ci usecase.DoubleAttackBlackjackInteractorIF, param DoubleAttackBlackjackWebInput, newOut func(string) *DoubleAttackBlackjackWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		if !requireParam(bc, w, newOut, param.Ante == nil, "param error: ante is required.") {
			return true
		}
		bustIt := 0
		if param.BustIt != nil {
			bustIt = *param.BustIt
		}
		bc.writePresenterResponse(w, ci.PlaceBet(*param.Ante, bustIt))
	case "a", "attack":
		// **見送り (0) を送れる必要がある。** 省略と区別が要る。
		if !requireParam(bc, w, newOut, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Attack(*param.Amount))
	case "h", "hit":
		bc.writePresenterResponse(w, ci.Hit())
	case "s", "stand":
		bc.writePresenterResponse(w, ci.Stand())
	case "d", "double":
		bc.writePresenterResponse(w, ci.Double())
	case "sp", "split":
		bc.writePresenterResponse(w, ci.Split())
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
