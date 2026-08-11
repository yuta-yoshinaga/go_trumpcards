//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func intPtrSnap(v int) *int { return &v }

func TestSnapWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0,"players":[],"message":""}`

	siMock := new(usecase.MockSnapInteractor)
	siMock.On("GetConfig").Return(domain.DefaultSnapConfig())
	siMock.On("Reset").Return(mockOutput)
	siMock.On("ResetWithConfig", domain.SnapConfig{PlayerCnt: 3, CpuDifficulty: domain.SnapCpuHard}).Return(mockOutput)
	siMock.On("Step").Return(mockOutput)
	siMock.On("Snap").Return(mockOutput)
	siMock.On("Tick").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewSnapWebController(func() uc.SnapInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.SnapWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with config", `{"command":"reset","sessionId":"s1","config":{"playerCnt":3,"cpuDifficulty":2}}`},
		{"step s", `{"command":"s","sessionId":"s1"}`},
		{"snap n", `{"command":"n","sessionId":"s1"}`},
		{"tick t", `{"command":"t","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	t.Run("quit q", func(t *testing.T) {
		exec(t, `{"command":"q","sessionId":"s1"}`).CodeIs(http.StatusOK)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	// **宣言は席を受け付けない。** 受け付けると CPU に誤宣言させられる。
	t.Run("snap ignores any seat the client sends", func(t *testing.T) {
		exec(t, `{"command":"n","sessionId":"s1","playerIdx":1}`).CodeIs(http.StatusOK)
		siMock.AssertNumberOfCalls(t, "Snap", 2)
	})
}

func TestSnapWebConfig_ClampsBothFields(t *testing.T) {
	cur := domain.SnapConfig{PlayerCnt: 2, CpuDifficulty: domain.SnapCpuNormal}
	for _, tc := range []struct {
		name          string
		players, diff *int
		wantPlayers   int
		wantDiff      domain.SnapCpuDifficulty
	}{
		{"nil keeps the current values", nil, nil, 2, domain.SnapCpuNormal},
		{"players below the minimum", intPtrSnap(1), nil, 2, domain.SnapCpuNormal},
		{"players above the maximum", intPtrSnap(9), nil, 2, domain.SnapCpuNormal},
		{"four players is kept", intPtrSnap(4), nil, 4, domain.SnapCpuNormal},
		{"difficulty below the minimum", nil, intPtrSnap(-1), 2, domain.SnapCpuNormal},
		{"difficulty above the maximum", nil, intPtrSnap(9), 2, domain.SnapCpuNormal},
		{"hard is kept", nil, intPtrSnap(2), 2, domain.SnapCpuHard},
		{"both at once", intPtrSnap(3), intPtrSnap(0), 3, domain.SnapCpuEasy},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.SnapWebInput
			input.Config = &controller.SnapWebConfig{PlayerCnt: tc.players, CpuDifficulty: tc.diff}
			var in controller.SnapWebInput
			b, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(b, &in); err != nil {
				t.Fatal(err)
			}
			cfg := controller.SnapConfigFromInputForTest(cur, in.Config)
			if cfg.PlayerCnt != tc.wantPlayers {
				t.Fatalf("PlayerCnt = %d, want %d", cfg.PlayerCnt, tc.wantPlayers)
			}
			if cfg.CpuDifficulty != tc.wantDiff {
				t.Fatalf("CpuDifficulty = %d, want %d", cfg.CpuDifficulty, tc.wantDiff)
			}
			// **丸めた結果はドメインに必ず受理される。**
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}
}
