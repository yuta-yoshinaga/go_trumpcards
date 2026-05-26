//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestChinesePokerWebController(t *testing.T) {
	mockOutput := `{"playerCards":[],"dealerCards":[],"playerFront":[],"playerMiddle":[],"playerBack":[],"dealerFront":[],"dealerMiddle":[],"dealerBack":[],"phase":0,"chips":0,"bet":0,"result":0,"frontResult":0,"middleResult":0,"backResult":0,"payout":0,"playerFrontRank":0,"playerMiddleRank":0,"playerBackRank":0,"dealerFrontRank":0,"dealerMiddleRank":0,"dealerBackRank":0,"playerRoyalty":0,"dealerRoyalty":0,"scoop":false,"message":""}`

	ciMock := new(usecase.MockChinesePokerInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("Bet", 100).Return(mockOutput)
	ciMock.On("SetHands", []int{0, 1, 2}, []int{3, 4, 5, 6, 7}).Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ChinesePokerInteractorIF { return ciMock }
	ctrl := controller.NewChinesePokerWebController(factory)
	defer ctrl.Stop()

	t.Run("reset", func(t *testing.T) {
		var input controller.ChinesePokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.ChinesePokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("set", func(t *testing.T) {
		var input controller.ChinesePokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"set","frontIndices":[0,1,2],"middleIndices":[3,4,5,6,7],"sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.ChinesePokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})
}
