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

func mustDilotiOutputJSON(msg string) string {
	out := &controller.DilotiWebOutput{
		Players:        []*controller.DilotiWebOutputPlayer{},
		Table:          []*controller.WebOutputCard{},
		Declarations:   []*controller.DilotiWebOutputDeclaration{},
		TakeOptions:    [][]*controller.DilotiWebOutputTake{},
		DeclareOptions: [][]*controller.DilotiWebOutputDeclCandidate{},
		CanTrail:       []bool{},
		HintTableIdxs:  []int{},
		HintDeclIdxs:   []int{},
		LastCapturer:   -1,
		WinnerIdx:      -1,
		HintHandIdx:    -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustDilotiOutputJSON: %v", err))
	}
	return string(b)
}

func TestDilotiWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	diMock := new(usecase.MockDilotiInteractor)
	diMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	diMock.On("Play", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewDilotiWebController(func() uc.DilotiInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.DilotiWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustDilotiOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", domain.DefaultDilotiConfig())
	})

	// **取る対象・宣言値は同じ要求に乗る。** 別便にすると「出したが取っていない」
	// 盤面が生まれる。
	t.Run("play carries the capture targets", func(t *testing.T) {
		run(t, `{"command":"play","handIndex":1,"action":"capture","tableIndices":[0,2],"declIndices":[1],"sessionId":"s1"}`,
			mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 1, domain.DilotiActionCapture, []int{0, 2}, []int{1}, 0)
	})
	t.Run("play carries the declared value", func(t *testing.T) {
		run(t, `{"command":"p","handIndex":0,"action":"declare","tableIndices":[1],"declValue":8,"sessionId":"s1"}`,
			mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 0, domain.DilotiActionDeclare, []int{1}, []int(nil), 8)
	})
	t.Run("play can lay off", func(t *testing.T) {
		run(t, `{"command":"play","handIndex":2,"action":"trail","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 2, domain.DilotiActionTrail, []int(nil), []int(nil), 0)
	})
	// **0 は「省略」ではない。** handIndex がポインタでないと、札を指定しない
	// 要求が先頭の札を出す手として通ってしまう。
	t.Run("play without a hand index is refused", func(t *testing.T) {
		var input controller.DilotiWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","action":"trail","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	// **手の種類が無い要求は通さない。** 通すと「取るつもりが場に置かれた」に
	// なる。
	t.Run("play without an action is refused", func(t *testing.T) {
		var input controller.DilotiWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","handIndex":0,"sessionId":"s1"}`), &input)
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
func TestDilotiWebConfig_ToConfig(t *testing.T) {
	i := func(n int) *int { return &n }

	for _, tt := range []struct {
		name       string
		in         controller.DilotiWebConfig
		wantTarget int
		wantDiff   domain.DilotiCpuDifficulty
	}{
		{"defaults", controller.DilotiWebConfig{},
			domain.DilotiDefaultTarget, domain.DilotiCpuDifficultyNormal},
		{"in range", controller.DilotiWebConfig{TargetScore: i(41), CpuDifficulty: i(0)},
			41, domain.DilotiCpuDifficultyEasy},
		{"too high a target falls back", controller.DilotiWebConfig{TargetScore: i(999)},
			domain.DilotiDefaultTarget, domain.DilotiCpuDifficultyNormal},
		{"too low a target falls back", controller.DilotiWebConfig{TargetScore: i(1)},
			domain.DilotiDefaultTarget, domain.DilotiCpuDifficultyNormal},
		{"an unknown difficulty falls back", controller.DilotiWebConfig{CpuDifficulty: i(9)},
			domain.DilotiDefaultTarget, domain.DilotiCpuDifficultyNormal},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			cfg := in.ToConfig()
			assert.Equal(t, tt.wantTarget, cfg.TargetScore)
			assert.Equal(t, tt.wantDiff, cfg.CpuDifficulty)
			assert.NoError(t, cfg.Validate())
		})
	}
	assert.Equal(t, domain.DefaultDilotiConfig(), controller.DilotiWebInput{}.ToConfig())
}

func TestDilotiCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	newMock := func() *usecase.MockDilotiInteractor {
		m := new(usecase.MockDilotiInteractor)
		m.On("GetConfig").Return(domain.DefaultDilotiConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewDilotiCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultDilotiConfig())
	})

	// **取る対象は同じ 1 行に書く。** 宣言は d を前置して区別する。
	t.Run("take parses table and declaration targets on the same line", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec("t 1 0 2 d1"))
		m.AssertCalled(t, "Play", 1, domain.DilotiActionCapture, []int{0, 2}, []int{1}, 0)
	})

	t.Run("take with only table numbers", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec("take 2 0"))
		m.AssertCalled(t, "Play", 2, domain.DilotiActionCapture, []int{0}, []int(nil), 0)
	})

	// **読めない番号は場に置く手に落とさない。** 落とすと `t 0 zz` が黙って
	// 場に足す手になり、取ったつもりの札が相手に残る。
	for _, arg := range []string{"zz", "-1", "0 zz", "dx"} {
		t.Run("a target that will not parse is refused: "+arg, func(t *testing.T) {
			m := newMock()
			out := controller.NewDilotiCuiController(m).Exec("t 0 " + arg)
			assert.Contains(t, out, msgStem("invalidFieldIndex"))
			m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}

	t.Run("declare carries the value", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec("d 0 8 1"))
		m.AssertCalled(t, "Play", 0, domain.DilotiActionDeclare, []int{1}, []int(nil), 8)
	})

	// **宣言値は 2〜10。** 断る文言も範囲を名指す。
	t.Run("declare refuses a value outside 2-10 and names the range", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewDilotiCuiController(m).Exec("d 0"),
			msgStem("diloti.declValueRequired"))
		for _, bad := range []string{"d 0 1", "d 0 11", "d 0 zz"} {
			out := controller.NewDilotiCuiController(m).Exec(bad)
			assert.Contains(t, out, msgStem("diloti.invalidDeclValue"), "入力 %q", bad)
		}
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("lay lays the card off", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec("l2 3"))
		m.AssertCalled(t, "Play", 3, domain.DilotiActionTrail, []int(nil), []int(nil), 0)
	})

	t.Run("play without a hand index is refused", func(t *testing.T) {
		m := newMock()
		assert.Contains(t, controller.NewDilotiCuiController(m).Exec("t"), msgStem("cardIndexRequired"))
		assert.Contains(t, controller.NewDilotiCuiController(m).Exec("t zz"), msgStem("invalidCardIndex"))
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	for _, cmd := range []string{"nextround", "nr"} {
		t.Run(cmd+" advances the round", func(t *testing.T) {
			m := newMock()
			assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec(cmd))
			m.AssertCalled(t, "NextRound")
		})
	}

	t.Run("setdifficulty and settarget reset with the new setting", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec("sd 0"))
		cfg := domain.DefaultDilotiConfig()
		cfg.CpuDifficulty = domain.DilotiCpuDifficultyEasy
		m.AssertCalled(t, "ResetWithConfig", cfg)

		m2 := newMock()
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m2).Exec("st 41"))
		cfg2 := domain.DefaultDilotiConfig()
		cfg2.TargetScore = 41
		m2.AssertCalled(t, "ResetWithConfig", cfg2)
	})

	// **勝負点は 21〜101。** 断られる数字を勧めないよう、拒否の文言も範囲を名指す。
	t.Run("a target outside 21-101 is refused and names the range", func(t *testing.T) {
		m := newMock()
		out := controller.NewDilotiCuiController(m).Exec("st 5")
		assert.Contains(t, out, msgStem("diloti.invalidTargetScore"))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.DilotiMinTarget))
		assert.Contains(t, out, fmt.Sprintf("%d", domain.DilotiMaxTarget))
		assert.Contains(t, controller.NewDilotiCuiController(m).Exec("st 999"),
			msgStem("diloti.invalidTargetScore"))
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint and log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec("h"))
		assert.Equal(t, mockOutput, controller.NewDilotiCuiController(m).Exec("l"))
	})

	t.Run("unknown command is reported", func(t *testing.T) {
		out := controller.NewDilotiCuiController(newMock()).Exec("zzzz")
		assert.NotEmpty(t, out)
		assert.NotEqual(t, mockOutput, out)
	})
}
