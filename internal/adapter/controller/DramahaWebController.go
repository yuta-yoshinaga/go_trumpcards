//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DramahaWebInput はドラマハの Web インプット。
//
// **共有の HoldemWebInput は広げない。** ドローがあるのはこのゲームだけで、
// 共有の型に `indices` を足すと、読まない 15 ゲームの API 面にも現れる。
// 埋め込みで足すぶんには他のゲームは何も変わらない。
type DramahaWebInput struct {
	HoldemWebInput
	// Indices は引き直すホールカードの**0 始まり**の位置。`draw` でのみ読む。
	// 省略/空は「交換しない」。
	Indices []int `json:"indices,omitempty"`
}

// DramahaWebController ドラマハWebコントローラークラス
type DramahaWebController = GameWebController[usecase.DramahaInteractorIF, DramahaWebInput, *HoldemWebOutput]

// NewDramahaWebController and NewDramahaWebControllerWithProvider are
// the standard and provider-backed constructors for DramahaWebController.
var NewDramahaWebController, NewDramahaWebControllerWithProvider = webControllerPair[usecase.DramahaInteractorIF, DramahaWebInput, *HoldemWebOutput](
	newDramahaDefaultOutput, dramahaDispatch,
)

func newDramahaDefaultOutput(msg string) *HoldemWebOutput {
	return &HoldemWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func dramahaDispatch(bc *baseController, w http.ResponseWriter, ogi usecase.DramahaInteractorIF, param DramahaWebInput, newDefault func(string) *HoldemWebOutput) bool {
	if dispatchPokerAction(bc, w, ogi, param.Command, param.Amount, param.HumanPlayMs) {
		return true
	}
	switch param.Command {
	case "d", "draw":
		// 省略も空も「交換しない」。**nil と [] を区別しない** —— どちらも
		// プレイヤーが「引かない」と言った形で、区別しても意味が無い。
		bc.writePresenterResponse(w, ogi.Draw(param.Indices))
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, ogi.ResetWithConfig(cfg, param.Profile))
	default:
		return dispatchLog(param.Command, bc, w, ogi.ActionLog)
	}
	return true
}
