//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PerseveranceWebInput パーシビアランスWebインプット
type PerseveranceWebInput struct {
	BaseWebInput
	From *PerseveranceWebZone `json:"from,omitempty"`
	To   *PerseveranceWebZone `json:"to,omitempty"`
}

// PerseveranceWebZone ゾーン指定
type PerseveranceWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// PerseveranceWebOutputTableauCard タブローカード出力
type PerseveranceWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// PerseveranceWebOutputHint ヒント出力
type PerseveranceWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// PerseveranceWebOutput パーシビアランスWebアウトプット
type PerseveranceWebOutput struct {
	Tableau    [][]*PerseveranceWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                    `json:"foundation"`
	Hint       *PerseveranceWebOutputHint            `json:"hint,omitempty"`
	// RedealsLeft 残りの再配り回数 (0-2)。0 になると再配りボタンは押せない。
	RedealsLeft int `json:"redealsLeft"`
	SolitaireWebOutputBase
	WebOutputBase
}

// PerseveranceWebController パーシビアランスWebコントローラークラス
type PerseveranceWebController = GameWebController[usecase.PerseveranceInteractorIF, PerseveranceWebInput, *PerseveranceWebOutput]

// NewPerseveranceWebController and NewPerseveranceWebControllerWithProvider are
// the standard and provider-backed constructors for PerseveranceWebController.
var NewPerseveranceWebController, NewPerseveranceWebControllerWithProvider = webControllerPair[usecase.PerseveranceInteractorIF, PerseveranceWebInput, *PerseveranceWebOutput](
	newPerseveranceDefaultOutput, perseveranceDispatch,
)

func newPerseveranceDefaultOutput(msg string) *PerseveranceWebOutput {
	return &PerseveranceWebOutput{
		Tableau:       make([][]*PerseveranceWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func perseveranceDispatch(bc *baseController, w http.ResponseWriter, bi usecase.PerseveranceInteractorIF, param PerseveranceWebInput, newDefault func(string) *PerseveranceWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return perseveranceMoveDispatch(bc, w, bi, param, newDefault)
	case "rd", "redeal":
		bc.writePresenterResponse(w, bi.Redeal())
	case "g", "giveup":
		bc.writePresenterResponse(w, bi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, bi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, bi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, bi.Reset, bi.Hint, bi.ActionLog)
	}
	return true
}

// perseveranceMoveDispatch は `move` を捌く。
//
// **共有の dispatchTopCardMove は使えない。**あれは「上札しか動かさないソリティア」
// (BakersDozen / BeleagueredCastle / Fortress / Somerset / StreetsAndAlleys)
// 専用で、クライアントの
// cardIndex を捨てて -1 を渡す契約になっている。Perseverance は同スート降順の並びを
// 一括で動かせるので、その契約のままだと**この game の看板ルールがサーバに届かない**。
//
// cardIndex は信用しないのではなく**ドメインに検査させる**: MoveTableauToTableau が
// 範囲と isRun を見て、並びになっていなければ 1 枚も動かさずに弾く。省略時は -1 に
// 落として従来どおり上札を動かす。
func perseveranceMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.PerseveranceInteractorIF, param PerseveranceWebInput, newDefault func(string) *PerseveranceWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	switch {
	case param.From.Zone == "tableau" && param.To.Zone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		cardIndex := -1
		if param.From.CardIndex != nil {
			cardIndex = *param.From.CardIndex
		}
		bc.writePresenterResponse(w, bi.MoveTableauToTableau(*param.From.Col, cardIndex, *param.To.Col))
	case param.From.Zone == "tableau" && param.To.Zone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
