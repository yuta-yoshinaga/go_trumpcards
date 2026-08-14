//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustHorseOutputJSON(msg string) string {
	out := &controller.HorseWebOutput{
		Seats:          []*controller.HorseWebOutputSeat{},
		CommunityCards: []*controller.WebOutputCard{},
		WinnerSeat:     -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustHorseOutputJSON: %v", err))
	}
	return string(b)
}

func TestHorseWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	hiMock := new(usecase.MockHorseInteractor)
	hiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	hiMock.On("Action", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
	hiMock.On("NextHand").Return(mockOutput)
	hiMock.On("Hint").Return(mockOutput)
	hiMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewHorseWebController(func() uc.HorseInteractorIF { return hiMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.HorseWebInput
		require.NoError(t, json.Unmarshal([]byte(body), &input))
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustHorseOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
		hiMock.AssertCalled(t, "ResetWithConfig", domain.DefaultHorseConfig())
	})

	// **綴りは 6 通り全部を通す。** 1 つでも取りこぼすと、その手だけが Web から
	// 打てなくなる。
	for _, tt := range []struct {
		body   string
		action int
		amount int
	}{
		{`{"command":"action","action":"fold","sessionId":"s1"}`, domain.HoldemActionFold, 0},
		{`{"command":"a","action":"f","sessionId":"s1"}`, domain.HoldemActionFold, 0},
		{`{"command":"action","action":"check","sessionId":"s1"}`, domain.HoldemActionCheck, 0},
		{`{"command":"action","action":"call","sessionId":"s1"}`, domain.HoldemActionCall, 0},
		{`{"command":"action","action":"bet","amount":40,"sessionId":"s1"}`, domain.HoldemActionBet, 40},
		{`{"command":"action","action":"raise","amount":80,"sessionId":"s1"}`, domain.HoldemActionRaise, 80},
		{`{"command":"action","action":"allin","sessionId":"s1"}`, domain.HoldemActionAllIn, 0},
	} {
		t.Run(tt.body, func(t *testing.T) {
			run(t, tt.body, mockOutput, http.StatusOK)
			hiMock.AssertCalled(t, "Action", tt.action, tt.amount, 0)
		})
	}

	t.Run("action missing", func(t *testing.T) {
		run(t, `{"command":"action","sessionId":"s1"}`,
			mustHorseOutputJSON("param error: action is required."), http.StatusBadRequest)
	})
	t.Run("action unknown", func(t *testing.T) {
		run(t, `{"command":"action","action":"surrender","sessionId":"s1"}`,
			mustHorseOutputJSON("param error: action must be fold, check, call, bet, raise or allin."),
			http.StatusBadRequest)
	})
	// **額の無いベットは 0 として通さない。** 0 のまま流すとドメインの汎用エラーに
	// なり、何を送り忘れたのか画面から分からない。
	for _, act := range []string{"bet", "raise"} {
		t.Run(act+" missing amount", func(t *testing.T) {
			run(t, `{"command":"action","action":"`+act+`","sessionId":"s1"}`,
				mustHorseOutputJSON("param error: amount is required for bet and raise."),
				http.StatusBadRequest)
		})
	}
	t.Run("next", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
		hiMock.AssertCalled(t, "NextHand")
	})
	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`,
			mustHorseOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
}

// **範囲外の設定は既定に落ちる。** 席数だけは丸めではなく一覧照合で、
// 4/6/9 以外はすべて既定へ戻る。
func TestHorseWebConfig_ToConfigNeverYieldsAnInvalidConfig(t *testing.T) {
	def := domain.DefaultHorseConfig()
	tests := []struct {
		name string
		body string
		want domain.HorseConfig
	}{
		{"未指定なら既定", `{}`, def},
		{"席数 4", `{"seats":4}`, domain.HorseConfig{
			Seats: 4, InitialChips: def.InitialChips, HandsPerDiscipline: def.HandsPerDiscipline,
		}},
		{"席数 6", `{"seats":6}`, domain.HorseConfig{
			Seats: 6, InitialChips: def.InitialChips, HandsPerDiscipline: def.HandsPerDiscipline,
		}},
		{"席数 9", `{"seats":9}`, domain.HorseConfig{
			Seats: 9, InitialChips: def.InitialChips, HandsPerDiscipline: def.HandsPerDiscipline,
		}},
		// **5 席は「範囲内だが作れない」。** 丸めでは通ってしまう値。
		{"席数 5 は既定へ", `{"seats":5}`, def},
		{"席数 3 は既定へ", `{"seats":3}`, def},
		{"席数 10 は既定へ", `{"seats":10}`, def},
		{"チップが範囲外", `{"initialChips":999999999}`, def},
		{"ハンド数が範囲外", `{"handsPerDiscipline":0}`, def},
		{"ハンド数が多すぎる", `{"handsPerDiscipline":99}`, def},
		{"範囲内はそのまま通る", `{"seats":6,"initialChips":2000,"handsPerDiscipline":3}`,
			domain.HorseConfig{Seats: 6, InitialChips: 2000, HandsPerDiscipline: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg controller.HorseWebConfig
			require.NoError(t, json.Unmarshal([]byte(tt.body), &cfg))
			got := cfg.ToConfig()
			assert.Equal(t, tt.want, got)
			// **結果はドメインの検証を通る。** 変換の結果を検証まで通さないと、
			// 「範囲内に見えるが不正」な設定が素通りする。
			assert.NoError(t, got.Validate())
		})
	}
}
