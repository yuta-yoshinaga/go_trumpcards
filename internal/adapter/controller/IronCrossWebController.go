//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// IronCrossWebInput アイアンクロスWebインプット
type IronCrossWebInput struct {
	BaseWebInput
	// Amount はベット / レイズの額。
	Amount *int `json:"amount,omitempty"`
	// Line は使う列 (1=縦, 2=横)。
	Line *int `json:"line,omitempty"`
}

// IronCrossWebOutCfg はアイアンクロスの設定
type IronCrossWebOutCfg struct {
	Seats        int `json:"seats"`
	InitialChips int `json:"initialChips"`
	Ante         int `json:"ante"`
}

// IronCrossWebOutputSeat は 1 席の状態
type IronCrossWebOutputSeat struct {
	Name    string `json:"name"`
	IsHuman bool   `json:"isHuman"`
	Chips   int    `json:"chips"`
	Bet     int    `json:"bet"`
	// Cards は手札 4 枚。**人間の席とショーダウン以外は伏せる。**
	Cards  []*WebOutputCard `json:"cards"`
	Folded bool             `json:"folded"`
	AllIn  bool             `json:"allIn"`
	IsTurn bool             `json:"isTurn"`
	// Line は選んだ列 (0=未選択, 1=縦, 2=横)。**CPU の分はショーダウンまで 0。**
	Line int `json:"line"`
	// HandRank と BestHand はショーダウン後のみ。
	HandRank  int              `json:"handRank"`
	BestHand  []*WebOutputCard `json:"bestHand"`
	WonAmount int              `json:"wonAmount"`
}

// IronCrossWebOutput アイアンクロスWebアウトプット
type IronCrossWebOutput struct {
	Phase int                       `json:"phase"`
	Seats []*IronCrossWebOutputSeat `json:"seats"`
	// Cross は十字の 5 枚。**伏せている位置は null で埋める** ── 詰めてしまうと
	// どの枝が開いたのか画面が復元できない。
	Cross []*WebOutputCard `json:"cross"`
	// RevealedCount は公開済みの枚数 (0..5)。
	RevealedCount int `json:"revealedCount"`
	// CrossTotal は十字の総枚数 (5)。
	CrossTotal int `json:"crossTotal"`
	// VerticalIndexes と HorizontalIndexes は各列が使う Cross の添字。
	VerticalIndexes   []int `json:"verticalIndexes"`
	HorizontalIndexes []int `json:"horizontalIndexes"`
	Pot               int   `json:"pot"`
	CurrentBet        int   `json:"currentBet"`
	ToCall            int   `json:"toCall"`
	RaiseCount        int   `json:"raiseCount"`
	CanRaise          bool  `json:"canRaise"`
	TurnSeat          int   `json:"turnSeat"`
	HumanSeat         int   `json:"humanSeat"`
	IsHumanTurn       bool  `json:"isHumanTurn"`
	// IsChoosing は縦横を選ぶ場面かどうか。
	IsChoosing     bool `json:"isChoosing"`
	HandNumber     int  `json:"handNumber"`
	RemainingCards int  `json:"remainingCards"`
	WinnerSeat     int  `json:"winnerSeat"`
	GameEndFlag    bool `json:"gameEndFlag"`

	Config *IronCrossWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// IronCrossWebController アイアンクロスWebコントローラークラス
type IronCrossWebController = GameWebController[usecase.IronCrossInteractorIF, IronCrossWebInput, *IronCrossWebOutput]

// NewIronCrossWebController and NewIronCrossWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewIronCrossWebController, NewIronCrossWebControllerWithProvider = webControllerPair[usecase.IronCrossInteractorIF, IronCrossWebInput, *IronCrossWebOutput](
	newIronCrossDefaultOutput, ironCrossDispatch,
)

func newIronCrossDefaultOutput(msg string) *IronCrossWebOutput {
	return &IronCrossWebOutput{
		Seats:             make([]*IronCrossWebOutputSeat, 0),
		Cross:             make([]*WebOutputCard, 0),
		VerticalIndexes:   domain.IronCrossLineIndexes(domain.IronCrossLineVertical),
		HorizontalIndexes: domain.IronCrossLineIndexes(domain.IronCrossLineHorizontal),
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

// ironCrossAction はコマンドとドメインのアクション定数の対応。
//
// **数値を書き写さない。** ドメインの定数をそのまま参照する。
var ironCrossAction = map[string]int{
	"fold": domain.IronCrossActionFold, "f": domain.IronCrossActionFold,
	"check": domain.IronCrossActionCheck, "k": domain.IronCrossActionCheck,
	"call": domain.IronCrossActionCall, "c": domain.IronCrossActionCall,
	"bet": domain.IronCrossActionBet, "b": domain.IronCrossActionBet,
	"raise": domain.IronCrossActionRaise, "r2": domain.IronCrossActionRaise,
}

// ironCrossLineCommand は列を選ぶコマンド名の対応。
//
// **`line` は本文の `line` を要求し、別名は自分で値を持つ。** 名前つきの
// 別名を用意しておかないと、「縦を選ぶ」だけのために本文を組む必要が出る。
var ironCrossLineCommand = map[string]domain.IronCrossLine{
	"vertical": domain.IronCrossLineVertical, "v": domain.IronCrossLineVertical,
	"horizontal": domain.IronCrossLineHorizontal, "h": domain.IronCrossLineHorizontal,
}

func ironCrossDispatch(bc *baseController, w http.ResponseWriter, ci usecase.IronCrossInteractorIF, param IronCrossWebInput, newOut func(string) *IronCrossWebOutput) bool {
	switch param.Command {
	case "fold", "f", "check", "k", "call", "c":
		// **額を取らない手。** 送られていても無視する。
		bc.writePresenterResponse(w, ci.Action(ironCrossAction[param.Command], 0))
	case "bet", "b", "raise", "r2":
		// **額は必須。** 0 を「省略」と同一視しないよう、未送信だけを弾く。
		if !requireParam(bc, w, newOut, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Action(ironCrossAction[param.Command], *param.Amount))
	case "vertical", "v", "horizontal", "h":
		bc.writePresenterResponse(w, ci.ChooseLine(int(ironCrossLineCommand[param.Command])))
	case "line":
		// **列は必須。** 未送信を「0 番の列」にしない ── 0 は「まだ選んでいない」。
		if !requireParam(bc, w, newOut, param.Line == nil, "param error: line is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.ChooseLine(*param.Line))
	case "next":
		bc.writePresenterResponse(w, ci.NextHand())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
