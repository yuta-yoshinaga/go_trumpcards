//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TuSacWebInput 四色牌Webインプット
type TuSacWebInput struct {
	BaseWebInput
	// Index は捨てる札の位置 (0 始まり)。
	Index *int `json:"index,omitempty"`
	// Indexes は場に出す札の位置 (0 始まり)。
	Indexes []int `json:"indexes,omitempty"`
}

// TuSacWebOutCfg は四色牌の設定
type TuSacWebOutCfg struct {
	Seats  int `json:"seats"`
	Rounds int `json:"rounds"`
}

// TuSacWebOutputMeld は場に出した組み合わせ
type TuSacWebOutputMeld struct {
	// Kind は 1=同色同種, 2=異色の車馬砲, 3=卒5枚。
	Kind int `json:"kind"`
	// Points はこの組み合わせの得点。
	Points int              `json:"points"`
	Cards  []*WebOutputCard `json:"cards"`
}

// TuSacWebOutputSeat は 1 席の状態
type TuSacWebOutputSeat struct {
	Name    string `json:"name"`
	IsHuman bool   `json:"isHuman"`
	// Cards は手札。**人間の席以外は空** ── 相手の手は最後まで見えない。
	Cards []*WebOutputCard `json:"cards"`
	// HandCount は手札の枚数。相手のぶんも枚数だけは分かる。
	HandCount int `json:"handCount"`
	// Melds は場に出した組み合わせ。**全員ぶん見える。**
	Melds      []*TuSacWebOutputMeld `json:"melds"`
	MeldPoints int                   `json:"meldPoints"`
	Score      int                   `json:"score"`
	RoundScore int                   `json:"roundScore"`
	IsTurn     bool                  `json:"isTurn"`
	WentOut    bool                  `json:"wentOut"`
}

// TuSacWebOutput 四色牌Webアウトプット
type TuSacWebOutput struct {
	Phase int                   `json:"phase"`
	Seats []*TuSacWebOutputSeat `json:"seats"`
	// DiscardTop は捨て札の一番上 (無ければ null)。
	DiscardTop   *WebOutputCard `json:"discardTop"`
	DiscardCount int            `json:"discardCount"`
	StockCount   int            `json:"stockCount"`
	TurnSeat     int            `json:"turnSeat"`
	HumanSeat    int            `json:"humanSeat"`
	IsHumanTurn  bool           `json:"isHumanTurn"`
	RoundNumber  int            `json:"roundNumber"`
	Rounds       int            `json:"rounds"`
	// WentOutSeat は上がった席 (-1 なら山切れ)。
	WentOutSeat int `json:"wentOutSeat"`
	// HandSize は配る枚数、DeckSize は札の総数。**画面に書き写させない。**
	HandSize int `json:"handSize"`
	DeckSize int `json:"deckSize"`
	// MeldPointsByKind は種別ごとの得点 (添字 = 種別)。
	MeldPointsByKind []int `json:"meldPointsByKind"`
	WinnerSeat       int   `json:"winnerSeat"`
	GameEndFlag      bool  `json:"gameEndFlag"`

	Config *TuSacWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// TuSacWebController 四色牌Webコントローラークラス
type TuSacWebController = GameWebController[usecase.TuSacInteractorIF, TuSacWebInput, *TuSacWebOutput]

// NewTuSacWebController and NewTuSacWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewTuSacWebController, NewTuSacWebControllerWithProvider = webControllerPair[usecase.TuSacInteractorIF, TuSacWebInput, *TuSacWebOutput](
	newTuSacDefaultOutput, tuSacDispatch,
)

func newTuSacDefaultOutput(msg string) *TuSacWebOutput {
	return &TuSacWebOutput{
		Seats:            make([]*TuSacWebOutputSeat, 0),
		WentOutSeat:      -1,
		HandSize:         domain.TuSacHandSize,
		DeckSize:         domain.TuSacDeckSize,
		MeldPointsByKind: tuSacMeldPointsByKind(),
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

// tuSacMeldPointsByKind は種別ごとの得点を添字順に並べる。
func tuSacMeldPointsByKind() []int {
	out := make([]int, int(domain.TuSacMeldKindMax)+1)
	for k := domain.TuSacMeldNone; k <= domain.TuSacMeldKindMax; k++ {
		out[int(k)] = domain.TuSacMeldPoints(k)
	}
	return out
}

func tuSacDispatch(bc *baseController, w http.ResponseWriter, ci usecase.TuSacInteractorIF, param TuSacWebInput, newOut func(string) *TuSacWebOutput) bool {
	switch param.Command {
	// **山と捨て札は別のコマンド。** 引き先を本文の真偽値にすると、
	// 送り忘れが「山から」に化けて、狙って拾った札が黙って流れる。
	case "draw", "d":
		bc.writePresenterResponse(w, ci.Draw(false))
	case "take", "t":
		bc.writePresenterResponse(w, ci.Draw(true))
	case "meld", "m":
		if !requireParam(bc, w, newOut, len(param.Indexes) == 0, "param error: indexes are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Meld(param.Indexes))
	case "discard", "x":
		// **位置は必須。** 0 を「省略」と同一視しない ── 0 は先頭の札。
		if !requireParam(bc, w, newOut, param.Index == nil, "param error: index is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Discard(*param.Index))
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
