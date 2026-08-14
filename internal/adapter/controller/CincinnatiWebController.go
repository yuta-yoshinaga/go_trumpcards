//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CincinnatiWebInput シンシナティWebインプット
type CincinnatiWebInput struct {
	BaseWebInput
	// Amount はベット / レイズの額。
	Amount *int `json:"amount,omitempty"`
}

// CincinnatiWebOutCfg はシンシナティの設定
type CincinnatiWebOutCfg struct {
	Seats        int `json:"seats"`
	InitialChips int `json:"initialChips"`
	Ante         int `json:"ante"`
}

// CincinnatiWebOutputSeat は 1 席の状態
type CincinnatiWebOutputSeat struct {
	Name    string `json:"name"`
	IsHuman bool   `json:"isHuman"`
	Chips   int    `json:"chips"`
	Bet     int    `json:"bet"`
	// Cards は手札 5 枚。**人間の席とショーダウン以外は伏せる。**
	Cards  []*WebOutputCard `json:"cards"`
	Folded bool             `json:"folded"`
	AllIn  bool             `json:"allIn"`
	IsTurn bool             `json:"isTurn"`
	// HandRank と BestHand はショーダウン後のみ。
	HandRank  int              `json:"handRank"`
	BestHand  []*WebOutputCard `json:"bestHand"`
	WonAmount int              `json:"wonAmount"`
}

// CincinnatiWebOutput シンシナティWebアウトプット
type CincinnatiWebOutput struct {
	Phase int                        `json:"phase"`
	Seats []*CincinnatiWebOutputSeat `json:"seats"`
	// Community は表向きのコミュニティ。**伏せている札は載せない。**
	Community []*WebOutputCard `json:"community"`
	// RevealedCount は公開済みの枚数 (0..5)。
	RevealedCount int `json:"revealedCount"`
	// CommunityTotal はコミュニティの総枚数 (5)。残り枚数を画面が出せるように。
	CommunityTotal int  `json:"communityTotal"`
	Pot            int  `json:"pot"`
	CurrentBet     int  `json:"currentBet"`
	ToCall         int  `json:"toCall"`
	RaiseCount     int  `json:"raiseCount"`
	CanRaise       bool `json:"canRaise"`
	TurnSeat       int  `json:"turnSeat"`
	HumanSeat      int  `json:"humanSeat"`
	IsHumanTurn    bool `json:"isHumanTurn"`
	HandNumber     int  `json:"handNumber"`
	RemainingCards int  `json:"remainingCards"`
	WinnerSeat     int  `json:"winnerSeat"`
	GameEndFlag    bool `json:"gameEndFlag"`

	Config *CincinnatiWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// CincinnatiWebController シンシナティWebコントローラークラス
type CincinnatiWebController = GameWebController[usecase.CincinnatiInteractorIF, CincinnatiWebInput, *CincinnatiWebOutput]

// NewCincinnatiWebController and NewCincinnatiWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewCincinnatiWebController, NewCincinnatiWebControllerWithProvider = webControllerPair[usecase.CincinnatiInteractorIF, CincinnatiWebInput, *CincinnatiWebOutput](
	newCincinnatiDefaultOutput, cincinnatiDispatch,
)

func newCincinnatiDefaultOutput(msg string) *CincinnatiWebOutput {
	return &CincinnatiWebOutput{
		Seats:         make([]*CincinnatiWebOutputSeat, 0),
		Community:     make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

// cincinnatiAction はコマンドとドメインのアクション定数の対応。
//
// **数値を書き写さない。** ドメインの定数をそのまま参照する ── ここに 0〜4 を
// 直書きすると、アクションの並びが 2 か所に存在することになる。
var cincinnatiAction = map[string]int{
	"fold": domain.CincinnatiActionFold, "f": domain.CincinnatiActionFold,
	"check": domain.CincinnatiActionCheck, "k": domain.CincinnatiActionCheck,
	"call": domain.CincinnatiActionCall, "c": domain.CincinnatiActionCall,
	"bet": domain.CincinnatiActionBet, "b": domain.CincinnatiActionBet,
	"raise": domain.CincinnatiActionRaise, "r2": domain.CincinnatiActionRaise,
}

func cincinnatiDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CincinnatiInteractorIF, param CincinnatiWebInput, newOut func(string) *CincinnatiWebOutput) bool {
	switch param.Command {
	case "fold", "f", "check", "k", "call", "c":
		// **額を取らない手。** 送られていても無視する。
		bc.writePresenterResponse(w, ci.Action(cincinnatiActionValue(param.Command), 0))
	case "bet", "b", "raise", "r2":
		// **額は必須。** 0 を「省略」と同一視しないよう、未送信だけを弾く。
		if !requireParam(bc, w, newOut, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Action(cincinnatiActionValue(param.Command), *param.Amount))
	case "next":
		bc.writePresenterResponse(w, ci.NextHand())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}

// cincinnatiActionValue はコマンド名をドメインのアクション値に直す。
func cincinnatiActionValue(cmd string) int { return cincinnatiAction[cmd] }
