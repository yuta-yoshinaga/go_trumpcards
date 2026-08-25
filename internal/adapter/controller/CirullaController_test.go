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

func mustCirullaOutputJSON(msg string) string {
	out := &controller.CirullaWebOutput{
		Players:         []*controller.CirullaWebOutputPlayer{},
		Table:           []*controller.WebOutputCard{},
		CaptureOptions:  [][][]int{},
		HintCaptureIdxs: []int{},
		LastCapturer:    -1,
		WinnerIdx:       -1,
		HintHandIdx:     -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCirullaOutputJSON: %v", err))
	}
	return string(b)
}

func TestCirullaWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	diMock := new(usecase.MockCirullaInteractor)
	diMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	diMock.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewCirullaWebController(func() uc.CirullaInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.CirullaWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustCirullaOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", domain.DefaultCirullaConfig())
	})

	// **出す札と取る札は同じ要求で届く。** 別便にすると「出したが取っていない」
	// 盤面が生まれる。
	t.Run("play carries the capture", func(t *testing.T) {
		run(t, `{"command":"play","handIndex":1,"captureIndices":[0,2],"sessionId":"s1"}`,
			mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 1, []int{0, 2})
	})
	t.Run("play without a capture lays the card off", func(t *testing.T) {
		run(t, `{"command":"p","handIndex":2,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 2, []int(nil))
	})
	// **0 は「省略」ではない。** handIndex がポインタでないと、札を指定しない
	// 要求が先頭の札を出す手として通ってしまう。
	t.Run("play without a hand index is refused", func(t *testing.T) {
		var input controller.CirullaWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "NextRound")
	})
	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
}

// 設定は範囲で丸め、組み立てた設定はドメインの検査を通る。
func TestCirullaWebConfig_ToConfig(t *testing.T) {
	i := func(n int) *int { return &n }

	for _, tt := range []struct {
		name       string
		in         controller.CirullaWebConfig
		wantTarget int
		wantDiff   domain.CirullaCpuDifficulty
	}{
		{"defaults", controller.CirullaWebConfig{},
			domain.CirullaDefaultTarget, domain.CirullaCpuDifficultyNormal},
		{"in range", controller.CirullaWebConfig{TargetScore: i(21), CpuDifficulty: i(0)},
			21, domain.CirullaCpuDifficultyEasy},
		{"too high a target falls back", controller.CirullaWebConfig{TargetScore: i(999)},
			domain.CirullaDefaultTarget, domain.CirullaCpuDifficultyNormal},
		{"too low a target falls back", controller.CirullaWebConfig{TargetScore: i(1)},
			domain.CirullaDefaultTarget, domain.CirullaCpuDifficultyNormal},
		{"an unknown difficulty falls back", controller.CirullaWebConfig{CpuDifficulty: i(9)},
			domain.CirullaDefaultTarget, domain.CirullaCpuDifficultyNormal},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			cfg := in.ToConfig()
			assert.Equal(t, tt.wantTarget, cfg.TargetScore)
			assert.Equal(t, tt.wantDiff, cfg.CpuDifficulty)
			assert.NoError(t, cfg.Validate())
		})
	}

	// 設定を積まない要求は既定のまま。
	assert.Equal(t, domain.DefaultCirullaConfig(), controller.CirullaWebInput{}.ToConfig())
}

func TestCirullaCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	newMock := func() *usecase.MockCirullaInteractor {
		m := new(usecase.MockCirullaInteractor)
		m.On("GetConfig").Return(domain.DefaultCirullaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewCirullaCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCirullaCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCirullaConfig())
	})

	// **取る札は同じ 1 行に書く。**
	t.Run("play parses the capture on the same line", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCirullaCuiController(m).Exec("p 1 0 2"))
		m.AssertCalled(t, "Play", 1, []int{0, 2})
	})

	t.Run("play with no table numbers lays the card off", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCirullaCuiController(m).Exec("play 2"))
		m.AssertCalled(t, "Play", 2, []int(nil))
	})

	// **読めない番号は場に置く手に落とさない。** 落とすと `p 0 zz` が黙って
	// 場に足す手になり、取ったつもりの札が相手に残る。
	for _, arg := range []string{"zz", "-1", "1 zz"} {
		t.Run("a table number that will not parse is refused: "+arg, func(t *testing.T) {
			m := newMock()
			out := controller.NewCirullaCuiController(m).Exec("p 0 " + arg)
			assert.Contains(t, out, msgStem("invalidFieldIndex"))
			m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
		})
	}

	t.Run("play without a hand index is refused", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewCirullaCuiController(m).Exec("p"), msgStem("cardIndexRequired"))
		assert.Contains(t, controller.NewCirullaCuiController(m).Exec("p zz"), msgStem("invalidCardIndex"))
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
	})

	for _, cmd := range []string{"nextround", "nr"} {
		t.Run(cmd+" advances the round", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewCirullaCuiController(m).Exec(cmd))
			m.AssertCalled(t, "NextRound")
		})
	}

	t.Run("setdifficulty and settarget reset with the new setting", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCirullaCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultCirullaConfig()
		cfg.CpuDifficulty = domain.CirullaCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)

		m2 := newMock()
		assert.Equal(t, mockOutput, controller.NewCirullaCuiController(m2).Exec("st 21"))
		cfg2 := domain.DefaultCirullaConfig()
		cfg2.TargetScore = 21
		m2.AssertCalled(t, "ResetWithConfig", cfg2)
	})

	// **勝負点は 11〜51。** 断られる数字を勧めないよう、拒否の文言も範囲を名指す。
	t.Run("a target outside 11-51 is refused and names the range", func(t *testing.T) {
		m := newMock()
		out := controller.NewCirullaCuiController(m).Exec("st 5")
		assert.Contains(t, out, msgStem("cirulla.invalidTargetScore"))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.CirullaMinTarget))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.CirullaMaxTarget))
		assert.Contains(t, controller.NewCirullaCuiController(m).Exec("st 99"),
			msgStem("cirulla.invalidTargetScore"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewCirullaCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewCirullaCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewCirullaCuiController(newMock()).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
