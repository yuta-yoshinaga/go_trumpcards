package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpiteAndMaliceWebInput Spite & Malice Web 入力
type SpiteAndMaliceWebInput struct {
	BaseWebInput
	From *SpiteAndMaliceWebZone `json:"from,omitempty"`
	To   *SpiteAndMaliceWebZone `json:"to,omitempty"`
}

// SpiteAndMaliceWebZone ゾーン指定 (hand|goal|side|foundation)
type SpiteAndMaliceWebZone struct {
	Zone string `json:"zone"`
	Idx  *int   `json:"idx,omitempty"`
}

// SpiteAndMaliceWebPlayer プレイヤー出力
type SpiteAndMaliceWebPlayer struct {
	Hand     []*WebOutputCard                               `json:"hand"`
	GoalTop  *WebOutputCard                                 `json:"goalTop,omitempty"`
	GoalSize int                                            `json:"goalSize"`
	Sides    [domain.SpiteAndMaliceSideCnt][]*WebOutputCard `json:"sides"`
	IsCpu    bool                                           `json:"isCpu"`
}

// SpiteAndMaliceWebHint ヒント出力
type SpiteAndMaliceWebHint struct {
	Source        string `json:"source"`
	Index         int    `json:"index"`
	FoundationIdx int    `json:"foundationIdx"`
	Discard       bool   `json:"discard"`
}

// SpiteAndMaliceWebOutput Spite & Malice Web 出力
type SpiteAndMaliceWebOutput struct {
	Phase           int                                                     `json:"phase"`
	Current         int                                                     `json:"current"`
	Players         [domain.SpiteAndMalicePlayerCnt]SpiteAndMaliceWebPlayer `json:"players"`
	Foundations     [domain.SpiteAndMaliceFoundationCnt][]*WebOutputCard    `json:"foundations"`
	FoundationTops  [domain.SpiteAndMaliceFoundationCnt]int                 `json:"foundationTops"`
	StockSize       int                                                     `json:"stockSize"`
	CompletedSize   int                                                     `json:"completedSize"`
	MoveCount       int                                                     `json:"moveCount"`
	Winner          int                                                     `json:"winner"`
	GoalSize        int                                                     `json:"goalSize"`
	CpuDifficulty   int                                                     `json:"cpuDifficulty"`
	CanAutoComplete bool                                                    `json:"canAutoComplete"`
	Hint            *SpiteAndMaliceWebHint                                  `json:"hint,omitempty"`
	WebOutputBase
}

// SpiteAndMaliceWebController Spite & Malice Web コントローラー
type SpiteAndMaliceWebController = GameWebController[usecase.SpiteAndMaliceInteractorIF, SpiteAndMaliceWebInput, *SpiteAndMaliceWebOutput]

// NewSpiteAndMaliceWebController and NewSpiteAndMaliceWebControllerWithProvider
// are the standard and provider-backed constructors for Spite & Malice.
var NewSpiteAndMaliceWebController, NewSpiteAndMaliceWebControllerWithProvider = webControllerPair[usecase.SpiteAndMaliceInteractorIF, SpiteAndMaliceWebInput, *SpiteAndMaliceWebOutput](
	newSpiteAndMaliceDefaultOutput, spiteAndMaliceDispatch,
)

func newSpiteAndMaliceDefaultOutput(msg string) *SpiteAndMaliceWebOutput {
	return &SpiteAndMaliceWebOutput{
		Winner:        -1,
		GoalSize:      domain.SpiteAndMaliceGoalSizeDefault,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func spiteAndMaliceDispatch(bc *baseController, w http.ResponseWriter, si usecase.SpiteAndMaliceInteractorIF, param SpiteAndMaliceWebInput, newDefault func(string) *SpiteAndMaliceWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return spiteAndMaliceMoveDispatch(bc, w, si, param, newDefault)
	case "d", "discard":
		return spiteAndMaliceDiscardDispatch(bc, w, si, param, newDefault)
	case "cpu":
		bc.writePresenterResponse(w, si.CpuStep())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, si.AutoComplete())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, si.Reset, si.Hint, si.ActionLog)
	}
	return true
}

func spiteAndMaliceMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.SpiteAndMaliceInteractorIF, param SpiteAndMaliceWebInput, newDefault func(string) *SpiteAndMaliceWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	if !requireParam(bc, w, newDefault, param.To.Zone != "foundation" || param.To.Idx == nil, "param error: to.zone must be 'foundation' with idx.") {
		return true
	}
	switch param.From.Zone {
	case "hand":
		if !requireParam(bc, w, newDefault, param.From.Idx == nil, "param error: from.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.PlayFromHand(*param.From.Idx, *param.To.Idx))
	case "goal":
		bc.writePresenterResponse(w, si.PlayFromGoal(*param.To.Idx))
	case "side":
		if !requireParam(bc, w, newDefault, param.From.Idx == nil, "param error: from.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.PlayFromSide(*param.From.Idx, *param.To.Idx))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid from.zone."))
	}
	return true
}

func spiteAndMaliceDiscardDispatch(bc *baseController, w http.ResponseWriter, si usecase.SpiteAndMaliceInteractorIF, param SpiteAndMaliceWebInput, newDefault func(string) *SpiteAndMaliceWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	if !requireParam(bc, w, newDefault, param.From.Zone != "hand" || param.From.Idx == nil, "param error: from must be hand with idx.") {
		return true
	}
	if !requireParam(bc, w, newDefault, param.To.Zone != "side" || param.To.Idx == nil, "param error: to must be side with idx.") {
		return true
	}
	bc.writePresenterResponse(w, si.Discard(*param.From.Idx, *param.To.Idx))
	return true
}
