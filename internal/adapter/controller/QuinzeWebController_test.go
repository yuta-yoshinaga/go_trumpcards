//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustQuinzeOutputJSON(msg string) string {
	out := &controller.QuinzeWebOutput{
		Seats:         []*controller.QuinzeWebOutputSeat{},
		NextBanker:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustQuinzeOutputJSON: %v", err))
	}
	return string(b)
}

func TestQuinzeWebController_Method(t *testing.T) {
	mockOutput := `{"seats":[],"bankerIdx":0,"isHumanBanker":false,"chips":1000,"activeSeat":0,"nextBanker":-1,"lastResult":"","phase":1,"targetPoints":15,"canHit":false,"canStand":false,"message":""}`

	siMock := new(usecase.MockQuinzeInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("Bet", 100).Return(mockOutput)
	siMock.On("Deal").Return(mockOutput)
	siMock.On("Hit").Return(mockOutput)
	siMock.On("Stand").Return(mockOutput)
	siMock.On("BankerHit").Return(mockOutput)
	siMock.On("BankerStand").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewQuinzeWebController(func() uc.QuinzeInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.QuinzeWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustQuinzeOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"bet", `{"command":"b","sessionId":"s1","amount":100}`},
		{"deal as banker", `{"command":"deal","sessionId":"s1"}`},
		{"hit", `{"command":"h","sessionId":"s1"}`},
		{"stand", `{"command":"s","sessionId":"s1"}`},
		{"banker hit", `{"command":"bh","sessionId":"s1"}`},
		{"banker stand", `{"command":"bs","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	for _, tc := range []struct{ name, body string }{
		{"bet missing amount", `{"command":"b","sessionId":"s1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}

func TestQuinzeWebOutput_JSONUsesPointKeys(t *testing.T) {
	out := &controller.QuinzeWebOutput{
		Seats:          []*controller.QuinzeWebOutputSeat{{Hand: &controller.QuinzeWebOutputHand{TotalPoints: 15}}},
		BankerHand:     &controller.QuinzeWebOutputHand{TotalPoints: 12},
		TargetPoints:   15,
		CpuStandPoints: 12,
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal Quinze output: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal Quinze output: %v", err)
	}
	for _, key := range []string{"targetPoints", "cpuStandPoints"} {
		if _, ok := m[key]; !ok {
			t.Errorf("serialized Quinze output is missing %q", key)
		}
	}
	handJSON, err := json.Marshal(out.Seats[0].Hand)
	if err != nil {
		t.Fatalf("marshal Quinze hand: %v", err)
	}
	var hand map[string]any
	if err := json.Unmarshal(handJSON, &hand); err != nil {
		t.Fatalf("unmarshal Quinze hand: %v", err)
	}
	if _, ok := hand["totalPoints"]; !ok {
		t.Error("serialized Quinze hand is missing \"totalPoints\"")
	}
	if _, ok := m["targetHalves"]; ok {
		t.Error("serialized Quinze output unexpectedly contains targetHalves")
	}
}
