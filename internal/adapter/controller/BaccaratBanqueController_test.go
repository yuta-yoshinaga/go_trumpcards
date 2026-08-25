//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustBaccaratBanqueOutputJSON(msg string) string {
	out := &controller.BaccaratBanqueWebOutput{
		Players:       []*controller.BaccaratBanqueWebOutputPlayer{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBaccaratBanqueOutputJSON: %v", err))
	}
	return string(b)
}

func newBaccaratBanqueMock(mockOutput string) *usecase.MockBaccaratBanqueInteractor {
	m := new(usecase.MockBaccaratBanqueInteractor)
	m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	m.On("Draw").Return(mockOutput)
	m.On("Stand").Return(mockOutput)
	m.On("NextCoup").Return(mockOutput)
	m.On("Retire").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)
	m.On("GetConfig").Return(domain.DefaultBaccaratBanqueConfig())
	return m
}

func TestBaccaratBanqueWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"banker"}`
	diMock := newBaccaratBanqueMock(mockOutput)

	ctrl := controller.NewBaccaratBanqueWebController(func() uc.BaccaratBanqueInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.BaccaratBanqueWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustBaccaratBanqueOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", domain.DefaultBaccaratBanqueConfig())
	})

	// **引くと止まるは別のコマンド。** 真偽値ひとつに畳むと、既定のまま届いた
	// 要求がどちらかに黙って倒れる。
	for _, tc := range []struct{ cmd, method string }{
		{"draw", "Draw"}, {"d", "Draw"}, {"stand", "Stand"}, {"s", "Stand"},
	} {
		t.Run(tc.cmd+" calls "+tc.method, func(t *testing.T) {
			m := newBaccaratBanqueMock(mockOutput)
			c := controller.NewBaccaratBanqueWebController(func() uc.BaccaratBanqueInteractorIF { return m })
			defer c.Stop()
			var input controller.BaccaratBanqueWebInput
			_ = json.Unmarshal([]byte(`{"command":"`+tc.cmd+`","sessionId":"s1"}`), &input)
			execRequest(t, c.Exec, &input).BodyIs(mockOutput)
			m.AssertCalled(t, tc.method)
			other := "Draw"
			if tc.method == "Draw" {
				other = "Stand"
			}
			m.AssertNotCalled(t, other)
		})
	}

	for _, cmd := range []string{"nextcoup", "nc"} {
		t.Run(cmd+" advances the coup", func(t *testing.T) {
			run(t, `{"command":"`+cmd+`","sessionId":"s1"}`, mockOutput, http.StatusOK)
			diMock.AssertCalled(t, "NextCoup")
		})
	}

	t.Run("retire gives up the bank", func(t *testing.T) {
		run(t, `{"command":"retire","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Retire")
	})

	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})

	// **範囲外は既定に戻す (webutil.BoundedIntPtr の約束)。** 丸めないのは、
	// 頼んだ額とも既定とも違う 第三の の値で遊ばせないため。
	t.Run("out-of-range config falls back to the default, in range it is kept", func(t *testing.T) {
		m := newBaccaratBanqueMock(mockOutput)
		c := controller.NewBaccaratBanqueWebController(func() uc.BaccaratBanqueInteractorIF { return m })
		defer c.Stop()
		var bad controller.BaccaratBanqueWebInput
		_ = json.Unmarshal([]byte(
			`{"command":"reset","sessionId":"s1","config":{"cpuDifficulty":9,"startChips":1,"betAmount":9999}}`), &bad)
		execRequest(t, c.Exec, &bad).BodyIs(mockOutput)
		assert.Equal(t, domain.DefaultBaccaratBanqueConfig(), bad.ToConfig())
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBaccaratBanqueConfig())

		// 負のコントロール: 範囲内なら既定ではなく頼んだ値が届く。
		var good controller.BaccaratBanqueWebInput
		_ = json.Unmarshal([]byte(
			`{"command":"reset","sessionId":"s1","config":{"cpuDifficulty":0,"startChips":5000,"betAmount":500}}`), &good)
		cfg := good.ToConfig()
		assert.Equal(t, domain.BaccaratBanqueCpuDifficultyEasy, cfg.CpuDifficulty)
		assert.Equal(t, 5000, cfg.StartChips)
		assert.Equal(t, domain.BaccaratBanqueMaxBet, cfg.BetAmount)
		assert.NotEqual(t, domain.DefaultBaccaratBanqueConfig(), cfg)
	})

	t.Run("no config falls back to the default", func(t *testing.T) {
		var input controller.BaccaratBanqueWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		assert.Equal(t, domain.DefaultBaccaratBanqueConfig(), input.ToConfig())
	})
}

func TestBaccaratBanqueCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"banker"}`

	for _, tc := range []struct{ cmd, method string }{
		{"draw", "Draw"}, {"d", "Draw"}, {"stand", "Stand"}, {"s", "Stand"},
		{"nextcoup", "NextCoup"}, {"nc", "NextCoup"}, {"retire", "Retire"},
	} {
		t.Run(tc.cmd+" calls "+tc.method, func(t *testing.T) {
			m := newBaccaratBanqueMock(mockOutput)
			assert.Equal(t, mockOutput, controller.NewBaccaratBanqueCuiController(m).Exec(tc.cmd))
			m.AssertCalled(t, tc.method)
		})
	}

	t.Run("sd resets with the new difficulty", func(t *testing.T) {
		m := newBaccaratBanqueMock(mockOutput)
		assert.Equal(t, mockOutput, controller.NewBaccaratBanqueCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultBaccaratBanqueConfig()
		cfg.CpuDifficulty = domain.BaccaratBanqueCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("sc and sb reset with the new stake", func(t *testing.T) {
		m := newBaccaratBanqueMock(mockOutput)
		assert.Equal(t, mockOutput, controller.NewBaccaratBanqueCuiController(m).Exec("sc 5000"))
		cfg := domain.DefaultBaccaratBanqueConfig()
		cfg.StartChips = 5000
		m.AssertCalled(t, "ResetWithConfig", cfg)

		m2 := newBaccaratBanqueMock(mockOutput)
		assert.Equal(t, mockOutput, controller.NewBaccaratBanqueCuiController(m2).Exec("sb 100"))
		cfg2 := domain.DefaultBaccaratBanqueConfig()
		cfg2.BetAmount = 100
		m2.AssertCalled(t, "ResetWithConfig", cfg2)
	})

	// **範囲外は断り、範囲を名指す。** 黙って丸めると、頼んだ額で遊べていない
	// ことに気づけない。
	t.Run("out-of-range chips and bets are refused and name the range", func(t *testing.T) {
		m := newBaccaratBanqueMock(mockOutput)
		out := controller.NewBaccaratBanqueCuiController(m).Exec("sc 1")
		assert.Contains(t, out, msgStem("baccaratbanque.invalidChips"))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.BaccaratBanqueMinChips))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.BaccaratBanqueMaxChips))

		assert.Contains(t, controller.NewBaccaratBanqueCuiController(m).Exec("sb 9999"),
			msgStem("baccaratbanque.invalidBet"))
		assert.Contains(t, controller.NewBaccaratBanqueCuiController(m).Exec("sc zz"),
			msgStem("baccaratbanque.invalidChips"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newBaccaratBanqueMock(mockOutput)
		assert.Equal(t, mockOutput, controller.NewBaccaratBanqueCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewBaccaratBanqueCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewBaccaratBanqueCuiController(newBaccaratBanqueMock(mockOutput)).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
